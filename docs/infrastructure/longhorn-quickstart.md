# Longhorn + Go Load Lab Quick Start

This document walks you through installing **Longhorn** (a dynamic storage provider) and the **Go Load Lab** Helm chart with file‐based logging on your Kubernetes cluster.

## 1. Prepare Ubuntu for Longhorn

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

## 2. Install Longhorn

1. **Add** the Longhorn Helm repo and update:
```bash
helm repo add longhorn https://charts.longhorn.io
helm repo update
```
2. **Install** Longhorn in a dedicated namespace (e.g., `longhorn-system`), specifying kubeletRootDir if needed:
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
3. Wait for all **Longhorn** Pods to become `Running`:
```bash
kubectl get pods -n longhorn-system
```
Once healthy, the **Longhorn** StorageClass should be available cluster‐wide.

## 3. Add the Go Load Lab Helm Repository

If you haven’t already:

```bash
helm repo add go-load-lab https://JoobyPM.github.io/go-load-lab
helm repo update
```

## 4. Install Go Load Lab with Persistent Logs

Enable logging and persistent volume usage in the Helm chart:

```bash
helm install go-load-lab-chart go-load-lab/go-load-lab-helmchart \
  --version 0.3.0 \
  --set logging.enabled=true \
  --set logging.logFile=/app/logs/go-load-lab.log \
  --set logging.persistentVolume.enabled=true \
  --set logging.persistentVolume.storageClass=longhorn \
  --set logging.persistentVolume.size=1Gi
```

Here’s what happens:

1. **Logging** is turned on (`LOG_FILE` env var).  
2. A **PersistentVolumeClaim** named `go-app-logs-pvc` is auto‐created via Longhorn.  
3. Logs write to `/app/logs/go-load-lab.log` on a persistent volume.

## 5. Validate & Access

1. **Check** that Pods are running:
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
4. Visit the Longhorn UI to monitor volumes, or see logs in `/app/logs/go-load-lab.log`.

kubectl debug pod/go-app-5c7c59d879-25ncx -n go-app \
  --image=ubuntu:22.04 \
  --target=go-app-container \
  -- bash
