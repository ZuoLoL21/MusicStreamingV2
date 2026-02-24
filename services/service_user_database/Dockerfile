FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy libs module to maintain relative path structure
COPY libs/ ./libs/

# Copy service files to services/user_database to match replace directive (../../libs)
COPY services/user_database/ ./services/user_database/

# Change to service directory for build
WORKDIR /app/services/user_database

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8080
CMD ["./server"]