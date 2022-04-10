# Build the manager binary
FROM docker.io/library/golang:1.18 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# All files are added here so that go embeds the correct vcs head status
COPY .gitignore ./.gitignore
COPY .github/ ./.github/
COPY . .
COPY .git/ ./.git/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags . -o manager && \
    go version -m ./manager > manager.version && \
    tail manager.version && \
    chmod 0555 ./manager && \
    sha256sum manager > manager.sha256 && \
    cat manager.sha256 && \
    chmod 0444 ./manager.sha256 ./manager.version


FROM scratch
LABEL \
  org.opencontainers.image.source="https://github.com/patrickmx/mailhog-operator" \
  org.opencontainers.image.title="Mailhog Operator" \
  org.opencontainers.image.description="deploy mailhogs on crc / oc" \
  io.k8s.description="deploy mailhogs on crc / oc" \
  io.openshift.tags="operator,mailhog" \
  io.openshift.min-memory="100Mi" \
  io.openshift.min-cpu="250m"
WORKDIR /
EXPOSE 8080 8081 9443
CMD ["/manager", "-config", "/operatorconfig/config.yml"]
COPY --from=builder /workspace/manager /workspace/manager.sha256 /workspace/manager.version /
COPY --from=builder /workspace/config/manager/controller_manager_config.yaml /operatorconfig/config.yml
USER 65532:65532
