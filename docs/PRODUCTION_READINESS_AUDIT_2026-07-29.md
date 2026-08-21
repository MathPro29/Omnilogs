# Executive Summary

OmniLogs **ยังไม่พร้อม Production** แม้เส้นทางหลัก `API -> NATS JetStream -> Worker -> Elasticsearch -> Search API -> React Log Explorer` มี implementation จริงและมีหลักฐาน runtime ว่าเคยประมวลผลสำเร็จครบ 827 logs

หลักฐานที่ผ่าน:

- `go test ./...` ผ่านทุก package ที่มี test
- `go build ./...` ผ่าน
- `go run ./cmd/openapi validate` ผ่าน
- `npm run build` ผ่าน
- `/health/live` และ `/health/ready` ตอบ 200
- Runtime ณ 2026-07-29: PostgreSQL มี 827 batches สถานะ `COMPLETED`, 827 index refs สถานะ `INDEXED`, Elasticsearch มีเอกสารรวม 827 รายการ และ NATS consumer มี `num_pending=0`, `num_ack_pending=0`
- Ingestion ที่ไม่มี API key ถูกปฏิเสธด้วย 401 และตอบ `X-Request-ID`

เหตุผลที่ยังไม่พร้อม:

1. Deployment config เป็น development และฝัง DB password, JWT secret และ encryption key ใน `backend/compose.yml:2-9`
2. PostgreSQL, Elasticsearch และ NATS เปิดพอร์ตสู่ host; Elasticsearch ปิด security/TLS ที่ `backend/compose.yml:93-127`
3. ไม่มี HTTPS/reverse proxy และไม่มี frontend production service
4. API container ใน Compose อยู่สถานะ `Created`; API ที่ตอบ port 2910 เป็น process นอก Compose จึงยังพิสูจน์ deployment topology เดียวกันไม่ได้
5. Ingestion สร้าง batch ใน PostgreSQLก่อน publish NATS ทีละ message โดยไม่มี transaction/outbox; publish กลาง batch ล้มเหลวทำให้ batch `FAILED` และ retry ด้วย idempotency key เดิมไม่ republish (`backend/api/internal/log_queues/usecase/usecase.go:85-90,122-158`)
6. `idempotency_key` ไม่มี unique constraint (`backend/api/models/log_queue_batch.go:15`)
7. Queue item inspection เป็น stub: `GetItemsByBatch` คืน array ว่าง และ `GetItemByID` คืน not found เสมอ (`backend/api/internal/log_queues/usecase/usecase.go:167-179`)
8. ไม่มี Dead Letter Queue แยก; permanent failures ถูกเก็บใน PostgreSQL แล้ว Ack NATS (`backend/api/internal/worker/usecase/batch_status.go:67-117`)
9. Payload limit ใช้ต่อ log แต่ไม่มี HTTP body/batch-count limit; `UploadRateLimit` ถูกสร้างแต่ไม่ผูกกับ ingestion route
10. Request timeout ไม่ถูกผูกกับ ingestion route; HTTP serverไม่มี `ReadTimeout`, `WriteTimeout`, `IdleTimeout`, `MaxHeaderBytes`
11. Fresh authenticated E2E และ browser verification ของ Log Explorer ยังทำไม่ได้โดยไม่ใช้/เปิดเผย credential จริง จึงเป็น `NEEDS TEST`
12. Frontend lint ไม่ผ่านหนึ่งจุด และไม่มี frontend test file

# Production Readiness Score

| Area | Status | Score | Evidence |
|---|---:|---:|---|
| API ingestion | ⚠️ PARTIAL | 52% | Auth/scope/validation/request ID มีจริง แต่ body limit, ingestion timeout, compression, public CORS และ HTTPS ไม่พร้อม |
| Queue | ⚠️ PARTIAL | 48% | Publish/consume/Ack/Nak/retry มีจริง แต่ไม่มี outbox, unique idempotency หรือ DLQ และ inspection เป็น stub |
| Elasticsearch | ⚠️ PARTIAL | 61% | Controlled mapping, bulk, search/filter/sort/pagination มีจริง; alias/nested/highlight ไม่มี และ aggregation ไม่อยู่ใน Log Search |
| Worker | ⚠️ PARTIAL | 67% | Concurrent preparation, masking, bulk, retry และ failure record มีจริง; reconciliation/health/DLQ ยังขาด |
| Log Explorer | ⚠️ PARTIAL | 64% | React ต่อ Search V2 จริงและแสดง raw/custom/metadata ได้; browser E2E, explicit sort และ field completeness ยังไม่พิสูจน์ |
| Security | ❌ FAIL | 24% | Dev secrets, plaintext DB, unauthenticated Elasticsearch/NATS และไม่มี TLS termination |
| Deployment/Operations | ❌ FAIL | 27% | Compose มี core services แต่ไม่มี frontend, API containerไม่รัน, ไม่มี restart/healthcheck สำหรับ API/worker |
| QA | ⚠️ PARTIAL | 58% | Backend test/build/OpenAPI และ frontend build ผ่าน; lint fail, frontend tests = 0, fresh E2E ยังไม่รัน |
| **Overall** | **❌ FAIL** | **50%** | มี functional core แต่มี security, delivery guarantee และ deployment blockers |

คะแนนเป็น checklist-based audit score ของสถานะที่ตรวจได้ ไม่ใช่ SLA หรือ load-test result

## API Ingestion Audit

| Check | Status | Evidence |
|---|---|---|
| API usable | ⚠️ PARTIAL | Routes `/api/v1/ingest/logs` และ `/api/v1/logs/ingest` มีจริง (`routes/log_queue_routes.go:29-32`); runtime API ตอบ health แต่ Compose API ไม่ได้รัน |
| Authentication/API key | ✅ PASS | SHA-256 lookup, active, expiry, revoked, permission และ scope checks (`middleware/audit_auth.go:85-162`) |
| Validation | ⚠️ PARTIAL | Gin binding + JSON object/depth/field/string/array limits (`dto/log_processing.go:8-31`, `log_queues/usecase/usecase.go:50-66`); ไม่มี max logs/batch |
| Error handling | ⚠️ PARTIAL | 400/401/403/413/500 mapping มีจริง; internal side-effect atomicity ยังไม่ปลอดภัย |
| Rate limit | ⚠️ PARTIAL | Global in-memory per-IP limiterมีจริง (`middleware/rate_limit.go:22-100`); ไม่ shared ระหว่าง replicas และ upload limiterไม่ถูกใช้ |
| Payload size | ⚠️ PARTIAL | default 256 KiB ต่อ log (`configs/env.go:123-127`); ไม่มี `MaxBytesReader` หรือ total batch limit |
| Timeout | ❌ FAIL | Config มี แต่ ingestion routeไม่ผูก `RequestTimeout`; serverมีเฉพาะ `ReadHeaderTimeout` (`bootstrap/api.go:43-47`) |
| Compression | ❌ FAIL / NOT IMPLEMENTED | ไม่พบ request gzip decompression หรือ response compression middleware |
| CORS | ⚠️ PARTIAL | รับ localhost/private IP เท่านั้น (`bootstrap/api.go:82-117`); production public origin ใช้ไม่ได้และไม่มี explicit allowlist config |
| HTTPS readiness | ❌ FAIL | API ใช้ `ListenAndServe`, compose ไม่มี TLS proxy; external servicesใช้ plaintext |
| Health check | ⚠️ PARTIAL | live/ready มีและ ready ตรวจ PostgreSQL + Elasticsearch; ไม่ตรวจ NATS และ error code ระบุ DB แม้ ES ล้ม (`routes/health_routes.go`) |
| Request ID | ✅ PASS | รับ/สร้าง `X-Request-ID` และตอบกลับ (`middleware/request_id.go:13-23`) |
| Trace ID | ⚠️ PARTIAL | Audit/Search อ่าน `X-Trace-ID`/`Traceparent`; ไม่มี trace context propagation ไป NATS/worker |
| Logging | ⚠️ PARTIAL | Gin, performance, recovery, slog และ audit มี; ไม่มี centralized telemetry/metrics proof |
| Retry | ⚠️ PARTIAL | Worker exponential delayed NAK สูงสุด default 3 retries (`worker/usecase/batch_status.go:67-117`); enqueue publish ไม่มี safe retry/outbox |

## Flow การรับ Log

| Stage | Status | Evidence |
|---|---|---|
| Client -> API | ⚠️ PARTIAL | React/API guide มี route จริง; fresh credentialed request ยัง `NEEDS TEST` |
| API key auth | ✅ PASS | Middleware บังคับ environment-bound key ก่อน ingestion |
| Validation | ✅ PASS | Binding, scope, hierarchy และ payload structural validation มีจริง |
| PostgreSQL batch -> NATS | ⚠️ PARTIAL | ทำจริง แต่ไม่ atomic และไม่มี outbox |
| NATS -> Worker | ✅ PASS | Durable pull consumer + explicit Ack (`configs/nats.go`, `worker/usecase/usecase.go:83-106`) |
| Processing | ✅ PASS | hierarchy, transform, validate, masking, controlled document preparation (`worker/usecase/log_preparer.go:115-257`) |
| Elasticsearch bulk | ✅ PASS | `_bulk` ใช้ deterministic `log_id` เป็น document ID (`worker/usecase/elasticsearch_builder.go:67-140`) |
| Search API | ✅ PASS (code) | Authz + ES query + pagination metadata (`main_logs/usecase/search.go`, `main_logs/handler/queries.go:83-135`) |
| React Log Explorer | ⚠️ PARTIAL | เรียก `/v1/logs/search` จริงและ infinite pagination (`frontend/src/features/log-explorer/services/log-search.service.ts:52-65`, `useLogExplorerController.ts:178-200`) |

จุดที่พบ:

- **TODO:** `frontend/src/routes/routes.tsx:50`
- **Stub:** queue consume usecase, item list และ item detail (`log_queues/usecase/usecase.go:163-179`)
- **Dead/legacy risk:** มี `.bak`, `logs-explorer-legacy` และ scratch commands จำนวนมาก; ไม่อยู่ใน supported runtime แต่ควรไม่ถูกส่งเข้า production image
- **Hardcode:** Compose secrets และ dev mode
- **Mock fallback:** Auth/User/Dashboard ใช้ mock เมื่อไม่มี `VITE_API_BASE_URL`; dashboard ยังแสดง “Live mockup”

## Elasticsearch Audit

| Capability | Status | Evidence |
|---|---|---|
| Index creation | ✅ PASS | สร้าง daily `search-v3` index ตาม product/environment |
| Alias | ❌ FAIL / NOT IMPLEMENTED | Model มี `WriteAlias` แต่ worker/search ใช้ prefix/pattern โดยตรง (`elastic_index_policy.go:11`, `main_logs/usecase/access.go:29-50`) |
| Mapping | ✅ PASS | Runtime index เป็น `dynamic:false`; controlled standard fields + flattened `data` |
| Dynamic fields | ⚠️ PARTIAL | arbitrary JSON เก็บใน `_source`; searchable ผ่าน `data` แบบ `flattened`, ไม่ใช่ typed dynamic mapping |
| Nested | ❌ FAIL / NOT IMPLEMENTED | ไม่มี nested mapping/query |
| Bulk insert | ✅ PASS | NDJSON bulk พร้อมตรวจ per-item status |
| Search | ✅ PASS (code/runtime data) | Search V1/V2 และ runtime indexed docs มีจริง |
| Pagination | ⚠️ PARTIAL | `from/size`, max page size 100; ไม่มี `search_after`, deep page อาจชน result-window |
| Sorting | ⚠️ PARTIAL | รองรับ sort field; V2 ไม่ validate sort field/order และยังไม่ runtime-tested |
| Filtering | ⚠️ PARTIAL | scope/time/custom/global filters มี; legacy GET บาง filter ยัง query `payload.*` ทั้งที่ payload mapping disabled |
| Highlight | ❌ FAIL / NOT IMPLEMENTED | UI highlight เป็น client-side; ES queryไม่มี `highlight` |
| Aggregation | ⚠️ PARTIAL | Dashboard repositoryมี aggregation แต่ Log Search/Explorer responseไม่มี aggregation |

Runtime Elasticsearch เป็น `yellow`, 1 node, 9 primary shards และ 9 unassigned replica shards สอดคล้องกับ single-node dev config แต่ไม่ใช่ production HA

## Queue Audit

| Capability | Status | Evidence |
|---|---|---|
| Publish | ⚠️ PARTIAL | JetStream publish ทำจริง แต่ DB/NATS ไม่ atomic |
| Consume | ✅ PASS | Durable pull consumer, batch fetch |
| Retry | ⚠️ PARTIAL | delayed NAK + exponential delay;ไม่มี configurable backoff policy |
| Dead Letter | ❌ FAIL / NOT IMPLEMENTED | permanent failure Ack แล้วเก็บ PostgreSQL failure เท่านั้น |
| Ack | ✅ PASS | Ack หลัง Elasticsearch และ metadata persistence สำเร็จ |
| Nack | ✅ PASS | `NakWithDelay` เมื่อ retryable |
| Reconnect | ⚠️ PARTIAL / UNVERIFIED | NATS clientมี default reconnect behavior แต่ไม่มี integration/fault test |
| Duplicate | ⚠️ PARTIAL | ES `_id=log_id` ลด duplicate ต่อ message; publish ซ้ำคนละ batchยังได้คนละ ID |
| Idempotency | ❌ FAIL | check-then-create ไม่มี unique DB constraint และ failed existing batchไม่ republish |

# Critical Issues

| Priority | Issue | Impact | Fix Required |
|---:|---|---|---|
| P0 | Dev secrets + unauthenticated plaintext infrastructure | Credential compromise, data exposure, unauthorized index/queue/DB access | External secret manager; rotate all shown values; TLS/auth; private network; remove host port exposure |
| P0 | No TLS/frontend production gateway/CORS allowlist | Real browser/client deploymentเชื่อมไม่ได้อย่างปลอดภัย | Deploy reverse proxy/load balancer with HTTPS; configure exact origins |
| P0 | DB batch creation and NATS publish not atomic | Partial batch/data loss; failed idempotent request cannot recover | Transactional outbox + relay, or persist queue items/outbox then publish/reconcile |
| P0 | Idempotency key not unique | Concurrent duplicate batches | Scoped unique constraint, conflict-safe insert, request fingerprint |
| P1 | No total HTTP body/max batch limit and no ingestion timeout | Memory/CPU/resource exhaustion | `MaxBytesReader`, max logs/batch, ingestion timeout, server timeouts |
| P1 | No DLQ/replay workflow | Permanent failures remain operationally stranded | Dedicated DLQ or explicit retry topic + admin replay/audit |
| P1 | Queue inspection endpoints are stubs | Operators cannot inspect item-level state | Implement repository-backed getters or remove misleading routes |
| P1 | Compose API not running, no restart/healthcheck | Deployment cannot self-heal or prove topology | Fix service lifecycle; add restart policy and API/worker health checks |
| P1 | Fresh E2E + Explorer browser test absent | Cannot prove current release from ingest through UI | Add isolated E2E fixture/credential and automated test |
| P2 | Frontend lint failure/no tests/large bundle | QA gate fail and slower client | Fix explicit `any`, add tests, route/code splitting |

# Missing Features

## Gap Analysis

| Feature | Current | Expected | Status | Impact | Priority | Fix Required |
|---|---|---|---|---|---:|---|
| API Authentication | API key hash/scope/expiry/revoke/permission | Same + TLS | ⚠️ PARTIAL | Key exposed in transit without TLS gateway | P0 | TLS |
| Atomic ingestion | DB then per-message NATS publish | Durable atomic acceptance | ❌ FAIL | Partial loss | P0 | Outbox |
| Idempotency | Non-unique lookup | Scoped unique + atomic replay behavior | ❌ FAIL | Duplicates/stuck failed batch | P0 | Schema + usecase |
| Payload protection | Per-log structural limits | HTTP total limit + batch count | ⚠️ PARTIAL | DoS | P1 | Middleware/DTO |
| DLQ | PostgreSQL failure rows | Operational DLQ and replay | ❌ FAIL | Failed logs stranded | P1 | Queue design |
| Queue inspection | Routes exist, usecase stub | Real batch/item results | ❌ FAIL | False operational surface | P1 | Implement |
| Alias/rollover | Stored fields only | Applied write/read alias or ILM | ❌ FAIL | Unsafe index evolution | P1 | ES lifecycle |
| Realtime update | Live Tail NATS Core broadcast | Tested scoped streaming | ⚠️ PARTIAL | Runtime behavior unverified | P1 | Browser/integration test |
| Explorer completeness | Raw JSON, summary, sections, custom fields | Every payload field + verified filters/sort | ⚠️ PARTIAL | Missing/mis-grouped fields possible | P1 | Contract/E2E tests |
| Highlight | Client-side text highlight | Server-side relevant highlighting | ❌ FAIL | Limited search UX | P2 | Optional ES highlight |
| Deployment | Development Compose | HA, secure, observable stack + frontend | ❌ FAIL | Cannot launch production safely | P0 | Production manifests |
| Frontend QA | Build passes, lint fails, 0 tests | Clean lint + automated tests | ❌ FAIL | Regression risk | P1 | QA suite |

# Checklist

## Functional Setup

- [ ] Deploy production-secured PostgreSQL
- [ ] Deploy production-secured Elasticsearch cluster
- [ ] Deploy production-secured NATS JetStream
- [ ] Deploy migration job and verify clean/upgrade migration
- [ ] Deploy API with healthcheck/restart policy
- [ ] Deploy worker with healthcheck/restart policy
- [ ] Deploy frontend behind HTTPS gateway
- [ ] Configure exact CORS origins
- [ ] Seed system roles/admin through controlled one-time job
- [ ] Create Product
- [ ] Create Project
- [ ] Create Feature
- [ ] Create Sub Feature
- [ ] Create Environment
- [ ] Generate environment-bound API key with `LOG_INGEST_CREATE`

## E2E Acceptance

- [ ] Send a uniquely identified sample payload
- [ ] Verify API returns 201, batch ID and request ID
- [ ] Poll batch status to `COMPLETED`
- [ ] Verify NATS pending/ack-pending returns to baseline
- [ ] Verify worker has no failure row for sample
- [ ] Verify Elasticsearch document exists by generated log ID
- [ ] Verify Search API returns the sample under exact product/environment scope
- [ ] Open Log Explorer and verify sample appears
- [ ] Verify Timestamp, Environment, Project, Feature/Sub Feature, Level, Trace, Request, Error, Latency, User and IP when present in sample
- [ ] Verify Header, Body, Metadata, Custom fields, JSON viewer and masked raw payload
- [ ] Verify filter, sort and pagination against known fixtures
- [ ] Verify unauthorized, expired, revoked, wrong-product and wrong-environment keys
- [ ] Verify duplicate idempotency requests under concurrency
- [ ] Verify retry, NATS reconnect, Elasticsearch outage and DLQ/replay

## Security/Operations

- [ ] Rotate hard-coded development credentials before any external exposure
- [ ] Remove direct public DB/ES/NATS ports
- [ ] Enable TLS and authentication for all service-to-service traffic
- [ ] Store secrets outside images/repository/manifests
- [ ] Verify sensitive masking and audit access end-to-end
- [ ] Add metrics/alerts for ingest rate, failures, lag, retry, ES health, DB pool and latency
- [ ] Add backup/restore and disaster-recovery test
- [ ] Run load, soak, failure and recovery tests
- [ ] Make lint/test/build/OpenAPI/E2E mandatory CI gates

# Deployment Guide

> ขั้นตอนนี้เป็น target guide; ห้ามใช้ `backend/compose.yml` ปัจจุบันเป็น production manifest

1. Provision PostgreSQL, Elasticsearch และ NATS ใน private network พร้อม TLS/auth/backup
2. สร้าง production secrets ใหม่สำหรับ DB, JWT และ `DATA_ENCRYPTION_KEY`; ห้าม reuse ค่าใน Compose
3. Build immutable API, worker, migrate และ frontend artifacts
4. Run `omnilogs-migrate` เป็น one-time job และยืนยัน exit 0
5. Run controlled seed commandเพื่อสร้าง system roles/initial admin; เก็บ credential ใน secret manager
6. Deploy worker แล้วตรวจ NATS consumer/DB/ES connectivity
7. Deploy API พร้อม `/health/live`, `/health/ready`, restart policy และ resource limits
8. Deploy frontend/reverse proxy พร้อม HTTPS และ exact CORS allowlist
9. สร้าง Product -> Project -> Feature -> Sub Feature -> Environment
10. Generate environment-bound API key และเก็บ raw keyครั้งเดียวใน secret manager
11. ส่ง sample log ที่มี unique marker
12. ตรวจ batch -> NATS -> worker -> Elasticsearch -> Search API -> Log Explorer ตาม E2E checklist
13. เปิด monitoring/alerts แล้วจึงรับ traffic จริงแบบ gradual rollout

# End-to-End Verification

| Verification | Result |
|---|---|
| Backend unit/package tests | ✅ PASS — `go test ./...` |
| Backend compile | ✅ PASS — `go build ./...` |
| OpenAPI | ✅ PASS — validation passed |
| Frontend production build | ✅ PASS — TypeScript + Vite |
| Frontend lint | ❌ FAIL — `ProductLogRoutingPanel.tsx:15:294`, `no-explicit-any` |
| Frontend automated tests | ❌ NOT IMPLEMENTED — 0 test/spec files |
| API live/readiness | ✅ PASS ณเวลาตรวจ |
| Unauthenticated ingestion rejection | ✅ PASS — 401 + `X-Request-ID` |
| Runtime historical pipeline consistency | ✅ PASS — DB batches 827 = DB refs 827 = ES docs 827 |
| NATS current lag | ✅ PASS ณเวลาตรวจ — pending 0, ack-pending 0, redelivered 0 |
| Fresh authenticated ingest -> search | ⚠️ NEEDS TEST — ไม่มี safe test API key ในขอบเขต audit |
| Search -> React browser rendering | ⚠️ NEEDS TEST — ไม่ได้ใช้ credential/session จริง |
| Failure/recovery E2E | ⚠️ NEEDS TEST |
| Performance/load/soak | ⚠️ NEEDS TEST |

หลักฐาน runtime ข้างต้นพิสูจน์ว่า flow เคยทำงานจริง แต่ไม่พิสูจน์ว่า release/configuration ปัจจุบันผ่าน fresh E2E เพราะ:

- Compose API containerไม่รัน
- API port 2910 มาจาก process นอก Compose
- ไม่ใช้หรือเปิดเผย raw API key/JWT จาก environment จริง
- ไม่สร้าง privileged test identity หรือแก้ข้อมูล production-like โดยไม่ได้รับอนุญาต

# Performance Improvements (Safe Only)

1. เพิ่ม frontend route-level code splitting; bundle ปัจจุบัน 1,951.42 kB (gzip 611.57 kB)
2. ผูก max batch count และ total request-body limitก่อน JSON decode เพื่อลด allocation
3. ใช้ `search_after` + stable tie-breaker สำหรับ deep pagination แทน `from`
4. จำกัด/validate Dynamic Search V2 field, operator, sort order, wildcard length และ page depth
5. ใช้ transactional outbox และ batch publisher เพื่อลด partial publish/reconciliation cost
6. ใช้ shared/distributed rate limiter เมื่อมีหลาย API replicas; in-memory map ปัจจุบันโตตาม client IP และไม่ consistent ข้าม replica
7. ตั้ง shard/replica/index template ให้สอดคล้อง cluster topology; single-node dev ปัจจุบัน yellow เพราะ replica unassigned
8. เพิ่ม metrics ก่อน optimize concurrency; worker ใช้ `2 * GOMAXPROCS` สำหรับ preparation และ bulk fetch 100 แต่ยังไม่มี load-test evidence

ไม่เสนอ micro-refactor ของ repository/service/usecase ที่ไม่มี benchmark เพราะยังไม่มีหลักฐานว่าจะเพิ่ม performance

# High Risk Changes (Skipped)

## Transactional outbox migration

- Impact: เปลี่ยน ingestion durability, idempotency, worker delivery และ schema
- Risk: สูง; อาจเปลี่ยน API acceptance semantics และเกิด duplicate/loss ระหว่าง migration
- Affected modules: `log_queues`, NATS config, worker, migrations, batch status
- Reason: ต้องออกแบบ compatibility/reconciliation และทดสอบ outage/concurrency
- Alternative: เพิ่ม outbox แบบ backward-compatible, dual-read transition และ migration runbook
- **Skipped due to high impact.**

## Elasticsearch alias/ILM migration

- Impact: index naming, retention, reads, writes และ archive
- Risk: สูง; mapping/alias ผิดอาจทำให้ search ขาดข้อมูลหรือเขียนผิด index
- Affected modules: worker index resolver, index policy repository, search index resolver, archive/retention
- Alternative: สร้าง template/alias ใหม่, backfill, shadow query แล้ว cut over
- **Skipped due to high impact.**

## Security/network topology replacement

- Impact: ทุก service และ deployment environment
- Risk: สูง; credential rotation/TLS rolloutผิดลำดับทำให้ outage
- Alternative: staged secret rotation, dual certificates, private endpoints และ canary
- **Skipped due to high impact.**

# Final Verdict

**NOT READY FOR PRODUCTION**

เหตุผลสั้น: core ingestion/search มีจริงและมี runtime evidence แต่ยังมี P0 ด้าน secrets/TLS/network, DB-to-NATS atomicity/idempotency, deployment topology และยังไม่มี fresh authenticated E2E ถึง Log Explorer
