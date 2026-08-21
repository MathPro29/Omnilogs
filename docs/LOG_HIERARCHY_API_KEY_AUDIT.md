# OmniLogs: Log Hierarchy and API Key Audit

**Audit date:** 2026-07-24 (Asia/Bangkok)  
**Scope:** current code and non-destructive runtime checks only. No product, API key, queue, Elasticsearch index, or schema was changed.

## Status vocabulary

Only these states are used in this report:

| Status | Meaning |
| --- | --- |
| PASS | Verified against the running environment or an executed automated test. |
| FAIL | Runtime evidence or code path proves the requested behaviour is not achieved. |
| MISSING | There is no implementation for the required step. |
| NOT VERIFIED | The code indicates behaviour, but this audit could not perform the required runtime check. |

## Relevant file tree

```text
testproduct/
├── index.html                                      # simulator form and default ingestion URL
└── scripts.js                                      # API-key request and optional project/category IDs

backend/api/
├── routes/
│   ├── log_queue_routes.go                         # POST /api/v1/ingest/logs
│   └── main_log_routes.go                          # GET /api/v1/logs
├── middleware/audit_auth.go                        # ProductAPIKeyAuthMiddleware
├── dto/
│   ├── api_key.go                                  # API-key request/response DTOs
│   ├── log_processing.go                           # ingest batch/item DTOs
│   └── message.go                                  # NATS LogMessage envelope
├── models/
│   ├── product.go
│   ├── product_environment.go
│   ├── project.go
│   ├── project_feature.go
│   ├── product_api_key.go
│   ├── log_queue_batch.go
│   ├── log_queue_item.go
│   ├── log_index_ref.go
│   └── log_object_storage_ref.go
└── internal/
    ├── api_keys/usecase/api_keys.go                # creates/revokes product/environment API keys
    ├── log_queues/
    │   ├── handler/handler.go                      # binds API request and applies key scope
    │   └── usecase/usecase.go                      # batch creation and NATS publish
    ├── worker/usecase/
    │   ├── usecase.go                              # JetStream pull worker
    │   ├── batch_processor.go                      # index and acknowledge flow
    │   ├── hierarchy_resolver.go                   # project/category ID/code resolution
    │   ├── validation.go                            # product/environment/project/category validation
    │   ├── log_preparer.go                         # normalization, masking, raw archive
    │   └── elasticsearch_builder.go                # Elasticsearch document/index construction
    ├── worker/repository/repository.go             # index refs and batch status
    ├── feature/usecase/
    │   ├── usecase.go                              # feature hierarchy writes
    │   └── feature_helpers.go                      # `path_ids` construction
    └── main_logs/
        ├── handler/queries.go                      # search request parsing
        ├── usecase/search.go                       # product-authorized ES search
        ├── usecase/query_builder.go                # product/environment/project/category filters
        ├── usecase/access.go                       # allowed index set and product authorization
        └── repository/repository.go                # Elasticsearch query execution

frontend/src/
├── features/log-explorer/services/log-search.service.ts # GET /v1/logs parameters
└── features/log-explorer/hooks/useLogExplorerController.ts  # UI scope selection/query key
```

## 1. Current hierarchy model

`Product` is identified by `product_id`; `ProductEnvironment` belongs to it through `product_id` ([models/product.go](../backend/api/models/product.go#L5), [models/product_environment.go](../backend/api/models/product_environment.go#L5)). `Project` carries `product_id` ([models/project.go](../backend/api/models/project.go#L3)). A feature and sub-feature are both `ProjectFeature` records: `parent_id` provides the tree, and each record carries `product_id`, `project_id`, `full_path`, `path_ids`, and `level` ([models/project_feature.go](../backend/api/models/project_feature.go#L3)).

On create/update, `rebuildFeaturePaths` sets a root to `path_ids = selfID`, and a child to `parent.path_ids + ',' + selfID` ([feature_helpers.go](../backend/api/internal/feature/usecase/feature_helpers.go#L19)). For example, a sub-feature with parent Feature A produces `feature_path_ids = "A,A.A"`; the event is stored once at the exact leaf category, not copied to its parents.

## 2. Current API-key model

| API-key property | Current implementation | Status |
| --- | --- | --- |
| Product | Required `ProductID` column; middleware sets `service_product_id`. | NOT VERIFIED at request level; code is explicit. |
| Environment | Optional `EnvironmentID` column; middleware sets `service_environment_id` only when the key is scoped. | NOT VERIFIED at request level; code is explicit. |
| Project / feature / sub-feature | No column or API-key binding. | MISSING |
| Source | No API-key binding. `source_id` can be supplied in the ingestion request. | MISSING |
| Allowed scopes / permissions | JSON `permissions` is stored, but ingestion middleware never reads or checks it. | FAIL |
| Expiration | Middleware rejects an expired key. | NOT VERIFIED at request level. |
| Revocation / inactive | Middleware requires `is_active = true`; revocation also rejects. | NOT VERIFIED at request level. |

The model contains only `product_id`, optional `environment_id`, `permissions`, `is_active`, `expires_at`, and `revoked_at` ([product_api_key.go](../backend/api/models/product_api_key.go#L8)). It does **not** contain `project_id`, `category_id`, or `source_id`.

`CreateAPIKey` validates that a selected environment belongs to the route product ([api_keys.go](../backend/api/internal/api_keys/usecase/api_keys.go#L13)), but creates no hierarchy binding. A key by itself therefore **cannot identify a Feature or Sub-feature**.

## 3. Current ingestion flow

| Step | File / route / function | Input | Output | Status |
| --- | --- | --- | --- | --- |
| Source | `testproduct/index.html`; `testproduct/scripts.js` `sendLogBatch` | Form API URL/key, optional project/category; generated logs | JSON batch | NOT VERIFIED end-to-end; key is browser-local only. |
| API route | `POST /api/v1/ingest/logs` and alias `POST /api/v1/logs/ingest` in [log_queue_routes.go](../backend/api/routes/log_queue_routes.go#L16) | HTTP JSON + API key | Calls `QueueHandler` | NOT VERIFIED direct request. |
| API-key validation | `ProductAPIKeyAuthMiddleware` in [audit_auth.go](../backend/api/middleware/audit_auth.go#L84) | `X-API-Key`, or Authorization bearer value | Context `service_product_id`, optional `service_environment_id` | NOT VERIFIED direct request. |
| Product/environment resolution | `QueueHandler` in [handler.go](../backend/api/internal/log_queues/handler/handler.go#L27) | Request `product_id`, `environment_id` and middleware context | Rejects mismatch; overwrites request with key values | NOT VERIFIED direct request. |
| Queue producer | `Enqueue` in [usecase.go](../backend/api/internal/log_queues/usecase/usecase.go#L43) | `IngestLogBatchRequest` | PostgreSQL batch row plus one JetStream message per log | PASS historically: 169 completed batches/log refs exist. |
| Worker consume | `RunOnce` / `processFetchedMessages` in [worker usecase.go](../backend/api/internal/worker/usecase/usecase.go#L83), [batch_processor.go](../backend/api/internal/worker/usecase/batch_processor.go#L13) | NATS `LogMessage` | Prepared, indexed, persisted, acknowledged message | PASS historically. |
| Hierarchy resolution | `resolveLogHierarchy` in [hierarchy_resolver.go](../backend/api/internal/worker/usecase/hierarchy_resolver.go#L21) | Payload IDs or codes | Canonical `project_id`, `category_id`, `feature_path_ids`, `feature_full_path` | PASS for existing indexed leaf data; see section 5. |
| Hierarchy validation | `validateLogHierarchy` in [validation.go](../backend/api/internal/worker/usecase/validation.go#L24) | Product/environment/project/category | Rejects mismatched or inactive project/category | PASS historically: invalid hierarchy failures are recorded. |
| Elasticsearch | `bulkIndexDocuments` / `buildElasticDocument` in [elasticsearch_builder.go](../backend/api/internal/worker/usecase/elasticsearch_builder.go#L67) | Normalized event | One ES document and one `log_index_refs` row | PASS historically and verified by `_count`. |
| Search API | `GET /api/v1/logs` in [main_log_routes.go](../backend/api/routes/main_log_routes.go#L15), handler `Search` | JWT + `product_id` and optional filters | Product-authorized ES query/results | NOT VERIFIED through HTTP UI session. |
| UI | `logSearchService.execute` and `useLogExplorerController` | Product/environment/project/category UI scope | `GET /v1/logs` request and rendered records | NOT VERIFIED interactively; frontend production build passes. |

### Product and environment source of truth

For an API-key route, Product always comes from the authenticated key. If a payload specifies a different `product_id`, `QueueHandler` returns forbidden; otherwise it overwrites `req.ProductID` with the key product ([handler.go](../backend/api/internal/log_queues/handler/handler.go#L35)).

For a key scoped to an environment, environment likewise comes from the key and a mismatching request value is forbidden ([handler.go](../backend/api/internal/log_queues/handler/handler.go#L42)). For a product-wide key with no `environment_id`, the request may choose an environment, but `Enqueue` validates it belongs to that product ([usecase.go](../backend/api/internal/log_queues/usecase/usecase.go#L63)). A request with no environment reaches the queue but worker processing fails because both product and environment are required before indexing ([log_preparer.go](../backend/api/internal/worker/usecase/log_preparer.go#L121)).

### Hierarchy source of truth

The worker accepts these payload fields, at the top level or within `metadata`:

- `project_id` or `project_code`
- `category_id`
- `sub_feature_code`, `category_code`, or `feature_code`

It resolves codes within the already authenticated `product_id`, then writes canonical IDs and the resolved category's `path_ids` into the payload ([hierarchy_resolver.go](../backend/api/internal/worker/usecase/hierarchy_resolver.go#L21)). Category validation requires `project_id`, and validates `category_id + project_id + product_id` together ([validation.go](../backend/api/internal/worker/usecase/validation.go#L69)). There is no endpoint/path routing rule and no category default from the API key.

## 4. Parent/child visibility and isolation

The log document contains top-level `product_id`, `environment_id`, optional `project_id`, `category_id`, and both `payload`/`data` copies of the dynamic event ([elasticsearch_builder.go](../backend/api/internal/worker/usecase/elasticsearch_builder.go#L140)). A category search first filters the exact `payload.category_id`, then also matches `payload.feature_path_ids` with exact/wildcard clauses ([query_builder.go](../backend/api/internal/main_logs/usecase/query_builder.go#L121)). Thus a parent category includes descendants without duplicating their documents.

Runtime `_count` checks against Elasticsearch were run against the existing Product 4 hierarchy. The stored leaf was category 5 with path `3,5` (parent category 3); all counts are from one product/environment index.

| Runtime query | Expected | Actual | Status | Evidence |
| --- | --- | ---: | --- | --- |
| Product 4 + Environment 7 | Leaf logs visible | 11 | PASS | ES `_count` |
| Product 4 + Environment 7 + Project 2 | Leaf logs visible | 11 | PASS | ES `_count` |
| Exact sub-feature 5 | 11 leaf logs | 11 | PASS | ES `_count` |
| Sibling sub-feature 6 | No leaf-5 logs | 0 | PASS | ES `_count` |
| Parent feature 3 via `feature_path_ids = 3,*` | Includes child 5 logs | 11 | PASS | ES `_count` |
| Product 4 + Environment 11 | No Environment-7 logs | 0 | PASS | ES `_count` |
| Product 8 | No Product-4 logs | 0 | PASS | ES `_count` |

Product isolation is also a mandatory top-level ES term filter ([query_builder.go](../backend/api/internal/main_logs/usecase/query_builder.go#L91)); the Search API separately authorizes the requested product before querying ([search.go](../backend/api/internal/main_logs/usecase/search.go#L13)). Environment filtering is a mandatory term only when `environment_id` is supplied ([query_builder.go](../backend/api/internal/main_logs/usecase/query_builder.go#L105)).

**Result:** parent visibility, sibling isolation, product isolation, and explicit environment isolation are PASS for the existing indexed data. They are not a new API-key end-to-end test because no existing key secret is stored in `index.html`.

## 5. Unknown JSON schema

`IngestLogItemRequest` accepts `data`, `fields`, or legacy `input_payload` as `json.RawMessage`; `Enqueue` only requires that the chosen value unmarshals to a JSON object ([log_processing.go](../backend/api/dto/log_processing.go#L8), [usecase.go](../backend/api/internal/log_queues/usecase/usecase.go#L48)). Therefore this object satisfies the request shape:

```json
{
  "status": 500,
  "endpoint": "/api/orders",
  "custom": { "a": 1 }
}
```

The worker preserves arbitrary fields under both `data` and `payload` in the Elasticsearch document ([elasticsearch_builder.go](../backend/api/internal/worker/usecase/elasticsearch_builder.go#L140)). Before masking, a successful worker run encrypts the original JSON into `log_object_storage_refs.encrypted_payload` ([log_preparer.go](../backend/api/internal/worker/usecase/log_preparer.go#L257)). For a failed event, `log_failures.error_details` receives the raw input JSON ([batch_status.go](../backend/api/internal/worker/usecase/batch_status.go#L67)).

If no project/category/code is supplied, hierarchy resolution returns without setting a hierarchy ([hierarchy_resolver.go](../backend/api/internal/worker/usecase/hierarchy_resolver.go#L39)); validation permits that as long as authenticated product and environment are present and valid ([validation.go](../backend/api/internal/worker/usecase/validation.go#L50)). Such a log is unclassified but remains searchable at Product and Environment scope; it has no category path to match feature/sub-feature filters.

| Question | Result | Status |
| --- | --- | --- |
| Does the API accept the example unknown object? | Code accepts arbitrary object values in `data`/`fields`/`input_payload`. | NOT VERIFIED by a new HTTP request. |
| Is original payload retained? | Yes on successful processing as AES-GCM encrypted archive; failure rows retain `error_details`. | PASS for successful historical events: 169 encrypted payload archive rows. |
| How is hierarchy assigned? | Explicit ID/code fields only; no endpoint/default/routing-rule assignment. | NOT VERIFIED by a new HTTP request. |
| What happens with no category? | Accepted as unclassified if product/environment are valid; visible only in non-category scopes. | NOT VERIFIED by a new HTTP request. |
| Is an explicit ES mapping/template present? | No mapping/template creation code was found in the audited ingestion path. Dynamic JSON relies on the running ES mapping. | MISSING |

## 6. `testproduct/index.html` audit

The simulator default endpoint is `http://localhost:2910/api/v1/ingest/logs` ([index.html](../testproduct/index.html#L57)). This is a valid registered API-key route. `sendLogBatch` posts JSON with `queue_key`, `source_type`, `source_platform`, `priority`, and `logs[]`; each event has `sequence_no`, `source_type`, `source_platform`, and `data` ([scripts.js](../testproduct/scripts.js#L135)). This matches `IngestLogBatchRequest` and `IngestLogItemRequest`.

It sends both `X-API-Key` and `Authorization: Bearer <key>` ([scripts.js](../testproduct/scripts.js#L179)); the backend accepts the first non-empty supported value. The simulator does **not** send `product_id` or `environment_id`: it depends on the API key. It sends project/category only when the optional form fields have values ([scripts.js](../testproduct/scripts.js#L142)).

There is no API key or product hard-coded in `index.html`. The key, endpoint override, project ID, and category ID are obtained from browser `localStorage` ([scripts.js](../testproduct/scripts.js#L52)). This audit did not read browser storage or secret values.

The simulator treats `HTTP 2xx` plus a response body as success, stores only `batch_id`, and does not poll batch status, inspect `log_failures`, search by log ID, or query Elasticsearch ([scripts.js](../testproduct/scripts.js#L192)). It is therefore **HTTP enqueue feedback only**, not an end-to-end assertion.

Runtime evidence shows recent batches for active key scope Product 8 / Environment 11 were accepted into the pipeline but all 23 permanently failed due to `category does not belong to the specified product and project`. This is consistent with stale project/category inputs being carried by the simulator, but the browser-local values themselves were not inspected.

| Test-client check | Result | Status |
| --- | --- | --- |
| Endpoint route exists | `/api/v1/ingest/logs` is registered. | PASS |
| Request DTO shape | `queue_key`, source metadata, and `logs[].data` match the backend DTO. | PASS |
| Existing key can be used from this browser | Key is not present in source and browser state was not inspected. | NOT VERIFIED |
| Simulator confirms Elasticsearch/search/UI | It has no such verification logic. | FAIL |
| Current simulator hierarchy fields match its API-key product | Recent scoped batches failed hierarchy validation. | FAIL |

## 7. Runtime health and test results

| Check | Result | Status |
| --- | --- | --- |
| API `/health/live` and `/health/ready` | HTTP 200 | PASS |
| PostgreSQL, NATS, Elasticsearch | Running; ES responded to read-only queries | PASS |
| Historical queue-to-index pipeline | 169 completed batches, 169 `INDEXED` refs, 169 encrypted raw archives | PASS |
| Queue item persistence | `log_queue_items` contains 0 rows; `Enqueue` never calls repository `CreateItems`. | FAIL |
| Recent scoped test-client-like batches | Product 8 / Environment 11: 23 permanently failed hierarchy validation | FAIL |
| Focused worker/search/document/handler Go tests | Passed | PASS |
| `go test ./...` | Passed | PASS |
| `go build ./...` | Passed | PASS |
| `frontend npm run build` | Passed | PASS |
| Browser/UI authenticated search | Browser automation unavailable in this environment; no JWT/UI session test executed. | NOT VERIFIED |

## 8. Missing components, defects, and security risks

1. **API-key permissions are not enforced — FAIL.** A key stores values such as `LOG_INGEST_CREATE`, but `ProductAPIKeyAuthMiddleware` validates hash/activity/expiration/revocation only; it does not inspect `Permissions` ([audit_auth.go](../backend/api/middleware/audit_auth.go#L84)).
2. **No API-key hierarchy/source scope — MISSING.** There is no `project_id`, `category_id`, `source_id`, or allowed-scope relation on `ProductAPIKey`. An environment-scoped key can still submit any valid project/category within its product.
3. **No active/deleted product validation at ingestion — FAIL.** Middleware reads the key alone, and worker validation checks environment/project/category but not the product record's active/non-deleted state. Runtime data contains active API keys pointing at a soft-deleted product; this permits ingestion that the normal product list/UI cannot surface.
4. **Queue-item persistence is disconnected — FAIL.** Repository methods and model exist, but `Enqueue` creates only `log_queue_batches` and publishes NATS messages; it does not call `CreateItems`. Runtime count is zero. Batch-item endpoints therefore cannot return a durable per-event queue record, and synthetic queue item IDs are used in references/failures.
5. **Missing explicit ES template/mapping — MISSING.** No index-template/mapping provisioning code was found. Unknown JSON is dynamically mapped, risking mapping conflicts/field explosion in production.
6. **No end-to-end correlation returned to test client — FAIL.** The ingest response returns batch metadata, not per-log IDs. `testproduct` cannot prove a submitted event reached ES/search.
7. **Unscoped product key can select any environment under its product — NOT VERIFIED as a live request.** This is intentional from the current code, but it must be an explicit contract decision; it is not a strict environment-bound key.

## 9. Recommended minimal fixes

1. Enforce API-key permissions for `LOG_INGEST_CREATE` before calling `QueueHandler`.
2. Decide API-key scope explicitly: retain product-wide keys only if they may select any product environment; otherwise require `environment_id`. Add optional project/category/source restrictions only if the product needs them.
3. Validate the authenticated product is active and not soft-deleted, and validate environment `is_active` as well as ownership.
4. Persist each `LogQueueItem` transactionally before publishing; make retry/failure/status endpoints read those durable rows.
5. Return accepted `log_id` values (or a polling URL carrying batch plus item IDs); update `testproduct` to poll terminal batch state and search a unique correlation ID.
6. Add an Elasticsearch index template with bounded dynamic fields (for example a controlled `flattened` area for arbitrary payload data), plus mappings for `product_id`, `environment_id`, `project_id`, `category_id`, and `feature_path_ids`.
7. Update the simulator to clear stale project/category values when its API key/product changes, and require users to select hierarchy values belonging to the key product.

## 10. API-key to future SDK/library path

An SDK does not require a redesign of the core transport: it can send the same API key and canonical `data` payload to the same endpoint. The SDK should add client-side configuration for product-scoped API key, environment, project code/ID, and category code/ID; prefer stable codes over database IDs. The backend changes above remain necessary because client-side validation is not a security boundary.

The minimum stable contract for an SDK is:

```text
API key -> authenticated product + optional fixed environment
SDK config/payload -> project_code + category_code (or IDs)
backend -> resolve within authenticated product, validate full hierarchy,
           write one document with product/environment/project/category/path IDs
search -> mandatory product filter; optional environment/project/category/path filters
```

## Final answers

1. **Can API key start log reception?** **PASS for authentication/enqueue historically; FAIL as a complete current test-client flow.** The system has accepted batches with API-key scopes, but current Product 8 / Environment 11 batches permanently fail at hierarchy validation.
2. **Is Product resolution correct?** **NOT VERIFIED by a direct new request; code rejects payload product mismatch and overwrites it from the key.** Product active/non-deleted validation is missing.
3. **Is Environment resolution correct?** **NOT VERIFIED by a direct new request.** A scoped key enforces its environment; an unscoped key accepts any environment belonging to its product. Active-environment validation is missing.
4. **Can logs flow through Product → Project → Feature → Sub-feature?** **PASS for existing indexed data.** Worker resolves/validates hierarchy and writes leaf path IDs; current simulator configuration failed this validation.
5. **Does parent scope work?** **PASS.** Runtime ES counts proved parent category 3 returns child category 5 logs while sibling 6 returns none.
6. **Do logs leak across Product?** **PASS for the tested ES data.** Search adds mandatory `product_id`; API-key handler rejects payload product mismatch. Product deletion validation remains a separate integrity risk.
7. **Do logs leak across Environment?** **PASS for the tested ES data when environment is filtered.** Scoped keys reject mismatch; product-wide keys are intentionally able to choose any valid environment of their product.
8. **Is unknown schema supported?** **NOT VERIFIED by new HTTP request.** Code accepts arbitrary JSON objects, persists successful raw payloads encrypted, and stores them dynamically in ES.
9. **What is missing now?** Permission enforcement, API-key hierarchy/source scope where required, active/deleted scope checks, durable queue items, explicit ES mapping/template, and end-to-end test-client verification.
10. **Does an SDK require a backend rewrite?** **No.** Keep the API-key transport and canonical payload; add stable project/category resolution and the minimal server-side protections above.
