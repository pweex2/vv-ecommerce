# VV-Ecommerce

A lightweight, microservices-based e-commerce backend system built with **Go** and the **Gin** framework.

## 🏗 Architecture

This project adopts a **Monorepo** structure managed by Go Workspaces (`go.work`). It consists of three decoupled microservices that communicate via HTTP.

### Tech Stack
- **Language**: Go 1.25+
- **Web Framework**: [Gin](https://github.com/gin-gonic/gin) (High-performance HTTP web framework)
- **Database**: MySQL (accessed via [GORM](https://gorm.io/))
- **Configuration**: Viper
- **Architecture**: Microservices, RESTful API

## 🚀 Services

| Service | Port | Description |
|---------|------|-------------|
| **Order Service** | `:8081` | Manages order creation, retrieval, and status updates. Orchestrates calls to Inventory and Payment services. |
| **Inventory Service** | `:8082` | Manages product stock levels. Handles stock deduction and checking. |
| **Payment Service** | `:8083` | Handles payment processing and recording. |

## 🛠 Prerequisites

- **Go** (1.22 or higher recommended)
- **MySQL** (Running on localhost:3306)

## 📦 Getting Started

### 1. Database Setup
Ensure you have a MySQL instance running. Create the following databases:
```sql
CREATE DATABASE order_db;
CREATE DATABASE inventory_db;
CREATE DATABASE payment_db;
```
*Note: The services are configured to use `root:root` by default. You can modify `configs/config.development.yaml` in each service if your credentials differ.*

### 2. Run the Services
You need to start each service in a separate terminal.

#### Start Inventory Service
```bash
cd services/inventory-service
go run ./cmd/inventory-service/main.go
```

#### Start Payment Service
```bash
cd services/payment-service
go run ./cmd/payment-service/main.go
```

#### Start Order Service
```bash
cd services/order-service
go run ./cmd/order-service/main.go
```

## 🔌 API Endpoints

### Order Service (`:8081`)
- `POST /orders` - Create a new order (requires `user_id`, `sku`, `total_amount`)
- `GET /orders?order_id={id}` - Get order details
- `PATCH /orders` - Update order status

### Inventory Service (`:8082`)
- `GET /inventories?product_id={id}` - List inventories for a product
- `GET /inventory/sku?sku={sku}` - Get specific inventory details
- `POST /inventory/decrease` - Decrease stock (Internal use)
- `POST /inventory/create` - Create initial inventory

### Payment Service (`:8083`)
- `POST /payments` - Process a payment
- `GET /payments?order_id={id}` - Get payment details

## 📂 Project Structure

```
vv-ecommerce/
├── pkg/                # Shared packages (Clients, Common Responses)
├── services/           # Microservices source code
│   ├── inventory-service/
│   ├── order-service/
│   └── payment-service/
├── go.work             # Go Workspace configuration
└── README.md
```

## ☁️ Deployment & Production Roadmap

Currently, this project is optimized for local development. To move towards a production-ready state (e.g., Docker, K8s), the following enhancements are planned:

- **🐳 Containerization**: Create multi-stage `Dockerfile`s for each service to ensure consistent runtime environments.
- **☸️ Orchestration**: Deploy using **Kubernetes** (or Docker Compose for dev-prod parity), utilizing `Deployments` and `Services`.
- **⚙️ Configuration**: Migrate from local YAML files to **Environment Variables** or **K8s ConfigMaps/Secrets** for better security and flexibility.
- **🔍 Observability**:
  - **Logging**: Implement structured JSON logging (e.g., Zap/Logrus) for aggregation (ELK/Loki).
  - **Tracing**: Integrate **OpenTelemetry/Jaeger** to trace requests across microservices.
  - **Metrics**: Expose Prometheus metrics for monitoring health and performance.
- **🛡️ API Gateway & BFF**: 
  - Use a gateway (e.g., Nginx, Kong) for global concerns like ingress, rate limiting, and SSL termination.
  - Implement a **BFF (Backend for Frontend)** layer (possibly GraphQL) to aggregate data and reduce round-trips for clients.
- **📡 Service Discovery**: Leverage Kubernetes native DNS or tools like Consul/Etcd to dynamically locate service instances without hardcoded URLs.
- **📨 Event-Driven Architecture**:
  - **Async Messaging**: Decouple critical paths (e.g., "Order Created" -> "Stock Deducted") using RabbitMQ or Kafka.
  - **Outbox Pattern**: Implement the Transactional Outbox Pattern to ensure data consistency between the database and the message broker.
- **🔒 Security**: 
  - **Zero Trust**: Implement **mTLS** for encrypted service-to-service communication.
  - **Auth**: Centralized **OAuth2/OIDC** (e.g., Keycloak) for user authentication and JWT verification.
- **🛡️ Resilience**: Apply **Circuit Breakers** (e.g., Hystrix/GoResilience) and **Retries** with exponential backoff to prevent cascading failures.
- **🗄️ Data Management**: Replace `AutoMigrate` with versioned **Database Migrations** (e.g., Golang-Migrate) for safe schema evolution.
- **📝 Documentation**: Auto-generate API documentation using **Swagger/OpenAPI** to keep API contracts up-to-date.
- **⚡ High Performance Communication**: Migrate internal service-to-service communication from HTTP/REST to **gRPC** (Protobuf) for lower latency and strict schema enforcement.
- **🚀 CI/CD**: Set up automated pipelines (GitHub Actions) for testing, building, and deploying.