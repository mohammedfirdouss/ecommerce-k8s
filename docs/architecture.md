# E-Commerce Application Architecture

## Overview

This document describes the architecture of the e-commerce microservices application deployed on Kubernetes.

## System Architecture

```
                         ┌──────────────┐
                         │   NGINX      │
                  :80 ── │   Ingress    │
                         │  Controller  │
                         └──────┬───────┘
                                │
           ┌────────────┬───────┼────────┬──────────────┐
           │            │       │        │              │
     ┌─────▼────┐ ┌─────▼────┐ ┌▼──────┐ ┌▼──────────┐ ┌▼─────────┐
     │  Auth    │ │ Product  │ │ Cart  │ │  Order    │ │ Payment  │
     │ :8081   │ │ :8082   │ │ :8083 │ │  :8084   │ │ :8085   │
     └────┬─────┘ └────┬─────┘ └──┬───┘ └──┬──┬────┘ └──┬──┬───┘
          │            │          │         │  │         │  │
          │            │          │         │  │ publish  │  │ consume
          │            │          │         │  └──┐   ┌──┘  │
          │            │          │         │     ▼   ▼     │
          │            │          │         │  ┌─────────┐  │
          │            │          │    consume │ RabbitMQ │ publish
          │            │          │    ◄───────│  :5672   │──────►
          │            │          │            └─────────┘
          │            │          │
     ┌────▼────────────▼──────────▼───┐
     │         PostgreSQL :5432       │
     │                                │
     │  Schemas:                      │
     │  ├── auth (users)              │
     │  ├── product (products)        │
     │  ├── cart (carts, cart_items)   │
     │  ├── orders (orders, items)    │
     │  └── payment (payments)        │
     └────────────────────────────────┘

     ┌────────────────────────────────┐
     │  Prometheus :9090 → Grafana    │
     │  (scrapes /metrics from all    │
     │   services via annotations)    │
     └────────────────────────────────┘
```

## Microservices

### Auth Service (port 8081)
- **Responsibility**: User registration, login, JWT token generation
- **Database Schema**: `auth` (users table)
- **Key Features**:
  - Password hashing with bcrypt
  - JWT token generation (24h expiry)
  - Token validation endpoint for other services

### Product Service (port 8082)
- **Responsibility**: Product catalog management
- **Database Schema**: `product` (products table)
- **Key Features**:
  - Full CRUD operations
  - Stock tracking

### Cart Service (port 8083)
- **Responsibility**: Shopping cart management
- **Database Schema**: `cart` (carts, cart_items tables)
- **Key Features**:
  - Per-user carts (identified via X-User-ID header)
  - Automatic cart creation
  - Item quantity aggregation

### Order Service (port 8084)
- **Responsibility**: Order creation and lifecycle management
- **Database Schema**: `orders` (orders, order_items tables)
- **RabbitMQ**:
  - **Publishes**: `order.created` events to `orders` exchange
  - **Consumes**: `payment.status` events from `payments` exchange
- **Key Features**:
  - Transaction-based order creation
  - Automatic status updates from payment events

### Payment Service (port 8085)
- **Responsibility**: Payment processing (mock)
- **Database Schema**: `payment` (payments table)
- **RabbitMQ**:
  - **Consumes**: `order.created` events from `orders` exchange
  - **Publishes**: `payment.status` events to `payments` exchange
- **Key Features**:
  - Configurable success rate (default 80%)
  - 2-second processing delay to simulate real payments

## Communication Patterns

### Synchronous (REST)
All services expose REST APIs accessed via the NGINX Ingress Controller. The ingress rewrites paths:
- `/api/auth/*` → auth-service:8081
- `/api/products/*` → product-service:8082
- `/api/cart/*` → cart-service:8083
- `/api/orders/*` → order-service:8084
- `/api/payments/*` → payment-service:8085

### Asynchronous (RabbitMQ)
```
Order Service ──(order.created)──► orders exchange ──► Payment Service
Payment Service ──(payment.status)──► payments exchange ──► Order Service
```

This decouples order creation from payment processing, allowing:
- Non-blocking order creation
- Independent scaling of payment processing
- Resilience to payment service downtime

## Database Design

Single PostgreSQL instance with per-service schemas provides:
- **Isolation**: Each service owns its schema
- **Simplicity**: Single database to manage for development
- **Migration path**: Can split into separate instances if needed

## Kubernetes Resources

### Deployments
- All microservices: 2 replicas each
- PostgreSQL: 1 replica with PVC
- RabbitMQ: 1 replica
- Prometheus: 1 replica
- Grafana: 1 replica

### Services
- All use ClusterIP (internal-only)
- External access via NGINX Ingress

### Health Checks
Every service exposes:
- `GET /healthz` — Liveness and readiness probe (checks DB connectivity)
- `GET /metrics` — Prometheus metrics endpoint

## Monitoring

### Prometheus
- Auto-discovers pods via Kubernetes annotations
- Scrapes `/metrics` endpoints every 15 seconds
- Stores 7 days of metrics data

### Grafana
- Pre-configured with Prometheus data source
- Default credentials: admin/admin
- Access via `kubectl port-forward`

## Security Considerations

- JWT tokens for authentication (24h expiry)
- JWT secret stored as Kubernetes Secret
- Services communicate internally via ClusterIP (not exposed externally)
- Database credentials should be replaced with Kubernetes Secrets in production
- RBAC configured for Prometheus service account
