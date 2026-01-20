# Dockerfile for billing-service
FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o billing-service ./cmd/api/main.go

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/billing-service .

EXPOSE 8080
CMD ["./billing-service"]
