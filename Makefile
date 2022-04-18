# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export VERSION=0.0.2)
# IMAGE_TAG_BASE defines the docker.io namespace and part of the image name for remote images.
# This variable is used to construct full image tags for bundle and catalog images.
# For example, running 'make bundle-build bundle-push catalog-build catalog-push' will build and push both
# operators.patrick.mx/mailhog-operator-bundle:$VERSION and operators.patrick.mx/mailhog-operator-catalog:$VERSION.
# VERSION and IMAGE_TAG_BASE from params.mak file
include ./params.mak

# CHANNELS define the bundle channels used in the bundle.
# Add a new line here if you would like to change its default config. (E.g CHANNELS = "candidate,fast,stable")
# To re-generate a bundle for other specific channels without changing the standard setup, you can:
# - use the CHANNELS as arg of the bundle target (e.g make bundle CHANNELS=candidate,fast,stable)
# - use environment variables to overwrite this value (e.g export CHANNELS="candidate,fast,stable")
CHANNELS ?= fest

# DEFAULT_CHANNEL defines the default channel used in the bundle.
# Add a new line here if you would like to change its default config. (E.g DEFAULT_CHANNEL = "stable")
# To re-generate a bundle for any other default channel without changing the default setup, you can:
# - use the DEFAULT_CHANNEL as arg of the bundle target (e.g make bundle DEFAULT_CHANNEL=stable)
# - use environment variables to overwrite this value (e.g export DEFAULT_CHANNEL="stable")
DEFAULT_CHANNEL ?= fast
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# BUNDLE_IMG defines the image:tag used for the bundle.
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMG=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:v$(VERSION)

# Image URL to use all building/pushing image targets
IMG ?= $(IMAGE_TAG_BASE):v$(VERSION)
IMG_LOCAL ?= $(IMAGE_TAG_BASE_LOCAL):v$(VERSION)
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.23

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: manifests-dryrun
manifests-dryrun: manifests
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	mkdir -p dry-run
	$(KUSTOMIZE) build config/default > dry-run/manifests.yaml

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: gofumpt ## Run go fmt against code.
	$(GOFUMPT) -l -w .

.PHONY: vet
vet: tidy ## Run go vet against code.
	go vet ./...

.PHONY: tidy
tidy: ## Run go vet against code.
	go mod tidy

.PHONY: sec
sec: gosec ## Run gosec
	$(GOSEC) ./...

.PHONY: test
test: manifests generate fmt vet lint ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./... -coverprofile cover.out

.PHONY: lint
lint: manifests generate fmt vet sec golangci-lint ## run linter with normal settings
	$(GOLANGCILINT) run

.PHONY: lint-strict
lint-strict: manifests generate fmt vet sec golangci-lint ## run linter with more strict tips
	$(GOLANGCILINT) run -E funlen,revive,dupl,lll,gocognit,cyclop

##@ Build

.PHONY: build
build: generate fmt vet lint ## Build manager binary.
	go build -tags . -o bin/manager

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./main.go -config config/manager/controller_manager_config.yaml

.PHONY: debug
debug: generate fmt vet manifests ## run with delve debugger
	go build -gcflags "all=-trimpath=$(shell go env GOPATH)" -o bin/manager main.go
	dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec ./bin/manager -config config/manager/controller_manager_config.yaml

.PHONY: docker-build
docker-build: test docker-refresh-base ## Build docker image with the manager.
	podman build -t ${IMG} .

.PHONY: docker-refresh-base
docker-refresh-base: ## refresh manager builder base image
	podman pull docker.io/library/golang:1.18

.PHONY: latest
latest: ## get information about the latest image / commit
	podman inspect ${IMG} | jq .[0].Id
	podman inspect ${IMG} | jq .[0].Created
	git rev-parse HEAD

##@ CRC Deploy

.PHONY: build-push-image-to-crc
build-push-image-to-crc: docker-build ## push the image from the local podman to imagestream
	$(KUSTOMIZE) build config/codeready | kubectl apply -f -
	podman login -u kubeadmin -p $(oc whoami -t) default-route-openshift-image-registry.apps-crc.testing --tls-verify=false
	oc registry login --insecure=true
	podman tag $(IMG) $(IMG_LOCAL)
	podman push --tls-verify=false default-route-openshift-image-registry.apps-crc.testing/mailhog-operator-system/mailhog:v$(VERSION)

.PHONY: crc-deploy
crc-deploy: crc-start crc-login-admin deploy build-push-image-to-crc latest ## set manager deployment to the local imagestream
	oc -n mailhog-operator-system patch deployment/mailhog-operator-controller-manager -p "{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"manager\",\"image\":\"$(IMG_LOCAL)\"}]}}}}"

##@ CRC Ad-Hoc Commands

.PHONY: crc-images
crc-images: ## show state of the local manager imagestream
	oc get is -n mailhog-operator-system

.PHONY: crc-pods
crc-pods: ## get pods in the local manager namespace
	oc get pods -n mailhog-operator-system

.PHONY: crc-logs
crc-logs: ## tail the logs of the latest manager pod
	oc -n mailhog-operator-system logs deployment/mailhog-operator-controller-manager -c manager -f

.PHONY: install-cert-manager
install-cert-manager: ## install cert manager
	kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.7.1/cert-manager.yaml

.PHONY: crc-login-admin
crc-login-admin: ## ensure kubeadmin is logged in
	$(shell crc console --credentials | awk -F"[']" '{print $$2}' | tail -n1)

.PHONY: crc-reset
crc-reset: ## destroy and recreate CRC
	crc delete -f
	crc start

.PHONY: crc-start
crc-start: ## ensure crc is started
	crc start

.PHONY: crc-creds
crc-creds: ## show crc credentials
	crc console --credentials

.PHONY: crc-restore-pinning
crc-restore-pinning: ## restore web console developer pinning. Need to login / pin something before
	oc -n openshift-console-user-settings \
		patch --type merge \
		cm/user-settings-$(shell oc get -o json users/kubeadmin | jq -r .metadata.uid) \
		-p '{"data":{"console.pinnedResources":"{\"admin\":[],\"dev\":[\"core~v1~ConfigMap\",\"apps~v1~Deployment\",\"apps.openshift.io~v1~DeploymentConfig\",\"mailhog.operators.patrick.mx~v1alpha1~MailhogInstance\",\"core~v1~Service\",\"core~v1~Pod\",\"route.openshift.io~v1~Route\"]}"}}'

.PHONY: crc-add-mongo
crc-add-mongo: ## deploy a matching mongo for the mongodb console example
	oc -n project new-app \
           -e MONGODB_USER=user \
           -e MONGODB_PASSWORD=password \
           -e MONGODB_DATABASE=mailhog \
           -e MONGODB_ADMIN_PASSWORD=admin \
           --name="mongodb" \
           -l "app.kubernetes.io/part-of=mailhog,app.openshift.io/runtime=mongodb" \
           registry.redhat.io/rhscl/mongodb-26-rhel7

.PHONY: all-catalogsources
all-catalogsources: ## list all catalogsources
	oc get --all-namespaces=true catalogsources

.PHONY: all-packagemanifests
all-packagemanifests: ## list all packagemanifests
	oc get --all-namespaces=true packagemanifests

.PHONY: all-clusterserviceversions
app-clusterserviceversions: ## list all clusterserviceversions
	oc get --all-namespaces=true csv

.PHONY: clean-leftover-bundles
clean-leftover-bundles: ## clean some temp files that get left over when working with bundles
	find . -name "bundle-*" -type d -exec rmdir {} \;
	find . -name "bundle_*" -type d -exec rm -rf {} \;

##@ Release

.PHONY: ship
ship: test build bundle ## create a tag from the current version in params.mak
	@if git show-ref --tags --quiet --verify -- "refs/tags/v$(VERSION)"; then \
    	echo "tag already exists"; \
    	exit 1; \
    fi
	git tag v$(VERSION)
	@echo "tag v$(VERSION) created, push it by running:"
	@echo "git push origin v$(VERSION)"

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
.PHONY: controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.8.0)

KUSTOMIZE = $(shell pwd)/bin/kustomize
.PHONY: kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v4@v4.5.2)

ENVTEST = $(shell pwd)/bin/setup-envtest
.PHONY: envtest
envtest: ## Download envtest-setup locally if necessary.
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@latest)

GOSEC = $(shell pwd)/bin/gosec
.PHONY: gosec
gosec: ## Download gosec locally if necessary. https://github.com/securego/gosec
	$(call go-install-tool,$(GOSEC),github.com/securego/gosec/v2/cmd/gosec@latest)

GOLANGCILINT = $(shell pwd)/bin/golangci-lint
.PHONY: golangci-lint
golangci-lint: ## Download golangci-lint locally if necessary. https://golangci-lint.run/
	$(call go-install-tool,$(GOLANGCILINT),github.com/golangci/golangci-lint/cmd/golangci-lint@v1.45.2)

GOFUMPT = $(shell pwd)/bin/gofumpt
.PHONY: gofumpt
gofumpt: ## Download golangci-lint locally if necessary. https://golangci-lint.run/
	$(call go-install-tool,$(GOFUMPT),mvdan.cc/gofumpt@latest)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-install-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

.PHONY: bundle
bundle: manifests kustomize ## Generate bundle manifests and metadata, then validate generated files.
	operator-sdk generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	operator-sdk bundle validate ./bundle

.PHONY: bundle-build
bundle-build: ## Build the bundle image.
	podman build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: bundle-clean
bundle-clean: crc-start crc-login-admin ## remove old manually installed bundle
	operator-sdk cleanup mailhog-operator
	sleep 10 # giving the old pod some seconds to terminate

.PHONY: bundle-run-develop
bundle-run-develop: crc-start crc-login-admin bundle-clean ## install latest develop bundle
	operator-sdk run bundle ghcr.io/patrickmx/mailhog-operator-bundle:develop

.PHONY: bundle-run-release
bundle-run-release: crc-start crc-login-admin bundle-clean ## install latest release bundle
	operator-sdk run bundle ghcr.io/patrickmx/mailhog-operator-bundle:latest

.PHONY: opm
OPM = ./bin/opm
opm: ## Download opm locally if necessary.
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.19.1/$${OS}-$${ARCH}-opm ;\
	chmod +x $(OPM) ;\
	}
else
OPM = $(shell which opm)
endif
endif

# A comma-separated list of bundle images (e.g. make catalog-build BUNDLE_IMGS=example.com/operator-bundle:v0.1.0,example.com/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
BUNDLE_IMGS ?= $(BUNDLE_IMG)

# The image tag given to the resulting catalog image (e.g. make catalog-build CATALOG_IMG=example.com/operator-catalog:v0.2.0).
CATALOG_IMG ?= $(IMAGE_TAG_BASE)-catalog:v$(VERSION)

# Set CATALOG_BASE_IMG to an existing catalog image tag to add $BUNDLE_IMGS to that image.
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif

# Build a catalog image by adding bundle images to an empty catalog using the operator package manager tool, 'opm'.
# This recipe invokes 'opm' in 'semver' bundle add mode. For more information on add modes, see:
# https://github.com/operator-framework/community-operators/blob/7f1438c/docs/packaging-operator.md#updating-your-existing-operator
.PHONY: catalog-build
catalog-build: opm ## Build a catalog image.
	$(OPM) index add --container-tool podman --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT)
