#!/usr/bin/env bash
set -euo pipefail

SERVICES=("auth" "product" "cart" "order" "payment")
CLUSTER_NAME="ecommerce"

echo "==> Building Docker images for all services"

for svc in "${SERVICES[@]}"; do
    echo "--- Building ecommerce/${svc}-service:latest"
    docker build -t "ecommerce/${svc}-service:latest" "./services/${svc}"
done

echo "--- Building ecommerce/ui:latest"
docker build -t "ecommerce/ui:latest" "./ui"

echo "==> Loading images into Kind cluster"
for svc in "${SERVICES[@]}"; do
    echo "--- Loading ecommerce/${svc}-service:latest"
    kind load docker-image "ecommerce/${svc}-service:latest" --name "${CLUSTER_NAME}"
done

echo "--- Loading ecommerce/ui:latest"
kind load docker-image "ecommerce/ui:latest" --name "${CLUSTER_NAME}"

echo "==> All images built and loaded!"
