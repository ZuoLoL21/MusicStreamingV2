FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy libs module to maintain relative path structure
COPY libs/ ./libs/

# Copy service files to services/file_storage to match replace directive (../../libs)
COPY services/file_storage/ ./services/file_storage/

# Change to service directory for build
WORKDIR /app/services/file_storage

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/local_server

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8000
CMD ["./server"]
