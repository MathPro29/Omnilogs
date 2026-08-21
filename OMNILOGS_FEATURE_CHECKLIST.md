# เช็กลิสต์ฟีเจอร์ OmniLogs ฝั่ง Backend

## สถานะ

- `[/]` = ทำเสร็จแล้ว ใช้งานได้จริง
- `[X]` = ยังไม่มี หรือยังใช้ไม่ได้
- `[ ]` = มีบางส่วนแล้ว แต่ยังไม่ครบหรือยังไม่สมบูรณ์

---

## 1. การยืนยันตัวตน

คำอธิบาย: ใช้สำหรับสมัครสมาชิก, เข้าสู่ระบบ, ออก token, ออกจากระบบ และรีเซ็ตรหัสผ่าน

- [ ] สมัครสมาชิก
      อธิบาย: ใช้งานได้เกือบครบ แต่ flow ยังไม่สมบูรณ์ เพราะ validation ของ `username` ยังขัดกับ contract บางส่วน
- [/] เข้าสู่ระบบ
  อธิบาย: login ด้วย email หรือ username และ password ได้จริง พร้อมออก access token และ refresh token
- [/] ต่ออายุ token
  อธิบาย: refresh token flow ฝั่ง backend ใช้งานได้จริง
- [/] ออกจากระบบ
  อธิบาย: revoke refresh token ได้จริง
- [/] ดูข้อมูลผู้ใช้ปัจจุบัน
  อธิบาย: ใช้ access token ดึงข้อมูล user ปัจจุบันได้
- [/] ขอรีเซ็ตรหัสผ่าน
  อธิบาย: สร้าง reset token และบันทึกลงฐานข้อมูลได้
- [/] ตั้งรหัสผ่านใหม่
  อธิบาย: เปลี่ยนรหัสผ่านใหม่และ revoke session เดิมทั้งหมดได้

---

## 2. การจัดการผู้ใช้

คำอธิบาย: ใช้ดูข้อมูลผู้ใช้และบริหารผู้ใช้ระดับระบบ

- [ ] ดูรายชื่อผู้ใช้ทั้งหมด
      อธิบาย: มี endpoint ใช้งานได้ แต่ยังติดเงื่อนไขสิทธิ์เพิ่มเติมจาก environment config
- [x] สร้างผู้ใช้โดย admin
      อธิบาย: ยังไม่มี API backend สำหรับ admin CRUD user โดยตรง
- [x] แก้ไขผู้ใช้
      อธิบาย: ยังไม่มี endpoint backend โดยตรง
- [x] ลบผู้ใช้
      อธิบาย: ยังไม่มี endpoint backend โดยตรง
- [x] ดูรายละเอียดผู้ใช้รายคน
      อธิบาย: ยังไม่มี endpoint backend โดยตรง

---

## 3. สิทธิ์และการอนุญาต

คำอธิบาย: ใช้กำหนดสิทธิ์ระดับ platform และระดับ product ว่าใครทำอะไรได้บ้าง

- [/] เปลี่ยน role ระดับ platform
  อธิบาย: admin สามารถเปลี่ยน role ของ user ได้จริง
- [/] ตรวจสอบสิทธิ์
  อธิบาย: ระบบคำนวณสิทธิ์จาก membership, scope, role permissions และ explicit rules ได้จริง
- [/] สร้าง permission rule เพิ่มเติม
  อธิบาย: เพิ่ม ALLOW หรือ DENY เฉพาะ user/resource/action ได้
- [/] ดู permission rule
  อธิบาย: อ่านรายการ permission rules ของ product ได้
- [/] แก้ไข permission rule
  อธิบาย: ปรับ effect, expiry และสถานะ active ได้
- [ ] บังคับสิทธิ์ได้ครบทุก route อย่างสม่ำเสมอ
      อธิบาย: มีหลายส่วนแล้ว แต่ยังไม่เป็นมาตรฐานเดียวกันทุก endpoint

---

## 4. การจัดการ Product

คำอธิบาย: ใช้สร้างและดูแล product หลักในระบบ

- [/] สร้าง Product
  อธิบาย: สร้าง product พร้อม environments และ role พื้นฐานได้จริง
- [/] ดูรายการ Product
  อธิบาย: list product ได้ตามสิทธิ์
- [/] ดูรายละเอียด Product
  อธิบาย: อ่านข้อมูล product รายตัวได้
- [/] แก้ไข Product
  อธิบาย: แก้ชื่อ, code และสถานะได้
- [/] ลบ Product
  อธิบาย: ลบ product รายตัวได้
- [ ] ลบ Product หลายรายการพร้อมกัน
      อธิบาย: มี flow แล้ว แต่ยังควรจัดให้สมบูรณ์และชัดเจนขึ้น

---

## 5. การจัดการ Project

คำอธิบาย: ใช้จัดการ project ภายใต้ product

- [/] สร้าง Project
  อธิบาย: สร้าง project ใต้ product ได้จริง
- [/] ดูรายการ Project
  อธิบาย: list projects ของ product ได้
- [/] ดูรายละเอียด Project
  อธิบาย: อ่าน project รายตัวได้
- [/] แก้ไข Project
  อธิบาย: แก้ชื่อและสถานะได้
- [x] ลบ Project
      อธิบาย: ตอนนี้ยังไม่มี route backend สำหรับลบ project

---

## 6. การจัดการ Feature / Category

คำอธิบาย: ใช้แบ่งโครงสร้างภายใน project เป็น feature หรือ category แบบลำดับชั้น

- [/] สร้าง Feature
  อธิบาย: สร้างได้ทั้ง root และ child feature
- [/] ดูรายการ Feature
  อธิบาย: ดึง tree, path และ level ของ feature ได้
- [/] แก้ไข Feature
  อธิบาย: เปลี่ยน parent, type, name และสถานะได้
- [/] ลบ Feature
  อธิบาย: ลบได้ตาม validation ของระบบ

---

## 7. การจัดการ Product Role

คำอธิบาย: ใช้สร้าง role ภายใน product และกำหนด permission ของ role นั้น

- [/] สร้าง Product Role
  อธิบาย: สร้าง role พร้อม permission list ได้จริง
- [/] ดู Product Role
  อธิบาย: ดู role และ permissions ได้
- [/] แก้ไข Product Role
  อธิบาย: ปรับชื่อและสิทธิ์ได้
- [/] ลบ Product Role
  อธิบาย: ลบ role ได้เมื่อไม่ติด constraint อื่น

---

## 8. การให้สิทธิ์เข้า Product

คำอธิบาย: ใช้ grant access ให้ user เข้า product, project หรือ feature ตามขอบเขตที่ต้องการ

- [/] สร้าง Membership
  อธิบาย: เพิ่ม user เข้า product พร้อม role ได้จริง
- [/] สร้าง Membership หลายคนพร้อมกัน
  อธิบาย: grant ให้หลาย user ได้ในครั้งเดียว
- [/] ดู Membership
  อธิบาย: ดูสมาชิกทั้งหมดของ product ได้
- [/] แก้ไข Membership
  อธิบาย: เปลี่ยน role, สถานะ และวันหมดอายุได้
- [/] ลบ Membership
  อธิบาย: ถอน user ออกจาก product ได้
- [/] สร้าง Scope
  อธิบาย: จำกัดสิทธิ์ระดับ product, project หรือ category ได้
- [/] ดู Scope
  อธิบาย: ดูขอบเขตสิทธิ์ของ membership ได้
- [/] แก้ไข Scope
  อธิบาย: ปรับ scope ได้
- [/] ลบ Scope
  อธิบาย: ถอนสิทธิ์เฉพาะ scope ได้

---

## 9. API Key

คำอธิบาย: ใช้สร้าง API key สำหรับ product และ environment

- [/] สร้าง API Key
  อธิบาย: สร้าง key สำหรับ product หรือ environment ได้
- [/] ดู API Key
  อธิบาย: ดูรายการ key ได้
- [/] แก้ไข API Key
  อธิบาย: ปรับชื่อ, expiry และสถานะได้
- [/] ยกเลิก API Key
  อธิบาย: revoke key ได้จริง
- [x] ใช้ API Key รับ log จากภายนอกโดยตรง
      อธิบาย: ยังไม่มี public ingest flow ที่สมบูรณ์ผ่าน API key

---

## 10. การรับ Log และ Queue

คำอธิบาย: ตอนนี้ backend รับ log ผ่าน queue แล้วให้ worker ไปประมวลผลต่อ

- [/] รับ Log Batch เข้า Queue
  อธิบาย: enqueue log batch ได้จริง
- [/] สั่ง Consume Queue
  อธิบาย: trigger worker ให้ประมวลผล queue ได้
- [/] มี Background Worker
  อธิบาย: worker ดึง queue ไป validate, transform และส่งเข้า Elasticsearch ได้
- [ ] รับ Log จากระบบภายนอกแบบครบ flow
      อธิบาย: โครงสร้าง queue/worker มีแล้ว แต่ external ingest flow ยังไม่ครบ

---

## 11. การค้นหา Log

คำอธิบาย: ใช้ค้น log ใน Elasticsearch และเปิดดูรายละเอียด log รายตัว

- [/] ค้นหา Log
  อธิบาย: ค้นตาม product, environment, project, category, level และ keyword ได้
- [/] ดูรายละเอียด Log
  อธิบาย: เปิดดู log รายตัวได้ และมี fallback จาก PostgreSQL เมื่อ ES document หาย
- [/] ดู Main Log จาก Audit Log
  อธิบาย: จาก audit สามารถโยงกลับไป main log ที่เกี่ยวข้องได้

---

## 12. Audit Logs

คำอธิบาย: ระบบบันทึก audit log อัตโนมัติหลังการเรียก API

- [/] บันทึก Audit อัตโนมัติ
  อธิบาย: middleware เขียน audit log ให้หลัง request จบ
- [/] ดูรายการ Audit Log
  อธิบาย: list audit logs ได้จริง
- [/] ดู Audit Log รายตัว
  อธิบาย: เปิดดู audit record รายตัวได้
- [/] ดู Audit Feed แบบย่อ
  อธิบาย: มี endpoint สำหรับดึง audit feed ตาม product/project/feature

---

## 13. การเข้าถึง Sensitive Log

คำอธิบาย: ใช้สำหรับข้อมูล log ที่เป็นความลับและต้องขอสิทธิ์ก่อนเปิดดู

- [/] สร้างคำขอเปิดดูข้อมูลลับ
  อธิบาย: user ขอ access ได้จริง
- [/] อนุมัติหรือปฏิเสธคำขอ
  อธิบาย: reviewer ที่มีสิทธิ์สามารถ review ได้
- [/] ดูรายการคำขอ
  อธิบาย: list request ได้จริง
- [/] เปิดดูค่าที่เข้ารหัส
  อธิบาย: reveal ค่า sensitive หลังผ่านเงื่อนไขได้
- [/] ดูประวัติการเข้าถึง
  อธิบาย: มี access history เก็บไว้จริง

---

## 14. Retention และ Archive

คำอธิบาย: ใช้จัดการอายุของ index, archive log และ restore log กลับมา

- [/] สร้าง Policy
  อธิบาย: สร้าง retention policy ได้จริง
- [/] ดู Policy
  อธิบาย: list policy ได้จริง
- [x] ดู Policy รายตัว
      อธิบาย: route มี แต่ handler ทำงานผิด flow
- [ ] แก้ไข Policy
      อธิบาย: มีแล้ว แต่ยังไม่สมบูรณ์เรื่องการใช้ path param และ request body
- [ ] ลบ Policy
      อธิบาย: มี flow แล้ว แต่ยังไม่สะอาดและไม่สม่ำเสมอ
- [/] ย้าย Log ไป Archive
  อธิบาย: push logs to archive ได้จริง
- [/] ดู Archive
  อธิบาย: ดู archive records ได้
- [/] กู้คืน Archive
  อธิบาย: restore log กลับเข้า Elasticsearch ได้จริง
- [/] ล้าง Log ทั้งหมด
  อธิบาย: clear logs และ indices ของ product ได้จริง

---

## 15. Dashboard และ Monitoring API

คำอธิบาย: ใช้ให้ backend เปิดข้อมูลเชิงสถิติและสถานะระบบ

- [/] ดูสถิติ Log
  อธิบาย: backend มี stats endpoint จาก Elasticsearch ใช้งานได้จริง
- [/] ดู Audit Feed
  อธิบาย: backend มี audit feed endpoint ใช้งานได้จริง
- [x] Dashboard Summary API
      อธิบาย: ยังไม่มี endpoint summary จริงใน backend
- [ ] ดูรายละเอียด Log ผ่าน dashboard route
      อธิบาย: มีบาง route แล้ว แต่ยังไม่ใช่ flow หลักและยังไม่ครบ
- [/] เช็กสถานะ Elasticsearch
  อธิบาย: readiness endpoint ใช้เช็ก DB และ ES ได้จริง

---

## 16. Queue Inspection

คำอธิบาย: ใช้ดูสถานะ batch และ item ที่อยู่ในคิว

- [/] ดูรายการ item ใน batch
  อธิบาย: เปิดดูรายการ queue item ภายใน batch ได้
- [/] ดูรายละเอียด queue item
  อธิบาย: ดู item รายตัวได้
- [/] ติดตาม retry และ failure
  อธิบาย: ระบบมี failure tracking และ retry logic แล้ว

---

## 17. System และ Health Check

คำอธิบาย: ใช้ตรวจว่าระบบพร้อมทำงานหรือไม่

- [/] Liveness Probe
  อธิบาย: เช็กว่า service ยัง live อยู่
- [/] Readiness Probe
  อธิบาย: เช็กว่า database และ Elasticsearch พร้อมใช้งาน
- [/] Swagger / OpenAPI
  อธิบาย: มีเอกสาร API ให้ใช้งาน
- [/] Middleware หลักของระบบ
  อธิบาย: มี recovery, request id, rate limit, auth และ audit middleware แล้ว

---

## ประเด็นที่ยังขาดหลัก ๆ ฝั่ง Backend

- [x] User CRUD API ยังไม่มี
- [x] Settings API ยังไม่มี
- [x] Dashboard Summary API ยังไม่มี
- [ ] Public ingestion ผ่าน API key ยังไม่ครบ flow
- [x] Metadata log หลายตารางมีแล้ว แต่ยังไม่มี API ใช้งานจริง
- [ ] Contract และรูปแบบ response ของบาง endpoint ยังไม่สม่ำเสมอ

---

## ลำดับที่ควรทำต่อ ฝั่ง Backend

- [ ] ปรับ auth/register contract ให้ตรงกัน
- [x] ทำ User CRUD API
- [x] ทำ Settings API
- [x] ทำ Dashboard Summary API
- [x] ทำ public ingest flow ผ่าน API key
- [ ] ทำ API สำหรับ log schema, masking, source และ ingestion policy
- [ ] ทำ retention policy handlers ให้สมบูรณ์และสม่ำเสมอ
