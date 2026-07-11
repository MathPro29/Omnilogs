# HR Admin Panel - Developer Guide

## Quick Start

```bash
bun install
bun dev          # เปิดที่ http://localhost:3000
# Mock login: admin / admin123
```

---

## Architecture Overview

```text
Browser -> React Router -> Layout (Auth/Admin) -> Page -> Service -> API/Mock
                               |                     |
                         ProtectedRoute          TanStack Query (server state)
                         PermissionGuard         Zustand (client state)
```

### Tech Stack

| หมวด | เทคโนโลยี |
|---------|-------------|
| Framework | React 18 + Vite 6 + TypeScript |
| UI | Ant Design (antd) + Tailwind CSS v3 |
| Routing | React Router v7 |
| Server State | TanStack Query v5 |
| Client State | Zustand v5 |
| Validation | Zod v4 |
| Date | Day.js |
| Icons | Heroicons |
| Animation | Motion for React |
| API | Axios |
| Loading | react-top-loading-bar |
| Font | Noto Sans Thai (Google Fonts) |

---

## Project Structure

```text
src/
  app/theme.ts              # Ant Design theme - ศูนย์กลางการกำหนดสี/ขนาด
  api/client.ts             # Axios instance + interceptors (401/403 auto-handle)
  components/
    global/                 # Component ที่ใช้ได้ทุกที่ (PageTransition, PermissionGuard, Skeletons)
    features/               # Component ที่ผูกกับ feature เฉพาะ
  constants/index.ts        # Permissions, API Endpoints, Query Keys, Routes, Pagination
  features/                 # Business modules (แยก components/hooks/utils ตาม domain)
  hooks/                    # Global hooks (usePermission, useLoadingBar)
  layouts/                  # AuthLayout (login), AdminLayout (sidebar+header+content)
  pages/                    # Login, Dashboard, Users, Settings, 403, 404
  routes/                   # ศูนย์กลาง routing ทั้งระบบ
    routes.tsx              # Route config (data-driven)
    menu.tsx                # Menu config + icons (sidebar items)
    helpers.ts              # Menu helpers (findMenuByPath, getFlatMenuItems)
    AppRoutes.tsx           # Route rendering (สร้าง <Routes> จาก config)
    ProtectedRoute.tsx      # Route guards (auth + permission)
    index.ts                # Re-export ทั้งหมด
  schemas/index.ts          # Zod schemas (login, createUser, editUser, settings, filter)
  services/                 # API services - รองรับ mock + real (dual-mode)
  store/                    # Zustand: auth.store (persist), ui.store, app.store
  styles/pages/             # CSS เฉพาะแต่ละหน้า
  types/index.ts            # TypeScript interfaces
  utils/                    # permission helpers (can/canAny/canAll), formatters
```

---

## Key Patterns

### 1. Theme System (src/app/theme.ts)

**ทุก antd component ต้องใช้ theme ที่นี่ ห้าม hardcode สี/ขนาดใน component**

```typescript
// ถูก - ใช้ theme token
export const antdThemeConfig = {
  token: { colorPrimary: '#6647F0' },
  components: {
    Input: { controlHeight: 36, activeBorderColor: '#6647F0' },
  },
};

// ผิด - hardcode สีใน component
<Input style={{ borderColor: '#6647F0' }} />  // ห้ามทำแบบนี้
```

### 2. TanStack Query (Server State)

**หลักการ:** ข้อมูลจาก API ทั้งหมดต้องใช้ TanStack Query จัดการ

**ช่วยอะไร:**

- Auto caching (staleTime: 5 นาที)
- Auto refetch เมื่อ data เก่า
- Loading/error state อัตโนมัติ
- Invalidation หลัง mutation -> auto refetch

**จุดที่ใช้:**

| หน้า | Query/Mutation | Query Key | Service |
|------|---------------|-----------|---------|
| DashboardPage | `useQuery` | `['dashboard', 'summary']` | `dashboardService.getSummary()` |
| UsersPage | `useQuery` | `['users', ...filters]` | `userService.getUsers(filters)` |
| UsersPage | `useMutation` | invalidate `['users']` | `userService.deleteUser(id)` |
| LoginPage | `useMutation` (manual) | - | `authService.login(credentials)` |

**Pattern:**

```typescript
// อ่านข้อมูล
const { data, isLoading } = useQuery({
  queryKey: [...QUERY_KEYS.USERS, filters],  // key ต้องรวม filters ด้วย
  queryFn: () => userService.getUsers(filters),
});

// แก้ไขข้อมูล
const deleteMutation = useMutation({
  mutationFn: userService.deleteUser,
  onSuccess: () => {
    message.success('ลบสำเร็จ');
    queryClient.invalidateQueries({ queryKey: QUERY_KEYS.USERS }); // refetch auto
  },
});
```

### 3. Zustand (Client State)

**หลักการ:** เฉพาะ state ที่ไม่ได้มาจาก API

| Store | Purpose | Persist? |
|-------|---------|----------|
| `auth.store` | user, token, permissions, isAuthenticated | Yes (localStorage) |
| `ui.store` | sidebar collapsed | No |
| `app.store` | page title, breadcrumbs | No |

### 4. Service Layer (Dual-Mode)

ทุก service ตรวจ `VITE_API_BASE_URL`:

- **ไม่มี** -> ใช้ mock data + delay (development)
- **มี** -> เรียก API จริงผ่าน Axios

### 5. Permission System (3 ระดับ)

```text
Route Level:    <ProtectedRoute requiredPermissions={[...]}>
Menu Level:     filterMenuByPermissions(menuConfig)
Component Level: <PermissionGuard permission="user:edit"><EditButton /></PermissionGuard>
```

**กฎ Permission:**

| `requiredPermissions` | ความหมาย |
|----------------------|---------|
| `undefined` (ไม่ระบุ) | ทุก role เข้าได้ (แค่ login) |
| `[]` (array ว่าง) | ทุก role เข้าได้ (แค่ login) |
| `['user:view']` | ต้องมี permission นี้ |
| `['user:view', 'user:edit']` | ต้องมีอย่างน้อย 1 (default) |
| `['user:view', 'user:edit']` + `requireAll: true` | ต้องมีทุกตัว |

### 6. Route System (Data-Driven)

ระบบ routing เป็น **data-driven** - กำหนด config ไว้ที่ `src/routes/` ระบบจะ generate routes + menu อัตโนมัติ

```text
src/routes/
  routes.tsx          -> Route config (หน้าไหนบ้าง + permission)
  menu.tsx            -> Menu config (sidebar items + icons)
  helpers.ts          -> Menu helpers (findMenuByPath, getFlatMenuItems)
  AppRoutes.tsx       -> Route rendering (อ่าน config -> สร้าง <Routes>)
  ProtectedRoute.tsx  -> Route guards (auth + permission)
  index.ts            -> Re-export ทั้งหมด
```

**App.tsx เหลือแค่ 3 providers:**

```typescript
export default function App() {
  return (
    <QueryClientProvider client={queryClient}>    // API cache
      <ConfigProvider theme={...} locale={...}>   // antd theme
        <BrowserRouter>                           // routing
          <AppRoutes />                           // <- ทุก route อยู่ที่นี่
        </BrowserRouter>
      </ConfigProvider>
    </QueryClientProvider>
  );
}
```

### 7. Route Layout Rules

| Route | Layout | หมายเหตุ |
|-------|--------|---------| 
| `/login` | AuthLayout | Gradient background |
| `/dashboard`, `/users`, `/settings` | AdminLayout | Sidebar + header |
| `/403` | AdminLayout | User ต้อง login อยู่แล้ว |
| `/404`, `*` | **ไม่ใช้ layout** | Standalone full-screen |

### 8. Component Organization

| Folder | หน้าที่ | ตัวอย่าง |
|--------|--------|---------|
| `components/global/` | ใช้ได้ทุกหน้า ไม่ผูกกับ feature | PageTransition, Skeletons |
| `components/features/` | ผูกกับ business module เฉพาะ | UserAvatar, DashboardChart |
| `features/<name>/` | Full module (components + hooks + utils) | features/payroll/ |

### 9. Loading Strategy

| สถานการณ์ | ใช้อะไร |
|-----------|--------|
| เปลี่ยนหน้า (route change) | react-top-loading-bar (สีม่วง 3px) |
| Fetch API data | Skeleton component (เลียนแบบ layout จริง) |
| หน้าแสดงผล | Motion animation (fade + slide) |

---

## Performance: useCallback และ useMemo

### หลักการตัดสินใจ

```text
ถามตัวเอง: "ถ้าไม่ใช้ useCallback/useMemo มีปัญหาจริงไหม?"
  ไม่มี -> ไม่ต้องใช้
  มี (re-render ซ้ำซ้อน / คำนวณหนัก) -> ใช้
```

> [!IMPORTANT]
> อย่าใส่ useCallback/useMemo ทุกที่! ใส่เฉพาะจุดที่มีปัญหา performance จริงเท่านั้น

### useMemo - จำผลลัพธ์ของการคำนวณ

**ใช้เมื่อ:** การคำนวณ/แปลงข้อมูลหนัก ที่ไม่ต้องการทำซ้ำทุก render

```typescript
// ===== ตัวอย่างที่ 1: Filter + คำนวณจาก data =====

function UsersPage() {
  const [search, setSearch] = useState('');
  const { data: users = [] } = useQuery({ ... });

  // useMemo: filter users ตาม search 
  // ถ้า users หรือ search ไม่เปลี่ยน จะไม่คำนวณใหม่
  const filteredUsers = useMemo(
    () => users.filter((u) => u.fullName.includes(search)),
    [users, search]
  );

  // useMemo: สถิติจาก filtered users
  const stats = useMemo(() => ({
    total: filteredUsers.length,
    active: filteredUsers.filter((u) => u.status === 'active').length,
    inactive: filteredUsers.filter((u) => u.status === 'inactive').length,
  }), [filteredUsers]);

  return (
    <div>
      <p>ใช้งาน: {stats.active} / ทั้งหมด: {stats.total}</p>
      <Table dataSource={filteredUsers} ... />
    </div>
  );
}
```

```typescript
// ===== ตัวอย่างที่ 2: แปลง data เป็น antd items =====

function AdminLayout() {
  const currentUser = useAuthStore((state) => state.currentUser);

  // useMemo: filter menu ตาม permission (ไม่ต้อง filter ซ้ำถ้า permissions ไม่เปลี่ยน)
  const filteredMenu = useMemo(
    () => filterMenuByPermissions(menuConfig),
    [currentUser?.permissions]  // recalculate เมื่อ permissions เปลี่ยน
  );

  // useMemo: แปลง MenuItem[] -> antd MenuProps['items'] 
  const antdMenuItems = useMemo(
    () => toAntdMenuItems(filteredMenu),
    [filteredMenu]
  );

  return <Menu items={antdMenuItems} />;
}
```

```typescript
// ===== ตัวอย่างที่ 3: สร้าง columns config สำหรับ Table =====

function UsersPage() {
  // useMemo: columns ไม่เปลี่ยน ไม่ต้องสร้างใหม่ทุก render
  const columns = useMemo<ColumnsType<User>>(() => [
    { title: 'ชื่อ', dataIndex: 'fullName', key: 'name' },
    { title: 'อีเมล', dataIndex: 'email', key: 'email' },
    { title: 'สถานะ', dataIndex: 'status', key: 'status',
      render: (status: string) => <Tag color={STATUS_COLORS[status]}>{STATUS_LABELS[status]}</Tag>,
    },
  ], []);

  return <Table columns={columns} />;
}
```

### useCallback - จำ function reference

**ใช้เมื่อ:** ส่ง function เป็น props ให้ child component ที่ใช้ `React.memo` หรือเป็น dependency ของ useEffect

```typescript
// ===== ตัวอย่างที่ 1: Handler ที่ส่งให้ child component =====

function UsersPage() {
  const queryClient = useQueryClient();

  // useCallback: ถ้า queryClient ไม่เปลี่ยน function reference จะไม่เปลี่ยน
  // -> child component ที่รับ onDelete จะไม่ re-render โดยไม่จำเป็น
  const handleDelete = useCallback((id: string) => {
    Modal.confirm({
      title: 'ยืนยันการลบ',
      content: 'ต้องการลบผู้ใช้นี้หรือไม่?',
      onOk: async () => {
        await userService.deleteUser(id);
        queryClient.invalidateQueries({ queryKey: QUERY_KEYS.USERS });
        message.success('ลบสำเร็จ');
      },
    });
  }, [queryClient]);

  // useCallback: navigate ไม่เปลี่ยน -> function reference คงที่
  const handleEdit = useCallback((id: string) => {
    navigate(`/users/${id}/edit`);
  }, [navigate]);

  return <UserTable onDelete={handleDelete} onEdit={handleEdit} />;
}

// Child component ใช้ React.memo เพื่อหลีกเลี่ยง re-render
const UserTable = React.memo(function UserTable({
  onDelete,
  onEdit,
}: {
  onDelete: (id: string) => void;
  onEdit: (id: string) => void;
}) {
  // component นี้จะ re-render เฉพาะเมื่อ props เปลี่ยนจริงเท่านั้น
  return <Table ... />;
});
```

```typescript
// ===== ตัวอย่างที่ 2: Debounced search =====

function UsersPage() {
  const [filters, setFilters] = useState<UserFilters>({ search: '' });

  // useCallback: function reference คงที่ -> debounce ทำงานถูกต้อง
  const handleSearch = useCallback(
    debounce((value: string) => {
      setFilters((prev) => ({ ...prev, search: value, page: 1 }));
    }, 300),
    []
  );

  return <Input.Search onChange={(e) => handleSearch(e.target.value)} />;
}
```

```typescript
// ===== ตัวอย่างที่ 3: Event handler ที่ใช้ใน useEffect dependency =====

function DashboardPage() {
  // useCallback: ใช้เป็น dependency ของ useEffect โดยไม่ทำให้ effect re-run ทุก render
  const fetchData = useCallback(async () => {
    const summary = await dashboardService.getSummary();
    setData(summary);
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);
}
```

### สรุปตาราง: เมื่อไหร่ใช้ / ไม่ใช้

| สถานการณ์ | ใช้ | ไม่ต้องใช้ |
|-----------|-----|----------|
| คำนวณ/filter ข้อมูลหนัก | `useMemo` | ถ้า data มีแค่ 10-20 items |
| สร้าง object/array ส่งเป็น props | `useMemo` | ถ้า child ไม่ใช้ `React.memo` |
| แปลง data format (mapping) | `useMemo` | ถ้า mapping เบา |
| ส่ง function เป็น props ให้ `React.memo` child | `useCallback` | ถ้า child ไม่ใช้ `React.memo` |
| Function เป็น dependency ของ `useEffect` | `useCallback` | ถ้าไม่ได้อยู่ใน deps |
| Debounced/throttled handler | `useCallback` | - |
| Simple event handler (onClick) ใน component เดียว | - | ไม่ต้องใช้ `useCallback` |
| ค่า primitive (string, number) | - | ไม่ต้องใช้ `useMemo` |

---

## Adding New Features

### เพิ่มหน้าใหม่ (แก้ 2 ไฟล์)

1. สร้าง page component ใน `src/pages/NewPage.tsx`
2. Export จาก `src/pages/index.ts`
3. เพิ่ม permission key + route path ใน `src/constants/index.ts`
4. เพิ่ม route config ใน **`src/routes/routes.tsx`**
5. เพิ่ม menu item ใน **`src/routes/menu.tsx`** (ถ้าต้องแสดงใน sidebar)
6. (optional) สร้าง CSS เฉพาะหน้า `src/styles/pages/new-page.css`

> [!TIP]
> **ไม่ต้องแก้ `App.tsx` หรือ `AdminLayout.tsx`** - ระบบ generate routes + menu อัตโนมัติ

### เพิ่ม API Service ใหม่

1. สร้าง `src/services/xxx.service.ts`
2. เพิ่ม API endpoints ใน `src/constants/index.ts` -> `API_ENDPOINTS`
3. เพิ่ม Query Keys ใน `src/constants/index.ts` -> `QUERY_KEYS`
4. เพิ่ม types ใน `src/types/index.ts`
5. Export จาก `src/services/index.ts`
6. รองรับ mock mode เสมอ

### เพิ่ม Component ใหม่

- ถ้าใช้ได้ทุกหน้า -> `src/components/global/`
- ถ้าผูกกับ feature แบบง่าย -> `src/components/features/`
- ถ้า component + hook + util ผูกกัน -> `src/features/<name>/`
- Export ผ่าน `src/components/index.ts` หรือ `src/features/<name>/index.ts`

---

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `VITE_API_BASE_URL` | No | `/api` | ถ้าไม่ตั้ง = ใช้ mock data |

---

## Coding Standards

1. TypeScript strict - ห้าม `any`
2. Named exports เท่านั้น (except `App` default export)
3. PascalCase สำหรับ components, camelCase สำหรับ utilities
4. Validation messages ภาษาไทย
5. ห้าม Emoji ในโค้ดและ UI -> ใช้ Heroicons / Ant Design Icons
6. ห้าม hardcode สี -> ใช้ CSS variable หรือ antd theme token
7. `useCallback` / `useMemo` ใช้เฉพาะจุดที่มีปัญหา performance จริง (ดูหัวข้อ Performance)
