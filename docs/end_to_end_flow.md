# OmniLogs end-to-end ingestion flow

## Flow

1. External service sends `POST /api/v1/ingest/logs` (or the backwards-compatible `/api/v1/logs/ingest`) with an API key.
2. `ProductAPIKeyAuthMiddleware` verifies the key hash, active/revoked/expiry state, `LOG_INGEST_CREATE` permission, and binds the product/environment/source scope.
3. The queue handler validates the request and payload scope. The queue use case validates each JSON object, creates the batch and its queue items, then publishes one JetStream message per queue item using its persisted `queue_item_id`.
4. The worker pull-consumes JetStream messages, validates/transforms/masks each payload, bulk-indexes documents to Elasticsearch, persists index references, deletes terminal queue items, and acknowledges messages.
5. `POST /api/v1/logs/search` queries Elasticsearch directly. React Log Explorer calls that endpoint through `logSearchService`; it displays its response rather than local/mock log records.

## Result checklist (2026-07-29)

| Check | Result | Evidence |
| --- | --- | --- |
| Ingestion route and API-key authentication | PASS (code review) | Both public ingestion aliases use `ProductAPIKeyAuthMiddleware`; middleware validates hash, state, permission, and scope. |
| Request/payload validation | PASS (code review) | Gin binding, payload size/object validation, product/environment and payload scope checks are present. |
| Queue persistence and publish | PASS (fixed + build) | Queue items are persisted before publish and every message carries its persisted queue item ID. |
| Worker consumption | PASS (code review) | Durable JetStream pull consumer fetches messages, indexes, persists results, and ACKs only successful messages. |
| Elasticsearch indexing | PASS (code review) | Worker uses Bulk API and only persists index references after successful item responses. |
| Search API returns Elasticsearch data | PASS (code review) | Search repository calls Elasticsearch and maps `_source` documents into the API response. |
| React Log Explorer matches Search API | PASS (build + code review) | Explorer calls `/v1/logs/search` and renders returned records; `npm run build` passed. |
| Backend verification | PASS | `go test ./...` and `go build ./...` passed. |
| Live HTTP ingest -> worker -> ES -> search test | FAIL (environment) | A pre-existing local `omnilogs-api.exe` owns port 2910. Rebuilt compose API could not bind that port, so the running API could not be confirmed to include this change. PostgreSQL, NATS, Elasticsearch, and worker were running. |

## Fix made

`backend/api/internal/log_queues/usecase/usecase.go` now creates `LogQueueItem` records before publishing, and uses the database-generated item IDs in JetStream messages. This restores the queue-item linkage used by worker index/failure records without changing the public API.
| Queue-item cleanup | PASS (fixed) | Successful items and items whose retry budget is exhausted are deleted only after terminal state is persisted; retrying items remain available for redelivery. |

## Cleanup performed

Removed 51 existing `log_queue_items` that already had successful Elasticsearch index references. No retrying or failed items were removed.