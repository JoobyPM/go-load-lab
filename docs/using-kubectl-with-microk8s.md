# How to Install and Configure kubectl with MicroK8s

## 1. Install kubectl
Follow the official Kubernetes documentation to install **kubectl** on Linux:
[Install kubectl on Linux](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/)

## 2. Use the Built-In MicroK8s kubectl (Simplest Approach)
MicroK8s includes a built-in `kubectl`, which you can call directly:
```bash
microk8s kubectl get pods --all-namespaces
```
Or create a shortcut (alias) so you can type `kubectl` as usual:
```bash
sudo snap alias microk8s.kubectl kubectl
```
Then test:
```bash
kubectl get nodes
kubectl get pods --all-namespaces
```

## 3. Use Your Installed kubectl with MicroK8s Config
If you prefer using your standalone `kubectl` installation, point it to the MicroK8s cluster by exporting the MicroK8s config:

1. **Export MicroK8s config to `~/.kube/config`**:
```bash
microk8s config > ~/.kube/config
```
Make sure `~/.kube/` exists (create it if needed).

2. **Verify the config**:
```bash
cat ~/.kube/config
```
You should see an entry that includes `server: https://127.0.0.1:16443` (the typical MicroK8s API endpoint).

3. **Test your standalone kubectl**:
```bash
kubectl get nodes
kubectl get pods --all-namespaces
```
You should now be able to manage your MicroK8s cluster using your own `kubectl` installation.