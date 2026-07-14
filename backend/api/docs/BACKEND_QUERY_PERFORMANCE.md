# Backend Query Performance Review

วันที่ตรวจสอบ: 2026-07-14

เอกสารนี้บันทึกผล audit และการปรับปรุงชุดแรก โดยเน้นจุดที่มีผลกระทบสูงและความเสี่ยงต่ำ ไม่เปลี่ยนโครงสร้าง response ของ request ที่สำเร็จ

## 1. ปัญหาที่ตรวจพบก่อนแก้ไข

| จุดที่พบ | ปัญหา | สาเหตุ | ผลกระทบ | แนวทาง | ระดับ |
|---|---|---|---|---|---|
| Main Log authorization | ใช้ `COUNT(*)` เพื่อตรวจว่ามี membership หรือไม่ | Query ต้องนับทั้งชุดทั้งที่ต้องการ boolean | เพิ่ม DB work ทุกการค้นหาและดูรายละเอียด Log | ใช้ `SELECT membership_id LIMIT 1` | สูง |
| Main Log และ Dashboard index resolution | อ่าน policy ทุก column ทุก request และ Dashboard ไม่ส่ง context | ใช้ model query แบบเต็มและไม่มี cancellation | DB round-trip ใช้ memory เกินจำเป็น และยกเลิกไม่ได้ | `SELECT index_prefix`, `WithContext`, dedupe pattern | สูง |
| Dashboard pagination | รับ limit/offset โดยไม่มี defensive normalization ใน usecase | พึ่งค่า binding จาก handler เพียงชั้นเดียว | ดึงข้อมูลมากเกินไปหรือใช้ deep offset ที่แพง | จำกัด limit สูงสุด 100 และ offset ขั้นต่ำ 0 | สูง |
| Dashboard Elasticsearch query | เงื่อนไข exact/range อยู่ใน `bool.must` | Elasticsearch คำนวณ scoring โดยไม่จำเป็น | CPU เพิ่มและ filter cache ใช้ได้ไม่เต็มที่ | ย้าย product/environment/project/category/time ไป `bool.filter` | สูง |
| Category filter | สร้าง wildcard หลาย clause ต่อ category และมี leading wildcard | Path IDs ถูกเก็บเป็น comma-separated string | Query cost โตตาม category และจำนวน document | จำกัด/ตัด ID ซ้ำสูงสุด 100; schema แบบ array เป็นงานถัดไป | สูง |
| Dynamic Custom Field | Field path จาก request ถูกนำไปประกอบชื่อ field | ไม่มี allowlist ของ segment | Query injection หรือ query ผิด field | validate segment และ normalize array path | สูง |
| Favorite reorder | Update ทีละ field ใน transaction | Query อยู่ใน loop | N database round-trips | Atomic `UPDATE ... CASE` หนึ่ง statement | สูง |
| Favorite field key | `COUNT` ซ้ำจนกว่าจะเจอ suffix ว่าง | Query อยู่ใน loop | จำนวน query โตตาม key ซ้ำ | Pluck key ครั้งเดียวแล้วหา suffix ใน memory | กลาง |
| Pending queue | ไม่มี index ตรงกับ `status='QUEUED' ORDER BY priority, received_at` | Index เดิมไม่รองรับ filter/sort นี้ | Scan และ sort โตเมื่อ queue สะสม | Partial index เฉพาะ QUEUED | สูง |
| Failed queue/batch status | Query ใช้ `batch_id` ร่วมกับ `status='FAILED'` | มี index batch เดี่ยวแต่ไม่ครอบคลุม status/item | Count/EXISTS อ่าน row เกินจำเป็น | Partial index `(batch_id, queue_item_id)` | สูง |
| Request/ES timeout | Search ไม่มี operation timeout ที่ตั้งค่าได้ | พึ่ง client connection และ context เดิม | Request ค้างและใช้ connection นาน | Configurable request deadline และ ES timeout | สูง |
| Observability | ไม่มี threshold กลางสำหรับ slow API/SQL/ES | Log กระจัดกระจาย | หา bottleneck จาก production ยาก | Threshold จาก environment และ log เฉพาะ slow/failed | กลาง |

## 2. สิ่งที่แก้ไข

- Main Log membership check เปลี่ยนจาก count เป็น existence query
- Policy lookup เลือกเฉพาะ `index_prefix`, ใช้ request context และตัด index pattern ซ้ำ
- Main Log และ Dashboard จำกัด Elasticsearch operation timeout
- ตรวจ `timed_out: true` จาก Elasticsearch และคืน timeout แทน partial result
- Dashboard normalize `limit` เป็น 1–100 และบังคับ offset ไม่ติดลบ
- Dashboard exact/range conditions ใช้ `bool.filter`; keyword search ยังอยู่ใน `must`
- Project/Category IDs ถูกตัดค่าซ้ำ ค่าติดลบ และจำกัดสูงสุด 100 ค่า
- Custom Field path รองรับ `orders[].items[].sku` โดยแปลงเป็น Elasticsearch field `orders.items.sku`
- ปฏิเสธ Custom Field path ที่มี segment นอก allowlist และปฏิเสธ path/value ที่ส่งมาไม่ครบคู่
- Favorite reorder เปลี่ยนจาก N updates เป็น atomic update ครั้งเดียว
- Favorite key resolution เปลี่ยนจาก N counts เป็นหนึ่ง query และ map lookup
- Read query สำคัญของ Favorite/Custom Field ผูกกับ request context
- เพิ่ม slow API, slow SQL และ slow Elasticsearch logging ที่ตั้งค่าจาก environment
- เพิ่ม partial indexes สองรายการแบบ `CREATE INDEX CONCURRENTLY`

## 3. โครงสร้างไฟล์ที่เพิ่มหรือแยก

```text
backend/api/
├── middleware/
│   ├── performance.go                    # log เฉพาะ slow/failed API request
│   ├── timeout.go                        # propagate request deadline
│   └── timeout_test.go
├── internal/dashboard/
│   ├── repository/
│   │   ├── observability.go              # slow Elasticsearch metrics/log
│   │   └── search_test.go                # filter context และ input bound
│   └── usecase/
│       ├── query_normalization.go         # pagination normalization
│       └── query_normalization_test.go
├── internal/main_logs/usecase/
│   ├── query_builder.go                   # filter validation/normalization
│   └── query_builder_test.go
└── migrate/
    ├── query_performance_indexes.go       # create/drop targeted indexes
    └── query_performance_indexes_test.go
```

## 4. Flow หลังปรับปรุง

### Main Log Search

```text
HTTP Request
→ Auth middleware
→ Request timeout middleware
→ Handler parse pagination/filter
→ Usecase normalize และ validate filter
→ Repository existence check สำหรับ membership
→ Repository resolve index prefix ด้วย context
→ Elasticsearch bool.filter + explicit operation timeout
→ ตรวจ Elasticsearch timed_out
→ Mapper
→ Response structure เดิม หรือ 400/504 สำหรับ invalid filter/timeout
```

### Dashboard Search/Stats

```text
HTTP Request
→ Auth + request deadline
→ Handler bind และ normalize limit/offset
→ Usecase
→ Repository resolve index pattern ด้วย context
→ Elasticsearch filter context + timeout
→ Parse result และ reject partial timed-out response
→ Response structure เดิม
```

### Favorite Reorder

```text
HTTP Request
→ Auth + request deadline
→ Validate duplicate field ID/display order
→ Build parameterized CASE expression
→ Atomic UPDATE หนึ่ง statement
→ ตรวจ RowsAffected เท่ากับจำนวนที่ส่งมา
→ 204 No Content
```

## 5. Performance และ Benchmark

ไม่สามารถวัด latency ก่อน/หลังของ endpoint จริงได้ใน workspace นี้ เพราะไม่มี production-size PostgreSQL/Elasticsearch dataset และ service topology สำหรับ load test จึงไม่สร้างตัวเลข endpoint สมมติ

Microbenchmark ปัจจุบันของ query construction บน Windows/Intel i5-9300H รัน `count=3`:

| Benchmark | เวลา | Memory | Allocations |
|---|---:|---:|---:|
| Main Log query builder | 15.7–18.9 µs/op | 21,053 B/op | 179 allocs/op |
| Dashboard query builder | 10.3–11.4 µs/op | 15,479–15,480 B/op | 131 allocs/op |

ตัวเลขนี้เป็น baseline ของโค้ดหลังแก้เท่านั้น ไม่ใช่ตัวแทน network/database latency และไม่ใช้กล่าวอ้างเปอร์เซ็นต์ improvement

ผลเชิงโครงสร้างที่ยืนยันได้:

- Favorite reorder ลดจาก N `UPDATE` เหลือ 1 `UPDATE`
- Favorite key collision ลดจาก N `COUNT` เหลือ 1 `SELECT field_key`
- Membership check หยุดเมื่อพบ row แรก แทนการ count ทั้งหมด
- Policy lookup ลด column ที่อ่านเหลือ `index_prefix`

## 6. Database Changes

| Index | Query ที่รองรับ | Column order | Write impact | วิธีสร้าง | ซ้ำกับเดิม |
|---|---|---|---|---|---|
| `idx_log_queue_batches_queued_order` | pending batch: `status='QUEUED' ORDER BY priority DESC, received_at ASC` | sort order ตรง query; status อยู่ใน partial predicate | ดูแล index เฉพาะ row ที่เป็น QUEUED และตอนเปลี่ยนสถานะ | Concurrently | ไม่มี index เดิมที่ตรง filter+sort |
| `idx_log_failures_failed_batch_item` | failed batch EXISTS และ worker count distinct item ต่อ batch | `batch_id` ก่อนเพื่อ lookup, `queue_item_id` ต่อเพื่อ count/index-only opportunity | ดูแลเฉพาะ failure ที่ status FAILED | Concurrently | index batch เดิมไม่ครอบคลุม status/item |

Rollback:

```go
migrate.RollbackQueryPerformanceIndexes(db)
```

ฟังก์ชันใช้ `DROP INDEX CONCURRENTLY IF EXISTS` ทั้งสองรายการ Migration จะข้ามเมื่อ Dialector ไม่ใช่ PostgreSQL

## 7. Configuration ใหม่

```env
DB_SLOW_QUERY_THRESHOLD_MS=500
API_REQUEST_TIMEOUT_SECONDS=15
API_SLOW_REQUEST_THRESHOLD_MS=2000
ELASTICSEARCH_QUERY_TIMEOUT_SECONDS=10
ELASTICSEARCH_SLOW_QUERY_THRESHOLD_MS=1000
```

ค่า connection pool เดิมถูกจัดทำตัวอย่างใน `.env.example` ให้ครบ แต่ไม่ได้เปลี่ยน default runtime

## 8. Verification

- Targeted tests ผ่าน: custom fields, main logs, dashboard, middleware, configs, migration และ routes
- Production binaries ผ่าน: `cmd/api`, `cmd/worker`, `cmd/migrate`
- `go build ./...` ยังไม่ผ่านเพราะ `scratch/cmd/run_dump.go` import package `scratch` ซึ่งเป็น program; เป็นปัญหาเดิมใน scratch utilities และไม่เกี่ยวกับ production binaries

## 9. ความเสี่ยงและงานถัดไป

1. Category hierarchy ยังใช้ leading wildcard กับ comma-separated path IDs ควรเพิ่ม normalized numeric array field ใน mapping แล้ว reindex ก่อนตัด fallback เดิม
2. Offset pagination ยังมีต้นทุนเมื่อ offset สูง แม้จำกัด page size แล้ว ควรเพิ่ม `search_after` แบบ opt-in โดยรักษา API เดิมช่วงเปลี่ยนผ่าน
3. Dashboard audit filter ที่อ่าน JSON และ `LIKE '%...%'` ต้องวัด `EXPLAIN (ANALYZE, BUFFERS)` จากข้อมูลจริงก่อนเลือก expression/GIN/trigram index
4. Audit secret และ sensitive access list บาง endpoint ยังไม่มี pagination การเพิ่มต้องออก API version หรือ backward-compatible metadata
5. Authorization helper ใน Product/Project/Feature/Environment มี query pattern ซ้ำ ควรรวมเป็น service เดียวและเพิ่ม query-count regression test
6. Repository รุ่นเก่าบางส่วนยังไม่รับ `context.Context`; request timeout จะยกเลิกได้เฉพาะ layer ที่ใช้ `WithContext`
7. Archive/restore และ clear-log เป็นงานยาว ควรย้ายเป็น background job พร้อม progress state แทนการบังคับ timeout สั้น
8. ต้องทำ load test จริงกับ List Logs, Search Logs, Dashboard Stats และ Failed Queue ก่อน deploy production
