# mailhog-operator

![mailhog-operator](hack/mailhog-operator-wdn.png "mailhog-operator")

A toy operator that deploys [mailhog](https://github.com/mailhog/MailHog) on [codeready containers](https://github.com/code-ready/crc).

[![containers (dev)](https://github.com/patrickmx/mailhog-operator/actions/workflows/containers_develop.yml/badge.svg)](https://github.com/patrickmx/mailhog-operator/actions/workflows/containers_develop.yml)
[![containers (tag)](https://github.com/patrickmx/mailhog-operator/actions/workflows/containers_tag.yml/badge.svg)](https://github.com/patrickmx/mailhog-operator/actions/workflows/containers_tag.yml)
[![GoDoc](https://godoc.org/goimports.patrick.mx/mailhog-operator?status.svg)](http://godoc.org/goimports.patrick.mx/mailhog-operator)
[![Go Report Card](https://goreportcard.com/badge/goimports.patrick.mx/mailhog-operator)](https://goreportcard.com/report/goimports.patrick.mx/mailhog-operator)

## Images

Check out the [latest releases](https://github.com/patrickmx/mailhog-operator/pkgs/container/mailhog-operator)

```bash
### Get the latest tagged image release
podman pull ghcr.io/patrickmx/mailhog-operator:develop
### Current pre-release image from master
podman pull ghcr.io/patrickmx/mailhog-operator:latest
```

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
