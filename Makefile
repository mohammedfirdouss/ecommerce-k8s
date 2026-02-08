.PHONY: setup teardown build deploy deploy-infra deploy-services deploy-monitoring undeploy test clean

CLUSTER_NAME := ecommerce
NAMESPACE := ecommerce

setup:
	./scripts/setup-cluster.sh

teardown:
	kind delete cluster --name $(CLUSTER_NAME)

build:
	./scripts/build-images.sh

deploy: deploy-infra deploy-services deploy-monitoring

deploy-infra:
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/secret.yaml
	kubectl apply -f k8s/postgres/
	kubectl apply -f k8s/rabbitmq/
	@echo "Waiting for PostgreSQL and RabbitMQ to be ready..."
	kubectl wait --namespace $(NAMESPACE) --for=condition=ready pod -l app=postgres --timeout=120s
	kubectl wait --namespace $(NAMESPACE) --for=condition=ready pod -l app=rabbitmq --timeout=120s

deploy-services:
	kubectl apply -f k8s/services/
	kubectl apply -f k8s/ingress.yaml

deploy-monitoring:
	kubectl apply -f k8s/monitoring/

undeploy:
	kubectl delete namespace $(NAMESPACE) --ignore-not-found

test:
	./scripts/test-e2e.sh

clean: undeploy teardown
