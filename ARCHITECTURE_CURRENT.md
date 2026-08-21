# Omnilogs — Current Architecture

ปัจจุบัน Omnilogs ใช้สถาปัตยกรรมแบบ **Modular Monolith ร่วมกับ Event-driven Worker** โดยมี React SPA เป็น frontend, Go/Gin เป็น backend API และแยกงานประมวลผล log แบบ asynchronous ไปยัง worker

## ภาพรวมระบบ

```text
ผู้ใช้งาน
  → React Frontend (React + TypeScript)
  → Go API (Gin)
  → Middleware (JWT / Permission / Audit / Rate Limit / Timeout)
  → Handler
  → Usecase
  → Repository
  → PostgreSQL / Elasticsearch
  → JSON Response
  → React Frontend
```

## Data และ Infrastructure

```text
Go API
  → PostgreSQL
      (User, Role, Membership, Permission, Scope,
       Policy, Queue Metadata, Audit)

Go API
  → Elasticsearch
      (Searchable Logs, Live Tail, Dashboard Aggregations)

Go API
  → NATS JetStream
      (Ingestion Messages)

Go Worker
  → Shared Archive Volume
      (Archived Log Files)
```

## Log Ingestion Flow

```text
App Logs / API Client
  → Omnilogs Ingestion API
  → Authenticate และ Validate Request
  → สร้าง Batch และ Queue Item ใน PostgreSQL
  → Publish Message ไปยัง NATS JetStream
  → Go Ingestion Worker
  → Validate Payload
  → Extract JSON Path / Custom Fields
  → Transform Data
  → Mask Sensitive Fields
  → Build Elasticsearch Document
  → Bulk Index เข้า Elasticsearch
  → อัปเดตสถานะใน PostgreSQL
      ├─ Completed
      ├─ Retry (กรณี temporary error)
      └─ Failure (กรณี permanent error)
```

## Log Search Flow

```text
ผู้ใช้งาน
  → React Frontend
  → GET /api/v1/logs พร้อม JWT และ Filters/Scope
  → Go API
  → ตรวจ Authentication และ Effective Permissions
  → Log Usecase
  → Scoped Elasticsearch Query
  → Logs / Pagination / Aggregations
  → บันทึก Audit ตามเงื่อนไข
  → JSON Response กลับ Frontend
```

## Backend Layer

```text
Routes
  → Middleware
  → Handler (HTTP parsing และ response mapping)
  → Usecase (Business rules และ authorization)
  → Repository (PostgreSQL / Elasticsearch access)
  → External Data Stores
```

Backend เป็น modular monolith แบ่งตาม domain เช่น authentication, products, projects, environments, scopes, memberships, dashboard, logs, audit, sensitive logs, custom fields, features และ archive

## Frontend Layer

```text
App Providers
  → Browser Router
  → ProtectedRoute / Permission Guard
  → Layouts (AuthLayout / AdminLayout)
  → Pages
  → Feature และ Shared Components
  → TanStack Query
  → Domain Services
  → Axios API Client
  → Backend API
```

- TanStack Query ใช้จัดการ server state, cache และ invalidation
- Zustand ใช้สำหรับ session และ client/UI state
- Ant Design และ Tailwind ใช้เป็น component และ styling system
- Backend authorization เป็น security boundary; frontend guards มีไว้ช่วยด้าน UX

## สรุปสั้น

```text
App Logs
  → Omnilogs API
  → PostgreSQL Queue
  → NATS JetStream
  → Worker (Validate → Transform → Mask)
  → Elasticsearch (Searchable Logs)
  → React Frontend (Search / Dashboard / Live Tail)
```
