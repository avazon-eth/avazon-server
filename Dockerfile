# Stage 1: Builder
FROM golang:1.21.0 AS builder

# Set environment variables for native ARM64 build
ENV CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=arm64

# Install necessary build dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN go build -o /app/main

# Stage 2: Final Image
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && \
    apt-get install -y \
    ca-certificates \
    libsqlite3-0 \
    && rm -rf /var/lib/apt/lists/*

# Set the working directory in the container
WORKDIR /app

# Copy the Go binary from the builder stage
COPY --from=builder /app/main /app/main

# (Optional) Copy the .env file if your application relies on it
COPY --from=builder /app/.env /app/.env

# Expose the application port
EXPOSE 8080

# Command to run the application
CMD ["./main"]
