# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 1. Copy go.mod and go.sum files
COPY services/api-gateway/go.mod services/api-gateway/go.sum ./services/api-gateway/

# 2. Download dependencies
WORKDIR /app/services/api-gateway
RUN go mod download

# 3. Copy source code
WORKDIR /app
COPY services/api-gateway/ ./services/api-gateway/

# 4. Build
WORKDIR /app/services/api-gateway
# 显式指定 -v 打印构建日志，确保构建成功
RUN CGO_ENABLED=0 GOOS=linux go build -v -o api-gateway cmd/api-gateway/main.go

# Run Stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/services/api-gateway/api-gateway .

# Expose port
EXPOSE 8000

# Run the application
CMD ["./api-gateway"]
