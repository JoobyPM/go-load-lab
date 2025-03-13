# Go Load Lab

A simple Go-based web server designed for Kubernetes and containerization practice. It exposes various endpoints for testing load, latency, liveness/readiness checks, and also serves a small UI for manual load testing.

## Features

- **Static file serving** (via `/`)
- **Health checks**:
  - **Liveness:** `/livez` (always returns `200 OK`)
  - **Readiness:** `/readyz` (depends on in-memory cache hydration)
- **Load testing endpoints**:
  - **`/havy-call?cpu=100m&duration=5`**: CPU load simulation
  - **`/wait?time=100`**: Wait simulation (latency)
  - **`/items`**: Paginated items (in-memory data)
- **Interactive UI** at `/` for quick testing

## Repository Structure

```
.
├── cmd/
│   └── server/
│       └── main.go        # Main entry point (starts the HTTP server)
├── internal/
│   ├── cache/             # Cache logic, item types, hydration
│   └── handlers/          # HTTP handler functions (health checks, load tests, etc.)
├── static/                # Static files (index.html, style.css, etc.)
├── helmchart/             # Helm chart for Kubernetes deployment
├── k8s/                   # (Optional) legacy Kubernetes manifests
├── docs/
│   ├── assets/
│   │   ├── ha-microk8s.puml  # PlantUML diagram
│   │   ├── ...               # Other assets (e.g. images)
│   │   └── ha-microk8s.png   # PNG image of the diagram
│   └── ha-microk8s.md        # Overview of HA MicroK8s cluster
├── Dockerfile             # Multi-stage build (Go + Distroless)
├── Makefile               # Docker tasks (build, push, run, multi-arch)
├── go.mod                 # Go module definition
├── LICENSE                # MIT License
└── README.md              # This file
```

## Quick Start (Local)

1. **Clone & Setup**

   ```bash
   git clone https://github.com/JoobyPM/go-load-lab.git
   cd go-load-lab
   go mod tidy
   ```

2. **Local Build & Run (Go)**

   ```bash
   cd cmd/server
   go build -o server
   ./server
   ```
   - The server listens on port `8080`.
   - Open [http://localhost:8080](http://localhost:8080) to access the UI.

3. **Local Build & Run (Docker)**

   ```bash
   # Build image
   docker build -t go-load-lab:local .
   
   # Run container
   docker run -p 8080:8080 --name go-load-lab go-load-lab:local
   ```
   - Visit [http://localhost:8080](http://localhost:8080) in your browser.

## Kubernetes Deployment

You can deploy via:

- **Helm**: Recommended. See [helmchart/README.md](./helmchart/README.md) for usage, including optional logging with persistent volumes.  
- **Legacy YAML**: Original manifests in [k8s/](./k8s). You can apply them directly with:
  ```bash
  kubectl apply -f k8s/
  ```

## HA MicroK8s Setup

For a **highly available MicroK8s cluster** (with multiple control-plane nodes, worker nodes, and an Ingress/LoadBalancer), see [docs/ha-microk8s.md](./docs/ha-microk8s.md).

### Running on Your Mac or Home Lab

- You can run MicroK8s in **VirtualBox VMs** on a MacBook (e.g., MacBook Pro M3).  
- **Minimum** recommended specs for each VM:  
  - **1 vCPU** and **2GB RAM** (absolute minimum)  
  - For better performance, allocate **2 vCPUs** and **4GB RAM** if you have enough resources.  
- Create multiple VMs (e.g., 2 or 3 control-plane nodes + 1 or 2 worker nodes) and join them in a single MicroK8s cluster.

## Contributing

- Open issues or PRs if you’d like to extend the application or add new endpoints.
- For major changes, please open an issue first to discuss.

## Why We Don’t Bundle Longhorn as a Subchart

Longhorn is a **cluster‐level storage solution**. Typically, you install it once (via its own Helm chart or YAML) so **all** workloads in the cluster can use it. We keep storage provisioning (like Longhorn) **separate** from this application chart so users can choose any storage class they prefer. This decoupling avoids unneeded complexity and ensures your cluster’s storage setup remains flexible.

If you want to install Longhorn, see [Longhorn’s official docs](https://longhorn.io/) or use its Helm chart:
```bash
helm repo add longhorn https://charts.longhorn.io
helm install longhorn longhorn/longhorn --namespace longhorn-system --create-namespace
```

Then in our Helm chart’s `values.yaml`, set:
```yaml
logging:
  enabled: true
  persistentVolume:
    enabled: true
    storageClass: longhorn
    size: 1Gi
```
…to enable persistent logging with Longhorn or any dynamic provisioner.

## License

This project is licensed under the [MIT License](LICENSE).