#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="ecommerce"

echo "==> Applying namespace and secrets"
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/secret.yaml

echo "==> Deploying PostgreSQL"
kubectl apply -f k8s/postgres/
echo "Waiting for PostgreSQL..."
kubectl wait --namespace "${NAMESPACE}" --for=condition=ready pod -l app=postgres --timeout=120s

echo "==> Deploying RabbitMQ"
kubectl apply -f k8s/rabbitmq/
echo "Waiting for RabbitMQ..."
kubectl wait --namespace "${NAMESPACE}" --for=condition=ready pod -l app=rabbitmq --timeout=120s

echo "==> Deploying microservices"
kubectl apply -f k8s/services/

echo "==> Deploying UI"
kubectl apply -f k8s/ui.yaml

echo "==> Deploying Ingress"
kubectl apply -f k8s/ingress.yaml

echo "==> Deploying monitoring"
kubectl apply -f k8s/monitoring/

echo "==> Waiting for all services to be ready..."
for svc in auth-service product-service cart-service order-service payment-service; do
    echo "Waiting for ${svc}..."
    kubectl wait --namespace "${NAMESPACE}" --for=condition=ready pod -l "app=${svc}" --timeout=120s || true
done

echo ""
echo "==> Deployment complete!"
echo "Services:"
kubectl get pods -n "${NAMESPACE}"
echo ""
echo "Access the API via: http://localhost/api/{auth,products,cart,orders,payments}"
