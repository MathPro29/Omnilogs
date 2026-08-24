# syntax=docker/dockerfile:1

# Stage 1: ขั้นตอน Build ใช้ร่วมกันสำหรับ API และ Worker ทุกตัว
FROM golang:1.25-bookworm AS builder

WORKDIR /app
ENV GOCACHE=/go-cache
ENV GOPROXY=https://proxy.golang.org,direct
# Keep compiler pressure predictable on developer machines. Override at build
# time with --build-arg GO_BUILD_PARALLELISM=<n> when more capacity is available.
ARG GO_BUILD_PARALLELISM=2

# คัดลอกไฟล์ dependency ของ Go ก่อน เพื่อให้ Docker cache layer นี้ไว้ใช้ซ้ำได้
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# คัดลอก source code ของแอปพลิเคชันเข้าไปใน image สำหรับ Build
COPY . .

# เตรียมโฟลเดอร์ cache ของ Go ให้เป็นตำแหน่งที่เขียนได้แน่นอน
# ไฟล์ชั่วคราวของ Go จะใช้ /tmp ซึ่งเขียนได้ใน builder image
# และช่วยหลีกเลี่ยงปัญหา /go-tmp ที่เป็น read-only บน Docker Desktop บางเครื่อง
RUN mkdir -p "$GOCACHE" /app/bin /tmp/go-build && chmod 1777 /tmp /tmp/go-build

# สร้าง Binary แบบ Static สำหรับ Linux โดยใช้ cache mount ของ BuildKit
# คำสั่ง test -s ตรวจว่า Binary ถูกสร้างจริงและมีขนาดมากกว่า 0 bytes
# เพื่อป้องกันกรณี Build layer เสียแต่ Docker รายงานว่าสำเร็จ
FROM builder AS api-builder
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/go-cache \
    GOMAXPROCS=$GO_BUILD_PARALLELISM CGO_ENABLED=0 GOOS=linux go build -p=$GO_BUILD_PARALLELISM -ldflags="-s -w" -o /app/bin/omnilogs-api ./cmd/api && \
    test -s /app/bin/omnilogs-api

FROM builder AS worker-builder
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/go-cache \
    GOMAXPROCS=$GO_BUILD_PARALLELISM CGO_ENABLED=0 GOOS=linux go build -p=$GO_BUILD_PARALLELISM -ldflags="-s -w" -o /app/bin/omnilogs-worker ./cmd/worker && \
    test -s /app/bin/omnilogs-worker

FROM builder AS retention-worker-builder
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/go-cache \
    GOMAXPROCS=$GO_BUILD_PARALLELISM CGO_ENABLED=0 GOOS=linux go build -p=$GO_BUILD_PARALLELISM -ldflags="-s -w" -o /app/bin/omnilogs-retention-worker ./cmd/retention-worker && \
    test -s /app/bin/omnilogs-retention-worker

FROM builder AS migrate-builder
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/go-cache \
    GOMAXPROCS=$GO_BUILD_PARALLELISM CGO_ENABLED=0 GOOS=linux go build -p=$GO_BUILD_PARALLELISM -ldflags="-s -w" -o /app/bin/omnilogs-migrate ./cmd/migrate && \
    test -s /app/bin/omnilogs-migrate

FROM builder AS seed-builder
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/go-cache \
    GOMAXPROCS=$GO_BUILD_PARALLELISM CGO_ENABLED=0 GOOS=linux go build -p=$GO_BUILD_PARALLELISM -ldflags="-s -w" -o /app/bin/omnilogs-seed ./cmd/seed && \
    test -s /app/bin/omnilogs-seed

# Stage 2: Image สำหรับรัน Ingest Worker
FROM debian:bookworm-slim AS worker

WORKDIR /app    

# ติดตั้ง CA certificates สำหรับการเชื่อมต่อ HTTPS ภายนอก
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

# คัดลอก Binary ของ Ingest Worker จากขั้นตอน Build มาไว้ใน Runtime image
COPY --from=worker-builder /app/bin/omnilogs-worker /usr/local/bin/omnilogs-worker

CMD ["omnilogs-worker"]

# Stage 3: Image สำหรับรัน Retention Worker
FROM debian:bookworm-slim AS retention-worker

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=retention-worker-builder /app/bin/omnilogs-retention-worker /usr/local/bin/omnilogs-retention-worker

CMD ["omnilogs-retention-worker"]

# Stage 4: Image สำหรับรัน Database Migration แบบ One-shot
FROM debian:bookworm-slim AS migrate

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=migrate-builder /app/bin/omnilogs-migrate /usr/local/bin/omnilogs-migrate

CMD ["omnilogs-migrate"]

FROM debian:bookworm-slim AS seed

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=seed-builder /app/bin/omnilogs-seed /usr/local/bin/omnilogs-seed

CMD ["omnilogs-seed"]

# Stage 5: Image สำหรับรัน API ซึ่งเป็น Target หลักของ Dockerfile
FROM debian:bookworm-slim AS api

WORKDIR /app

# ติดตั้ง CA certificates และ curl สำหรับการเชื่อมต่อ HTTPS
# curl ถูกใช้โดย API healthcheck ใน docker-compose.yml
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates curl && \
    rm -rf /var/lib/apt/lists/*

# คัดลอก Binary ของ API จากขั้นตอน Build มาไว้ใน Runtime image
COPY --from=api-builder /app/bin/omnilogs-api /usr/local/bin/omnilogs-api

# คัดลอกเอกสาร Swagger และ OpenAPI schema ที่ API ต้องใช้ตอนรัน
COPY docs/swagger ./docs/swagger
COPY api/openapi.yaml ./api/openapi.yaml

EXPOSE 2910
CMD ["omnilogs-api"]
