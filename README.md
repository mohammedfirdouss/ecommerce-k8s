# E-Commerce Microservices on Kubernetes

A microservices-based e-commerce application built with **Go**, deployed on **Kubernetes** using **Kind**, with **PostgreSQL**, **RabbitMQ**, **Prometheus**, and **Grafana**.

## Architecture

The application consists of 5 microservices:

| Service | Port | Description |
|---------|------|-------------|
| **auth-service** | 8081 | User registration, login, JWT authentication |
| **product-service** | 8082 | Product catalog CRUD operations |
| **cart-service** | 8083 | Shopping cart management |
| **order-service** | 8084 | Order creation and tracking |
| **payment-service** | 8085 | Mock payment processing |

### Communication
- **Synchronous**: REST (HTTP/JSON) via Kubernetes ClusterIP services
- **Asynchronous**: RabbitMQ for order→payment events and payment→order status updates
- **Service Discovery**: Kubernetes DNS-based service discovery
- **External Access**: NGINX Ingress Controller

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Go 1.22+](https://golang.org/dl/) (for local development)

## Quick Start

```bash
# 1. Create Kind cluster with NGINX Ingress
make setup

# 2. Build Docker images and load into Kind
make build

# 3. Deploy everything (infra + services + monitoring)
make deploy

# 4. Run end-to-end tests
make test

# 5. Tear down
make clean
```

## Accessing the Application

After deployment, use port-forwarding to access the services:

### Web UI (Recommended)
```bash
# Start port-forward for the web UI
kubectl port-forward -n ecommerce svc/ui 8080:80
```
Then open: **http://localhost:8080**

### Full Application (UI + APIs via Ingress)
```bash
# Start port-forward for the Ingress controller
kubectl port-forward -n ingress-nginx svc/ingress-nginx-controller 8080:80
```
Then access:
- **Web UI**: http://localhost:8080/ui/
- **APIs**: http://localhost:8080/api/{auth,products,cart,orders,payments}/

### Monitoring
```bash
# Prometheus (metrics)
kubectl port-forward -n ecommerce svc/prometheus 9090:9090

# Grafana (dashboards) - login: admin/admin
kubectl port-forward -n ecommerce svc/grafana 3000:3000
```

### GitHub Codespaces / Remote Environments
If running in GitHub Codespaces or a remote environment:
1. Run the port-forward command
2. Go to the **PORTS** tab in VS Code
3. Find port `8080` and click the globe icon to open in browser

### Quick Access Commands
```bash
# Check all pods are running
kubectl get pods -n ecommerce

# View service endpoints
kubectl get svc -n ecommerce

# View ingress configuration
kubectl get ingress -n ecommerce

# Tail logs from a service
kubectl logs -f -n ecommerce deployment/auth-service
```

## API Endpoints

All endpoints are accessible via the Ingress at `http://localhost`.

### Auth Service (`/api/auth`)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth/register` | Register a new user |
| POST | `/api/auth/login` | Login and get JWT token |
| GET | `/api/auth/validate` | Validate JWT token (requires auth) |
| GET | `/api/auth/healthz` | Health check |

### Product Service (`/api/products`)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/products/` | List all products |
| GET | `/api/products/{id}` | Get product by ID |
| POST | `/api/products/` | Create a product |
| PUT | `/api/products/{id}` | Update a product |
| DELETE | `/api/products/{id}` | Delete a product |
| GET | `/api/products/healthz` | Health check |

### Cart Service (`/api/cart`)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/cart/` | Get user's cart (X-User-ID header) |
| POST | `/api/cart/items` | Add item to cart |
| DELETE | `/api/cart/items/{id}` | Remove item from cart |
| DELETE | `/api/cart/` | Clear cart |
| GET | `/api/cart/healthz` | Health check |

### Order Service (`/api/orders`)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/orders/` | Create order (X-User-ID header) |
| GET | `/api/orders/` | List user's orders |
| GET | `/api/orders/{id}` | Get order details |
| GET | `/api/orders/healthz` | Health check |

### Payment Service (`/api/payments`)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/payments/{orderID}` | Get payment status |
| GET | `/api/payments/healthz` | Health check |

## Monitoring

- **Prometheus**: Scrapes `/metrics` from all services (port 9090)
- **Grafana**: Dashboard UI (port 3000, admin/admin)

Access via port-forward:
```bash
kubectl port-forward -n ecommerce svc/prometheus 9090:9090 &
kubectl port-forward -n ecommerce svc/grafana 3000:3000 &
```

## Development

### Local Development
Each service can be run locally:
```bash
cd services/auth
export DB_HOST=localhost DB_PORT=5432 DB_USER=ecommerce DB_PASSWORD=ecommerce_pass DB_NAME=ecommerce
go run .
```

### Adding a New Service
1. Create service directory under `services/`
2. Implement Go service with `/healthz` and `/metrics` endpoints
3. Create Dockerfile
4. Add K8s deployment and service manifests under `k8s/services/`
5. Add Ingress path rule in `k8s/ingress.yaml`
6. Add to `scripts/build-images.sh`

## Tech Stack

- **Language**: Go 1.22 (chi router, sqlx, JWT)
- **Database**: PostgreSQL 16 (single instance, per-service schemas)
- **Message Broker**: RabbitMQ 3 (AMQP for async communication)
- **Container Runtime**: Docker
- **Orchestration**: Kubernetes (Kind)
- **Ingress**: NGINX Ingress Controller
- **Monitoring**: Prometheus + Grafana