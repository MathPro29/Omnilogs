# Omnilogs Frontend Developer Guide

คู่มือนี้อธิบายโครงสร้างและแนวทางพัฒนา frontend ของ Omnilogs ซึ่งเป็น React application สำหรับ dashboard การจัดการผู้ใช้ การสำรวจ logs, live tail และ audit logs

## เริ่มต้นใช้งาน

รันคำสั่งจากโฟลเดอร์ `frontend/`

```bash
npm install
npm run dev       # Vite: http://localhost:3000
npm run build     # ตรวจ TypeScript และ build production
npm run lint      # ตรวจ ESLint
npm run preview   # preview build ที่สร้างแล้ว
```

หากไม่กำหนด `VITE_API_BASE_URL` ระบบ auth, dashboard และ users จะใช้ข้อมูล mock ใน development โดย login ได้ด้วย `admin / admin123` บริการ audit logs ไม่มี mock mode และต้องเชื่อมต่อ backend

## สถาปัตยกรรม

```text
Browser
  -> App providers (TanStack Query, Ant Design, React Router)
  -> AppRoutes -> GuestRoute / ProtectedRoute -> AuthLayout หรือ AdminLayout
  -> Page -> Service -> apiClient -> Omnilogs API
                         |
                         +-> Zustand ใช้สำหรับ session และ UI state
```

### เทคโนโลยีหลัก

| ส่วน | เทคโนโลยี |
| --- | --- |
| Framework | React 19, TypeScript, Vite 6 |
| UI | Ant Design 6, Tailwind CSS 4 |
| Routing | React Router DOM 7 |
| Server state | TanStack Query 5 |
| Client/session state | Zustand 5 |
| HTTP | Axios |
| Validation | Zod 4 |
| Date | Day.js |
| Icons | Heroicons, `@ant-design/icons`, Lucide React |
| Motion/loading | Motion, `react-top-loading-bar` |

## โครงสร้างไฟล์

```text
src/
  main.tsx                 # entry point
  App.tsx                  # providers และ BrowserRouter
  api/client.ts            # Axios, bearer token และ 401/403 handling
  app/theme.ts             # Ant Design theme และ locale ภาษาไทย
  components/
    global/                # shell/navigation, transition, loading และ permission UI
    features/              # components ที่ใช้ร่วมกับ feature เฉพาะ
  constants/index.ts       # routes, permissions, endpoints, query keys และค่ากลาง
  features/                # จุดรวม feature ที่กำลังขยายในอนาคต
  hooks/                   # usePermission และ useLoadingBar
  layouts/                 # AuthLayout และ AdminLayout
  pages/                   # route-level screens
  routes/                  # route config, menu config และ guards
  schemas/                 # Zod schemas สำหรับ auth และ forms
  services/                # service layer ที่เรียก API หรือ mock ตาม config
  store/                   # auth, UI และ app Zustand stores
  styles/                  # global, layout และ page-scoped CSS
  types/                   # shared TypeScript types
  utils/                   # permission helpers และ formatters
```

## Routes และหน้าจอ

`src/routes/AppRoutes.tsx` เป็นจุดประกอบ routes ส่วน `src/routes/routes.tsx` เก็บ protected admin routes และ `src/routes/menu.tsx` เก็บรายการ sidebar

| กลุ่ม | Routes |
| --- | --- |
| Guest | `/login`, `/register`, `/forgot-password` |
| Application | `/dashboard`, `/users`, `/settings`, `/logs-live`, `/logs-explorer`, `/audit-logs` |
| Error | `/403`, `/404` และ fallback `*` |

`AuthLayout` ใช้กับหน้าสำหรับผู้ยังไม่ login และ `AdminLayout` ใช้กับหน้าภายในระบบที่มี sidebar/header การเพิ่มหน้าใหม่ต้องเพิ่ม route ใน `routes.tsx` และเพิ่ม menu ใน `menu.tsx` หากต้องการให้ค้นพบได้จาก navigation อย่าเพิ่ม route ไว้เพียงฝั่งเดียว

## Data และ API

- Component และ page ห้ามเรียก Axios โดยตรง ให้เรียกผ่าน service ใน `src/services/`
- ใช้ `apiClient` จาก `src/api/client.ts` เพื่อให้ได้ base URL, timeout 30 วินาที และ bearer token interceptor
- `401` จะล้าง auth state แล้ว redirect ไป `/login`; `403` จะ redirect ไป `/403`
- ใช้ response envelope ของ backend (`success`, `message`, `data`, `meta`) ตาม handler/OpenAPI จริง ห้ามเดารูปแบบ pagination ใหม่
- เพิ่ม endpoint ที่ `API_ENDPOINTS` และเพิ่ม query key ที่ `QUERY_KEYS` ใน `src/constants/index.ts`
- ใช้ TanStack Query สำหรับข้อมูลจาก server รวม filter และ scope ทั้งหมดไว้ใน `queryKey`; หลัง mutation ให้ invalidate query ที่เกี่ยวข้อง
- ใช้ Zustand เฉพาะ session/UI state ไม่คัดลอกข้อมูลจาก query มาเก็บซ้ำใน store

บริการที่มีอยู่คือ `authService`, `dashboardService`, `userService` และ `auditService` โดย backend login จะเรียก `/auth/login` แล้วเรียก `/auth/me` เพื่อประกอบ user model ฝั่ง frontend

## Authorization และ scope

การซ่อนเมนูหรือปุ่มเป็นเพียง usability; backend เป็น security boundary เสมอ ให้ใช้ permission helper (`can`, `canAny`, `canAll`) และ `PermissionGuard`/`ProtectedRoute` แทนการตรวจ role name กระจัดกระจาย

```tsx
<PermissionGuard permission="user:edit">
  <Button>แก้ไขผู้ใช้</Button>
</PermissionGuard>
```

ต้องจัดการ state ให้แยกกันชัดเจน ได้แก่ loading, empty, error, forbidden, ไม่มี scope, membership หมดอายุ และ feature ถูกปิดใช้งาน สำหรับข้อมูลที่ผูกกับ product/project/environment ให้ใส่ scope ใน query key และ invalidate/reset ข้อมูลเมื่อเปลี่ยน scope

## เพิ่ม feature หรือหน้าใหม่

1. เพิ่ม type ใน `src/types/` และ schema ใน `src/schemas/` หากมี form/request validation
2. สร้าง service ใน `src/services/<domain>.service.ts` และ export จาก `src/services/index.ts`
3. เพิ่ม endpoint/query key ใน `src/constants/index.ts`
4. สร้าง page ใน `src/pages/` และ export จาก `src/pages/index.ts`
5. เพิ่ม route ใน `src/routes/routes.tsx` พร้อม permission ที่จำเป็น
6. เพิ่ม menu ใน `src/routes/menu.tsx` หากหน้าต้องแสดงใน sidebar
7. เพิ่ม component ที่ใช้ร่วมกันไว้ใน `components/global/`, component ของ feature ไว้ใน `components/features/` หรือ `features/<name>/`
8. เพิ่ม CSS เฉพาะหน้าใน `src/styles/pages/` เมื่อ theme และ utility classes ไม่เพียงพอ

ทุก async screen ต้องมี loading, empty, error/retry และ permission-denied state ตารางข้อมูลต้องรองรับ narrow viewport ด้วยการ scroll หรือ detail drawer และ filter ที่เปลี่ยนต้อง reset pagination

## Environment และความปลอดภัย

Frontend อ่านค่าผ่าน `import.meta.env` เท่านั้น

| ตัวแปร | ค่าเริ่มต้น | ความหมาย |
| --- | --- | --- |
| `VITE_API_BASE_URL` | `/api` ใน `apiClient` | base URL สำหรับ API; หากไม่กำหนด service ที่รองรับจะใช้ mock mode |

ห้ามอ่าน แสดงผล หรือ commit `.env` และห้ามใส่ token, password, secret หรือ raw sensitive log ลง console, analytics หรือ client-side storage

## มาตรฐานโค้ดและ UI

- เปิดใช้ TypeScript strict และหลีกเลี่ยง `any`; ใช้ named exports ยกเว้น entry point หรือ convention ของ third party
- ใช้ Ant Design เป็น component หลัก และใช้ Tailwind สำหรับ layout/spacing; token สีและขนาดกลางอยู่ใน `src/app/theme.ts` หรือ CSS variables
- ใช้ `Noto Sans Thai Looped` และข้อความที่ผู้ใช้เห็นควรเป็นภาษาไทยให้สอดคล้องกับ locale ที่ตั้งไว้
- ใช้ `useMemo`/`useCallback` เมื่อมีการคำนวณหนัก ต้องรักษา reference ให้ memoized child หรือเป็น dependency ที่จำเป็น ไม่ใช้เป็นค่าเริ่มต้นทุก component
- ปุ่มที่มีผลกระทบสูงต้องมี confirmation และแสดงผลสำเร็จ/ผิดพลาดอย่างชัดเจน
- icon-only control ต้องมี `aria-label`, ทุก form ต้องมี label และทุกการทำงานต้องใช้ keyboard ได้
- รักษา white-first visual system, contrast ที่อ่านง่าย, focus state ที่เห็นได้ และรองรับ `prefers-reduced-motion`
- ห้าม log token, password, encryption material, audit secret, decrypted value หรือ raw sensitive log payload

## Verification ก่อนส่งงาน

จาก `frontend/` ให้รัน:

```bash
npm run build
npm run lint
```

เมื่อเปลี่ยน route ให้ตรวจ guest/authenticated/forbidden/direct URL และเมื่อเปลี่ยน API ให้ตรวจ endpoint, envelope, pagination, auth header, error handling และ scope isolation กับ backend contract หากเปลี่ยน API contract ให้แก้ source contract ฝั่ง backend และรัน OpenAPI validation/sync ตามคู่มือ backend ห้ามแก้ generated spec ด้วยมือ

