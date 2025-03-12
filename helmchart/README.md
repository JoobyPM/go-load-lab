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
- **`hpa.enabled`**: Enable/disable HorizontalPodAutoscaler, etc.

Example advanced override:
```bash
helm install go-load-lab ./helmchart \
  --set replicaCount=3 \
  --set image.repository=myuser/go-load-lab \
  --set image.tag="2.0.0" \
  --set ingress.host=my-custom-domain.com \
  --set hpa.enabled=false
```

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
├── values.yaml        # Default values (image, replicas, etc.)
└── templates/
    ├── namespace.yaml
    ├── deployment.yaml
    ├── service.yaml
    ├── ingress.yaml
    └── hpa.yaml
```

## License

Please see [LICENSE](../LICENSE) in the project root for details.