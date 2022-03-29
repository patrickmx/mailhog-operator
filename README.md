# mailhog-operator

![mailhog-operator](hack/mailhog-operator-wdn.png "mailhog-operator")

A toy operator that deploys [mailhog](https://github.com/mailhog/MailHog) on [codeready containers](https://github.com/code-ready/crc).

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
