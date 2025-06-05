# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o job-distributor ./main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/job-distributor .

# Create config directory
RUN mkdir -p /etc/job-distributor

# Hardcode default environment variables for easy Kurtosis deployment
ENV JD_PORT=50051
ENV JD_LOG_LEVEL=info
ENV JD_ENABLE_METRICS=false
ENV JD_METRICS_PORT=8080

# Expose the gRPC port (Job Distributor runs on 50051, not 8080)
EXPOSE 50051

# Expose metrics port
EXPOSE 8080

# Health check on the gRPC port
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD nc -z localhost 50051 || exit 1

# Command to run with hardcoded defaults
CMD ["./job-distributor"] 