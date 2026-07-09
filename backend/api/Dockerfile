# syntax=docker/dockerfile:1

# Stage 1: Build stage (shared for both API and Worker)
FROM golang:1.25-bookworm AS builder

WORKDIR /app
ENV GOCACHE=/go-cache
ENV GOTMPDIR=/go-tmp

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy application source code
COPY . .

# Keep Go's build cache and temporary files off /tmp so Docker builds
# remain stable even when the default temp filesystem is mounted read-only.
RUN mkdir -p "$GOCACHE" "$GOTMPDIR"

# Build statically linked binaries for Linux
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/omnilogs-api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/omnilogs-worker ./cmd/worker
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/omnilogs-migrate ./cmd/migrate

# Stage 2: Final runner stage for API
FROM debian:bookworm-slim AS api

WORKDIR /app

# Install CA certificates for external HTTPS requests
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy built binary from builder
COPY --from=builder /bin/omnilogs-api /usr/local/bin/omnilogs-api

# Copy API swagger documentation and schema files
COPY docs/swagger ./docs/swagger
COPY api/openapi.yaml ./api/openapi.yaml

EXPOSE 2910
CMD ["omnilogs-api"]

# Stage 3: Final runner stage for Worker
FROM debian:bookworm-slim AS worker

WORKDIR /app    

# Install CA certificates for external HTTPS requests
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy built binary from builder
COPY --from=builder /bin/omnilogs-worker /usr/local/bin/omnilogs-worker

CMD ["omnilogs-worker"]

# Stage 4: Final runner stage for Migrate
FROM debian:bookworm-slim AS migrate

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /bin/omnilogs-migrate /usr/local/bin/omnilogs-migrate

CMD ["omnilogs-migrate"]
