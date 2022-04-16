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

## CR Examples

Some example CR configurations can be found as OC Console Examples in [examples.yaml](config/codeready/mailhogInstance_console_examples.yaml)

## Development

```bash
### Create project
# As developer in the crc console add a project named "project"
### Deploy the operator to crc (tested on fedora)
make crc-deploy
### Check the MailhogInstance type in the web console, a sample should be ready to go
```

## License

[Apache 2](LICENSE)
