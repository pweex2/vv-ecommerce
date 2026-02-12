# VV-Ecommerce: High-Concurrency Microservices Backend

An advanced learning project for e-commerce backend system engineered with **Go**, **Gin**, and **Microservices Architecture**. 

This project explores industry-standard patterns for solving **Distributed Data Consistency**, **High Concurrency**, and **Full-Link Observability**.

## ✨ Key Features & Engineering Highlights

*   **Microservices Architecture**: Decoupled services (Order, Inventory, Payment, Gateway) communicating via RESTful APIs and asynchronous messaging.
*   **Distributed Transactions**: Implemented **Saga Pattern** and **Transactional Outbox Pattern** to ensure eventual consistency, solving the classic "Dual-Write" problem between Database and Message Queue.
*   **Observability**: End-to-end distributed tracing using **OpenTelemetry (OTel)** and **Jaeger**.
*   **High Reliability**:
    *   **Idempotent Consumers**: Prevents duplicate processing of messages.
    *   **Smart Retry**: Exponential backoff strategies for transient failures.
    *   **Concurrency Safety**: Optimized Outbox Processor using **MySQL `FOR UPDATE SKIP LOCKED`** to support horizontal scaling without race conditions.
*   **DevOps Ready**: Fully containerized with **Docker Compose** and automated CI/CD pipelines via **GitHub Actions**.

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
| **Frontend** | `http://localhost:3000` | **Web UI Dashboard**. User interface for managing orders and products. |
| **API Gateway** | `http://localhost:8000` | **Main Entry Point**. All API requests go here. |
| **MySQL** | `localhost:3306` | Database (User: `root`, Pass: `root`) |
| **RabbitMQ UI** | `http://localhost:15672` | Message Queue Dashboard (User: `guest`, Pass: `guest`) |
| **Redis** | `localhost:6379` | Cache |
| **Jaeger UI** | `http://localhost:16686` | Distributed Tracing Dashboard |

---

## 🏗 Architecture

This project adopts a **Monorepo** structure managed by Go Workspaces (`go.work`). It consists of three decoupled microservices that communicate via HTTP and an API Gateway.

### Tech Stack
- **Language**: Go 1.25+
- **Web Framework**: [Gin](https://github.com/gin-gonic/gin)
- **Database**: MySQL (accessed via [GORM](https://gorm.io/))
- **Messaging**: RabbitMQ (Event-Driven Architecture)
- **Caching**: Redis
- **Observability**: OpenTelemetry, Jaeger
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
| **Frontend** | `:3000` | React-based User Interface. |
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

## 💡 Architecture Decisions (The "Why")

*Common Interview Questions & Answers based on this architecture:*

### 1. Why Saga Pattern instead of 2PC (Two-Phase Commit)?
*   **Trade-off**: We prioritized **Availability** and **Performance** over Strong Consistency.
*   **Reasoning**: 2PC locks resources across services (Order, Inventory, Payment), causing high latency and "Blocking" if one service fails. Saga (Orchestration) allows local transactions to commit immediately, using **Compensating Transactions** (Rollbacks) to handle failures eventually.

### 2. Why RabbitMQ over Kafka?
*   **Trade-off**: We prioritized **Reliability** and **Complex Routing** over Extreme Throughput.
*   **Reasoning**:
    *   **Reliability**: RabbitMQ's Confirm/Ack mechanism ensures zero message loss (critical for Financial/Inventory data).
    *   **Routing**: We need features like **Dead Letter Queues (DLQ)** and **Delay Queues** (for Order Timeout cancellation), which RabbitMQ supports natively. Kafka is better for log streaming/analytics, not transactional messaging.

### 3. Why Optimistic Locking (`version` field) vs Pessimistic Locking (`SELECT FOR UPDATE`)?
*   **Trade-off**: We prioritized **Throughput** over Conflict Prevention.
*   **Reasoning**: In e-commerce, reading stock is frequent, but writing (buying) is less frequent. Pessimistic locking blocks all readers, killing performance. Optimistic locking allows high concurrency, failing only when the write actually conflicts (using `CAS` or `WHERE version = old_version`).
    *   *Note*: For the Outbox Processor, we DO use `FOR UPDATE SKIP LOCKED` (Pessimistic) because that is a specific "Job Queue" pattern where we strictly need to prevent multiple consumers from processing the same event.

---

## 🗺️ Deployment & Production Roadmap

### 📚 Learning Roadmap: From Junior to Senior

This section outlines the gap analysis and planned improvements to transform this project from a "functional demo" to a "production-grade financial system".

#### 1. Reliability Engineering (The "Dirty Work")
- [ ] **Circuit Breaker**: Implement `Hystrix` or `Resilience4j` patterns to prevent cascading failures when downstream services (e.g., Payment Gateway) are slow.
- [ ] **Rate Limiting**: Protect APIs using Token Bucket or Leaky Bucket algorithms (Redis-based) to handle traffic spikes.
- [ ] **Dead Letter Queues (DLQ)**: Handle poison pill messages that exceed max retries, ensuring they don't block the queue.
- [ ] **Graceful Shutdown**: Ensure all in-flight requests and DB connections are closed properly on SIGTERM.

#### 2. Quality Assurance & Testing (Critical Gap)
- [ ] **Unit Tests**: Add `*_test.go` for all Domain Services (Order, Inventory, Payment) with >80% coverage. Use `stretchr/testify` for assertions and `vektra/mockery` for mocking interfaces.
- [ ] **Integration Tests**: Test the full Saga flow (Order -> Inventory -> Payment) using Docker Compose and a real DB/MQ.
- [ ] **Error Handling Refactor**: Replace string-based errors (`errors.New("text")`) with **Sentinel Errors** (typed errors) to allow robust error checking (`errors.Is()`).

#### 3. Performance Optimization
- [ ] **Database Sharding**: Move from single DB to horizontal sharding (e.g., by `UserID` or `OrderID`) to support millions of rows.
- [ ] **Multi-Level Caching**: Implement Local Cache (Go map) + Remote Cache (Redis) to prevent "Hot Key" issues and Cache Penetration.
- [ ] **Connection Pooling**: Tune DB and Redis connection pools (MaxOpenConns, MaxIdleConns) based on load testing.
- [ ] **Go Profiling (pprof)**: Integrate `net/http/pprof` to visualize Goroutine leaks, Memory allocations, and CPU hotspots under load.

#### 4. Fintech Domain Knowledge (The "Money" Part)
- [ ] **Reconciliation (对账)**: Implement a daily batch job to verify internal `Payment` records against external Gateway reports (Mocked CSV).
- [ ] **Idempotency Keys**: Enforce strict idempotency on critical Payment APIs using Redis keys to prevent double-charging.
- [ ] **Audit Logging**: Immutable logs for every financial state change for compliance.

#### 4. Architecture Evolution (Microservices 2.0)
- [ ] **Service Discovery**: Replace hardcoded URLs with **Consul** or **Etcd** to support dynamic scaling (Service Registry & Discovery).
- [ ] **RPC Protocol**: Migrate internal service-to-service communication from HTTP/JSON to **gRPC/Protobuf** for higher performance and type safety.
- [ ] **Distributed Locking**: Implement Redis-based distributed locks (Redlock) for critical sections not suitable for DB locking (e.g., user-level frequency control).
- [ ] **Configuration Management**: Move from `.env` files to a centralized config server (e.g., Nacos/Etcd) with hot-reload capabilities.

#### 5. Advanced Financial Architecture (The "Deep Water" Zone)
- [ ] **TCC Pattern**: Implement **Try-Confirm-Cancel** for scenarios requiring stronger consistency than Saga (e.g., Cross-border asset transfer).
- [ ] **Data Security**: Implement field-level encryption (AES-256) for PII (Personally Identifiable Information) using a KMS mock.
- [ ] **Traffic Control**: Implement **Canary Deployment** (Grey Release) logic at the Gateway to route 1% of traffic to new service versions.
- [ ] **Disaster Recovery**: Simulate a "Region Failover" scenario where the primary DB goes down, and the system automatically switches to a standby replica (Chaos Engineering).

### Phase 1: Containerization (Current)
- [x] Dockerize all services (Multi-stage builds)
- [x] Docker Compose for local development orchestration
- [x] Environment variable configuration (.env)

### Phase 2: Observability & Monitoring
- [x] **Distributed Tracing**: Integrate Jaeger/OpenTelemetry to visualize TraceIDs across services.
- [ ] **Metrics**: Expose Prometheus metrics (`/metrics`) for request latency, error rates, and queue depth.
- [ ] **Logging**: Centralized logging (ELK Stack or Loki) to aggregate logs from all containers.

### Phase 3: CI/CD & Automation
- [x] **CI Pipeline**: GitHub Actions to run tests and linting on PRs.
- [ ] **Image Publishing**: Auto-build and push Docker images to Registry (Docker Hub/ECR) on merge.

### Phase 4: Kubernetes (K8s) Migration
- [ ] Create Helm Charts or K8s Manifests (Deployment, Service, Ingress).
- [ ] Implement **Liveness & Readiness Probes** for zero-downtime deployments.
- [ ] **Secrets Management**: Move sensitive `.env` data to K8s Secrets or HashiCorp Vault.

### Phase 5: Security & Resilience
- [ ] **API Gateway Auth**: Implement JWT validation at the Gateway level.
- [ ] **Rate Limiting**: Protect services using Redis-based rate limiting in the Gateway.
- [ ] **Circuit Breaking**: Enhance clients with Hystrix/Resilience4j patterns.
