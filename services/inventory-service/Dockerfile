# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 1. Copy go.mod and go.sum files
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY services/inventory-service/go.mod services/inventory-service/go.sum ./services/inventory-service/

# 2. Download dependencies
WORKDIR /app/services/inventory-service
RUN go mod download

# 3. Copy source code
WORKDIR /app
COPY pkg/ ./pkg/
COPY services/inventory-service/ ./services/inventory-service/

# 4. Build
WORKDIR /app/services/inventory-service
RUN go build -o inventory-service ./cmd/inventory-service/main.go

# Run Stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/services/inventory-service/inventory-service .
COPY --from=builder /app/services/inventory-service/configs ./configs

EXPOSE 8082

CMD ["./inventory-service"]
