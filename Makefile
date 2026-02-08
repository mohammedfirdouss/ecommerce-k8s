.PHONY: setup teardown build deploy deploy-infra deploy-services deploy-monitoring undeploy test clean gitops-install gitops-deploy gitops-status

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
	kubectl apply -f k8s/ui.yaml
	kubectl apply -f k8s/ingress.yaml

deploy-monitoring:
	kubectl apply -f k8s/monitoring/

undeploy:
	kubectl delete namespace $(NAMESPACE) --ignore-not-found

test:
	./scripts/test-e2e.sh

clean: undeploy teardown

# ---- GitOps with ArgoCD ----

gitops-install:
	./scripts/install-argocd.sh

gitops-deploy:
	kubectl apply -f k8s/argocd/project.yaml
	kubectl apply -f k8s/argocd/app-of-apps.yaml
	@echo "ArgoCD will now sync all applications from Git."
	@echo "Monitor: kubectl port-forward svc/argocd-server -n argocd 9090:443"

gitops-status:
	@echo "==> ArgoCD Applications:"
	kubectl get applications -n argocd
	@echo ""
	@echo "==> Ecommerce Pods:"
	kubectl get pods -n $(NAMESPACE)
