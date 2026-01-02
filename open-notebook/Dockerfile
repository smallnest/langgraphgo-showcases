# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git make build-base

# Copy go mod files from root if needed? 
# The project is at /Users/smallnest/ai/langgraphgo
# The showcase is at /Users/smallnest/ai/langgraphgo/showcases/open-notebook
# The go.mod in showcases/open-notebook has a replace for langgraphgo.
# So we need to copy the whole repo or handle the replace.

# Since we are in the showcase directory context during build:
# We need to copy the parent directory's code if we want to build it properly.
# But Docker context usually doesn't allow going up.

# For simplicity, assuming the build context is the root of the repo 
# OR the showcase handles its own dependencies.
# Given the previous Dockerfile, it seems it was designed to run from the showcase dir.

COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
# Use CGO_ENABLED=0 since modernc.org/sqlite is used
RUN CGO_ENABLED=0 GOOS=linux go build -o open-notebook main.go

# Runtime stage
FROM alpine:3.20

WORKDIR /app

# Install runtime dependencies
# Add python3 and pip for markitdown
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    python3 \
    py3-pip \
    libmagic

# Install markitdown
RUN pip install --break-system-packages markitdown

# Create data directory
RUN mkdir -p /data

# Copy binary from builder
COPY --from=builder /app/open-notebook .

# Expose port
EXPOSE 8080

# Set environment variables
ENV STORE_PATH=/data/checkpoints.db
ENV SQLITE_PATH=/data/vector.db
ENV SERVER_HOST=0.0.0.0
ENV SERVER_PORT=8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Run the application
CMD ["./open-notebook", "-server"]