# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 1. Copy go.mod and go.sum files
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY services/payment-service/go.mod services/payment-service/go.sum ./services/payment-service/

# 2. Download dependencies
WORKDIR /app/services/payment-service
RUN go mod download

# 3. Copy source code
WORKDIR /app
COPY pkg/ ./pkg/
COPY services/payment-service/ ./services/payment-service/

# 4. Build
WORKDIR /app/services/payment-service
RUN go build -o payment-service ./cmd/payment-service/main.go

# Run Stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/services/payment-service/payment-service .
COPY --from=builder /app/services/payment-service/configs ./configs

EXPOSE 8083

RUN chmod +x ./payment-service

CMD ["./payment-service"]
