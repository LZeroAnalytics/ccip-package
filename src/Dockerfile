FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git build-base linux-headers

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the deployer
RUN go build -o ccip-deployer ./cmd/deployer

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary
COPY --from=builder /app/ccip-deployer .

# Make executable
RUN chmod +x ccip-deployer

# Create configs directory
RUN mkdir -p configs

ENTRYPOINT ["/app/ccip-deployer"] 