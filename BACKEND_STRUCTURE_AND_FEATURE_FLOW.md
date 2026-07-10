# Backend Structure and Feature Flow

## 1. Overview

Backend นี้เป็นส่วน API หลักของระบบ **OmniLogs** (ระบบศูนย์รวมการจัดการและจัดเก็บ Log ประสิทธิภาพสูง) พัฒนาด้วยภาษา Go และใช้ Gin Web Framework ทำหน้าที่จัดการและประสานงานส่วนประกอบต่าง ๆ ได้แก่:

- **Authentication & Platform Management**: สมัครสมาชิก ล็อกอิน จัดการผู้ใช้ แพลตฟอร์มโรล (God, Owner, Superadmin)
- **Product, Project, and Feature Access**: จัดการข้อมูลสินค้า โปรเจกต์ย่อย รวมถึงกำหนด Product Memberships, Roles และ Permissions (RBAC)
- **Log Ingestion & Queue Processing**: รับ Payload Log จำนวนมากจากภายนอก ส่งเข้า Queue (NATS JetStream) และมี Background Worker คอยหยิบไปประมวลผลก่อนนำไป Index ใน Elasticsearch
- **Log Search & Sensitive Data Decryption**: ค้นหาข้อมูล Log แบบ Full-text ค้นหาแบบเรียลไทม์ (Live Tail) ตลอดจนการทำ Masking ปิดบังข้อมูลสำคัญ และการขออนุมัติเพื่อถอดรหัสเปิดดูข้อมูลอ่อนไหว (Sensitive Data Access Request)
- **Log Retention & Archiving**: ตรวจสอบอายุข้อมูลของ Log และทำการบีบอัดเก็บเป็นไฟล์สำรอง (ZIP Archives) ก่อนลบ Index ใน Elasticsearch เพื่อประหยัดพื้นที่จัดเก็บข้อมูล

---

## 2. Backend Folder Tree

โครงสร้างไดเรกทอรีหลักของโปรเจกต์ (ฝั่ง Backend API) มีดังนี้:

```text
backend/api/
├─ cmd/
│  ├─ api/
│  │  └─ main.go                 # จุดเริ่มต้นรัน REST API Server
│  └─ worker/
│     └─ main.go                 # จุดเริ่มต้นรัน Background Worker Daemon
├─ configs/                      # ตั้งค่าระบบ เช่น การเชื่อมต่อ Database, NATS, ES, Environment Variables
├─ dto/                          # Data Transfer Object (โครงสร้างข้อมูลรับ-ส่ง)
├─ internal/                     # Domain Module ต่าง ๆ ตาม Clean Architecture
│  ├─ api_keys/                  # จัดการ Product API Keys
│  ├─ archive_utils/             # ฟังก์ชันจัดการการบีบอัดและสำรองข้อมูล Elastic Index
│  ├─ audit_logs/                # บันทึกประวัติการกระทำต่าง ๆ บนแพลตฟอร์ม
│  ├─ audit_secret/              # จัดการคีย์สำหรับการสืบค้นข้อมูลลับ
│  ├─ auth/                      # การยืนยันตัวตนระดับแพลตฟอร์มและจัดการสมาชิก
│  ├─ background_working/        # ตัวกลางจัดตั้งค่า API Server และ Worker (แพ็กเกจ bootstrap)
│  ├─ dashboard/                 # คำนวณสถิติและดึงข้อมูลสรุปหน้าแดชบอร์ด
│  ├─ elastic_index_policy/      # กำหนดโครงสร้าง Index Shards, Replicas และนโยบายลบข้อมูล
│  ├─ environment/               # จัดการสภาพแวดล้อมใช้งานของ Product (เช่น Dev, Production)
│  ├─ feature/                   # จัดการ Feature ภายใต้โปรเจกต์
│  ├─ global_auth/               # จัดการสิทธิ์การใช้งานระดับ Product (Roles & Permissions)
│  ├─ log_archive/               # แสดงรายการและกู้คืนไฟล์ Log สำรอง
│  ├─ log_queues/                # รับ Log เข้าคิวเพื่อรอประมวลผล
│  ├─ main_logs/                 # บริการค้นหา Log และดึง Log จาก Elasticsearch/PostgreSQL
│  ├─ product/                   # จัดการ Product และสิทธิ์การเข้าถึงข้อมูลสินค้า
│  ├─ project/                   # จัดการโปรเจกต์ภายใต้ Product
│  ├─ queue/                     # ตัวเชื่อมต่อคิวและจัดการโครงสร้าง Payload คิว (NATS Wrapper)
│  ├─ scopes/                    # ตรวจสอบขอบเขตการทำงาน (Scope Checking)
│  ├─ sensitive_log/             # จัดการคำขอและอนุมัติเข้าถึงข้อมูล Sensitive
│  └─ worker/                    # ตรรกะประมวลผล จัดการ Masking ความลับ และ Index ข้อมูลลง Elastic
├─ middleware/                   # Middlewares สำหรับ Gin เช่น Rate Limit, Auth, Audit Logger
├─ migrate/                      # สคริปต์ทำ Database Migration
├─ models/                       # โครงสร้างตาราง PostgreSQL (GORM Models)
├─ responses/                    # ตัวจัดการแสดงผล Error และข้อความตอบกลับ
├─ routes/                       # ตัวจัดเส้นทาง API Route
├─ utils/                        # เครื่องมือและฟังก์ชันช่วยเหลือทั่วไป (เช่น การแฮช, สร้าง Token JWT)
├─ main.go                       # จุดรันหลัก (มักจะเรียกใช้ bootstrap.RunAPIServer)
└─ go.mod                        # ไฟล์จัดการ Go Module dependencies
```

### คำอธิบายหน้าที่ของแต่ละส่วน:

- **cmd/**: จุดเริ่มต้นหลักของแอปพลิเคชัน แยกการรันเป็น API server และ Worker Daemon ออกจากกันเพื่อรองรับการขยายตัว (Scale)
- **configs/**: โหลดค่าตัวแปรระบบ (`.env`) และทำหน้าที่เชื่อมต่อระบบฐานข้อมูล (PostgreSQL via GORM, Elasticsearch และ NATS JetStream)
- **routes/**: กำหนด API Route แยกตามกลุ่มการทำงาน และนำ Middleware มาดักกรองพารามิเตอร์ต่าง ๆ
- **middleware/**: ตัวประมวลผลก่อนเข้า Handler เช่น การยืนยันสิทธิ์ Token JWT, การจำกัดปริมาณคำขอ (Rate Limit) และการทำบันทึก Audit ประวัติการทำธุรกรรม
- **internal/ (Domain Modules)**: โค้ดของแต่ละ Feature ถูกแยกโฟลเดอร์ออกจากกันอย่างชัดเจน โดยด้านในแต่ละโมดูลจะแบ่งโครงสร้างตาม **Clean Architecture**:
  - `handler/`: รับ HTTP Request, Bind JSON และตรวจสอบ Input ข้อมูลเบื้องต้น
  - `usecase/` (หรือ Service): จัดการ Business Logic หลักของ Feature นั้น ๆ
  - `repository/`: การดึง/เขียนข้อมูลไปยัง Database หรือ Elasticsearch
- **models/**: กำหนด Schema ตารางฐานข้อมูล PostgreSQL ทั้งหมดเพื่อใช้งานร่วมกับ GORM ในการทำ Query และ Migration

---

## 3. Backend Request Lifecycle

การไหลของข้อมูลเมื่อ Request วิ่งเข้ามายัง Backend API:

```text
Client / Frontend
  → Route Configuration
  → Global Middleware (Recovery, CORS, Request ID, Global Rate Limit)
  → Audit Logger Middleware (บันทึก Audit Log ลง DB)
  → Auth Middleware (ตรวจสอบ JWT Token)
  → Authorization checks (ตรวจสอบสิทธิ์ Product Role / Platform Role)
  → Module Handler (รับ HTTP Request และ Bind DTO)
  → Module Usecase (รัน Business Logic)
  → Module Repository (ดึง/เขียนข้อมูลลง PostgreSQL / Elasticsearch)
  → Module Usecase (ประมวลผลข้อมูลลัพธ์)
  → Module Handler (แปลง DTO และจัดรูปแบบ Response)
  → Response กลับไปให้ Client / Frontend
```

### รายละเอียดในแต่ละชั้น:

1. **Route**: ดักจับ URL และระบุเส้นทางผ่าน HTTP Method (GET, POST, PUT, DELETE, PATCH)
2. **Middleware**: ตรวจสอบและดักสิทธิ์ เช่น ยืนยัน Token JWT ที่ส่งมาผ่าน Header `Authorization` และสกัดข้อมูล User ID ลงใน Context ของ Gin เพื่อให้ชั้นถัดไปใช้งานต่อได้
3. **Handler**: ดึงค่าพารามิเตอร์ (Path, Query, JSON Body) มาทำการ Bind เข้าหา DTO Struct และตรวจสอบความถูกต้องเบื้องต้น (Validation)
4. **Usecase / Service**: ปฏิบัติตาม Business Logic ประสานงานเรียกใช้ Repository หรือคำสั่งเข้ารหัสข้อมูล และเรียกใช้งานระบบบันทึกประวัติ (Audit)
5. **Repository**: ติดต่อกับ Database Client เพื่อรัน SQL หรือยิงคำสั่ง Elasticsearch Query/Bulk API
6. **Response**: นำผลลัพธ์จาก Usecase มาแปลงให้อยู่ในรูป JSON มาตรฐานและส่งกลับไปทางเครือข่าย

---

## 4. Feature Flow

_(ดูรายละเอียดในหัวข้อที่ 5 สำหรับ Feature Flows ที่ใช้งานจริงทั้งหมด)_

---

## 5. Required Feature Flows

นี่คือรายละเอียดและลำดับขั้นตอนการทำงานของ Feature ต่าง ๆ ที่มีอยู่จริงในระบบ OmniLogs:

---

### 5.1 Authentication Flow

#### Purpose

ใช้ตรวจสอบและยืนยันตัวตนของผู้ใช้งานระบบแพลตฟอร์ม ประกอบด้วยการสมัครสมาชิก (Register), การล็อกอินเข้าสู่ระบบ (Login), การสร้าง Token ใหม่เมื่อหมดอายุ (Refresh Token), การออกจากระบบ (Logout) และการกู้คืนรหัสผ่าน (Forgot/Reset Password)

#### Related Files

- `routes/auth_routes.go`
- `internal/auth/handler/auth.go`
- `internal/auth/usecase/auth_flow.go`
- `internal/auth/usecase/password_flow.go`
- `internal/auth/repository/user.go`
- `internal/auth/repository/session.go`
- `models/user.go`
- `models/auth_session.go`
- `models/password_reset_token.go`
- `middleware/auth.go`

#### Flow Diagram

```text
Frontend User
  → POST /api/v1/auth/login
  → AuthRateLimit Middleware (ตรวจสอบ Rate limit)
  → AuthHandler.Login()
  → AuthUsecase.Login()
  → UserRepository.FindByEmail() (ตรวจสอบข้อมูลใน PostgreSQL)
  → PasswordService.Compare() (เทียบแฮชรหัสผ่าน)
  → Token Generation (สร้าง JWT Access Token & Refresh Token)
  → SessionRepository.Create() (บันทึก Session ลง DB)
  → Return tokens & user profile to Frontend
```

#### Step-by-Step Explanation

1. ผู้ใช้ส่งข้อมูลอีเมลและรหัสผ่านเข้ามาทาง `/api/v1/auth/login`
2. ระบบตรวจสอบความถี่ในการส่งคำขอผ่าน `AuthRateLimit` Middleware
3. `AuthHandler.Login` อ่านค่า JSON Body และตรวจสอบโครงสร้างข้อมูล
4. Handler ส่งข้อมูลไปที่ `AuthUsecase.Login`
5. Usecase สั่งดึงข้อมูล User จากตาราง `users` ผ่าน `UserRepository.FindByEmail`
6. ตรวจสอบว่าผู้ใช้เป็นสถานะ Active หรือไม่ และนำรหัสผ่านที่กรอกมาแฮชเพื่อเทียบกับ `password_hash` ในฐานข้อมูล
7. หากข้อมูลถูกต้อง ระบบจะสร้าง JWT Access Token และ Refresh Token
8. บันทึก Session ลงตาราง `auth_sessions` ด้วยสถานะพร้อมใช้งาน
9. Handler ส่งผลลัพธ์เป็น Access Token, Refresh Token และข้อมูลผู้ใช้กลับไปให้ Frontend

#### Input

- **Body (JSON)**: `email`, `password` (สำหรับ Login) / `username`, `email`, `password`, `full_name` (สำหรับ Register)

#### Output

- **Response (JSON)**: `access_token`, `refresh_token`, ข้อมูลรายละเอียดผู้ใช้

#### Database / Storage Used

- **PostgreSQL**: ตาราง `users`, `auth_sessions`, `password_reset_tokens`

---

### 5.2 Authorization / Permission Flow

#### Purpose

ควบคุมการเข้าสิทธิ์ของหน้าบ้านและระบบภายนอก (RBAC) เพื่อให้มั่นใจว่าผู้ใช้หรือบริการนั้น ๆ มีสิทธิ์ทำงานตามที่ร้องขอจริง ไม่ว่าจะเป็นระดับแพลตฟอร์ม (God, Owner, Superadmin) หรือสิทธิ์ระดับกลุ่มสินค้า (Product Roles & Permissions)

#### Related Files

- `routes/admin_auth_routes.go`
- `routes/product_routes.go`
- `middleware/auth.go`
- `internal/global_auth/handler/actions.go`
- `internal/global_auth/usecase/actions.go`
- `internal/global_auth/repository/repositoty.go` (ตรวจสอบสิทธิ์)
- `models/platform_role.go`
- `models/product_membership.go`
- `models/product_role_permission.go`

#### Flow Diagram

```text
Client Request
  → Route Endpoint
  → UserAuthMiddleware (แกะ JWT และกำหนด Context)
  → RequireAdminPlatformRole() / CheckPermission handler (ตรวจสอบบทบาทสิทธิ์)
  → GlobalAuthRepository.CountProductMembership() (ถ้าเป็น Route จัดการ Product)
  → Proceed to Handler/Usecase OR Block & Abort with 403 Forbidden
```

#### Step-by-Step Explanation

1. ลูกค้าส่ง Request ไปยัง API Endpoint ที่ได้รับการป้องกัน
2. `UserAuthMiddleware` แกะข้อมูล JWT Token จาก Header `Authorization: Bearer <token>`
3. เก็บข้อมูล `userId`, `role`, `email`, `roleId` ลงใน Gin Context
4. สำหรับระบบแอดมิน: `RequireAdminPlatformRole` เช็กว่าผู้ใช้มี Role เป็น `god`, `owner` หรือ `superadmin` หรือไม่ หากไม่ใช่จะส่ง 403 Forbidden กลับทันที
5. สำหรับการเข้าถึง Product/Project/Log: ในชั้น Usecase จะมีการเรียกใช้การเช็กสิทธิ์สมาชิก เช่น `CountProductMembership` เพื่อตรวจสอบว่าผู้ใช้คนนั้นเป็นสมาชิกของ Product ID ดังกล่าวจริงและได้รับอนุญาตตาม Action หรือไม่
6. หากตรวจสอบผ่าน จะทำงานใน Usecase ต่อไป หากไม่ผ่านจะ Reject ด้วย Error Forbidden

#### Input

- **Header**: `Authorization: Bearer <token>`
- **Body / Context**: `product_id`, `resource` (เช่น LOG, API_KEY), `action` (เช่น READ, CREATE)

#### Output

- **Context Variables**: `userId`, `role`
- **Response (เมื่อไม่ผ่าน)**: HTTP Status 401 Unauthorized หรือ 403 Forbidden

#### Database / Storage Used

- **PostgreSQL**: ตาราง `platform_memberships`, `platform_roles`, `product_memberships`, `product_roles`, `product_role_permissions`, `user_role_permission_rules`

---

### 5.3 Product Management Flow

#### Purpose

ใช้สำหรับบริหารจัดการกลุ่มสินค้าหลัก (Product) โดยทำหน้าที่เป็น Root Node ในการแยกแยะความเป็นเจ้าของของข้อมูล Log, สมาชิก, โปรเจกต์, คีย์ความปลอดภัย และเก็บข้อมูลประวัติการทำงาน

#### Related Files

- `routes/product_routes.go`
- `internal/product/handler/handler.go`
- `internal/product/usecase/product_usecase.go`
- `internal/product/repository/product_repository.go`
- `models/product.go`
- `models/product_membership.go`

#### Flow Diagram

```text
User Manager
  → POST /api/v1/products
  → UserAuthMiddleware (ตรวจสอบสิทธิ์การล็อกอิน)
  → ProductHandler.CreateProduct()
  → ProductUsecase.Create()
  → ProductRepository.Create() (บันทึกข้อมูลลง Postgres)
  → Return Status Created & Product detail
```

#### Step-by-Step Explanation

1. ผู้ใช้ส่งชื่อ Product และข้อมูลประกอบมาที่ `/api/v1/products`
2. Handler ตรวจสอบความถูกต้องของอินพุต
3. Usecase จัดตั้งค่าเบื้องต้นและบันทึก Product ใหม่ลงฐานข้อมูล PostgreSQL ผ่าน Repository
4. ระบบสร้าง Product ID เพื่อใช้ระบุสิทธิ์ในโมดูลย่อยถัด ๆ ไป
5. ส่งคืนรายละเอียดสินค้าที่จัดเก็บสำเร็จแล้วให้กับผู้ใช้

#### Input

- **Body (JSON)**: `product_name`, `description`

#### Output

- **Response (JSON)**: `product_id`, `product_name`, `description`, `created_at`

#### Database / Storage Used

- **PostgreSQL**: ตาราง `products`

---

### 5.4 Project / Feature Management Flow

#### Purpose

ใช้สำหรับจัดการหน่วยงานย่อยและ Feature ภายใต้ Product เพื่อกรองข้อมูล Log ให้มีโครงสร้างที่ชัดเจนและแยกสิทธิ์การดูข้อมูลตามรายแอปพลิเคชันหรือทีมพัฒนาย่อยได้

#### Related Files

- `routes/product_routes.go`
- `internal/project/handler/handler.go` / `internal/feature/handler/handler.go`
- `internal/project/usecase/usecase.go` / `internal/feature/usecase/usecase.go`
- `internal/project/repository/repository.go` / `internal/feature/repository/repository.go`
- `models/project.go`
- `models/project_feature.go`

#### Flow Diagram

```text
Frontend Developer
  → POST /api/v1/products/:productId/projects/:projectId/features
  → ProjectHandler / FeatureHandler.Create()
  → ProjectUsecase / FeatureUsecase.Create()
  → Check hierarchy in Repository (ตรวจสอบความสอดคล้อง)
  → PostgreSQL INSERT
  → Return Feature JSON Response
```

#### Step-by-Step Explanation

1. ผู้ใช้ส่งข้อมูลขอสร้างโปรเจกต์ หรือสร้างฟีเจอร์ลงมาที่ Endpoint ที่ระบุรหัสสินค้า
2. Usecase ตรวจสอบความสอดคล้อง (Hierarchy Validation) ว่าโปรเจกต์นั้นอยู่ในสิทธิ์ของ Product นั้นจริง
3. บันทึกข้อมูลโปรเจกต์ลงในตาราง `projects` หรือฟีเจอร์ลงในตาราง `project_features`
4. ส่งผลลัพธ์ข้อมูลกลับให้ผู้ใช้งาน

#### Input

- **Path Params**: `productId`, `projectId`
- **Body (JSON)**: `feature_name`, `description`

#### Output

- **Response (JSON)**: รายละเอียดโครงสร้าง Project หรือ Feature ที่สร้างสำเร็จ

#### Database / Storage Used

- **PostgreSQL**: ตาราง `projects`, `project_features`

---

### 5.5 API Key Flow

#### Purpose

ใช้สำหรับจัดการสร้าง แก้ไข และถอนสิทธิ์ API Key ประจำ Product โดย API Key นี้ทำหน้าที่เป็นสิทธิ์การเข้าถึงแทนบัญชีผู้ใช้งาน เพื่ออำนวยความสะดวกให้ระบบภายนอกสามารถยิง Log หรือค้นหาข้อมูลได้โดยอัตโนมัติ

#### Related Files

- `routes/product_routes.go`
- `internal/api_keys/handler/handler.go`
- `internal/api_keys/usecase/api_keys.go`
- `internal/api_keys/usecase/authorization.go`
- `models/product_api_key.go`

#### Flow Diagram

```text
Product Admin
  → POST /api/v1/products/:productId/api-keys
  → APIKeysHandler.CreateAPIKey()
  → APIKeysUsecase.CreateAPIKey() (ตรวจสิทธิ์ & สร้างคีย์สุ่ม)
  → Generate prefix & SHA-256 Hash
  → Save hashed value to PostgreSQL product_api_keys table
  → Return raw API Key (omni_xxx...) to Admin (โชว์ครั้งแรกครั้งเดียว)
```

#### Step-by-Step Explanation

1. แอดมินส่งคำขอสร้างคีย์พร้อมกำหนด Permissions ลงมาที่ระบบจัดการคีย์
2. Usecase เช็กสิทธิ์ว่าผู้สร้างมีสิทธิ์จัดการคีย์หรือไม่ผ่าน `authorize()`
3. ระบบจะสร้างสตริงคีย์สุ่มนำหน้าด้วย `omni_` (ความยาว 32 ไบต์สุ่ม)
4. ทำการแฮชสตริงตัวจริงด้วยอัลกอริทึม SHA-256
5. เก็บเฉพาะ Key Prefix และแฮชผลลัพธ์ `key_hash` ลงใน PostgreSQL เพื่อความปลอดภัยระดับสูง
6. ส่งสตริงคีย์ลับตัวเต็มกลับให้ผู้ใช้แสดงผลเพียงครั้งเดียวบน Frontend
7. **สถานะการทำงานจริง**: API Key CRUD ได้รับการพัฒนาเสร็จสมบูรณ์ แต่*ยังไม่ได้ถูกเชื่อมโยงเป็นด่านตรวจ (Middleware) สำหรับเส้นทาง Log Ingestion* ในขณะนี้ ในระบบจริงยังคงใช้การกรองผ่าน JWT (UserAuthMiddleware)

#### Input

- **Path Param**: `productId`
- **Body (JSON)**: `key_name`, `permissions`, `expires_at`

#### Output

- **Response (JSON)**: รายละเอียดโมเดลคีย์พร้อมฟิลด์ `api_key` (สตริงคีย์ลับดิบที่ห้ามทำหาย)

#### Database / Storage Used

- **PostgreSQL**: ตาราง `product_api_keys`

---

### 5.6 Log Ingestion Flow

#### Purpose

ทำหน้าที่รับข้อมูล Log บรรจุรวมมาเป็นกลุ่ม (Batch Log Ingestion) จากผู้ให้บริการหรือซอฟต์แวร์ภายนอก เพื่อนำเข้าสู่กระบวนการจัดเก็บอย่างปลอดภัย ผ่านการลงคิวและแจ้งผลลัพธ์การได้รับข้อมูลสำเร็จอย่างรวดเร็ว

#### Related Files

- `routes/log_queue_routes.go`
- `internal/log_queues/handler/handler.go`
- `internal/log_queues/usecase/usecase.go`
- `internal/log_queues/repository/repository.go`
- `internal/queue/message.go` (LogMessage wrapper)
- `models/log_queue_batch.go`

#### Flow Diagram

```text
Log Sender / Agent
  → POST /api/v1/queues
  → UserAuthMiddleware (ตรวจสอบสิทธิ์)
  → QueueHandler.QueueHandler()
  → QueueUsecase.Enqueue()
  → Persist metadata in PostgreSQL as "QUEUED"
  → Loop items: Wrap into queue.LogMessage & Publish to NATS JetStream
  → Return status HTTP 201 & Batch ID to Agent
```

#### Step-by-Step Explanation

1. ลูกค้าผู้ส่ง Log ทำการล็อกอินและส่ง JSON Payload กลุ่ม Log มาที่ `/api/v1/queues`
2. Handler รับข้อมูลและถอดโครงสร้างเป็น `IngestLogBatchRequest`
3. Usecase ตรวจสอบความถูกต้องระหว่าง Product ID และ Environment ID
4. คำนวณนโยบายการเก็บรักษาข้อมูล (Retention Days) จากนโยบายของระบบ
5. สร้างแถวประวัติในตาราง `log_queue_batches` ด้วยสถานะเริ่มต้นว่า `QUEUED`
6. ทำการวนลูป Log แต่ละรายการ นำมาห่อหุ้มในออบเจกต์ `queue.LogMessage` และสั่งพับลิชข้อมูลเข้าสู่ **NATS JetStream Queue Engine**
7. หากการนำส่งเข้า NATS สำเร็จในทุกคำขอ คิวจะเก็บไว้ในคิวรันงาน ส่วนผู้ส่งจะได้รับ HTTP 201 Created พร้อม Batch ID กลับไปทันทีโดยไม่ต้องรอให้ทำการบันทึกลง Elastic เสร็จ
8. หากนำเข้า NATS ไม่สำเร็จ ระบบจะอัปเดตสถานะของ Batch ใน PostgreSQL เป็น `FAILED` พร้อมเก็บข้อความ Error

#### Input

- **Body (JSON)**: `product_id`, `environment_id`, `source_type`, `logs` (อาเรย์ของ Raw logs payload)

#### Output

- **Response (JSON)**: รายละเอียด Batch ID, จำนวน Log ทั้งหมดในชุดข้อมูล และสถานะการนำเข้าคิว

#### Database / Storage Used

- **PostgreSQL**: ตาราง `log_queue_batches`
- **Queue**: NATS JetStream Subject/Stream

---

### 5.7 Log Search Flow

#### Purpose

ช่วยให้ผู้ใช้งานสามารถทำความเข้าใจ ค้นหาตัวกรอง ค้นหาแบบเต็มข้อความ (Full-text Search) บน Elasticsearch Cluster และรองรับการดึงข้อมูล Log ตัวหลัก รวมถึงการเปิดดู Log สตรีมมิ่งสด (Live Tail) ผ่านหน้าเว็บได้อย่างรวดเร็ว

#### Related Files

- `routes/main_log_routes.go`
- `internal/main_logs/handler/handler.go`
- `internal/main_logs/usecase/usecase.go`
- `internal/main_logs/usecase/query_builder.go` (ดึงลอจิกการต่อ Query ออกมา)
- `internal/main_logs/usecase/mapper.go` (ดึงการแปลงออบเจกต์ ES Hit ออกมา)
- `internal/main_logs/repository/repository.go` (แยกชั้นยิง API ออกมา)
- `models/log_index_ref.go`

#### Flow Diagram

```text
Web UI User
  → GET /api/v1/logs?product_id=1&query=error
  → MainLogHandler.Search()
  → MainLogUsecase.Search() (เช็กสิทธิ์ผู้ใช้)
  → query_builder.go: buildSearchQuery() (ประมวลผล Elasticsearch JSON query)
  → MainLogRepository.SearchElastic()
  → Get raw hits from Elasticsearch
  → mapper.go: mapToResponse() (แปลงรูปแบบข้อมูล)
  → Return list of logs to Frontend
```

#### Step-by-Step Explanation

1. ผู้ใช้ระบุเงื่อนไขการค้นหาผ่าน Query String เช่น Product ID, คีย์เวิร์ด, ช่วงเวลา และระดับของ Log
2. Handler ตรวจสอบพารามิเตอร์ และส่งต่อให้ `MainLogUsecase.Search`
3. Usecase เช็กสิทธิ์ความเป็นสมาชิก Product ของผู้ขอใช้งาน
4. เรียกใช้ฟังก์ชัน `buildSearchQuery` ใน `query_builder.go` เพื่อสร้าง Query Body JSON ของ Elasticsearch ที่ซับซ้อน (มีทั้ง Filter, Multi-match, Range และ Sort)
5. `MainLogRepository.SearchElastic` ยิงคำขอไปที่ Elasticsearch Server และส่งข้อมูล Hit ดิบกลับมา
6. แปลงเอกสารผ่าน `mapper.go` กลับไปเป็น DTO response ในระบบ
7. _สำหรับการดึง Log รายไอดี (`GetByID`)_: ระบบจะเช็กตาราง `log_index_refs` ใน PostgreSQL ก่อน เพื่อดูว่าไฟล์ถูกสร้างที่ Index ไหนใน Elastic จากนั้นจึงตามไปยิงดึง หากไม่พบตัวใน Elastic ระบบจะวิ่งไปดึง Backup Payload สำรองจาก Postgres เพื่อป้องกันข้อมูลหาย
8. ส่งคืนผลลัพธ์รายการ Log ให้กับ Frontend

#### Input

- **Query Params**: `product_id`, `environment_id`, `query`, `level`, `page`, `per_page`

#### Output

- **Response (JSON)**: `total` (จำนวนที่พบทั้งหมด), `logs` (ข้อมูลฟิลด์ Log ต่าง ๆ เช่น message, timestamp, level)

#### Database / Storage Used

- **Elasticsearch**: Index ของ Elastic ตามฟอร์แมต `omnilogs-product-<id>-*`
- **PostgreSQL**: ตาราง `log_index_refs`

---

### 5.8 Sensitive Data Access Flow

#### Purpose

ควบคุมความเป็นส่วนตัวและความปลอดภัยของข้อมูลสำคัญ (เช่น พาสเวิร์ด, โทเค็น, ข้อมูลบัตรเครดิต) ที่ถูกแปลงบิดเบือน (Masked) ใน Log โดยอนุญาตให้ผู้ใช้มีสิทธิ์ยื่นคำขอเพื่อดูข้อมูลจริง และผ่านกระบวนการอนุมัติ รวมถึงสืบสวนประวัติย้อนหลังได้

#### Related Files

- `routes/product_routes.go`
- `internal/sensitive_log/handler/handler.go`
- `internal/sensitive_log/usecase/usecase.go`
- `internal/sensitive_log/repository/repository.go`
- `models/sensitive_log_access_request.go`
- `models/sensitive_log_access_history.go`
- `models/log_sensitive_field_secret.go`

#### Flow Diagram

```text
Dev User
  → POST /api/v1/products/:productId/sensitive-logs/reveal
  → sensitiveHandler.RevealSensitiveValue()
  → sensitiveUsecase.RevealValue()
  → Check approval status in sensitive_log_access_requests table
  → Retrieve encrypted secret from log_sensitive_field_secrets
  → Decrypt using AES-256 GCM (with env.DataEncryptionKey)
  → Insert log into sensitive_log_access_histories (เพื่อทำ Audit)
  → Return raw decrypted string value to Frontend
```

#### Step-by-Step Explanation

1. นักพัฒนาตรวจพบ Log ที่มีข้อมูลอ่อนไหวที่ถูก Mask ไว้ (เช่น `******`) จึงสร้างคำขอเข้าถึงข้อมูลระบุรหัสงาน ผ่าน API ขอเข้าถึงข้อมูล
2. เมื่อผู้มีสิทธิ์อนุมัติ (เช่น God, Owner) ตรวจสอบและให้สิทธิ์อนุมัติผ่านระบบ
3. นักพัฒนายิงคำขอถอดรหัสมาที่ `/api/v1/products/:productId/sensitive-logs/reveal`
4. Usecase ตรวจสอบสถานะการอนุมัติว่าถูกอนุมัติจริง และยังอยู่ในช่วงเวลาที่แอดมินเปิดให้เข้าถึงข้อมูล (ไม่หมดอายุ)
5. ค้นหาคีย์ลับที่ถูกเข้ารหัสไว้ประจำ Log แถวนั้นในตาราง `log_sensitive_field_secrets`
6. ใช้คีย์หลักระดับแอปพลิเคชัน (`DataEncryptionKey`) ทำการถอดรหัสความลับผ่านขั้นตอนเข้ารหัสแบบสมมาตร AES-256
7. เขียนข้อมูลประวัติการถอดรหัสลงในตาราง `sensitive_log_access_histories` เพื่อเป็นหลักฐานสืบสวน
8. ส่งสตริงข้อมูลดั้งเดิมที่ถอดรหัสแล้วกลับไปเพื่อแสดงผลบนเว็บ

#### Input

- **Body (JSON)**: `request_id` (คำขอที่ได้รับอนุมัติแล้ว)

#### Output

- **Response (JSON)**: `decrypted_value` (ข้อมูลสตริงความลับตัวจริง)

#### Database / Storage Used

- **PostgreSQL**: ตาราง `sensitive_log_access_requests`, `sensitive_log_access_histories`, `log_sensitive_field_secrets`

---

### 5.9 Queue / Worker Flow

#### Purpose

ทำงานเป็น Background Worker คอยหยิบคิว Log จาก NATS มาตรวจสอบความสอดคล้อง ดำเนินการลบ/ปิดบังข้อมูลลับอัตโนมัติ (Masking) ก่อนบันทึกลงระบบสืบค้นข้อมูลหลัก (Elasticsearch Bulk) และคอยดูแลนโยบายวันหมดอายุของดัชนี (Retention Policy Maintenance)

#### Related Files

- `cmd/worker/main.go`
- `internal/background_working/worker.go` (ตัว Bootstrap)
- `internal/worker/usecase/usecase.go`
- `internal/worker/usecase/batch_processor.go`
- `internal/worker/usecase/validation.go`
- `internal/worker/usecase/elasticsearch_builder.go`
- `internal/archive_utils/archive.go`
- `models/log_archive.go`

#### Flow Diagram

```text
NATS JetStream Queue
  → Worker Daemon (เช็กคิวทุกช่วงวินาที)
  → Fetch batch of messages (ดึงข้อความจากคิว JetStream)
  → check schema rules & sensitive config definitions
  → Masking engine (ตรวจจับ ฟิลด์แฮช / ฟิลด์เข้ารหัสข้อมูลอ่อนไหว)
  → buildElasticDocument() (สร้าง ES Document & คีย์ลับลง Postgres)
  → Elasticsearch Bulk API (ส่งบันทึกเป็นก้อนใหญ่ใน ES)
  → BulkPersistSuccesses() (บันทึก Index Ref และคีย์ลับลง PostgreSQL สำเร็จ)
  → Acknowledge NATS messages & Broadcast to live tail NATS subject
```

#### Step-by-Step Explanation

1. Worker เริ่มทำงาน ดำเนินคำสั่งรันแบบ Loop ตรวจหาคิวใหม่ใน NATS JetStream
2. เมื่อได้รับข้อความในรอบประมวลผล ดำเนินการคลี่ข้อมูลชุด Log
3. วนลูปอ่านข้อมูลและนำโมเดลเช็กระดับ Feature/Category ว่าถูกต้องกับเงื่อนไขของระบบหรือไม่
4. ในกรณีที่มีการตั้งกฎ Sensitive Masking Rules ของ Product:
   - ฟิลด์ที่กำหนดความปลอดภัยจะถูกแฮชเป็น SHA-256 หรือเข้ารหัสเก็บไว้ใน Postgres `log_sensitive_field_secrets`
   - ในเนื้อหา Payload หลักของ Log จะถูกแทนที่ด้วยข้อความ Masking เช่น `[MASKED]` หรือสตริงที่อ่านยากเพื่อความปลอดภัยเมื่อเก็บใน Elastic
5. แปลงเอกสาร Log และจัดเตรียม Metadata
6. รันคำสั่ง Bulk ส่งบันทึกให้กับ Elasticsearch
7. เมื่อ Elastic บันทึกสำเร็จ:
   - บันทึกพิกัดของ Log (Index ID, Doc ID) ในตาราง `log_index_refs`
   - ทำการบันทึกคีย์ลับในตาราง `log_sensitive_field_secrets`
8. ส่งคำสั่ง Acknowledge (Ack) ยืนยันกับ NATS เพื่อนำเอาข้อความออกจากคิว และบรอดแคสต์ข้อมูล Log นั้นไปยังช่องสัญญาณเรียลไทม์ `omnilogs.logs.live.<product_id>` เพื่อให้ระบบ Live Tail ดึงไปแสดงผลได้ทันที
9. **ระบบการบำรุงรักษา (Periodic Maintenance)**: ทุก ๆ 30 วินาที Worker จะเข้าลูปเช็กอายุ Index ใน Elastic
   - ดึงรายชื่อ Index ที่มีอายุเกินกว่าค่า `retention_days` ที่ตั้งไว้ในระบบ
   - สั่งบล็อกการเขียนข้อมูลในดัชนีนั้นชั่วคราว (`blocks.write: true`)
   - เรียกระบบ `archive_utils.ArchiveIndex` เพื่อดาวน์โหลดและบีบอัดข้อมูลออกมาเขียนเป็นไฟล์เก็บในเซิร์ฟเวอร์ (เช่น data/archives/) พร้อมเขียนประวัติลงตาราง `log_archives`
   - ทำการลบ Index ที่หมดอายุออกจาก Elasticsearch อย่างปลอดภัย

#### Sensitive Data Masking Flow

Flow นี้มีหน้าที่ป้องกันไม่ให้ข้อมูลอ่อนไหวถูกจัดเก็บหรือแสดงผลใน Elasticsearch โดยตรง ขณะที่ยังสามารถเปิดดูค่าจริงได้เมื่อผู้ใช้มีสิทธิ์และได้รับอนุมัติตามกระบวนการ Sensitive Data Access

```text
Log Payload จาก NATS JetStream
  → อ่าน Sensitive Field Definitions และ Log Masking Rules ของ Product
  → เดินสำรวจข้อมูลแบบ Nested Object / Array พร้อมระบุ Field Path
  → ตรวจจับฟิลด์ที่ตรงตาม Key หรือ Path ที่กำหนดเป็น Sensitive
  → เข้ารหัสค่าจริงด้วย AES-256-GCM
  → บันทึก Secret แยกใน PostgreSQL พร้อม Log ID, Field Key และ Field Path
  → แทนค่าจริงใน Payload ด้วยค่า Mask เช่น [REDACTED]
  → ส่ง Payload ที่ถูก Mask ไปสร้างเอกสารใน Elasticsearch
```

ขั้นตอนการทำงานมีดังนี้:

1. Worker โหลดรายการฟิลด์อ่อนไหวที่เปิดใช้งาน (`is_sensitive = true` และ `is_active = true`) รวมถึงกฎ `LogMaskingRules` ของ Product และเก็บไว้ใน Cache เพื่อลดการอ่านข้อมูลซ้ำ
2. ระบบวนตรวจสอบ Payload แบบ Recursive ครอบคลุมทั้ง Object และ Array พร้อมสร้างตำแหน่งของฟิลด์ เช่น `user.password` หรือ `items[0].token`
3. หากชื่อฟิลด์หรือ Field Path ตรงกับ Sensitive Field Definition หรือ Masking Rule ระบบจะถือว่าเป็นข้อมูลอ่อนไหวและจะไม่ส่งค่าจริงเข้า Elasticsearch
4. ค่าจริงจะถูกเข้ารหัสด้วย `AES-256-GCM` และสร้าง `secret_id` สำหรับอ้างอิง จากนั้นเตรียมบันทึกลงตาราง `log_sensitive_field_secrets`
5. ค่าใน Payload ที่จะทำดัชนีจะถูกแทนด้วยค่า Mask ของ Rule หากมีการกำหนดไว้ หรือใช้ค่าเริ่มต้น `[REDACTED]`
6. ระบบส่ง Payload ที่ถูก Mask ไปยัง Elasticsearch ส่วนค่าจริงและ Metadata สำหรับค้นคืนจะถูกเก็บแยกใน PostgreSQL
7. การถอดรหัสทำได้ผ่าน Sensitive Data Access Flow เท่านั้น โดยต้องตรวจสอบสิทธิ์ คำขออนุมัติ และบันทึกประวัติการเปิดดูใน `sensitive_log_access_histories`

หลักการด้านความปลอดภัย:

- ห้ามเขียนค่าจริงของข้อมูลอ่อนไหวลงใน Elasticsearch, Live Tail, Error Log หรือ Audit Log
- หากการเข้ารหัสหรือการบันทึก Secret ล้มเหลว ต้องไม่ยืนยันข้อความ NATS สำเร็จ เพื่อให้ระบบ Retry และป้องกันข้อมูลสูญหาย
- การค้นคืนต้องอ้างอิงทั้ง `log_id` และ `field_path` เพื่อไม่ให้เปิดเผย Secret ผิดรายการ

#### Input

- **Source**: NATS JetStream messages (Queue payload)

#### Output

- **Destination**: Elasticsearch indexed documents, PostgreSQL ref records, ZIP archive files

#### Database / Storage Used

- **Queue**: NATS JetStream
- **Elasticsearch**: Indices ของ Log ตามรายวัน
- **PostgreSQL**: ตาราง `log_queue_batches`, `log_index_refs`, `log_sensitive_field_secrets`, `log_archives`
- **Local Disk**: พื้นที่สำหรับจัดเก็บไฟล์สำรอง Log บีบอัด (.zip/.json)

---

### 5.10 Monitoring Flow

#### Purpose

ตรวจวัดสถานะการประมวลผล สุขภาพของระบบคิว สรุปความเร็วและประสิทธิภาพของระบบ Log รายวัน เพื่อรายงานปัญหาให้ผู้ควบคุมตรวจสอบได้ทันท่วงที

#### Related Files

- `routes/health_routes.go`
- `routes/dashboard_routes.go`
- `internal/dashboard/handler/handler.go`
- `internal/dashboard/usecase/usecase.go`
- `internal/dashboard/repository/repository.go`

#### Flow Diagram

```text
System Administrator / DevOps
  → GET /health/ready
  → Health Router Exec readyCheck()
  → Ping PostgreSQL & Ping Elasticsearch
  → Return HTTP 200 (Ready) or HTTP 503 (Service Unavailable)
```

#### Step-by-Step Explanation

1. ระบบมอนิเตอร์ภายนอก (เช่น Kubernetes Probe, Prometheus หรือ Uptime Kuma) ยิงขอข้อมูลมายัง Endpoint `/health/live` หรือ `/health/ready`
2. ระบบเช็กเส้นทางสด (Live): ส่งคืนผลลัพธ์ตกลงทันทีเพื่อบอกว่าโปรแกรมยังทำงานอยู่
3. ระบบเช็กความพร้อมทำงาน (Ready): เรียกใช้งานฟังก์ชัน `readyCheck`
4. รันคำสั่งตรวจสอบสถานะฐานข้อมูล PostgreSQL (`PingContext`) และสถานะคลัสเตอร์ Elasticsearch (`Info()`)
5. หากเช็กผ่านทั้งหมด จะส่งคืนสถานะ HTTP 200 OK และตอบว่า ready
6. หากตรวจพบว่ามีจุดใดตอบช้าหรือขัดข้อง จะส่ง HTTP 503 Service Unavailable และแจกแจงข้อผิดพลาดกลับไป
7. นอกจากนี้หน้าแดชบอร์ดมีระบบดึงสถิติความเร็วการบันทึก Log รายวัน ผ่าน API `/api/v1/dashboard/stats` และดึงประวัติ Audit logs ทั้งหมดด้วยการดึงข้อมูลจากตาราง `system_audit_logs`

#### Input

- **HTTP GET Request**: `/health/ready` / `/api/v1/dashboard/stats`

#### Output

- **Response (JSON)**: ค่าความพร้อมทำงานของ Database / ข้อมูลสถิติกราฟ Logs / ข้อมูลรายการ Audit Logs

#### Database / Storage Used

- **PostgreSQL**: ตาราง `system_audit_logs` (สืบค้นประวัติระบบ)
- **Elasticsearch**: ข้อมูลคลัสเตอร์สถานะการตอบรับ (Info check)

---

## 6. Example Output Format

สำหรับสถาปัตยกรรมและรายละเอียดต่าง ๆ ทางผู้พัฒนาได้ทำการแสดงตัวอย่างการเรียกใช้งาน Flow และระบุพารามิเตอร์ Input/Output ของ Features หลัก ๆ ไว้ครบถ้วนในหัวข้อที่ 5 ของเอกสารนี้แล้ว เพื่อความสะดวกรวดเร็วในการอ่านและตามรอยโค้ด

---

## 7. Final Summary

ระบบ Backend ของ OmniLogs ได้รับการออกแบบโครงสร้างการทำงานแบบแยกส่วนหน้าที่ (Separation of Concerns) ตามหลักของ Clean Architecture โดยแยกกระบวนการทำงานหลักออกเป็นกลุ่มโมดูลภายใต้ไดเรกทอรี `internal/`

1. **การรับและส่งผ่านข้อมูล (Ingestion & Queue)**: ใช้สถาปัตยกรรมแบบ Event-driven โดยฝั่ง API Server จะทำตัวเป็นผู้ผลิตสาร (Producer) เพื่อรับความต้องการและส่งเข้าคิว NATS JetStream ทันที เพื่อป้องกันไม่ให้เว็บเซิร์ฟเวอร์เกิดสภาวะคอขวดและสามารถตอบรับผลกลับหาผู้ส่งได้อย่างรวดเร็ว
2. **การทำงานเบื้องหลัง (Background Worker)**: ทำงานในรูปแบบ Daemon แยกอิสระ ดึงข้อมูลจากคิวแบบเป็นกลุ่ม (Batch) เพื่อจัดการระบบความปลอดภัยของข้อมูล (Masking Secrets) เข้ารหัสข้อมูลและนำส่งไปสร้างดัชนีใน Elasticsearch ทำให้มีความสามารถประมวลผลสูง (High Throughput)
3. **การเข้าถึงและรักษาความปลอดภัย (Auth & Sensitive Decryption)**: การควบคุมความปลอดภัยแยกสองชั้น ชั้นภายนอกใช้ JWT และชั้นภายในใช้ตรรกะ Product Roles/Permissions (RBAC) ร่วมกับกลไกถอดรหัสที่มีระบบเก็บประวัติ Audit Log อย่างเข้มงวด
4. **ความเข้ากันได้ของระบบ (Compatibility)**: โครงสร้าง Backend นี้สนับสนุนประสิทธิภาพที่ยืดหยุ่น การแบ่งเลเยอร์ Handler -> Usecase -> Repository ช่วยให้ทีมงานพัฒนาระบบสามารถทำการเขียน Unit Test หรือเปลี่ยนย้ายระบบย่อยได้ง่ายในภายหลังโดยไม่ส่งผลกระทบต่อส่วนอื่นของระบบ (Non-breaking Change)
