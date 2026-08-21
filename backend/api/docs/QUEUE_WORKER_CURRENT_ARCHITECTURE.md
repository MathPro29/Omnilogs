# Queue and Worker Current Architecture

This document describes how the queue and worker pipeline currently works in this codebase.
It is intended as a handoff document for AI-assisted architecture review and message broker selection.

## Purpose

The current pipeline accepts log payloads from the API, stages them into an internal queue, and processes them into Elasticsearch while also writing metadata and sensitive-data artifacts into PostgreSQL.

## Main Components

- API queue routes: `backend/api/routes/log_queue_routes.go`
- Queue handler: `backend/api/internal/log_queues/handler/handler.go`
- Queue ingest use case: `backend/api/internal/log_queues/usecase/usecase.go`
- Queue persistence repository: `backend/api/internal/log_queues/repository/repository.go`
- Worker bootstrap: `backend/api/cmd/worker/main.go`
- Worker runtime: `backend/api/internal/bootstrap/worker.go`
- Worker processor: `backend/api/internal/worker/processor.go`
- Worker use cases:
  - `backend/api/internal/worker/usecase/usecase.go`
  - `backend/api/internal/worker/usecase/batch_processor.go`
  - `backend/api/internal/worker/usecase/retry_queue.go`
  - `backend/api/internal/worker/usecase/validation.go`
  - `backend/api/internal/worker/usecase/elasticsearch_builder.go`
- Worker repository: `backend/api/internal/worker/repository/repository.go`

## Current End-to-End Flow

1. A client calls `POST /api/v1/queues`.
2. The request is validated by `QueueHandler`.
3. `Enqueue(...)` validates product/environment relationships.
4. Logs are not written to PostgreSQL immediately one-by-one.
5. Logs are first pushed into an in-memory channel (`jobChan`) that is created per `productId-environmentId` key.
6. A per-key in-process worker buffers logs in memory.
7. The buffered logs are flushed into PostgreSQL as:
   - one row in `log_queue_batches`
   - many rows in `log_queue_items`
8. A worker process later claims one batch from PostgreSQL.
9. The worker loads pending items in that batch.
10. Each item is processed sequentially.
11. The worker validates hierarchy and payload shape.
12. The worker masks sensitive fields and stores secrets in PostgreSQL.
13. The worker builds one Elasticsearch document and indexes it.
14. The worker writes a `log_index_ref` row in PostgreSQL.
15. The worker stores the original encrypted payload in PostgreSQL-backed object storage metadata.
16. The processed queue item is deleted from `log_queue_items`.
17. When all items in the batch are attempted, the batch status is updated.

## Queue Ingest Design

### In-memory staging before DB

The queue ingest layer currently uses an in-memory buffer:

- Worker map key: `productId-environmentId`
- Buffer type: Go channel
- Channel capacity: `10000` items per key
- Default flush size: `100`
- Default flush interval: `3 seconds`

This means the system currently behaves like:

- API request
- in-memory buffering
- periodic batch flush to PostgreSQL

### Important implication

The first durable write does **not** happen at API receive time.
Durability begins only after the in-memory buffer flushes into PostgreSQL.

If the process crashes before flush:

- buffered logs in memory can be lost

If the channel becomes full:

- API request goroutines will block on channel send

## Queue Data Model

### `log_queue_batches`

Represents one logical batch of logs to be processed together.

Observed fields used in code:

- `batch_id`
- `product_id`
- `environment_id`
- `queue_key`
- `source_type`
- `source_platform`
- `status`
- `priority`
- `received_at`
- `available_at`
- `processing_started_at`
- `processed_at`
- `locked_by`
- `locked_at`
- `processing_attempts`
- `error_message`

### `log_queue_items`

Represents each log payload inside a batch.

Observed fields used in code:

- `queue_item_id`
- `batch_id`
- `sequence_no`
- `input_payload`
- `status`
- `worker_id`
- `retry_count`
- `max_retry_count`
- `next_retry_at`
- `processing_started_at`
- `processed_at`
- `error_message`

## Worker Runtime Model

### Deployment model

There are two distinct runtimes:

- API server
- worker process

The worker can run independently via `backend/api/cmd/worker/main.go`.

### Batch claim model

The worker:

- polls PostgreSQL
- claims one eligible batch at a time
- uses row locking with `FOR UPDATE SKIP LOCKED`
- marks the batch as `PROCESSING`

### Item processing model

Inside a claimed batch:

- items are loaded in sequence order
- items are processed one-by-one
- processing is sequential, not parallel

### Retry and stuck lock handling

The worker also periodically:

- checks retryable items/batches
- resets batches back to retryable states when needed
- recovers stuck locks based on timeout
- runs retention cleanup logic

## Current Status Lifecycle

### Batch status

Observed batch statuses include:

- `QUEUED`
- `PROCESSING`
- `COMPLETED`
- `RETRY_PENDING`
- `PARTIAL`
- `FAILED`

### Item status

Observed item statuses include:

- `PENDING`
- `PROCESSING`
- `RETRY_PENDING`
- `PROCESSED`
- failure terminal behavior is handled through failure records and state transitions

## Current Processing Responsibilities Per Log

For each individual log, the worker may perform all of the following:

- mark item as processing
- parse payload JSON
- validate product/environment/project/category hierarchy
- load sensitive field definitions
- load masking rules
- mask payload recursively
- store extracted sensitive secrets in PostgreSQL
- build Elasticsearch document
- index one document into Elasticsearch
- write `log_index_ref`
- encrypt and store original payload metadata in PostgreSQL
- delete processed queue item

This means the pipeline is not only a queue consumer.
It is also:

- a validation pipeline
- a masking pipeline
- an indexing pipeline
- a metadata persistence pipeline
- a sensitive-data persistence pipeline

## Manual Consume Path

There is also a manual consume endpoint:

- `POST /api/v1/queues/consume`

This endpoint calls `RunOnce()` and processes only one batch.

This is useful for:

- manual triggering
- testing
- UI-driven import flows

But it also means the API layer currently knows about direct queue consumption.

## Storage Systems Used Today

### PostgreSQL

PostgreSQL is currently used for:

- queue batches
- queue items
- retry state
- batch locks
- sensitive secrets
- log index references
- encrypted original payload metadata
- archive metadata

### Elasticsearch

Elasticsearch is currently used for:

- searchable log documents
- main log retrieval
- dashboard queries

## Current Scaling Characteristics

### Good fit for

- MVP or internal systems
- low to medium throughput
- workloads where policy validation matters more than raw ingestion speed
- product/environment partitioning with moderate concurrency

### Weak fit for

- sustained very high throughput
- strong durability guarantees at ingress time
- very large fan-out consumer groups
- replay-heavy or stream-processing-heavy architectures

## Key Constraints That Matter for Message Broker Selection

An AI selecting a broker should consider these current constraints:

1. The system currently batches by `productId-environmentId`.
2. Ordering is meaningful at least within a batch and likely within a source stream.
3. Retry and delayed retry behavior already exist.
4. Sensitive-data handling happens in the worker, not in the API.
5. The pipeline is write-heavy on PostgreSQL today.
6. The current design mixes queue state and application metadata in the same database.
7. The current API path includes an optional direct-consume behavior.
8. The worker processes one batch at a time and one item at a time.
9. Elasticsearch indexing is currently single-document, not bulk.
10. Durability at ingress time is currently weaker than a real broker-backed architecture.

## Known Architectural Risks

- In-memory buffer loss before flush
- API blocking when channel is full
- PostgreSQL becoming both queue store and metadata bottleneck
- Sequential worker throughput
- High per-log processing cost
- Single-document Elasticsearch writes
- Tight coupling between queueing and downstream processing rules

## Questions an AI Should Answer When Choosing a Broker

- Is strict ordering needed per product, environment, source, or feature hierarchy?
- Is at-least-once enough, or is exactly-once-like behavior required?
- What is the expected sustained ingest rate and peak burst rate?
- How long must messages survive if workers are down?
- Is delayed retry a hard requirement inside the broker, or can the app keep managing retries?
- Should replay/backfill be a first-class capability?
- Is operational simplicity more important than maximum throughput?
- Is the team comfortable operating Kafka-class infrastructure, or is a simpler broker preferred?
- Should broker partitioning align with `productId-environmentId`?

## Suggested Target Characteristics for a Future Broker-Based Design

If this system migrates to a real message broker, the target design would ideally provide:

- durable write on ingress
- consumer horizontal scaling
- ordering strategy per key
- delayed retry or DLQ support
- replay support
- queue lag observability
- backpressure handling
- decoupling between API accept path and consume path

## Short Summary for Broker Evaluation

Current architecture summary:

- API receives logs
- logs enter in-memory per-key channel
- periodic flush writes batches/items to PostgreSQL
- worker claims one batch from PostgreSQL
- worker processes logs sequentially
- worker validates, masks, stores secrets, indexes into Elasticsearch, and stores metadata

This is currently a PostgreSQL-backed queue pipeline with an in-process memory buffer in front of it, not a true broker-based streaming architecture.
