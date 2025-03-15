# Preparing Infrastructure for Go Load Lab

This guide provides a **high‐level** sequence for preparing your environment before deploying the **Go Load Lab** application. It covers:

1. Installing **MicroK8s** (or ensuring a compatible Kubernetes cluster).
2. Adding a **Load Balancer** (MetalLB) for bare‐metal IP allocation.
3. Enabling **Persistent Storage** (Longhorn).
4. Optionally installing **Traefik** if you want a DNS‐challenge–based HTTPS solution.

> **Note**: Each step links to a more detailed sub‐guide within `docs/infrastructure/`.

---

## 1. MicroK8s Basics (Optional)

- If you’re new to MicroK8s or want a **highly available** setup with multiple nodes, see [ha-microk8s.md](./ha-microk8s.md).
- For single‐node usage or minimal labs, you can simply:
  ```bash
  snap install microk8s --classic
  microk8s status --wait-ready
  ```
- Enable relevant MicroK8s add‐ons if desired (e.g., `ingress`, `metrics-server`).  
  ```bash
  microk8s enable ingress
  microk8s enable metrics-server
  # Disable the built-in metallb if you want to install the official chart.
  microk8s disable metallb
  ```

---

## 2. Install MetalLB (Official Helm Chart)

If you need a **LoadBalancer** on bare metal (home lab, VM, etc.), we recommend the **official** MetalLB chart. Follow the instructions in:

- [install-metallb-helm.md](./install-metallb-helm.md)

…which covers:

1. Disabling the built-in MicroK8s metallb add-on  
2. Installing “vanilla” MetalLB with Helm (`helm install metallb metallb/metallb`)  
3. Creating IP pools via IPAddressPool / L2Advertisement

This lets your **LoadBalancer** Services get an external IP from a specified range (e.g. `192.168.68.230-192.168.68.239`).

---

## 3. Install Longhorn for Persistent Storage

For dynamic block storage, see our:

- [longhorn-quickstart.md](./longhorn-quickstart.md)

It details:

1. Ubuntu dependencies (iSCSI, NFS, etc.)  
2. Installing Longhorn (`helm install longhorn longhorn/longhorn ...`)  
3. Verifying the Longhorn pods & StorageClass

This is particularly useful if you plan to enable **file‐based logging** or other persistent volumes in Go Load Lab. If you have another preferred storage solution, you can skip Longhorn.

---

## 4. Optionally Install Traefik (DNS Challenge Edition)

If you’d like an **Ingress controller** that can automatically provision **Let’s Encrypt** certificates (e.g., via Route 53 DNS challenge), see:

- [traefik.md](./traefik.md)

That guide explains:

1. Installing Traefik from its official Helm chart  
2. Using the **LoadBalancer** IP from MetalLB  
3. Configuring DNS challenge (Route 53), storing ACME data on Longhorn, and exposing 80/443

*Note*: If you prefer the built‐in MicroK8s ingress or a different controller, you can skip this step.

---

## 5. Deploy Go Load Lab

Once your infrastructure is set:

1. **Clone** the repository (or add our Helm repo).  
2. **Install** the Helm chart (`helm install go-load-lab ./helmchart`) or from the remote chart.  
3. **Confirm** everything is running:  
   ```bash
   kubectl get pods -n go-app
   kubectl get svc -n go-app
   kubectl get ingress -n go-app
   ```

You can now test load endpoints, attach persistent volumes, enable HTTPS, and more. See the main [README.md](../../README.md) or the [helmchart/README.md](../../helmchart/README.md) for advanced options.

---

### Summary

- **MicroK8s**: optional or use any K8s distribution you prefer.  
- **MetalLB**: official chart for a true LoadBalancer.  
- **Longhorn**: for dynamic RWX storage (or any other storage solution).  
- **Traefik**: for DNS-challenge TLS if desired.  

When all pieces are in place, proceed to deploy **Go Load Lab** with any optional features (logging, HPA, Ingress) you’d like.