FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install protoc and build tools
RUN apk add --no-cache protobuf protoc make git

# Install protoc plugins via Go
ENV GOPATH=/go
ENV PATH=$PATH:/go/bin
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

# Install sqlc
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Create services/device directory
RUN mkdir -p services/device

# Copy go.mod
COPY services/device/go.mod ./services/device/
WORKDIR /app/services/device
RUN go mod download

# Copy SQL files
WORKDIR /app
COPY services/device/sql ./services/device/sql
COPY services/device/sqlc.yaml ./services/device/

# Generate sqlc code
WORKDIR /app/services/device
RUN sqlc generate || echo "sqlc generate failed - will use placeholder"

# Copy proto files
WORKDIR /app
COPY services/device/proto ./services/device/proto

# Generate proto code
WORKDIR /app/services/device
RUN mkdir -p proto/device
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    --proto_path=./proto \
    proto/device/device.proto

# Copy source code
COPY services/device .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/services/device/device-service ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/services/device/device-service .

EXPOSE 9090 8080

CMD ["./device-service"]

