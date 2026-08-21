# OmniLogs Retention Backend Implementation

วันที่ตรวจสอบ: 2026-08-18

เอกสารนี้สรุป backend implementation สำหรับ Retention Policies, historical logs, daily archive, verification, safe deletion, restore, RBAC และ audit โดยยึด architecture เดิมของ OmniLogs: Go/Gin + GORM/PostgreSQL + Elasticsearch + NATS worker.

## สถานะความครบถ้วน

| ส่วนงาน | สถานะ | หลักฐาน/หมายเหตุ |
|---|---|---|
| Retention Policy | PASS | Product + Environment scope, calendar units, create/update/list/toggle/delete, migration/backfill และ full Go test |
| Historical Logs | PASS | Elasticsearch aggregation แบบ day/week/month/year พร้อม category/feature/sub-feature filters; live Elasticsearch ยังไม่ได้รันใน environment นี้ |
| Archive Job | PASS | Durable job metadata, daily NDJSON + GZIP streaming, scope manifest และ checksum; มี streaming integration-style unit test |
| Verification | PASS | SHA-256 + GZIP/NDJSON manifest/count verification ก่อนเปลี่ยนเป็น `VERIFIED`; live stack ยังไม่ได้รัน |
| Safe Delete | PASS | Delete เฉพาะ `VERIFIED`, query scope เดียวกัน, count guard ก่อน/หลัง `DeleteByQuery` |
| Restore | PASS | `SEARCH_ONLY` และ `RESTORE_TO_ACTIVE`, `SKIP_EXISTING`/`OVERWRITE`, preserves document ID และ environment scope |
| RBAC | PASS | ใช้ global-auth เดิม เพิ่ม `RETENTION`/`ARCHIVE` permissions และ environment-specific `UserPermission` เป็น gate เพิ่มเติม |
| Audit | PASS | archive start/complete/fail/verify/delete และ restore start/complete/fail ผ่าน system audit + route audit middleware |
| Docker build | FAIL (local environment) | Docker base image ที่มีอยู่ในเครื่องมี `/usr/local/go/bin/go` ขนาด 0 bytes; Dockerfile เพิ่ม toolchain/output guard แล้ว แต่ต้อง refresh/pull base image ก่อน build จริง |

## Policy model และ calendar semantics

`models.LogRetentionPolicy` รองรับ policy ที่ผูกกับ `(product_id, environment_id)` และยังคง legacy fields เพื่อ backward compatibility:

- Active retention: `active_retention_value` + `active_retention_unit`
- Archive: `archive_enabled`, `archive_after_value/unit`
- Delete หลัง archive: `delete_active_after_archive`
- Archive retention: `archive_retention_value/unit` หรือ `archive_never_delete`
- Existing logs: `apply_to_existing_logs`

หน่วยที่รับรองคือ `DAY`, `WEEK`, `MONTH`, `YEAR` และรับ plural legacy values ได้ด้วย การคำนวณใช้ `time.AddDate` ผ่าน `internal/retention_policy/helper` จึงไม่แปลงเดือนเป็น 30 วันหรือปีเป็น 365 วัน ตัวอย่าง leap day และ calendar month มี unit tests รองรับ

Migration ทำ AutoMigrate สำหรับ fields/job ใหม่, backfill calendar fields จาก legacy policy และสร้าง partial unique index สำหรับ active policy ต่อ product/environment ใน PostgreSQL

## Historical logs และ preview

Provider Elasticsearch อยู่ที่ `internal/retention_policy/provider` และสร้าง query ที่ต้องมีทั้ง:

- `product_id`
- `environment_id`
- `@timestamp` range
- optional `category_id`, `feature_id`, `sub_feature_id`

Historical response รวม oldest/newest log, document count, estimated payload bytes, archive status และ buckets ที่ group ด้วย calendar interval `day`, `week`, `month` หรือ `year`

Routes:

```text
GET  /api/v1/products/:productId/retention-policies/historical?environment_id=...&group_by=day|week|month|year
POST /api/v1/products/:productId/retention-policies/preview
```

Preview แสดง archive/delete candidates และระบุว่า delete ต้องมี verified archive ก่อนเสมอ

## Archive pipeline

Archive ถูกแยกเป็น durable job และ daily archive records:

```text
PENDING
  -> EXPORTING
  -> COMPRESSING
  -> ARCHIVED
  -> VERIFYING
  -> VERIFIED
  -> DELETING_ACTIVE (เมื่อ policy เปิด)
  -> COMPLETED
```

ไฟล์เก็บตาม daily layout:

```text
<root>/archives/product-<product>/environment-<environment>/YYYY/MM/DD.ndjson.gz
```

`internal/archive_utils.StreamArchive` ใช้ Elasticsearch scroll ทีละ 500 documents ไม่โหลด logs ทั้งหมดเข้า memory โดยเขียน NDJSON line ต่อ line และมี manifest ต้นไฟล์/ท้ายไฟล์ที่บันทึก scope, date range, document count และ original byte count

แต่ละ `LogArchive` บันทึก:

- product/environment และ category/feature/sub-feature scope
- document count, original/compressed bytes
- format `NDJSON`, compression `GZIP`
- SHA-256 checksum
- `manifest_valid`, verification/deletion timestamps และ status

การสร้าง archive ซ้ำในวันเดียวกันจะ reuse เฉพาะ scope เดียวกัน หากพบ archive วันเดียวกันแต่ category/feature scope ต่างกันจะ fail เพื่อป้องกันการเขียนทับไฟล์ daily ที่มีอยู่

Routes:

```text
POST /api/v1/products/:productId/log-archives
GET  /api/v1/products/:productId/log-archives?environment_id=...&group_by=day|week|month|year
GET  /api/v1/products/:productId/log-archives/:archiveId?environment_id=...
GET  /api/v1/products/:productId/archive-jobs?environment_id=...
GET  /api/v1/products/:productId/archive-jobs/:jobId?environment_id=...
```

## Verification และ safe deletion

`POST /log-archives/:archiveId/verify` ทำตามลำดับนี้:

1. ตรวจ product/environment ownership และ permission
2. คำนวณ SHA-256 ของไฟล์เทียบ metadata
3. เปิด GZIP และตรวจ manifest scope/version
4. อ่าน NDJSON จนครบและตรวจ document count
5. เปลี่ยน archive เป็น `VERIFIED`
6. ถ้า policy เปิด `delete_active_after_archive` ให้เรียก scoped `DeleteByQuery` เท่านั้น

Safe delete ใช้ query เดียวกับ archive (product, environment, date range และ filters) และต้องผ่าน count guard สองชั้น:

- candidate count ใน Elasticsearch ต้องเท่ากับ archive document count ก่อนลบ
- deleted count ต้องเท่ากับ expected count หลังลบ

legacy `DeleteArchivedLogs` ถูกปิดการใช้งานเพราะ signature เดิมไม่สามารถรับรอง product/environment/date scope ได้ ส่วน legacy `push-to-archives` ถูกลดบทบาทให้ archive-only และไม่ลบ active logs โดยตรง

## Restore

`POST /log-archives/:archiveId/restore` รองรับ:

- `SEARCH_ONLY`: restore ไป index `omnilogs-archive-search-<archiveId>`
- `RESTORE_TO_ACTIVE`: restore ไป active daily index ของ product/environment
- `SKIP_EXISTING`: ใช้ bulk create และนับ conflict เป็น skipped
- `OVERWRITE`: ใช้ bulk index เพื่อเขียนทับ document เดิม

Restore บังคับ archive status เป็น `VERIFIED`, ตรวจ checksum/manifest ซ้ำ และ preserve `_id`, `_index`, product/environment scope ของเอกสาร

## RBAC และ audit

ทุก policy/archive operation รับ actor จาก request context และตรวจ:

1. environment ต้อง active และเป็นของ product ที่ระบุ
2. platform admin override ตาม global-auth เดิม
3. product role/membership permission ผ่าน `RETENTION` หรือ `ARCHIVE`
4. ถ้ามี legacy `UserPermission` row สำหรับ user/product/environment นั้น row ดังกล่าวเป็น environment-specific gate เพิ่มเติม

เพิ่ม resource types `RETENTION` และ `ARCHIVE` ใน permission validation เพื่อให้ role เดิมสามารถกำหนดสิทธิ์ CREATE/READ/UPDATE/DELETE/VERIFY/RESTORE ได้ตาม action ที่เรียกใช้

Audit actions ที่บันทึก ได้แก่:

```text
ARCHIVE_STARTED
ARCHIVE_COMPLETED
ARCHIVE_FAILED
ARCHIVE_VERIFIED
HISTORICAL_LOGS_DELETED
RESTORE_STARTED
RESTORE_COMPLETED
RESTORE_FAILED
```

## Worker integration

worker retention maintenance โหลดเฉพาะ active policy ที่เปิด archive, สร้าง daily archive, verify ให้เป็น `VERIFIED`, แล้วจึง purge expired archive files ตาม calendar retention หรือข้ามเมื่อ `archive_never_delete=true` การลบ active logs ไม่ได้อยู่ใน worker โดยตรงและไม่สามารถเกิดก่อน verification ได้

## Verification commands

คำสั่งที่ผ่าน:

```text
go test -p 1 ./...
go vet ./internal/archive_utils ./internal/retention_policy/... ./internal/log_archive/... ./internal/global_auth/...
go build ./cmd/api ./cmd/worker ./cmd/migrate
```

มี unit tests สำหรับ calendar arithmetic, Elasticsearch scroll environment guard และ streaming archive manifest/checksum verification

Docker Compose configuration ตรวจ syntax ผ่านก่อนเริ่ม build แต่ Docker build จริงใน environment นี้ยังติด base image cache ที่ corrupted (`golang:1.25-bookworm` มี Go executable ขนาด 0 bytes) Dockerfile มี explicit `test -s` guard เพื่อ fail ชัดเจนและไม่รายงาน build สำเร็จทั้งที่ไม่มี binary; หลัง refresh image แล้วควร rerun:

```text
docker compose -f compose.yml build --no-cache
docker compose -f compose.yml up -d
```

