# mailhog-operator

![mailhog-operator](hack/mailhog-operator-wdn.png "mailhog-operator")

A toy operator that deploys [mailhog](https://github.com/mailhog/MailHog) on [codeready containers](https://github.com/code-ready/crc).

[![containers (dev)](https://github.com/patrickmx/mailhog-operator/actions/workflows/containers_develop.yml/badge.svg)](https://github.com/patrickmx/mailhog-operator/actions/workflows/containers_develop.yml)
[![containers (tag)](https://github.com/patrickmx/mailhog-operator/actions/workflows/containers_tag.yml/badge.svg)](https://github.com/patrickmx/mailhog-operator/actions/workflows/containers_tag.yml)
[![Tag](https://img.shields.io/github/v/tag/patrickmx/mailhog-operator?sort=semver)](https://github.com/patrickmx/mailhog-operator/tags)
[![GoDoc](https://godoc.org/goimports.patrick.mx/mailhog-operator?status.svg)](http://godoc.org/goimports.patrick.mx/mailhog-operator)
[![Go Report Card](https://goreportcard.com/badge/goimports.patrick.mx/mailhog-operator)](https://goreportcard.com/report/goimports.patrick.mx/mailhog-operator)

## Images

Check out the [latest releases](https://github.com/patrickmx/mailhog-operator/pkgs/container/mailhog-operator)

### Just the operator

```bash
### Get the latest tagged image release
podman pull ghcr.io/patrickmx/mailhog-operator:latest
### Current pre-release development image
podman pull ghcr.io/patrickmx/mailhog-operator:develop
```

### Bundle

```bash
operator-sdk run bundle ghcr.io/patrickmx/mailhog-operator-bundle:latest
operator-sdk run bundle ghcr.io/patrickmx/mailhog-operator-bundle:develop
```

### Catalog Source

```bash
# Install the CatalogSource on oc/crc/origin
oc -n openshift-marketplace create -f config/catalogsource/mailhog-catalogsource.yaml
# To watch multiple namespaces or watch a different one than where the operator is, add a separate operator group
```

## CR Examples

Some example CR configurations can be found as OC Console Examples in [console_examples.yaml](config/codeready/mailhogInstance_console_examples.yaml) or the [bare minimal cr](config/samples/mailhog_v1alpha1_mailhoginstance.yaml).

```bash
# Load Openshift Console Examples:
oc create -f config/codeready/mailhogInstance_console_examples.yaml
```

## Development

Build current repo code and deploy to a local CRC installation:

```bash
### Deploy the operator to crc (tested on fedora)
make crc-deploy
### Check the MailhogInstance type in the web console, a sample should be ready to go
```

## License

[Apache 2](LICENSE)
