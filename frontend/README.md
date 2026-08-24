# OmniLogs Frontend

React 19 + Vite + TypeScript frontend ?????? OmniLogs ??? Ant Design, Tailwind CSS, TanStack Query, Zustand, Zod ??? React Router

## Commands

?????????????? `frontend/`:

```bash
npm install
npm run dev
npm run lint
npm run build
npm run preview
```

## Docker Compose

Frontend และ Backend ใช้ Compose แยกกัน โดยควรเปิด Backend ก่อนเพื่อให้ API พร้อมรับ request:

```bash
# terminal 1: Omnilogs/backend
docker compose up -d --build

# terminal 2: Omnilogs/frontend
docker compose up -d --build
```

Frontend เปิดที่ `http://localhost:5173` และจะ proxy `/api/*` ไปยัง
`http://host.docker.internal:2910` ตามค่าเริ่มต้น หาก Backend อยู่คนละเครื่อง
หรือใช้คนละ domain ให้กำหนด `API_UPSTREAM` ใน `.env` ของ frontend แล้ว build ใหม่
เช่น `API_UPSTREAM=https://api.example.com`

## ?????????????

- `src/pages/` ? route-level orchestration
- `src/components/global/` ? shared UI
- `src/components/features/` ? presentation components ?????? feature
- `src/features/` ? hooks, workflow, services, stores, schemas, types ??? feature utilities
- `src/services/` ? domain services ????????????????
- `src/api/` ? Axios client ??? auth/error interceptors
- `src/store/` ? session ??? application UI state
- `src/utils/` ? utilities ????

## ??????

- [Frontend refactor, architecture ??? feature flows](docs/FRONTEND_REFACTOR_FLOW.md)
- [Developer guide](docs/DEVELOPER_GUIDE.md)
- [Ant Design CSS reference](docs/ANT_DESIGN_CSS_REFERENCE.md)

??????????: component ???????? Axios ??????, server state ??? TanStack Query, client/session state ??? Zustand ??? backend authorization ???? security boundary
