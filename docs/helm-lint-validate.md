# Linting & Validating the Helm Chart

This document briefly explains how to **lint** and **validate** the Helm chart in this repository.

## 1. Helm Lint

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

## 2. Render & Validate Manifests

You can also render the chart templates and validate them with external tools:

```bash
helm template . --namespace my-namespace > rendered.yaml
```

Then pass `rendered.yaml` to a **Kubernetes schema validator** such as [kubeconform](https://github.com/yannh/kubeconform):

```bash
kubeconform -strict rendered.yaml
```

### **Note**: Bonus Tips
- For additional checks, you can also **render** your templates and pipe them through Kubernetes validators:
  ```bash
  helm template /path/to/chart | kubeconform -strict
  ```

This checks whether the generated manifests match official Kubernetes API schemas.

### Why Validate Manifests?

- **Catches** advanced schema issues (e.g. outdated `apiVersion`).
- **Ensures** your chart’s resources conform to the cluster’s Kubernetes version.

## 3. CI/CD Integration

- In a CI environment (e.g. GitHub Actions, GitLab CI), you can run these commands automatically:
  - **Lint** the chart via `helm lint`.
  - **Render** via `helm template` and pipe to a schema validator.
- Tools like [chart-testing](https://github.com/helm/chart-testing) can automate linting, validation, and even install tests.

---

**Summary**:  
1. **`helm lint .`** → quick check for basic mistakes.  
2. **`helm template . | kubeconform`** → thorough validation against schema.  
3. **CI/CD** → automate repeated checks to maintain chart quality.