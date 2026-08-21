# Log hierarchy API-key ingestion

#compose worker : docker compose -f backend/compose.yml up -d --build worker

## Flow

`POST /api/v1/ingest/logs` authenticates `X-API-Key`, requires `LOG_INGEST_CREATE`, and fixes `product_id` and `environment_id` from that key. A request or payload that supplies a different product/environment is rejected. Ingestion keys must be bound to an environment.

The worker resolves `project_id`/`project_code` and `category_id`/`category_code` in the key product. A category resolves its project and must belong to the same product/project. Valid category logs are `CLASSIFIED`; payloads with no hierarchy are retained as `UNCLASSIFIED`.

If an exact hierarchy is not supplied or no longer exists, the worker can
automatically match active features in the authenticated product from stable
route and hierarchy hints such as `route_key`, `request_path`,
`route_pattern`, `feature_name`, and routing dimensions. Matching is
deterministic: normalized exact matches win, followed by contained route tokens
and small spelling differences. Tied or low-confidence results are never
guessed; the log remains at Product + Environment with
`routing_method=PRODUCT_ENVIRONMENT_DEFAULT`.

Elasticsearch documents contain `routing_status` and `routing_method` (`EXPLICIT_ID`, `EXPLICIT_CODE`, or `NONE`). Category path IDs preserve feature-to-sub-feature parent searches while product and environment are top-level filters.

## Status polling

After enqueueing, poll `GET /api/v1/ingest/batches/{batch_id}` with the same API key. The endpoint is restricted to that key's product/environment. Terminal batch states are `INDEXED` and `FAILED`.

## Test-page use

Open `testproduct/index.html`, enter an environment-bound API key, optionally enter Project/Category IDs or codes, and connect. The page shows the returned `batch_id`, inferred routing state, and polls to `INDEXED` or `FAILED`. Changing the API key clears hierarchy fields.

## Verification

- PASS: focused Go package tests for middleware, worker usecase, queue packages, route package, and search query usecase.
- BLOCKED: full PostgreSQL/NATS/Elasticsearch integration cases were not run because the local stack was not started in this session.
- NOT IMPLEMENTED: automatic routing, routing-rule storage, source-default hierarchy, SDK/library, and database schema changes.
