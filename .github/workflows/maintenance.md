# Maintenance

## How to update supported Kubernetes

- Update Kubernetes version in `Makefile.versions`.
- Update `k8s.io/*` packages version in `go.mod`.

## How to update supported Flatcar Container Linux

NecoPerf supports one Flatcar Linux version at a time.
If a new stable version of Flatcar Linux is released, please do the following steps.

1. Check the [stable release](https://www.flatcar.org/releases) of Flatcar Linux.
    * e.g. "3602.2.2" is released as Stable.
2. Find the tags for the container image of the [flatcar-sdk-amd64](https://github.com/orgs/flatcar/packages/container/package/flatcar-sdk-amd64) corresponding to the stable version of Flatcar Linux.
    * e.g. "3602.0.0" is the corresponding tag.
    * Note that the container images of the flatcar-sdk-amd64 are not updated as frequently as the Flatcar releases.
3. Update `FLATCAR_VERSION` in `Makefile.versions`.
    * e.g. Use "3602.0.0" as `FLATCAR_VERSION`.

If the contents of the container image have changed, especially in terms of the files layout, please update the relevant Dockerfiles.

## How to update dependencies

- Update `Makefile.versions`.
- Update `Dockerfile.cli` and `Dockerfile.daemon`.
- Update go.mod.
- Update GitHub Actions.
