# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
WORKDIR /app/services/billing-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/billing-service ./cmd/api/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata curl
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/bin/billing-service .

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./billing-service"]
