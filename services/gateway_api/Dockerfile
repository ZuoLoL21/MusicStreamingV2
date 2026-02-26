# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy shared libs first
COPY libs/ ./libs/

# Copy service code
COPY services/gateway_api/ ./services/gateway_api/

# Build the binary
WORKDIR /app/services/gateway_api
RUN go mod tidy
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway_api ./cmd/server

# Runtime stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/services/gateway_api/gateway_api .

EXPOSE 8080

CMD ["./gateway_api"]
