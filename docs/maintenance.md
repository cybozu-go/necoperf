# Maintenance

## Prerequisites

CLI tools (kind, kustomize, kubectl, setup-envtest, protoc and its plugins, mdbook, staticcheck) are managed by [aqua](https://aquaproj.github.io/) in `aqua.yaml`.
Their checksums are recorded in `aqua-checksums.json` and verified on installation.
The Makefiles execute the tools via `aqua exec`.

1. Install aqua. See [Install aqua](https://aquaproj.github.io/docs/install) for details.
2. Put the `aqua` command in a directory included in `PATH`.
3. Install the tools.
   ```console
   $ make setup
   ```

## How to update dependencies

Most versions are derived from the supported Kubernetes version and the Go version, so update them in the following order.

```
Kubernetes version
 ├─ k8s.io/* in go.mod
 │   └─ controller-runtime in go.mod
 │       └─ setup-envtest in aqua.yaml
 │           └─ ENVTEST_K8S_VERSION in Makefile.versions
 ├─ kind in aqua.yaml
 │   └─ KIND_NODE_VERSION in Makefile.versions
 └─ kubectl in aqua.yaml

Go version in go.mod
 └─ golang image in Dockerfile.cli and Dockerfile.daemon
```

1. Decide the Kubernetes version and the Go version to support.
2. Update `go.mod`.
    - Update the Go version.
    - Update `k8s.io/*` packages to the version corresponding to the Kubernetes version (e.g. `v0.35.x` for Kubernetes 1.35).
    - Update `sigs.k8s.io/controller-runtime` to the version compatible with the `k8s.io/*` packages.
    - Update the other modules.
    - Run `go mod tidy`.
3. Update the `ghcr.io/cybozu/golang` image in `Dockerfile.cli` and `Dockerfile.daemon` to match the Go version in `go.mod`.
4. Update `aqua.yaml`.
    1. Update the standard registry first, because new package versions may not be defined in an old registry.
    2. Update the packages.
        - Update `kubernetes-sigs/kind` to the version which supports `kindest/node` of the Kubernetes version.
        - Update `kubernetes/kubectl` to the same version as the Kubernetes version.
        - Update `kubernetes-sigs/controller-runtime/setup-envtest` to the same version as controller-runtime in `go.mod`.
        - Update the other packages to the latest versions.
    3. Update `aqua-checksums.json`.
       Since `require_checksum` is enabled, packages without checksums fail to install.
       ```console
       $ aqua update-checksum -prune
       ```
    4. Install the tools.
       ```console
       $ make setup
       ```

    Notes:
    - Write the version as the tag name, including its prefix if any (e.g. `kubernetes-sigs/kustomize@kustomize/v5.8.1`).
    - You can select a version interactively with `aqua update -s <package>` (e.g. `aqua update -s kubernetes/kubectl`).
    - Do not run `aqua update` without arguments, because it updates all packages including kubectl and setup-envtest to the latest versions.
5. Update `Makefile.versions`.
    - Update `ENVTEST_K8S_VERSION` to the Kubernetes version available for envtest.
      ```console
      $ # Use this command to list the available k8s versions for envtest
      $ aqua exec -- setup-envtest list
      ```
    - Update `KIND_NODE_VERSION` to the `kindest/node` image of the Kubernetes version listed in the release note of the kind in `aqua.yaml`.
6. Update GitHub Actions in `.github/workflows/*.yaml`.
    - Update the actions and the tools installed by them (e.g. `aqua_version` of `aquaproj/aqua-installer`, `version` of `goreleaser/goreleaser-action`).
    - Update `runs-on` if a new Ubuntu runner image is available.
7. Check that the build and the tests pass.
    - Run the checks and the unit tests.
      ```console
      $ make check-generated lint test
      ```
    - Run the e2e tests.
      ```console
      $ make -C e2e start test
      ```
    - Stop the kind cluster.
      ```console
      $ make -C e2e stop
      ```

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
