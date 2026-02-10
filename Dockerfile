# Build frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Build backend
FROM golang:1.25-alpine AS backend-builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o vibe-server .
RUN CGO_ENABLED=1 go build -o seed-db ./cmd/seed

# Runtime
FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app

# Copy backend binary
COPY --from=backend-builder /app/vibe-server .
COPY --from=backend-builder /app/seed-db .

# Copy entrypoint
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

# Copy frontend dist
COPY --from=frontend-builder /app/frontend/dist ./static

# Create data directory
RUN mkdir -p /app/data

ENV PORT=8080
ENV DATABASE_PATH=/app/data/vibe.db
ENV ENABLE_SCRAPER=false

EXPOSE 8080

ENTRYPOINT ["./entrypoint.sh"]
