# vv-ecommerce

## Project Overview
vv-ecommerce is a microservices-based e-commerce platform designed to handle core online shopping functionalities. The initial setup includes foundational configuration files and a placeholder for the order service, with plans to expand to additional microservices (e.g., product catalog, payment processing, user management).

## Prerequisites
- Docker Desktop (for running containerized services via `docker-compose.yml`)
- Visual Studio Code (recommended IDE, use `vv-ecommerce.code-workspace` for workspace setup)
- Go (optional, for local development of microservices like `order-service`)

## Setup Instructions
1. Clone the repository :
   ```bash
   git clone https://github.com/pweex2/vv-ecommerce
   ```
2. Navigate to the project root directory:
   ```cmd
   cd vv-ecommerce
   ```

## Running the Project
Use Docker Compose to start all containerized services:
```cmd
docker-compose up --build
```

## Project Structure
- `README.md`: Project documentation (you're here)
- `docker-compose.yml`: Container orchestration configuration
- `vv-ecommerce.code-workspace`: VS Code workspace setup
- `services/order-service`: Placeholder for the order management microservice (to be implemented)

## Next Steps
- Implement core functionality for the `order-service` (e.g., create order, update order status)
- Add additional microservices (product-service, payment-service)
- Integrate a database (e.g., PostgreSQL, MongoDB) for persistent data storage
- Implement user authentication and authorization