# Intern Go API

Go API ตัวอย่างสำหรับเด็กฝึกงานเรียนรู้ backend architecture โดยใช้ Gin, GORM, PostgreSQL, JWT, bcrypt, rate limit และ upload provider แบบ mock/R2

โปรเจกต์นี้ตั้งใจให้ไล่ flow ง่าย ไม่ซ่อน logic มากเกินไป และแยกความรับผิดชอบชัดเจนตามแนวคิด:

```text
route -> middleware -> handler -> usecase -> repository/service
```

## Project Structure

```text
api/
  main.go
  configs/
    env.go
    db.go
  routes/
    auth_routes.go
    todo_routes.go
    upload_routes.go
  internal/
    auth/
      handler/
      usecase/
      repository/
    todos/
      handler/
      usecase/
      repository/
    uploads/
      handler/
      usecase/
      service/
  middleware/
    auth.go
    rate_limit.go
    request_id.go
  models/
    user.go
    todo.go
  dto/
    auth.go
    todo.go
    upload.go
  utils/
    response.go
    jwt.go
    password.go
```

## Folder Responsibilities

`main.go`

เริ่มต้น application เท่านั้น: load env, connect database, AutoMigrate, create Gin router, register global middleware และเรียก register routes

`routes/`

เป็นจุดประกอบ dependency ของแต่ละ feature เช่นสร้าง repository, usecase, handler แล้วผูก endpoint เข้ากับ handler เพื่อให้ `main.go` สะอาดที่สุด

`handler/`

รับผิดชอบ HTTP layer เช่น bind request, validate input, อ่าน path/query/body/form-data, อ่าน user จาก context และ map error เป็น response

`usecase/`

รับผิดชอบ business logic เช่นตรวจ password, สร้าง token, ตรวจ todo ownership และเรียก repository/service ผ่าน interface

`repository/`

รับผิดชอบ database query เท่านั้น ไม่ควรรู้เรื่อง Gin, HTTP status code หรือ business rule

`service/`

รับผิดชอบ external provider เช่น mock uploader หรือ Cloudflare R2 uploader

`dto/`

เก็บ request/response shape ที่ API ใช้คุยกับ client

`models/`

เก็บ GORM database models

`middleware/`

เก็บ middleware เช่น auth, rate limit และ request id

`utils/`

เก็บ helper กลาง เช่น response envelope, JWT และ password hashing

## Run PostgreSQL

PostgreSQL อยู่ใน folder `sourse-data`

```bash
docker network create external-local-net
cd ../sourse-data
docker compose up -d
```

ค่า Docker Compose หลัก:

```text
container: intern-postgres
image: postgres:16
database: intern_api
username: postgres
password: postgres
port: 5432:5432
network: external-local-net
```

## Run API

```bash
cd ../api
cp .env.example .env
go mod tidy
go run .
```

Default URL:

```text
http://localhost:8903
```

## Environment Summary

Application:

```text
APP_PORT=8903
```

Database:

```text
DB_HOST=localhost
DB_PORT=5432
DB_NAME=intern_api
DB_USERNAME=postgres
DB_PASSWORD=postgres
```

JWT:

```text
JWT_SECRET=change-me
ACCESS_TOKEN_EXPIRE_SECONDS=900
REFRESH_TOKEN_EXPIRE_SECONDS=604800
```

Rate limit:

```text
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=60
RATE_LIMIT_WINDOW_SECONDS=60
AUTH_RATE_LIMIT_REQUESTS=10
AUTH_RATE_LIMIT_WINDOW_SECONDS=60
UPLOAD_RATE_LIMIT_REQUESTS=20
UPLOAD_RATE_LIMIT_WINDOW_SECONDS=60
```

Upload:

```text
UPLOAD_PROVIDER=mock
UPLOAD_MOCK_BASE_URL=https://mock-upload.local/files
```

Cloudflare R2:

```text
R2_ENDPOINT=
R2_REGION=auto
R2_ACCESS_KEY_ID=
R2_SECRET_ACCESS_KEY=
R2_BUCKET=
R2_PUBLIC_BASE_URL=
R2_OBJECT_PREFIX=local/uploads
```

## Packages

`github.com/gin-gonic/gin`

HTTP server, routing, middleware, JSON binding และ multipart upload

`github.com/go-playground/validator/v10`

อ่าน validation error จาก Gin binding แล้วแปลงเป็น field-level error details

`gorm.io/gorm`

ORM สำหรับ query database และ AutoMigrate

`gorm.io/driver/postgres`

PostgreSQL driver สำหรับ GORM

`github.com/joho/godotenv`

โหลด `.env` ตอน local development

`github.com/golang-jwt/jwt/v5`

สร้างและตรวจ JWT access/refresh token

`golang.org/x/crypto/bcrypt`

hash password และตรวจ password ตอน login

`github.com/aws/aws-sdk-go-v2`

AWS SDK core type เช่น `aws.Config`

`github.com/aws/aws-sdk-go-v2/credentials`

สร้าง static credentials สำหรับ R2

`github.com/aws/aws-sdk-go-v2/service/s3`

upload file ไป Cloudflare R2 ผ่าน S3-compatible API

หมายเหตุ: packages ที่เป็น `// indirect` ใน `go.mod` คือ dependency ลูกของ package หลัก ไม่ได้ import เองโดยตรง

## Database Models

`User`

```text
id
name
email unique
passwordHash
createdAt
updatedAt
deletedAt
```

`Todo`

```text
id
userId
title
description
isDone
createdAt
updatedAt
deletedAt
```

ทั้งสอง model ใช้ `gorm.Model` จึงมี `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt` ให้โดยอัตโนมัติ และ delete จะเป็น soft delete

## Response Contract

ทุก response ใช้ envelope กลางเพื่อให้ frontend handle ได้ง่าย

Object success:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "Intern"
  }
}
```

List success:

```json
{
  "success": true,
  "data": [],
  "meta": {
    "page": 1,
    "perPage": 10,
    "total": 0,
    "totalPage": 0
  }
}
```

Error:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "request validation failed",
    "status": 400,
    "details": [
      {
        "field": "email",
        "rule": "email",
        "message": "email must be a valid email"
      }
    ]
  }
}
```

หลักการ:

```text
data คือ payload โดยตรง เป็น object หรือ array ก็ได้
meta ใช้เฉพาะ list response สำหรับ pagination
error.code ใช้ UPPER_SNAKE_CASE และควร stable
error.message เป็นข้อความอ่านโดยคน
error.status คือ HTTP status code
error.details ใช้ใส่ validation field errors หรือรายละเอียดเพิ่มเติม
X-Request-ID อยู่ใน response header ไม่อยู่ใน body
```

## Auth APIs

`POST /auth/register`

Public endpoint ใช้ auth rate limit, validate name/email/password, hash password ด้วย bcrypt และ return user โดยไม่ส่ง `passwordHash`

```bash
curl -X POST http://localhost:8903/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"Intern","email":"intern@example.com","password":"password123"}'
```

`POST /auth/login`

ตรวจ email/password และ return access token กับ refresh token

```bash
curl -X POST http://localhost:8903/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"intern@example.com","password":"password123"}'
```

`POST /auth/refresh-token`

รับ refresh token และออก access token ใหม่ โดยรับเฉพาะ token ที่มี `tokenType=refresh`

```bash
curl -X POST http://localhost:8903/auth/refresh-token \
  -H 'Content-Type: application/json' \
  -d '{"refreshToken":"<refreshToken>"}'
```

`GET /auth/me`

Protected endpoint ต้องส่ง access token

```bash
curl http://localhost:8903/auth/me \
  -H 'Authorization: Bearer <accessToken>'
```

## Todo APIs

ทุก todo endpoint เป็น protected endpoint และ user เห็น/แก้ไขได้เฉพาะ todo ของตัวเอง

`GET /todos?page=1&perPage=10`

ดึง todo แบบ pagination

```bash
curl 'http://localhost:8903/todos?page=1&perPage=10' \
  -H 'Authorization: Bearer <accessToken>'
```

`GET /todos/:todoId`

ดึง todo รายการเดียว

```bash
curl http://localhost:8903/todos/1 \
  -H 'Authorization: Bearer <accessToken>'
```

`POST /todos`

สร้าง todo

```bash
curl -X POST http://localhost:8903/todos \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <accessToken>' \
  -d '{"title":"Learn Go architecture","description":"Trace route to repository"}'
```

`PUT /todos/:todoId`

แก้ไข todo

```bash
curl -X PUT http://localhost:8903/todos/1 \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <accessToken>' \
  -d '{"title":"Learn usecase layer","isDone":true}'
```

`DELETE /todos/:todoId`

ลบ todo แบบ soft delete

```bash
curl -X DELETE http://localhost:8903/todos/1 \
  -H 'Authorization: Bearer <accessToken>'
```

## Upload API

`POST /uploads/file`

Protected endpoint ใช้ upload-specific rate limit รับ multipart form field ชื่อ `file`

```bash
curl -X POST http://localhost:8903/uploads/file \
  -H 'Authorization: Bearer <accessToken>' \
  -F 'file=@./example.png'
```

Mock response:

```json
{
  "success": true,
  "data": {
    "fileName": "example.png",
    "fileUrl": "https://mock-upload.local/files/<generated-file-name>",
    "contentType": "image/png",
    "size": 12345,
    "provider": "mock"
  }
}
```

R2 object key format:

```text
{R2_OBJECT_PREFIX}/{yyyy}/{mm}/{dd}/{uuid}-{safe-file-name}
```

## JWT Design

JWT claims:

```text
userId
email
tokenType
exp
iat
```

`tokenType` มี 2 ค่า:

```text
access
refresh
```

`AuthMiddleware` รับเฉพาะ access token ส่วน refresh endpoint รับเฉพาะ refresh token

โปรเจกต์นี้ไม่ได้เก็บ refresh token ใน DB เพื่อให้ง่ายต่อการเรียนรู้ ถ้าต่อยอด production ควรเพิ่ม refresh token storage, token rotation และ revoke flow

## Rate Limit Design

ใช้ in-memory fixed window rate limit

```text
global: ใช้กับทุก route
auth: ใช้กับ /auth/register, /auth/login, /auth/refresh-token
upload: ใช้กับ /uploads/file
```

Key ใช้:

```text
client IP + route group
```

ถ้าเกิน limit จะตอบ HTTP 429:

```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "too many requests",
    "status": 429
  }
}
```

Response headers:

```text
X-RateLimit-Limit
X-RateLimit-Remaining
X-RateLimit-Reset
```

## Add New API Workflow

1. สร้าง DTO request/response ใน `dto/`
2. ถ้าต้อง query DB ให้เพิ่ม repository interface และ implementation
3. เพิ่ม usecase interface และ business logic
4. เพิ่ม handler method สำหรับ bind/validate/response
5. เพิ่ม route ใน `routes/<feature>_routes.go`
6. ประกอบ repository/usecase/handler ใน route file
7. เลือก middleware ว่า public, protected, global rate limit หรือ custom rate limit
8. เพิ่ม test หรือ curl example

ตัวอย่าง API ง่ายที่ไม่ต้องมี repository/usecase:

```go
router.GET("/health", func(c *gin.Context) {
	utils.Success(c, 200, gin.H{"status": "ok"})
})
```

## Layer Rules

Handler ควรทำ:

```text
bind request
validate request
อ่าน path/query/body/form-data
อ่าน userId จาก context
เรียก usecase
map error เป็น HTTP response
```

Usecase ควรทำ:

```text
business rule
permission/ownership rule
เรียก repository หรือ service ผ่าน interface
คืน error ที่ handler map ต่อได้
```

Repository ควรทำ:

```text
database query เท่านั้น
ไม่รู้จัก Gin context
ไม่รู้จัก HTTP status code
ไม่ตัดสิน business rule
```

Service ควรทำ:

```text
คุยกับ external provider
ซ่อนรายละเอียด provider ไว้หลัง interface
ทำให้เปลี่ยน mock/r2 ได้จาก env
```

## Why Wire Dependencies In Routes

โปรเจกต์นี้ให้ `routes` เป็นจุดประกอบ dependency ของแต่ละ feature:

```text
route creates repository
route creates usecase
route creates handler
route binds endpoint
```

ข้อดีคือ `main.go` สั้นและอ่านง่าย แต่ละ route file เห็น dependency ของ feature ตัวเองครบ และเพิ่ม feature ใหม่ได้โดยไม่ทำให้ `main.go` ใหญ่ขึ้น

## Production Notes

ถ้าจะต่อยอดใช้งานจริง ควรพิจารณาเพิ่ม:

```text
structured logger
database migration tool เช่น golang-migrate
unit tests สำหรับ usecase
integration tests สำหรับ repository
Redis rate limit สำหรับหลาย instance
refresh token storage และ revoke flow
CORS config
graceful shutdown
Dockerfile สำหรับ API
CI pipeline
```
