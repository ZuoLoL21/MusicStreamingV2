# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy shared libs first
COPY libs/ ./libs/

# Copy service code
COPY services/gateway_recommendation/ ./services/gateway_recommendation/

# Build the binary
WORKDIR /app/services/gateway_recommendation
RUN go mod tidy
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway ./cmd/server

# Runtime stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/services/gateway_recommendation/gateway .

EXPOSE 8090

CMD ["./gateway"]
