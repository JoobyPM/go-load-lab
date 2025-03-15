Below is an **updated** guide for installing MetalLB “the official way” (from its Helm chart or manifests), **tailored** to a MicroK8s setup. These instructions incorporate key points from the **MetalLB official documentation**, but add MicroK8s‐specific context where useful.

---

## 1. Prerequisites & Checks

1. **Disable the built‐in MicroK8s Add‐On**  
   If you previously enabled the MetalLB add‐on in MicroK8s, **disable** it:
   ```bash
   microk8s disable metallb
   ```
   We will now install “vanilla” MetalLB from the official upstream.

2. **Check Your Network Mode**  
   - MetalLB can work in either `iptables` or `ipvs` proxy mode.  
   - If you’re using IPVS **and** your Kubernetes version is ≥1.14.2, you must enable **strict ARP** (see below). Otherwise, if you’re using the default iptables proxy (typical in MicroK8s), you can skip the strict ARP step.

3. **(Optional) Strict ARP for IPVS**  
   If your MicroK8s cluster is in **IPVS** mode, do:
   ```bash
   kubectl edit configmap -n kube-system kube-proxy
   ```
   Ensure the config includes:
   ```yaml
   apiVersion: kubeproxy.config.k8s.io/v1alpha1
   kind: KubeProxyConfiguration
   mode: "ipvs"
   ipvs:
     strictARP: true
   ```
   Then restart the affected Pods or wait for them to reconcile.

4. **Confirm your environment**  
   - If you are on a **cloud** VM, check the [MetalLB cloud compatibility docs](https://metallb.universe.tf/installation/clouds/) – many clouds do not allow ARP or BGP announcements.  
   - For bare metal or “home lab” MicroK8s, you’re usually good to proceed.

---

## 2. Install MetalLB via **Helm** (Recommended)

Below we install the official Helm chart from the official MetalLB repository.

### 2.1. Add the MetalLB Helm Repository

```bash
helm repo add metallb https://metallb.github.io/metallb
helm repo update
```

### 2.2. Install the Chart

```bash
helm install metallb metallb/metallb \
  --namespace metallb-system \
  --create-namespace
```

This will create:
- A **controller** Deployment (which assigns IP addresses to Services).
- A **speaker** DaemonSet (which does ARP/BGP to advertise those addresses).
- All necessary RBAC resources.

> **Note**: If your cluster enforces Pod Security Admission, you may need to label `metallb-system` with privileged labels:
> 
> ```yaml
>   labels:
>     pod-security.kubernetes.io/enforce: privileged
>     pod-security.kubernetes.io/audit: privileged
>     pod-security.kubernetes.io/warn: privileged
> ```
> 
> This is because the speaker Pod requires elevated permissions to manipulate network interfaces.

### 2.3. (Optional) Customize Values

If you want to override defaults (e.g. FRR mode, enabling debug logs, etc.), create a `values.yaml`:

```yaml
speaker:
  logLevel: debug
  frr:
    enabled: true

controller:
  logLevel: info
```

Then re‐install (or upgrade):
```bash
helm install metallb metallb/metallb -n metallb-system -f values.yaml
```

---

## 3. Provide MetalLB a **ConfigMap** to Assign IPs

MetalLB does **not** come with any default IP pools. You must define them. For **Layer2** mode on a typical “bare‐metal” or VM network:

```yaml
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: my-ip-pool
  namespace: metallb-system
spec:
  addresses:
  - 192.168.68.230-192.168.68.239   # Example, adapt to your LAN or VM network
---
apiVersion: metallb.io/v1beta1
kind: L2Advertisement
metadata:
  name: l2-adv
  namespace: metallb-system
spec: {}
```

Apply these:
```bash
kubectl apply -f metallb-config.yaml
```
Hence, any **LoadBalancer** Service in your cluster can now request a load‐balancer IP from the range `192.168.68.230-239`.

---

## 4. Confirm MetalLB is Working

1. **Check the Pods**:
   ```bash
   kubectl get pods -n metallb-system -o wide
   ```
   - You should see 1 **controller** Pod (running) and N **speaker** Pods (one on each node).

2. **Create a Test LoadBalancer Service**  
   For instance:
   ```yaml
   apiVersion: v1
   kind: Service
   metadata:
     name: my-nginx
     namespace: default
   spec:
     type: LoadBalancer
     ports:
       - port: 80
         targetPort: 80
     selector:
       app: my-nginx
   ```
   Once deployed, run:
   ```bash
   kubectl get svc my-nginx
   ```
   The `EXTERNAL-IP` column should show an IP from `192.168.68.230-239` (or whichever range you set). If it does, you’re good!

---

## 5. In a Nutshell

1. **Disable** microk8s’s built‐in `metallb` add‐on.  
2. **(If IPVS)**: enable strict ARP.  
3. **Helm Install** official chart → `helm install metallb metallb/metallb`.  
4. **Define** an IPAddressPool + L2Advertisement (or BGP config) so MetalLB can allocate addresses.  
5. **Enjoy** real LoadBalancer IPs on your bare‐metal MicroK8s cluster.

For more advanced usage (FRR mode, BGP sessions, etc.), consult the [MetalLB official docs](https://metallb.universe.tf). 

---

### Additional Notes

- If you prefer to install via **manifests** (YAML) or Kustomize, see [MetalLB’s official installation docs](https://metallb.universe.tf/installation/).  
- If you want to use **FRR** or **FRR-K8s** modes, set the appropriate Helm values or apply the relevant manifest (`metallb-frr.yaml`, `metallb-frr-k8s.yaml`) per the docs.  
- Always keep an eye on [release notes](https://github.com/metallb/metallb/releases) if upgrading to new major/minor versions.

That’s all – you have a “pure upstream” MetalLB in your MicroK8s cluster!