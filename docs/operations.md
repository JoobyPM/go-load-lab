# Operations & Maintenance Guide

This document covers **four** essential operational topics for the Go Load Lab (or similar) application:

1. [Debugging Logs in a Distroless Container](#1-debugging-logs-in-a-distroless-container)  
2. [Linting & Validating the Helm Chart](#2-linting--validating-the-helm-chart)  
3. [Manually Unloading (Cordon/Drain) a Kubernetes Node](#3-manually-unloading-a-kubernetes-node)  
4. [Using MicroK8s & Longhorn Quick Start](#4-using-microk8s--longhorn-quick-start)

---

## 1. Debugging Logs in a Distroless Container

When running **Go Load Lab** (or any application) in a **Distroless** container, you cannot simply `exec` into a shell environment. Below are some approaches to **view or debug** the log file (`/app/logs/go-load-lab.log`) in Kubernetes.

### 1.1. Use `kubectl debug` with an Ephemeral Container

The **kubectl debug** command can inject a temporary debug container (e.g. Ubuntu) into the same Pod, letting you run normal shell tools:

1. Identify your Pod (e.g., `go-app-abc123-xyz` in namespace `go-app`):
   ```bash
   kubectl get pods -n go-app
   ```
2. Inject a debug container:
   ```bash
   kubectl debug pod/go-app-abc123-xyz -n go-app \
     --image=ubuntu:22.04 \
     --target=go-app-container \
     -- bash
   ```
   - **`--target=go-app-container`** ensures the debug container shares the same volume mounts as the main container.

3. In the ephemeral container shell:
   ```bash
   cat /app/logs/go-load-lab.log
   ```
   or
   ```bash
   tail -f /app/logs/go-load-lab.log
   ```
   This way, you can read logs or run other debugging commands.

### 1.2. Temporarily Switch to a Non‐Distroless Image

If **kubectl debug** is not available, you could **edit** your Pod spec (or the Helm values) to use an Ubuntu‐based (or Alpine) image that has a shell. For example:

```yaml
image:
  repository: ubuntu
  tag: "22.04"
```

Then deploy and use:
```bash
kubectl exec -it go-app-abc123-xyz -n go-app -- bash
cat /app/logs/go-load-lab.log
```

This approach modifies your actual container image (not recommended for production), but it can help in a pinch for troubleshooting.

### 1.3. Mount the Logs Volume Elsewhere

If you’re persisting logs via a **PVC** (e.g. with Longhorn), you can mount that same PVC in a separate “debug” Pod or on a host path, then view the file externally. For example, create a new Pod with the same volume claim:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: logs-viewer
  namespace: go-app
spec:
  containers:
  - name: logs-viewer
    image: ubuntu:22.04
    command: [ "sleep", "infinity" ]
    volumeMounts:
    - name: logs-volume
      mountPath: /app/logs
  volumes:
  - name: logs-volume
    persistentVolumeClaim:
      claimName: go-app-logs-pvc
```

Then:
```bash
kubectl exec -it logs-viewer -n go-app -- bash
cd /app/logs
tail -f go-load-lab.log
```

### 1.4. Summary (Distroless Logs)
- **`kubectl debug`** is the quickest way to see logs in a Distroless container.  
- **Switching to a shell-based image** or **mounting the same volume** are alternatives if needed.

Use whichever method suits your environment, but remember that **Distroless** images are intentionally minimal—so ephemeral debug containers are usually the safest and most direct approach for log inspection.

---

## 2. Linting & Validating the Helm Chart

This document briefly explains how to **lint** and **validate** the Helm chart in this repository.

### 2.1. Helm Lint

Helm provides a built-in linting command to catch syntax errors or missing values:

```bash
# From the helmchart/ directory:
helm lint .
```

Typical **helm lint** checks include:
- YAML syntax correctness
- Presence of required keys in `Chart.yaml`
- Template rendering issues

If everything is correct, you’ll see a success message like:
```
1 chart(s) linted, 0 chart(s) failed
```
Otherwise, fix the warnings/errors before deploying.

### 2.2. Render & Validate Manifests

You can also render the chart templates and validate them with external tools:

```bash
helm template . --namespace my-namespace > rendered.yaml
```

Then pass `rendered.yaml` to a **Kubernetes schema validator** such as [kubeconform](https://github.com/yannh/kubeconform):

```bash
kubeconform -strict rendered.yaml
```

> **Note**: For additional checks, you can also **render** your templates and pipe them through Kubernetes validators:
> ```bash
> helm template /path/to/chart | kubeconform -strict
> ```
> This checks whether the generated manifests match official Kubernetes API schemas.

#### Why Validate Manifests?
- **Catches** advanced schema issues (e.g. outdated `apiVersion`).  
- **Ensures** your chart’s resources conform to the cluster’s Kubernetes version.

### 2.3. CI/CD Integration

- In a CI environment (e.g. GitHub Actions, GitLab CI), you can run these commands automatically:
  - **Lint** the chart via `helm lint`.
  - **Render** via `helm template` and pipe to a schema validator.
- Tools like [chart-testing](https://github.com/helm/chart-testing) can automate linting, validation, and even install tests.

#### Summary (Helm Lint & Validate)
1. **`helm lint .`** → quick check for basic mistakes.  
2. **`helm template . | kubeconform`** → thorough validation against schema.  
3. **CI/CD** → automate repeated checks to maintain chart quality.

---

## 3. Manually Unloading a Kubernetes Node

This quick guide covers how to temporarily move all workloads off a node (e.g., for maintenance) and then allow new Pods again.

### 3.1. Steps

1. **Cordon the node**
   ```bash
   kubectl cordon <node-name>
   ```
   > This marks the node as unschedulable, preventing *new* pods from being scheduled onto it.

2. **Drain the node**
   ```bash
   kubectl drain <node-name> --ignore-daemonsets --delete-emptydir-data
   ```
   > This evicts all *non-DaemonSet* pods on the node, forcing them to be rescheduled on other nodes.  
   > `--delete-emptydir-data` allows emptyDir volumes to be wiped when pods are evicted.

3. **Uncordon the node**
   ```bash
   kubectl uncordon <node-name>
   ```
   > This makes the node schedulable again. New pods may land there once it’s back in service.

### 3.2. Troubleshooting & Tips

- **Pods did not reschedule**  
  - Ensure other nodes have enough free capacity (check CPU/memory requests vs. node allocatable).
  - Verify no conflicting `nodeSelector`, `affinity`, or `taints` are blocking pods from moving.

- **Control-plane nodes unevenly loaded**  
  - If one control-plane node has no pods while another is overloaded, confirm there are **no `NoSchedule` taints** on the “unused” node (e.g., `node-role.kubernetes.io/master:NoSchedule`). Removing that taint lets workloads schedule there.
  - Consider a [descheduler](https://github.com/kubernetes-sigs/descheduler) if you want automatic rebalance of long-running pods.

- **Maintenance tasks**  
  - Once the node is drained, you can safely do OS/kernel upgrades or hardware checks.  
  - When ready, uncordon so workloads can move back.

#### Example Usage
(substitute your domain/node name):
```bash
# Suppose the node is vs3.domain.com
kubectl cordon vs3.domain.com
kubectl drain vs3.domain.com --ignore-daemonsets --delete-emptydir-data
# ...perform maintenance...
kubectl uncordon vs3.domain.com
```

---

## 4. Using MicroK8s & Longhorn Quick Start

> **Note**: The file named `using-kubectl-with-microk8s.md` originally contained instructions for installing **Longhorn** on a MicroK8s cluster and then deploying **Go Load Lab** with persistent logs. Below is the content merged for convenience.

---

### 4.1. Introduction

This guide helps you prepare **Ubuntu** for Longhorn, install **Longhorn** itself, and then deploy the **Go Load Lab** Helm chart with file-based logging. The steps assume a MicroK8s cluster (or any standard K8s environment), but are most often used in a home-lab or minimal environment.

### 4.2. Prepare Ubuntu for Longhorn

1. **Update packages**:
   ```bash
   sudo apt-get update
   ```
2. **Install necessary dependencies** (iSCSI, NFS, cryptsetup/dmsetup):
   ```bash
   sudo apt-get install -y open-iscsi nfs-common cryptsetup dmsetup
   sudo systemctl enable iscsid
   sudo systemctl start iscsid
   ```
3. **(Optional) Enable the iSCSI TCP kernel module** if available:
   ```bash
   sudo modprobe iscsi_tcp
   lsmod | grep iscsi
   echo "iscsi_tcp" | sudo tee /etc/modules-load.d/iscsi_tcp.conf
   ```
4. **Run the Longhorn environment check** to verify everything:
   ```bash
   curl -sSfL https://raw.githubusercontent.com/longhorn/longhorn/v1.8.1/scripts/environment_check.sh | bash
   ```
   - If there are errors (e.g., missing `iscsi_tcp`), fix them, then re-run the check until it passes.

### 4.3. Install Longhorn (with Helm)

1. **Add** the Longhorn Helm repo and update:
   ```bash
   helm repo add longhorn https://charts.longhorn.io
   helm repo update
   ```
2. **Install** Longhorn in a dedicated namespace (e.g., `longhorn-system`), specifying your MicroK8s `kubeletRootDir`:
   ```bash
   helm install longhorn longhorn/longhorn \
     --namespace longhorn-system \
     --create-namespace \
     --set defaultSettings.kubeletRootDir="/var/snap/microk8s/common/var/lib/kubelet" \
     --set csi.kubeletRootDir="/var/snap/microk8s/common/var/lib/kubelet" \
     --set global.cattle.psp.enabled=false \
     --set defaultSettings.defaultReplicaCount=2 \
     --set defaultSettings.guaranteedEngineCPU=0 \
     --set defaultSettings.replicaSoftAntiAffinity=true \
     --set nfs.defaultClass=true
   ```
   - Wait for all **Longhorn** Pods to become `Running`:
     ```bash
     kubectl get pods -n longhorn-system
     ```

### 4.4. Deploy Go Load Lab with Persistent Logs

If you haven’t already added the chart:
```bash
helm repo add go-load-lab https://JoobyPM.github.io/go-load-lab
helm repo update
```

Now install it:

```bash
helm install go-load-lab-chart go-load-lab/go-load-lab-helmchart \
  --version 0.3.0 \
  --set logging.enabled=true \
  --set logging.logFile=/app/logs/go-load-lab.log \
  --set logging.persistentVolume.enabled=true \
  --set logging.persistentVolume.storageClass=longhorn \
  --set logging.persistentVolume.size=1Gi
```

**What Happens**:

1. **File-based logging** is turned on (`LOG_FILE` env var).  
2. A **PVC** named `go-app-logs-pvc` is auto-created with the **longhorn** StorageClass.  
3. Logs are written to `/app/logs/go-load-lab.log`, persisting across restarts.

### 4.5. Validate & Access

1. **Check** the Pods:
   ```bash
   kubectl get pods -n go-app
   ```
2. **Look** for your logs PVC:
   ```bash
   kubectl get pvc -n go-app
   ```
3. If `service.type=LoadBalancer`, see which IP was assigned:
   ```bash
   kubectl get svc go-app-service -n go-app
   ```
4. **Access** the Longhorn UI at `http://<NodeIP>:<NodePort>` or Ingress. Monitor volumes, check replication, etc.

#### Debugging Distroless logs
Because the container is Distroless, consider `kubectl debug` or a debug sidecar if you need to read the logs from inside the Pod. (Refer to the [Debugging Logs in a Distroless Container](#1-debugging-logs-in-a-distroless-container) section above.)

**Congratulations!** You have **Longhorn** installed and **Go Load Lab** running with persistent logs on MicroK8s.