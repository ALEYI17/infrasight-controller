# 🎛️ InfraSight Controller

The **InfraSight Controller** is a Kubernetes-native way to deploy and manage the InfraSight eBPF telemetry agents across your cluster nodes.

It leverages [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) to define a custom controller and CRD (`EbpfDaemonSet`) that enables fine-grained control over how the eBPF agents (from [`ebpf_loader`](https://github.com/ALEYI17/ebpf_loader)) are deployed via DaemonSets.

## 🚀 Overview

This controller automates the deployment of eBPF agent DaemonSets in a Kubernetes cluster. Each agent is configured via the `EbpfDaemonSet` custom resource and deployed with the necessary permissions, host volumes, and runtime settings.

## 📦 Features

- ⚙️ Built with **Kubebuilder**
- 🎯 Deploys a **DaemonSet** with necessary volumes and security contexts
- ✅ Includes **webhooks** for:
  - Defaulting unset values
  - Validating configuration (e.g., image format, resource limits, node selectors)
- 🔄 Reconciles updates to the `EbpfDaemonSet` CR
- 🧪 Supports eBPF probes like `execve`, `accept`, and others

## 🧾 `EbpfDaemonSetSpec` Definition

```go
type EbpfDaemonSetSpec struct {
  Image         string                      `json:"image,omitempty"`
  NodeSelector  map[string]string           `json:"nodeSelector,omitempty"`
  Tolerations   []corev1.Toleration         `json:"tolerations,omitempty"`
  Resources     corev1.ResourceRequirements `json:"resources,omitempty"`
  RunPrivileged bool                        `json:"runPrivileged,omitempty"`
  EnableProbes  []string                    `json:"enableProbes"`
  ServerAddress string                      `json:"serverAddress"`
  ServerPort    string                      `json:"serverPort"`
}
````

Each field allows customizing the agent behavior, deployment affinity, and runtime options.

## 🧪 Project Status

> 🚧 **Alpha stage** — `v1alpha1`
>
> This project is under active development. APIs may change.

## 🛠️ Deployment Instructions

At the moment, the controller is not published as a Helm chart or image registry, so to deploy it manually:

1. Clone the repository:

   ```bash
   git clone https://github.com/ALEYI17/infrasight-controller.git
   cd infrasight-controller
   ```

2. Build and push the image:

   ```bash
   make docker-build docker-push IMG=<your-registry>/<image>:<tag>
   ```

3. Deploy cert-manager (required for webhook certificates):

   ```bash
   kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.17.2/cert-manager.yaml
   ```

4. Deploy the controller to your cluster:

   ```bash
   make deploy IMG=<your-registry>/<image>:<tag>
   ```

5. To undeploy:

   ```bash
   make undeploy
   ```

You can follow the [Kubebuilder book tutorial](https://book.kubebuilder.io/cronjob-tutorial/running) for a full walkthrough.

## 📄 Example CR

A sample `EbpfDaemonSet` is available at:

```
config/samples/ebpf_v1alpha1_ebpfdaemonset.yaml
```

Example snippet:

```yaml
apiVersion: ebpf.monitoring.dev/v1alpha1
kind: EbpfDaemonSet
metadata:
  name: example-ebpfds
spec:
  image: aley17/ebpf_loader:latest
  enableProbes:
    - execve
    - accept
  serverAddress: ebpf-server.default.svc
  serverPort: "8080"
```

## 📚 Related Repositories

This is part of the **[InfraSight](https://github.com/ALEYI17/InfraSight)** platform:

- [`infrasight-controller`](https://github.com/ALEYI17/infrasight-controller): Kubernetes controller to manage agents
- [`ebpf_loader`](https://github.com/ALEYI17/ebpf_loader): Agent that collects and sends eBPF telemetry from nodes
- [`ebpf_server`](https://github.com/ALEYI17/ebpf_server): Receives and stores events (e.g., to ClickHouse)
- [`ebpf_deploy`](https://github.com/ALEYI17/ebpf_deploy): Helm charts to deploy the stack

