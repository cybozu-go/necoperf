# Maintenance

## How to update supported Kubernetes

- Update Kubernetes version in `Makefile.versions`.
- Update `k8s.io/*` packages version in `go.mod`.

## How to update supported Flatcar Container Linux

NecoPerf supports one Flatcar Linux version at a time.
If a new stable version of Flatcar Linux is released, please do the following steps.

- Check the [stable release](https://www.flatcar.org/releases) of Flatcar Linux
- Find the tags for the container image of the [flatcar-sdk-amd64](https://github.com/orgs/flatcar/packages/container/package/flatcar-sdk-amd64) corresponding to the stable version of Flatcar Linux
- Update `FLATCAR_VERSION` in `Makefile.versions`.

If Flatcar Container Linux Container Image has changed, please fix the relevant source code.

## How to update dependencies

- Update `Makefile.versions`.
- Update `Dockerfile.cli` and `Dockerfile.daemon`.
- Update go.mod.
- Update GitHub Actions.
