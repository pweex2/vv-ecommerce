# VV-Ecommerce

A lightweight, microservices-based e-commerce backend system built with **Go** and the **Gin** framework.

## 🐳 Quick Start (Docker) - Recommended

The easiest way to run the project is using Docker Compose. This ensures all services (MySQL, Redis, RabbitMQ, and Microservices) are wired up correctly and handles cross-platform compatibility (Windows/Mac/Linux).

### 1. Prerequisites
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) installed

### 2. Configuration
Copy the example environment file to create your local configuration:

```bash
# Mac/Linux
cp .env.example .env

# Windows (PowerShell)
copy .env.example .env
```

(Optional) Edit `.env` if you need to change ports or credentials. The defaults are usually fine.

### 3. Run
Start the entire system with one command:

```bash
docker-compose up --build
```

### 4. Access
| Service | URL / Port | Description |
| :--- | :--- | :--- |
| **API Gateway** | `http://localhost:8000` | **Main Entry Point**. All API requests go here. |
| **MySQL** | `localhost:3306` | Database (User: `root`, Pass: `root`) |
| **RabbitMQ UI** | `http://localhost:15672` | Message Queue Dashboard (User: `guest`, Pass: `guest`) |
| **Redis** | `localhost:6379` | Cache |

---

## 🏗 Architecture

This project adopts a **Monorepo** structure managed by Go Workspaces (`go.work`). It consists of three decoupled microservices that communicate via HTTP and an API Gateway.

### Tech Stack
- **Language**: Go 1.25+
- **Web Framework**: [Gin](https://github.com/gin-gonic/gin)
- **Database**: MySQL (accessed via [GORM](https://gorm.io/))
- **Messaging**: RabbitMQ (Event-Driven Architecture)
- **Caching**: Redis
- **Gateway**: Custom Go-based Reverse Proxy

## 🌐 Configuration & Networking

This project handles different environments using **Environment Variables** (`.env`).

| Configuration | Local (Docker) | Production (K8s / Cloud) |
| :--- | :--- | :--- |
| **Database Host** | `mysql` (Docker Service Name) | `db-prod.cluster-xyz.aws.com` |
| **Service Discovery** | `http://inventory-service:8082` | `http://inventory-service` (K8s DNS) |
| **API Gateway** | `localhost:8000` | `api.vv-ecommerce.com` |

## 🔄 Distributed Transaction & Consistency

This project implements the **Saga Pattern (Orchestration-based)** and **Transactional Outbox Pattern** to ensure data consistency across microservices.

### The "Order Creation" Saga
1. **Order Service**: Creates an order in `PENDING` state and writes an event to the `outbox_events` table (in the same DB transaction).
2. **Outbox Processor**: Asynchronously reads from `outbox_events` and publishes messages to **RabbitMQ**.
3. **Inventory Service**: Consumes message, deducts stock.
4. **Payment Service**: Consumes message, processes payment.
5. **Compensation**: If any step fails, compensating events are triggered to rollback changes (e.g., restore stock).

## 🛡️ Standardized Error Handling

- **AppError**: A unified error struct used across all services.
- **Retry Logic**: Smart retry mechanisms for transient errors (e.g., timeouts) vs. permanent errors (e.g., invalid input).

## 🚀 Services Overview

| Service | Internal Port | Description |
|---------|---------------|-------------|
| **API Gateway** | `:8000` | Routes requests to internal services. **Publicly Exposed**. |
| **Order Service** | `:8081` | Manages orders. Orchestrates Sagas. |
| **Inventory Service** | `:8082` | Manages stock levels. |
| **Payment Service** | `:8083` | Handles payments. |

---

## 🛠 Manual Local Development (Optional)

If you prefer to run services manually (without Docker Compose for apps), ensure you have the infrastructure running:

### 1. Start Infrastructure
```bash
# Only start infra (MySQL, Redis, MQ)
docker-compose up -d mysql redis rabbitmq
```

### 2. Run Services
You need to start each service in a separate terminal.

#### Start API Gateway
```bash
cd services/api-gateway
go run ./cmd/api-gateway/main.go
```

#### Start Order Service
```bash
cd services/order-service
go run ./cmd/order-service/main.go
```

(Repeat for Inventory and Payment services)

## 🔌 API Endpoints (via Gateway)

Base URL: `http://localhost:8000`

- `POST /api/v1/orders` - Create a new order
- `GET /api/v1/orders/:id` - Get order details
- `POST /api/v1/inventory/deduct` - Deduct stock (Internal/Debug)

---

## 🗺️ Deployment & Production Roadmap

### Phase 1: Containerization (Current)
- [x] Dockerize all services (Multi-stage builds)
- [x] Docker Compose for local development orchestration
- [x] Environment variable configuration (.env)

### Phase 2: Observability & Monitoring
- [ ] **Distributed Tracing**: Integrate Jaeger/OpenTelemetry to visualize TraceIDs across services.
- [ ] **Metrics**: Expose Prometheus metrics (`/metrics`) for request latency, error rates, and queue depth.
- [ ] **Logging**: Centralized logging (ELK Stack or Loki) to aggregate logs from all containers.

### Phase 3: CI/CD & Automation
- [ ] **CI Pipeline**: GitHub Actions to run tests and linting on PRs.
- [ ] **Image Publishing**: Auto-build and push Docker images to Registry (Docker Hub/ECR) on merge.

### Phase 4: Kubernetes (K8s) Migration
- [ ] Create Helm Charts or K8s Manifests (Deployment, Service, Ingress).
- [ ] Implement **Liveness & Readiness Probes** for zero-downtime deployments.
- [ ] **Secrets Management**: Move sensitive `.env` data to K8s Secrets or HashiCorp Vault.

### Phase 5: Security & Resilience
- [ ] **API Gateway Auth**: Implement JWT validation at the Gateway level.
- [ ] **Rate Limiting**: Protect services using Redis-based rate limiting in the Gateway.
- [ ] **Circuit Breaking**: Enhance clients with Hystrix/Resilience4j patterns.
