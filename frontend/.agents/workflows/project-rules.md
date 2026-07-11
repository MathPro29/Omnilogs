---
description: Project rules and conventions for HR Admin Panel development
---

# HR Admin Panel - Development Rules

// turbo-all

## Tech Stack Rules

- Frontend: React + Vite + TypeScript (SPA)
- UI Components: Ant Design (antd) - ทุก component ต้องใช้ theme ที่ `src/app/theme.ts`
- Styling: Tailwind CSS สำหรับ layout, spacing, responsive
- Routing: React Router (nested layout, protected routes)
- Form Validation: Zod
- Date/Time: Day.js
- Icons: Heroicons (ห้ามใช้ Emoji)
- Animation: Motion for React
- API Client: Axios (พร้อม interceptors)
- Server State: TanStack Query
- Client State: Zustand
- Loading: react-top-loading-bar สำหรับ route change + Skeleton สำหรับ data loading

## Theme System (src/app/theme.ts)

ทุก antd component ต้องใช้ theme ที่กำหนดไว้ใน `src/app/theme.ts` ห้าม hardcode สี/ขนาด ลงในแต่ละ component โดยตรง

### Colors

```text
Primary:   #6647F0
Sidebar:   #100446, #1C0B68, #2A158A
Purple:    #3C22AC, #4F33CE, #7B68EE, #8D74FF, #AE9CFF, #CFC4FF, #F0EDFF
Error:     #FF3D89
Success:   #00BE8F
Warning:   #D4B106
Info:      #1890FF
Text:      #000000D9 (primary), #00000073 (secondary)
```

### Control Sizes

- Input, Select, InputNumber, DatePicker, Button (ขนาดกลาง): `controlHeight: 36px`
- TextArea: `min-height: 80px` (กำหนดผ่าน CSS)

### Input Focus/Error Behavior

- Focus: border เปลี่ยนเป็น `#6647F0` + outline shadow `rgba(102, 71, 240, 0.12)`
- Hover: border เปลี่ยนเป็น `#AE9CFF`
- Error: outline shadow `rgba(255, 61, 137, 0.15)`

## State Management Rules

1. **TanStack Query** สำหรับ **server state เท่านั้น**: API data, lists, details, mutations
2. **Zustand** สำหรับ **client state เท่านั้น**: auth, UI preferences, permissions
3. ห้ามใช้ Zustand เก็บ API data list/detail โดยตรง
4. แบ่ง Zustand store ตาม concern: auth.store, ui.store, app.store

### TanStack Query Usage Pattern

```typescript
// Query: อ่านข้อมูล (GET)
const { data, isLoading } = useQuery({
  queryKey: [...QUERY_KEYS.XXX, filters],
  queryFn: () => someService.getData(filters),
});

// Mutation: เปลี่ยนแปลงข้อมูล (POST/PUT/DELETE)
const mutation = useMutation({
  mutationFn: someService.createData,
  onSuccess: () => {
    message.success('สำเร็จ');
    queryClient.invalidateQueries({ queryKey: QUERY_KEYS.XXX });
  },
});
```

**Query Keys ที่ใช้:** `QUERY_KEYS.DASHBOARD_SUMMARY`, `QUERY_KEYS.USERS`, `QUERY_KEYS.USER_DETAIL(id)`, `QUERY_KEYS.ROLES`, `QUERY_KEYS.SETTINGS`

## Authentication / Authorization Rules

1. Client เป็น UX Layer, API เป็น Security Layer
2. ใช้ permission เป็นแกนหลักของ authorization ฝั่ง UI
3. role เป็นเพียงชุดรวมของ permissions
4. route guard, menu guard, action guard ต้องใช้ permission ชุดเดียวกัน
5. ทุก endpoint ฝั่ง API ต้องตรวจสิทธิ์ซ้ำ
6. ProtectedRoute ตรวจ: ยังไม่ login -> /login, ไม่มี permission -> /403

## Permission Helpers

- `can(permission)` - ตรวจ permission เดียว
- `canAny(permissions)` - ตรวจว่ามีอย่างน้อยหนึ่ง
- `canAll(permissions)` - ตรวจว่ามีทุกตัว
- `<PermissionGuard>` component สำหรับ conditional rendering

## Route & Menu System Rules

1. ระบบ routing เป็น **data-driven** - config อยู่ที่ `src/routes/`
2. Route config อยู่ที่ `src/routes/routes.tsx` - เพิ่มหน้าใหม่ที่นี่
3. Menu config อยู่ที่ `src/routes/menu.tsx` - เพิ่ม/แก้ sidebar ที่นี่
4. Menu helpers อยู่ที่ `src/routes/helpers.ts` (findMenuByPath, getFlatMenuItems)
5. Route rendering อยู่ที่ `src/routes/AppRoutes.tsx` (ไม่ต้องแก้)
6. **App.tsx เหลือแค่ providers** ไม่มี route logic
7. ทุก menu item ต้องมี requiredPermissions (undefined/[] = ทุก role เห็น)
8. ใช้ filterMenuByPermissions() กรอง menu ตาม user permissions
9. รองรับ nested submenu

## Loading Rules

1. ทุกหน้าที่มีการ fetch API ต้องใช้ **Skeleton** loading
2. ใช้ **react-top-loading-bar** สำหรับ route change
3. ใช้ **Motion** สำหรับ page transition และ animation ให้ลื่นไหล
4. Skeleton ต้องตรงกับ layout จริงของหน้า (DashboardSkeleton, TableSkeleton, FormSkeleton, DetailSkeleton)

## UI / Design Rules

1. ใช้ antd เป็น component base - **ต้อง**ใช้ theme จาก `src/app/theme.ts` เท่านั้น
2. ใช้ Tailwind CSS สำหรับ layout, spacing, responsive
3. ใช้ Motion เฉพาะจุดที่ช่วยให้ UI ลื่น
4. เน้น clean, professional, enterprise-style UI
5. Light theme เป็นหลัก
6. **ห้ามใช้ Emoji ทุกกรณี** -> ใช้ Heroicons หรือ Ant Design Icons แทน
7. ตอบกลับและเข้าใจบริบทเป็นภาษาไทยเท่านั้น
8. Font หลัก: **Noto Sans Thai** (Google Fonts)
9. ห้ามกำหนดสีให้ scrollbar -> ใช้ default ของ browser
10. ห้าม hardcode สี ใน component -> ใช้ CSS variable หรือ theme token

## Route Layout Rules

1. Login -> ใช้ `AuthLayout` (gradient background)
2. Admin pages (Dashboard, Users, Settings) -> ใช้ `AdminLayout` (sidebar + header)
3. 403 Forbidden -> อยู่ใน `AdminLayout` (user ต้อง login อยู่แล้ว)
4. **404 Not Found -> standalone (ไม่ใช้ layout ใด)** - เป็นหน้า full-screen
5. Root `/` -> redirect ไป `/dashboard`

## Global vs Feature - หลักการแบ่ง

**กฎ:** "ถ้าลบ feature นี้ออก ของชิ้นนี้ยังมีคนใช้ไหม?"
- ใช่ -> Global (`src/utils/`, `src/hooks/`, `src/components/global/`)
- ไม่ -> Feature (`src/features/<name>/`)

### Component Structure

```text
src/components/
  global/          -> ใช้ได้ทุกหน้า (PageTransition, PermissionGuard, Skeletons)
  features/        -> Component เฉพาะ feature แบบง่าย (UserAvatar, RoleTag)
  index.ts         -> Re-export ทั้งหมด
```

### Feature Module Structure

```text
src/features/
  users/
    components/    -> UI เฉพาะ user (UserForm, UserDetailCard)
    hooks/         -> Hooks เฉพาะ user (useUserFilters, useUserMutation)
    utils/         -> Functions เฉพาะ user (userValidation)
    types.ts       -> Types เฉพาะ
    index.ts       -> Public API
  payroll/
    utils/         -> calculateOvertimePay(), generatePayslipPDF()
    ...
```

### ตัวอย่างการแบ่ง

| รายการ | ที่อยู่ | เหตุผล |
|--------|--------|--------|
| `formatDate()` | `src/utils/` | ทุก module ใช้ |
| `usePermission()` | `src/hooks/` | ไม่ผูกกับ feature |
| `PageTransition` | `src/components/global/` | ทุกหน้าใช้ |
| `useUserFilters()` | `src/features/users/hooks/` | เฉพาะ user |
| `calculateOvertimePay()` | `src/features/payroll/utils/` | เฉพาะ payroll |
| `LeaveRequestForm` | `src/features/leave/components/` | เฉพาะลางาน |

## Folder Structure

```text
src/
  app/          -> App entry, providers, theme config (antd theme อยู่ที่นี่)
  api/          -> Axios client, interceptors
  assets/       -> Static assets
  components/   -> Reusable UI components (แบ่ง global/ + features/)
  constants/    -> Permissions, API endpoints, query keys, status labels
  features/     -> Business modules (แยก components/hooks/utils ตาม domain)
  hooks/        -> Global hooks (usePermission, useLoadingBar)
  layouts/      -> AuthLayout, AdminLayout
  pages/        -> Page components
  routes/       -> ศูนย์กลาง routing ทั้งระบบ
    routes.tsx  -> Route config (data-driven)
    menu.tsx    -> Menu config + icons
    helpers.ts  -> Menu helpers
    AppRoutes.tsx -> Route rendering
    ProtectedRoute.tsx -> Route guards
  schemas/      -> Zod schemas (login, user, settings)
  services/     -> API services (auth, user, dashboard) - รองรับ mock + real
  store/        -> Zustand stores (auth, ui, app)
  styles/       -> Page-specific CSS
    pages/      -> CSS แยกตามหน้า (login.css, dashboard.css, users.css, ...)
    index.css   -> รวม import ทั้งหมด
  types/        -> TypeScript types/interfaces
  utils/        -> Helpers (permission, formatter)
```

## Page CSS Convention

1. CSS เฉพาะหน้าอยู่ใน `src/styles/pages/<page-name>.css`
2. ใช้ class wrapper เช่น `.login-page`, `.dashboard-page`, `.users-page`
3. เพิ่มหน้าใหม่ -> สร้างไฟล์ CSS ใหม่ แล้ว import ใน `src/styles/index.css`
4. Global styles อยู่ใน `src/index.css` เท่านั้น

## Coding Conventions

1. TypeScript ให้ครบ ห้าม any
2. ใช้ named exports
3. ตั้งชื่อไฟล์ใช้ PascalCase สำหรับ components, camelCase สำหรับ utilities
4. ทุก service ต้องรองรับ mock mode สำหรับ development
5. Validation message ภาษาไทย อ่านง่าย เหมาะกับระบบธุรกิจ
6. หลีกเลี่ยง hardcode role logic กระจายทั่วระบบ

## Performance Rules (useCallback / useMemo)

**หลักการ:** ใช้เฉพาะจุดที่มีปัญหา performance จริง อย่าใช้มั่ว

### useMemo - จำผลลัพธ์การคำนวณ

ใช้เมื่อ:
- Filter/map/sort data ที่มีจำนวนเยอะ
- สร้าง object/array ส่งเป็น props ให้ React.memo child
- แปลง data format (เช่น MenuItem[] -> antd MenuProps)
- สร้าง columns config สำหรับ Table

ไม่ต้องใช้เมื่อ:
- Data มีแค่ 10-20 items
- ค่าเป็น primitive (string, number)
- Child ไม่ได้ใช้ React.memo

```typescript
// ใช้ useMemo - filter ข้อมูลเยอะ
const filteredUsers = useMemo(
  () => users.filter((u) => u.fullName.includes(search)),
  [users, search]
);

// ไม่ต้อง useMemo - ค่า primitive
const userName = currentUser?.fullName || 'User';  // ไม่ต้อง wrap
```

### useCallback - จำ function reference

ใช้เมื่อ:
- ส่ง function เป็น props ให้ React.memo child
- Function เป็น dependency ของ useEffect
- Debounced/throttled handler

ไม่ต้องใช้เมื่อ:
- Simple onClick handler ใน component เดียว
- Inline function ที่ไม่ได้ส่งต่อ

```typescript
// ใช้ useCallback - ส่งให้ React.memo child
const handleDelete = useCallback((id: string) => {
  userService.deleteUser(id);
}, []);

// ไม่ต้อง useCallback - ใช้ใน component เดียว
const handleClick = () => setOpen(true);  // ไม่ต้อง wrap
```

## Service Layer Pattern

```typescript
// ทุก service ต้องรองรับ 2 modes
const USE_MOCK = !import.meta.env.VITE_API_BASE_URL;

export const someService = {
  getData: async (): Promise<SomeType> => {
    if (USE_MOCK) {
      await new Promise((r) => setTimeout(r, 800)); // จำลอง delay
      return MOCK_DATA;
    }
    const response = await apiClient.get<ApiResponse<SomeType>>(API_ENDPOINTS.SOME.PATH);
    return response.data.data;
  },
};
```

## Development Workflow

1. ทดสอบ: `bun dev` (port 3000)
2. Mock login: admin / admin123
3. ไม่ต้องตั้ง VITE_API_BASE_URL จะใช้ mock data อัตโนมัติ
