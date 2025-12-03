# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files for workspace
COPY go.work go.work.sum ./
COPY shared/go.mod shared/go.sum ./shared/
COPY services/web-app/go.mod services/web-app/go.sum ./services/web-app/
COPY services/user-service/go.mod services/user-service/go.sum ./services/user-service/
COPY services/api-gateway/go.mod services/api-gateway/go.sum ./services/api-gateway/

# Download dependencies
RUN cd services/web-app && go mod download

# Copy source code
COPY shared/ ./shared/
COPY services/web-app/ ./services/web-app/

# Build the application
RUN cd services/web-app && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /web-app cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /web-app .

# Copy templates and static files
COPY --from=builder /app/services/web-app/templates ./templates/
COPY --from=builder /app/services/web-app/static ./static/

# Expose HTTP port
EXPOSE 3000

# Run the application
CMD ["./web-app"]
