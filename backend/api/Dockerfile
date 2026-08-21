# syntax=docker/dockerfile:1

# Stage 1: Build stage (shared for both API and Worker)
FROM golang:1.25-bookworm AS builder

WORKDIR /app
ENV GOCACHE=/go-cache
ENV GOPROXY=https://proxy.golang.org,direct

# Cache Go modules
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# Copy application source code
COPY . .

# Keep the Go build cache in a stable writable directory. Go temporary build
# files use /tmp, which is writable in the builder image and avoids the
# read-only /go-tmp path on some Docker Desktop BuildKit workers.
RUN mkdir -p "$GOCACHE" /app/bin /tmp/go-build && chmod 1777 /tmp /tmp/go-build

# Build statically linked binaries for Linux using build cache mounts.
# The explicit toolchain check prevents a corrupted base-image layer from
# producing a false-positive `go build` step with no binary output.
FROM builder AS api-builder
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/omnilogs-api ./cmd/api && \
    test -s /app/bin/omnilogs-api

FROM builder AS worker-builder
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/omnilogs-worker ./cmd/worker && \
    test -s /app/bin/omnilogs-worker

FROM builder AS retention-worker-builder
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/omnilogs-retention-worker ./cmd/retention-worker && \
    test -s /app/bin/omnilogs-retention-worker

FROM builder AS migrate-builder
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/omnilogs-migrate ./cmd/migrate && \
    test -s /app/bin/omnilogs-migrate

# Stage 3: Final runner stage for Worker
FROM debian:bookworm-slim AS worker

WORKDIR /app    

# Install CA certificates for external HTTPS requests
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy built binary from builder
COPY --from=worker-builder /app/bin/omnilogs-worker /usr/local/bin/omnilogs-worker

CMD ["omnilogs-worker"]

# Stage 3b: Final runner stage for Retention Worker
FROM debian:bookworm-slim AS retention-worker

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=retention-worker-builder /app/bin/omnilogs-retention-worker /usr/local/bin/omnilogs-retention-worker

CMD ["omnilogs-retention-worker"]

# Stage 4: Final runner stage for Migrate
FROM debian:bookworm-slim AS migrate

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=migrate-builder /app/bin/omnilogs-migrate /usr/local/bin/omnilogs-migrate

CMD ["omnilogs-migrate"]

# Stage 5: Final runner stage for API (Default target)
FROM debian:bookworm-slim AS api

WORKDIR /app

# Install CA certificates for external HTTPS requests
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy built binary from builder
COPY --from=api-builder /app/bin/omnilogs-api /usr/local/bin/omnilogs-api

# Copy API swagger documentation and schema files
COPY docs/swagger ./docs/swagger
COPY api/openapi.yaml ./api/openapi.yaml

EXPOSE 2910
CMD ["omnilogs-api"]
