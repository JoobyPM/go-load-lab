# Part 2: Installing & Exposing Traefik with MetalLB (DNS Challenge Edition)

Below is a guide to:

1. Install “vanilla” Traefik from the official Helm chart,  
2. Use MetalLB for a LoadBalancer IP,  
3. Obtain Let’s Encrypt certificates via a **Route 53** DNS challenge,  
4. Store ACME data on **Longhorn** (RWX volume),  
5. Avoid “bind: permission denied” issues by listening on **high ports (8000/8443)** internally but exposing **80/443** on the Service.


## 1. Prerequisites

1. **MetalLB is installed** (with a namespace `metallb-system` and at least one IP pool, e.g. `192.168.68.230-192.168.68.239`).  
2. **Helm ≥3.9** and **kubectl** are installed on your workstation (or your CI), and your kube‐config points to the MicroK8s cluster.  
3. If you used the built‐in MicroK8s Traefik add‐on (`microk8s enable ingress`), **disable** it:
   ```bash
   microk8s disable ingress
   ```
   We want the “vanilla” Traefik from the **official Helm chart**, not the MicroK8s add-on.

4. In **Route 53**, create (or confirm) an **A** record pointing `[my.domain.com]` and/or `*.my.domain.com` to the **external IP** assigned by MetalLB (e.g. `192.168.68.230`).  
   - If your local IP is not publicly routable, you’ll at least need internal DNS resolution.  


## 2. Add the Traefik Helm Repo & Update

```bash
helm repo add traefik https://traefik.github.io/charts
helm repo update
```


## 3. Prepare IAM Credentials (For Route 53 DNS Challenge)

Traefik needs AWS credentials to create the `_acme-challenge.my.domain.com` TXT records in Route 53. Typically, you store these in a Kubernetes **Secret**:

```bash
kubectl create secret generic route53-credentials-secret \
  --namespace=traefik \
  --from-literal=aws_access_key_id='YOUR_ACCESS_KEY_ID' \
  --from-literal=aws_secret_access_key='YOUR_SECRET_ACCESS_KEY'
```

*(Replace `YOUR_ACCESS_KEY_ID` / `YOUR_SECRET_ACCESS_KEY` with your actual key/secret.)*

Your IAM user/policy should allow:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    { "Effect": "Allow", "Action": "route53:GetChange", "Resource": "arn:aws:route53:::change/*" },
    { "Effect": "Allow", "Action": "route53:ListHostedZonesByName", "Resource": "*" },
    {
      "Effect": "Allow",
      "Action": [ "route53:ListResourceRecordSets" ],
      "Resource": [ "arn:aws:route53:::hostedzone/XYZ" ]
    },
    {
      "Effect": "Allow",
      "Action": [ "route53:ChangeResourceRecordSets" ],
      "Resource": [ "arn:aws:route53:::hostedzone/XYZ" ],
      "Condition": {
        "ForAllValues:StringEquals": {
          "route53:ChangeResourceRecordSetsNormalizedRecordNames": [
            "_acme-challenge.my.domain.com",
            "_acme-challenge.*.my.domain.com"
          ],
          "route53:ChangeResourceRecordSetsRecordTypes": [ "TXT" ]
        }
      }
    }
  ]
}
```

*(HostedZone ARNs may differ. Adjust domain names as needed.)*


## 4. Create Your `values.yaml` for Traefik

Below are the **working values** that you’ve validated, featuring:

- High ports (8000/8443) inside the container → Expose 80/443 on the Service
- Let’s Encrypt DNS challenge via Route 53
- Longhorn RWX for ACME storage
- Non‐root container with an initContainer to fix `acme.json` permissions

```yaml
#################################################################
# Minimal Traefik Helm values with key changes for your setup
#################################################################

# ------------------------------------------------------------------------------
# 1) Image
# ------------------------------------------------------------------------------
image:
  registry: docker.io
  repository: traefik
  # tag left blank => chart will use its default (appVersion)
  pullPolicy: IfNotPresent

# ------------------------------------------------------------------------------
# 2) Deployment (or DaemonSet)
#    - Set to "Deployment" and 1 replica by default.
#    - Add an initContainer to fix acme.json permissions.
# ------------------------------------------------------------------------------
deployment:
  enabled: true
  kind: Deployment
  replicas: 1

  # Fix acme.json perms across restarts
  initContainers:
    - name: volume-permissions
      image: busybox:latest
      command: ["sh", "-c", "touch /data/acme.json && chmod 600 /data/acme.json"]
      volumeMounts:
        - name: data
          mountPath: /data

# ------------------------------------------------------------------------------
# 3) Ports
#    - If your environment disallows binding low ports as non‐root,
#      we listen on 8000/8443 in the container, but the LB exposes 80/443 externally.
# ------------------------------------------------------------------------------
ports:
  web:
    port: 8000
    expose:
      default: true
    exposedPort: 80
    protocol: TCP

  websecure:
    port: 8443
    expose:
      default: true
    exposedPort: 443
    protocol: TCP
    tls:
      enabled: true

# ------------------------------------------------------------------------------
# 4) Service
#    - Use LoadBalancer type so MetalLB can assign an IP
# ------------------------------------------------------------------------------
service:
  enabled: true
  single: true
  type: LoadBalancer
  annotations: {}
  labels: {}

# ------------------------------------------------------------------------------
# 5) Persistence
#    - Enable a PVC to store acme.json so certs persist across restarts
#    - Use Longhorn’s default (RWX) class
# ------------------------------------------------------------------------------
persistence:
  enabled: true
  name: data
  accessMode: ReadWriteMany
  size: 1Gi
  storageClass: longhorn
  subPath: ""
  annotations: {}

# ------------------------------------------------------------------------------
# 6) Certificates Resolvers (Let’s Encrypt via Route 53 DNS Challenge)
# ------------------------------------------------------------------------------
certificatesResolvers:
  route53resolver:
    acme:
      email: "anon@example.com"  # Anonymize your real email
      storage: "/data/acme.json"
      dnsChallenge:
        provider: route53
        delayBeforeCheck: 30  # Or 'propagation.delayBeforeChecks' for newer Traefik

# ------------------------------------------------------------------------------
# 7) AWS Credentials (from a Kubernetes Secret) for Route 53
# ------------------------------------------------------------------------------
env:
  - name: AWS_ACCESS_KEY_ID
    valueFrom:
      secretKeyRef:
        name: route53-credentials-secret
        key: aws_access_key_id

  - name: AWS_SECRET_ACCESS_KEY
    valueFrom:
      secretKeyRef:
        name: route53-credentials-secret
        key: aws_secret_access_key

  - name: AWS_REGION
    value: "us-east-1"  # or your region

# ------------------------------------------------------------------------------
# 8) Providers
#    - We generally keep Kubernetes Ingress + CRD providers
# ------------------------------------------------------------------------------
providers:
  kubernetesCRD:
    enabled: true
    allowCrossNamespace: false
    allowExternalNameServices: false
    allowEmptyServices: true

  kubernetesIngress:
    enabled: true
    allowExternalNameServices: false
    allowEmptyServices: true
    publishedService:
      enabled: true

# ------------------------------------------------------------------------------
# 9) (Optional) IngressClass
# ------------------------------------------------------------------------------
ingressClass:
  enabled: true
  isDefaultClass: true

# ------------------------------------------------------------------------------
# 10) Misc/Defaults
# ------------------------------------------------------------------------------
rbac:
  enabled: true

hostNetwork: false

podSecurityContext:
  runAsGroup: 65532
  runAsNonRoot: true
  runAsUser: 65532

securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop: [ "ALL" ]
    add:
      - NET_BIND_SERVICE
  readOnlyRootFilesystem: true
```

*(Adjust the domain/email as needed. Note that specifying `NET_BIND_SERVICE` is optional here if you’re not binding `<1024` inside the container.)*


## 5. Install (or Upgrade) Traefik With These Values

```bash
kubectl create namespace traefik

helm install traefik traefik/traefik \
  --namespace=traefik \
  -f traefik.values.yaml
```

- Traefik will spin up a **LoadBalancer Service** named `traefik`.
- MetalLB assigns an IP from your pool (e.g., `192.168.68.230`).
- Traefik listens on container ports 8000 (HTTP) and 8443 (HTTPS), but the Service exposes 80/443 externally.

### 5.1. Verify the LoadBalancer IP

```bash
kubectl get svc -n traefik
```
You should see something like:
```
NAME       TYPE           CLUSTER-IP     EXTERNAL-IP       PORT(S)                       AGE
traefik    LoadBalancer   10.152.183.3   192.168.68.230    80:32122/TCP,443:32132/TCP    2m
```

### 5.2. Confirm DNS

Point `my.domain.com` (and `*.my.domain.com`, if desired) to that IP. Test with `dig` or `nslookup`:

```bash
dig my.domain.com
```
Should show `192.168.68.230`.


## 6. Let Traefik Issue Certificates

Create an Ingress referencing the `route53resolver`. For example:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: whoami-ingress
  annotations:
    traefik.ingress.kubernetes.io/router.tls: "true"
    traefik.ingress.kubernetes.io/router.tls.certresolver: "route53resolver"
spec:
  rules:
    - host: my.domain.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: whoami
                port:
                  number: 80
  tls:
    - hosts:
      - my.domain.com
```

Then visit `https://my.domain.com`. Traefik does a DNS challenge with Route 53, obtains a Let’s Encrypt cert, and serves it via the ACME storage in `/data/acme.json`.

> **Wildcard**  
> If you want `*.my.domain.com`, you can similarly use DNS challenge. Just ensure your host rules are set appropriately (`app.my.domain.com`) or a real wildcard route.


## 7. Summary

- Installed **Traefik** from the official Helm chart (with the `traefik.values.yaml` you see above).  
- **MetalLB** gave a LoadBalancer IP (on ports 80/443).  
- **Route 53** DNS challenge provided Let’s Encrypt certificates.  
- **Longhorn** RWX volume persisted the ACME data across restarts.  

Now you can route multiple services behind Traefik using Ingress or IngressRoute objects, each with their own domain or subdomain. If you prefer direct binding on ports 80/443 inside the container and run as root, you can remove the high‐port workaround—but the approach above is more secure and fully functional.

**Happy HTTPS load balancing with Traefik, MetalLB, and Route 53!**