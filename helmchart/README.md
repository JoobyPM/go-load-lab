# Helm Chart – Go Load Lab

This directory contains a **Helm chart** to deploy the [Go Load Lab](https://github.com/JoobyPM/go-load-lab) application. It creates a Deployment, Service, Ingress, and optional Horizontal Pod Autoscaler (HPA) on **MicroK8s** or any standard Kubernetes cluster.

## Prerequisites

1. **Helm 3.x** installed (either standalone or via `microk8s helm3`).
2. **Kubernetes** cluster:
   - For **MicroK8s**, see the quick start steps below.
   - For other clusters, ensure you have the necessary permissions to install Helm charts and create cluster resources.

## Quick Start (MicroK8s)

1. **Enable necessary add-ons** in MicroK8s:
   ```bash
   microk8s enable ingress
   microk8s enable metrics-server
   microk8s enable metallb:192.168.68.230-192.168.68.239
   ```
   - **ingress**: Adds the NGINX Ingress Controller.  
   - **metrics-server**: Needed for HPA to gather CPU metrics.  
   - **metallb**: For LoadBalancer IP allocation (using `192.168.68.230-192.168.68.239` as an example).

2. **(Optional) Enable local registry**:
   ```bash
   microk8s enable registry
   ```
   - Makes it easier to push local Docker images to `localhost:32000`.

3. **(Optional) Enable Helm 3 in MicroK8s**:
   ```bash
   microk8s enable helm3
   ```
   - Then use `microk8s helm3` in place of `helm`. For example:
     ```bash
     microk8s helm3 install go-load-lab ./helmchart \
       --set image.repository=localhost:32000/go-load-lab \
       --set image.tag="1.0.1"
     ```
     Alternatively, you can use a local Helm 3 install (without MicroK8s) if you prefer.

4. **Build & push the Docker image** (if you want to customize locally):
   ```bash
   docker build -t localhost:32000/go-load-lab:1.0.1 .
   docker push localhost:32000/go-load-lab:1.0.1
   ```
   Or use a Docker Hub repository if you prefer.

5. **Install the chart** (assuming you have Helm 3 available):
   ```bash
   # Using a local Helm install
   helm install go-load-lab ./helmchart \
     --set image.repository=localhost:32000/go-load-lab \
     --set image.tag="1.0.1"
   ```
   The above command:
   - Installs your chart under the release name `go-load-lab`.
   - Overrides the default image repository/tag with your local image.

6. **Confirm**:
   ```bash
   microk8s kubectl get pods -n go-app
   microk8s kubectl get svc -n go-app
   microk8s kubectl get ingress -n go-app
   ```
   - The chart defaults to a **LoadBalancer** Service, so MetalLB should assign an IP from your configured range.
   - An Ingress named `go-app-ingress` will route traffic from `go-app.your-domain.com` (adjustable via `values.yaml`).

7. **Test**:
   - If `go-app.your-domain.com` is mapped to your node’s IP (or to the assigned MetalLB IP), open:
     ```
     http://go-app.your-domain.com
     ```
   - If you prefer direct IP: see the Service’s external IP with:
     ```
     microk8s kubectl get svc go-app-service -n go-app
     ```

## Customization

Use `--set` or edit `values.yaml` to adjust:

- **`replicaCount`**: Number of replicas for the Deployment.  
- **`image.repository`**, **`image.tag`**: Docker image reference.  
- **`service.type`**: `LoadBalancer`, `ClusterIP`, or `NodePort`.  
- **`ingress.enabled`**: Toggle Ingress on/off; set your own hostname.  
- **`ingress.traefikTls.certresolver`**: If you’d like **Traefik** to issue and serve HTTPS certificates automatically, set this to something like `route53resolver` (or your chosen resolver). Doing so adds the Traefik TLS annotations and a `tls:` block referencing your Ingress host.  
- **`hpa.enabled`**: Enable/disable HorizontalPodAutoscaler, etc.

### Example advanced override

```bash
helm install go-load-lab ./helmchart \
  --set replicaCount=3 \
  --set image.repository=myuser/go-load-lab \
  --set image.tag="2.0.0" \
  --set ingress.host=my-custom-domain.com \
  --set hpa.enabled=false
```

### Example for Traefik HTTPS

```bash
helm install go-load-lab ./helmchart \
  --set ingress.enabled=true \
  --set ingress.host=my-tls-app.domain.com \
  --set ingress.traefikTls.certresolver=route53resolver
```

This automatically adds:
```
traefik.ingress.kubernetes.io/router.tls: "true"
traefik.ingress.kubernetes.io/router.tls.certresolver: "route53resolver"
```
…plus a `tls:` block referencing `my-tls-app.domain.com`.

## Logging to File (Optional)

By default, this chart logs **only** to stdout. If you’d like to log to a file:

1. **Enable logging** in `values.yaml` (or via `--set`) by setting:
   ```yaml
   logging:
     enabled: true
     logFile: "/app/logs/go-load-lab.log"
     persistentVolume:
       enabled: true
       storageClass: "longhorn"  # Or another StorageClass
       size: "1Gi"
   ```
2. When `logging.enabled=true`, a **ConfigMap** sets `LOG_FILE` in the container environment, and the application writes logs to that path.
3. If `logging.persistentVolume.enabled=true`, a **PVC** called `go-app-logs-pvc` is created, which gets mounted at `/app/logs`. This makes logs **persistent**.
4. Make sure your cluster has a dynamic provisioner (like [Longhorn](https://longhorn.io/)) to satisfy the PVC request.
5. Example Helm install with file logging enabled:
   ```bash
   helm install go-load-lab ./helmchart \
     --set logging.enabled=true \
     --set logging.logFile=/app/logs/go-load-lab.log \
     --set logging.persistentVolume.enabled=true \
     --set logging.persistentVolume.storageClass=longhorn \
     --set logging.persistentVolume.size=2Gi
   ```

> **Note**: If you **don’t** enable the logging features, no ConfigMap or PVC is created, and the app logs only to stdout.

## Installing Longhorn (If Needed)

If you want to use **Longhorn** as your dynamic storage provider, you must install it **separately**. For example, via Helm:

```bash
helm repo add longhorn https://charts.longhorn.io
helm repo update
helm install longhorn longhorn/longhorn --namespace longhorn-system --create-namespace
```

After it’s installed and running in your cluster, you can reference the **`longhorn`** StorageClass from this chart (as shown above) to automatically provision persistent volumes for logs or other data. We **do not** bundle Longhorn as a subchart because storage solutions are typically cluster-level infrastructure, and many users already have their own storage classes. Keeping storage separate makes this chart more flexible and avoids coupling the application deployment to a specific storage provider.

## Updating / Uninstalling

- **Upgrade**:
  ```bash
  helm upgrade go-load-lab ./helmchart
  ```
- **Uninstall**:
  ```bash
  helm uninstall go-load-lab
  ```
  This removes all resources defined by this chart. The `go-app` **namespace** will remain unless you remove it manually.

## Contents

```
helmchart/
├── Chart.yaml         # Chart metadata
├── values.yaml        # Default values (image, replicas, logging, etc.)
└── templates/
    ├── configmap-logs.yaml
    ├── pvc-logs.yaml
    ├── namespace.yaml
    ├── deployment.yaml
    ├── service.yaml
    ├── ingress.yaml
    └── hpa.yaml
```

## License

Please see [LICENSE](../LICENSE) in the project root for details.