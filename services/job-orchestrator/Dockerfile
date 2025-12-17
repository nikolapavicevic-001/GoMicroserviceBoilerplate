FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY services/job-orchestrator/go.mod services/job-orchestrator/go.sum ./
RUN go mod download

# Copy source code
COPY services/job-orchestrator/ .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/job-orchestrator ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/job-orchestrator .

EXPOSE 8080

CMD ["./job-orchestrator"]

