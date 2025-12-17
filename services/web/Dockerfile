FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY services/web/go.mod services/web/go.sum ./services/web/
WORKDIR /app/services/web
RUN go mod download

COPY services/web .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/services/web/web-service .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/services/web/web-service .
COPY --from=builder /app/services/web/templates ./templates

EXPOSE 8080

CMD ["./web-service"]

