# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o open-notebook backend/*.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    sqlite-libs \
    tzdata

# Create data directory
RUN mkdir -p /data

# Copy binary from builder
COPY --from=builder /app/open-notebook .
COPY --from=builder /app/frontend ./frontend

# Expose port
EXPOSE 8080

# Set environment variables
ENV STORE_PATH=/data/checkpoints.db
ENV SQLITE_PATH=/data/vector.db

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Run the application
CMD ["./open-notebook", "-server"]
