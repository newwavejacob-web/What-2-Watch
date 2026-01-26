# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies for SQLite (CGO)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy dependency files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO enabled for SQLite
RUN CGO_ENABLED=1 go build -o vibe-server .
RUN CGO_ENABLED=1 go build -o seed-db ./cmd/seed

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binaries
COPY --from=builder /app/vibe-server .
COPY --from=builder /app/seed-db .

# Create data directory
RUN mkdir -p /app/data

# Environment variables
ENV PORT=8080
ENV DATABASE_PATH=/app/data/vibe.db
ENV ENABLE_SCRAPER=false

EXPOSE 8080

# Default command
CMD ["./vibe-server"]
