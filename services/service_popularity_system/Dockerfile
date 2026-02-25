FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy libs module to maintain relative path structure
COPY libs/ ./libs/

# Copy service files to services/service_popularity_system to match replace directive (../../libs)
COPY services/service_popularity_system/ ./services/service_popularity_system/

# Change to service directory for build
WORKDIR /app/services/service_popularity_system

# Tidy dependencies to ensure go.sum is correct
RUN go mod tidy
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/local_server

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8080
CMD ["./server"]