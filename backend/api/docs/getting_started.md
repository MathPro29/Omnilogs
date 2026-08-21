# Production connection guide

1. Create Product — **PASS**
2. Create Project — **PASS**
3. Create Feature — **PASS**
4. Create Sub Feature — **PASS**
5. Create Environment — **PASS**
6. Generate an API Key bound to that Environment with `LOG_INGEST_CREATE` — **PASS**
7. Send a sample payload — **PASS**

```bash
curl --request POST "$OMNILOGS_URL/api/v1/ingest/logs" \
  --header "X-API-Key: $OMNILOGS_API_KEY" \
  --header "Content-Type: application/json" \
  --data '{"queue_key":"orders","source_type":"SERVICE","source_platform":"HTTP","logs":[{"sequence_no":1,"source_type":"SERVICE","source_platform":"HTTP","data":{"timestamp":"2026-07-29T12:00:00Z","level":"INFO","service":"orders","message":"order created","trace_id":"trace-1","request_id":"req-1","project_code":"orders","feature_code":"create"}}]}'
```

8. Check Queue — batch response is `201`; status endpoint is `/api/v1/ingest/batches/{batchId}`. **PASS**
9. Check Worker — worker consumes the JetStream message; retry only transient failures. **PASS**
10. Check Elasticsearch — successful item creates an index reference/document. **PASS**
11. Open Search API — call `POST /api/v1/logs/search` as an authorized user. **PASS**
12. Open Log Explorer — select the same Product/Environment and search the message/trace/request ID. **PASS**