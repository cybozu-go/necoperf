Design Document
===============

NecoPerf provides an easy mechanism for cpu profiling in Kubernetes multi-tenancy without giving users strong permissions.

## Context and Scope

Currently, it is possible to retrieve cpu profiling using [Linux perf](https://perf.wiki.kernel.org/index.php/Main_Page) for application running on Kubernetes.
But it requires a lot of manual operations and strong permissions.
To solve these problems, we provide NecoPerf, a mechanism to more easily retrieve cpu profiling for application.
NecoPerf can automate many manual operations.

### Goals

- Provides a system for users to easily run perf command and retrieve cpu profiling for application
- NecoPerf users can specify options when running perf command

### Non-goals

- Support for various operating systems (initial implementation supports Flatcar Linux only)
- Support TLS
- Profiling of child processes
  - e.g. container run by [tini](https://github.com/krallin/tini)
- Continuous Profiling
- Processing and visualization of profile data, including conversion to [FlameGraph](https://github.com/brendangregg/FlameGraph)

## Proposal

### User Stories

This section describes the actual flow of a situation when a user uses perf command to retrieve profiling.

- The assumption is that the Kubernetes cluster in User stories is used in a multi-tenancy
  - There is a team managing the cluster and several teams using the cluster
  - The team that uses the cluster is called the tenant
  - Tenant do not have strong privileges
  - The team managing the Kubernetes cluster is called cluster admin

- Tenant are aware that there are performance issues with their application and want to do a cpu profile them using perf command to identify bottlenecks.
  However, a lot of things need to be done manually, as the following steps are required to run perf command
  1. Install a perf that is compatible with the kernel version of the host operating system in the container image
  2. Modify the manifest to add a sidecar or ephemeral container with the necessary permissions to run perf
  3. The user of tenant enters a sidecar or ephemeral container and executes perf command against the target container to retrieve the cpu profile

- Cluster Admin wants to minimize the permissions granted to the tenant.
  However, to run perf, the tenant needs to be able to grant  `CAP_SYS_ADMIN` and `CAP_SYS_PTRACE` permissions to the pod, which violates the principle of least privilege.

NecoPerf does not require manual operations and allows for easy cpu profiling for application using perf command.

### Constraints

- Debug symbols are required for perf to resolve symbols.
  These debug symbols must be included in the container image to be profiled
- NecoPerf performs cpu profiling based on the PID of the application, so if the target process is killed during profiling, profiling will not continue and will terminate

### Risk and Mitigations

- Security Risk
  - Originally, `CAP_SYSLOG`, `CAP_SYS_ADMIN`, `CAP_SYS_CHROOT` and other permissions are required to run perf command, but using necoperf is safe because it is not necessary to give those permissions to tenant.
  On the other hand, deploying necoperf requires the permission to run `CAP_SYSLOG` and CRI APIs, so it should be managed correctly so that ordinary users cannot misuse it.
- Performance Risk
  - To prevent tenant from running perf for long periods, the NecoPerf validates the values from the user request

## The actual design

The first implementation creates a gRPC server that simply runs perf command on the specified container id and returns the profiling results.
The perf command is used to retrieve profiling and convert the retrieved profiling data.

We also create a command line tool as a client to send requests to the gRPC server.
This command line tool queries the Kubernetes API server based on the pod and container name entered by the user and retrieves the container id.
The command line tool sends a profiling request to the gRPC server based on the retrieved container id.

```console
necoperf-client -n <namespace> <pod-name> -c <container name> -o <output directory>
```

### API

```protobuf
service NecoPerf {
    rpc Profile(ProfileRequest) returns (ProfileResponse);
}

message ProfileRequest {
    string container_id = 1;
    int64 timeout_seconds = 2;
}

message ProfileResponse {
    bytes data = 1;
}
```

### System Context Diagram

NecoPerf system overview diagram is shown below.

```mermaid
graph TD;
    User-->|exec|necoperf-client
    necoperf-client-->|GET|k8s-api-server[kube-apiserver]
    necoperf-client -->|gRPC call|necoperf-daemon

subgraph node1
    necoperf-daemon-->|CRI call|CRI
    perf-->|profile|pod[target pod]
    perf-.->|export/read|perf.data((necoperf.data))
    perf-.->|export|perf.script((necoperf.script))
    necoperf-daemon-->|exec|perf
    subgraph daemonset
        necoperf-daemon
    end
end

subgraph your-pod
    necoperf-client-.->|export|result((result))
end
```

## Alternatives

This section lists some existing systems and explains why they are not used.

- [IBM/perf-sidecar-injector](https://github.com/IBM/perf-sidecar-injector)
  - perf-sidecar-injector is a mutating webhook that adds a perf container as a sidecar container
  - perf-sidecar-injector requires privileged permission to run the perf container
  - To access the target container from the sidecar, the Pod setting `shareProcessNamespace` must be enabled.
    Enabling `shareProcessNamespace` settings allows other containers in the pod to see environment variables and file systems.
    Some tenant may not accept this case.
- [yahoo/kubectl-flame](https://github.com/yahoo/kubectl-flame)
  - kubectl-flame is a kubectl plugin that allows profiling of applications on kubernetes
  - kubectl-flame performs profiling of NodeJS applications by using perf.
  - The command-line arguments of kubectl-flame's perf are hard-coded and the arguments cannot be changed except for the execution time.
<https://github.com/yahoo/kubectl-flame/blob/master/agent/profiler/perf.go#L60>
  - kubectl-flame only supports docker runtime and does not support containerd runtime.
<https://github.com/yahoo/kubectl-flame/issues/51>
- [iovisor/kubectl-trace](https://github.com/iovisor/kubectl-trace)
  - kubectl-trace is a kubectl plugin to schedule bpftrace programmers against Pods on a Kubernetes cluster
  - kubectl-trace only supports tracing against Pods and does not support profiling
- [giannisalinetti/perf-utils](https://github.com/giannisalinetti/perf-utils)
  - The container image of perf-utils installs tools for performance analysis and troubleshooting for immutable systems such as Fedora CoreOS
  - perf-utils does not install a perf compatible with the host kernel version

Explains the problems with the sidecar container method and the Ephemeral Container method.

- The sidecar container method requires the sidecar container to be deployed beforehand.
  If you deploy the sidecar container later, you must tolerate the pod to restart.
- As of Kubernetes 1.26, once an Ephemeral Container is added to a Pod, it cannot be changed or removed
  > Like regular containers, you may not change or remove an ephemeral container after you have added it to a Pod.
  [Ephemeral Container](https://kubernetes.io/docs/concepts/workloads/pods/ephemeral-containers/#understanding-ephemeral-containers)
- The cluster administrator needs to authorise the tenant to set permissions such as `CAP_SYS_ADMIN` to Pod
- It is difficult for tenant to prepare a version of perf command that is compatible with the host OS
