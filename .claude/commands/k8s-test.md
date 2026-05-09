# /k8s-test — Build, push, and deploy Mortar to the test cluster

Deploys the current working tree to the `mortar-test` namespace on the homelab
Kubernetes cluster. Accepts an optional image tag; defaults to the current short
git SHA.

**Usage:** `/k8s-test [tag]`

---

## Steps

Set the kubeconfig shorthand at the top of every `kubectl` call:
```
KUBECONFIG=~/src/homelab/kubeconfig.yml
```

### 1. Resolve the image tag

```bash
TAG=${1:-$(git rev-parse --short HEAD)}
IMAGE="ghcr.io/nbellowe/mortar:$TAG"
echo "Deploying $IMAGE"
```

### 2. Build and push the Docker image

```bash
docker build --platform linux/amd64 -t "$IMAGE" .
docker push "$IMAGE"
```

### 3. Check that secret.yaml exists

```bash
if [ ! -f k8s/test/secret.yaml ]; then
  echo "ERROR: k8s/test/secret.yaml not found."
  echo "Copy k8s/test/secret.yaml.example → k8s/test/secret.yaml and fill in API keys."
  exit 1
fi
```

### 4. Apply manifests

Apply resources individually so that the secret (which is gitignored and not in
`kustomization.yaml` tracking) is included:

```bash
KUBECONFIG=~/src/homelab/kubeconfig.yml kubectl apply -f k8s/test/namespace.yaml
KUBECONFIG=~/src/homelab/kubeconfig.yml kubectl apply -f k8s/test/secret.yaml
KUBECONFIG=~/src/homelab/kubeconfig.yml kubectl apply -f k8s/test/service.yaml
KUBECONFIG=~/src/homelab/kubeconfig.yml kubectl apply -f k8s/test/deployment.yaml
```

### 5. Update the image tag on the deployment

```bash
KUBECONFIG=~/src/homelab/kubeconfig.yml \
  kubectl set image deployment/mortar mortar="$IMAGE" -n mortar-test
```

### 6. Wait for rollout

```bash
KUBECONFIG=~/src/homelab/kubeconfig.yml \
  kubectl rollout status deployment/mortar -n mortar-test --timeout=120s
```

### 7. Port-forward and report

```bash
KUBECONFIG=~/src/homelab/kubeconfig.yml \
  kubectl port-forward svc/mortar 3000:3000 -n mortar-test
```

Report to the user: **Mortar test deploy ready at http://localhost:3000**

---

## Tear down

```bash
KUBECONFIG=~/src/homelab/kubeconfig.yml kubectl delete namespace mortar-test
```

---

## Troubleshooting

- **ImagePullBackOff**: the image tag wasn't pushed, or ghcr.io auth is stale.
  Run `docker login ghcr.io` and retry.
- **Pod pending**: check `kubectl describe pod -n mortar-test` for scheduling issues.
- **Plugin errors in logs**: verify API keys in `k8s/test/secret.yaml` match the
  current SOPS secrets (`sops -d ~/src/homelab/kubernetes/apps/media/media-secrets/secret.sops.yaml`).
- **Connection refused to upstream service**: cross-namespace DNS requires the full
  `service.media.svc.cluster.local` form — confirm URLs in `secret.yaml`.
