# API connection checklist

1. Generate API Key — create a key bound to one Product and Environment with `LOG_INGEST_CREATE`. **PASS**
2. Call API — `POST /api/v1/ingest/logs` (or compatible `/api/v1/logs/ingest`) using `X-API-Key: omni_keys_...`. **PASS**
3. Authentication — validates prefix/hash, active state, revocation, expiry, and permission. **PASS**
4. Queue — request Product/Environment/Source must match the key scope; valid payloads are persisted and published to JetStream. **PASS**
5. Worker — pull-consumes messages, validates/transforms/masks, and retries transient failures. **PASS**
6. Elasticsearch — successful documents are bulk indexed. **PASS**
7. Search API — `POST /api/v1/logs/search` reads Elasticsearch with product access enforcement. **PASS**
8. Log Explorer — calls the Search API and renders its returned documents. **PASS**

Rejected connections: missing/invalid/inactive/revoked/expired keys return 401; a key without `LOG_INGEST_CREATE` or with a mismatched Product/Environment/Source returns 403.