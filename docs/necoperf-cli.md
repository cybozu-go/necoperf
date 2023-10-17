# necoperf-cli command reference

```console
necoperf-cli <subcommand> args...
```

- [`necoperf-cli profile PODNAME`](#necoperf-cli-profile-podname)

## `necoperf-cli profile PODNAME`

Perform profiling for container on pod.

| Option | Default value |Description |
|:-------|:--------------|:-----------|
| `--necoperf-namespace`|`necoperf`| Namespace in which necoperf-daemon is running|
| `-n`,`--namespace` | `default` | Namespace in which the pod being profiled is running |
| `--container` ||Specify the container name to profile. If no container name is specified, the first container of the pod is set as the target of profiling.|
| `--timeout` |`30s`| Time to run cpu profiling on server|
| `--outputDir` |`/tmp`|Directory for output of profiling results|
