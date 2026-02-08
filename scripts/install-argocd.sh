#!/usr/bin/env bash
set -euo pipefail

echo "==> Installing ArgoCD"

# Create namespace
kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -

# Install ArgoCD (server-side apply needed for large CRDs)
kubectl apply -n argocd --server-side --force-conflicts -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

echo "==> Waiting for ArgoCD server to be ready..."
kubectl wait --namespace argocd \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/name=argocd-server \
  --timeout=180s

# Disable TLS on ArgoCD server for local access (Kind)
kubectl patch deployment argocd-server -n argocd \
  --type='json' \
  -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--insecure"}]' \
  2>/dev/null || true

echo "==> ArgoCD installed successfully!"
echo ""
echo "To get the admin password:"
echo "  kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d; echo"
echo ""
echo "To access the ArgoCD UI:"
echo "  kubectl port-forward svc/argocd-server -n argocd 9090:443"
echo "  Open https://localhost:9090 (username: admin)"
