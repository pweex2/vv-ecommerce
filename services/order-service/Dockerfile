# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 1. Copy go.mod and go.sum files
# We need to copy pkg's go.mod because it is referenced by replace directive
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY services/order-service/go.mod services/order-service/go.sum ./services/order-service/

# 2. Download dependencies
WORKDIR /app/services/order-service
RUN go mod download

# 3. Copy source code
WORKDIR /app
COPY pkg/ ./pkg/
COPY services/order-service/ ./services/order-service/

# 4. Build
WORKDIR /app/services/order-service
RUN go build -o order-service ./cmd/order-service/main.go

# Run Stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/services/order-service/order-service .
COPY --from=builder /app/services/order-service/configs ./configs

EXPOSE 8081

RUN chmod +x ./order-service

CMD ["./order-service"]
