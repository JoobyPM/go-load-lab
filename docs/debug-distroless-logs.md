# Debugging Logs in a Distroless Container

When running **Go Load Lab** (or any application) in a **Distroless** container, you cannot simply `exec` into a shell environment. Below are some approaches to **view or debug** the log file (`/app/logs/go-load-lab.log`) in Kubernetes.

---

## 1. Use `kubectl debug` with an Ephemeral Container

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

---

## 2. Temporarily Switch to a Non-Distroless Image

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

---

## 3. Mount the Logs Volume Elsewhere

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

---

## Summary

- **`kubectl debug`** is the quickest way to see logs in a Distroless container.  
- **Switching to a shell-based image** or **mounting the same volume** are alternatives if needed.  

Use whichever method suits your environment, but remember that **Distroless** images are intentionally minimal—so ephemeral debug containers are usually the safest and most direct approach for log inspection.