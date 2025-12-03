# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files for workspace
COPY go.work go.work.sum ./
COPY shared/go.mod shared/go.sum ./shared/
COPY services/api-gateway/go.mod services/api-gateway/go.sum ./services/api-gateway/
COPY services/user-service/go.mod services/user-service/go.sum ./services/user-service/
COPY services/web-app/go.mod services/web-app/go.sum ./services/web-app/

# Download dependencies
RUN cd services/api-gateway && go mod download

# Copy source code
COPY shared/ ./shared/
COPY services/api-gateway/ ./services/api-gateway/

# Build the application
RUN cd services/api-gateway && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /api-gateway cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /api-gateway .

# Expose HTTP port
EXPOSE 8080

# Run the application
CMD ["./api-gateway"]
