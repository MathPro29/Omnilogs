---
description: Project rules and conventions for Omnilogs frontend development
---

# Omnilogs Frontend Development Rules

## Project stack

- Use React 19, TypeScript, and Vite 6.
- Use Ant Design 6 for UI components and `src/app/theme.ts` for shared theme and Thai locale.
- Use Tailwind CSS 4 for layout, spacing, and responsive composition.
- Use React Router DOM 7 for routing.
- Use Zod 4 for new form and request validation.
- Use Axios only through `src/api/client.ts`.
- Use TanStack Query 5 for server state.
- Use Zustand 5 only for client/session/UI state.
- Use Day.js for date handling.
- Use Heroicons, Ant Design Icons, or Lucide React. Do not use emoji as UI icons.
- Use Motion only when it improves feedback or continuity.
- Use `react-top-loading-bar` for route transitions and Skeleton components for data loading.

## Source of truth

Before changing behavior, check the current implementation and backend contract. Keep this file and `docs/DEVELOPER_GUIDE.md` aligned with the code. Do not document routes, services, or features that do not exist.

The frontend currently contains dashboard, user management, settings, Logs Explorer, Live Tail, and Audit Logs screens. It is not a separate HR Admin Panel.

## Project structure

```text
src/
  main.tsx                 # application entry point
  App.tsx                  # providers and BrowserRouter
  api/client.ts            # Axios instance and auth/error interceptors
  app/theme.ts             # Ant Design theme and Thai locale
  components/global/       # shared shell, navigation, guards, loading UI
  components/features/     # reusable feature-oriented components
  constants/index.ts       # routes, permissions, endpoints, query keys
  hooks/                   # shared hooks
  layouts/                 # AuthLayout and AdminLayout
  pages/                   # route-level screens
  routes/                  # route config, menu config, helpers, guards
  schemas/                 # Zod schemas
  services/                # API and supported mock-mode services
  store/                   # auth, UI, and app stores
  styles/                  # global, layout, and page CSS
  types/                   # shared TypeScript types
  utils/                   # permission helpers and formatters
```

Do not create speculative folders such as payroll or leave modules. Add a feature folder only when the feature has enough components, hooks, or utilities to justify it.

## State management

- TanStack Query owns API data, lists, details, loading/error state, and mutations.
- Zustand owns authentication/session data, permissions, sidebar state, page title, and breadcrumbs.
- Do not copy server data into Zustand.
- Include every filter and active product/project/environment scope in a query key.
- Invalidate related queries after mutations.
- Reset or invalidate scope-bound data when the scope changes.

Example:

```tsx
const usersQuery = useQuery({
  queryKey: [...QUERY_KEYS.USERS, filters, scope],
  queryFn: () => userService.getUsers(filters, scope),
});
```

## API and security

- Components and pages must not call Axios directly. Add or use a domain service in `src/services/`.
- Add reusable endpoint paths to `API_ENDPOINTS` and query keys to `QUERY_KEYS` in `src/constants/index.ts`.
- Preserve the backend response envelope: `success`, `message`, `data`, and optional `meta`.
- Use the configured `apiClient` so bearer-token handling and 401/403 behavior remain consistent.
- A 401 clears auth and redirects to `/login`; a 403 redirects to `/403`.
- Backend authorization is the security boundary. Frontend guards only improve usability.
- Never log or persist passwords, tokens, secrets, encryption material, decrypted values, audit secrets, or raw sensitive log payloads.
- Never read, print, copy, or commit `.env` files.

## Authorization

Use effective permissions rather than scattered checks such as `role === 'owner'`.

- Use `can`, `canAny`, and `canAll` from `src/utils/permission.ts`.
- Use `ProtectedRoute` for route-level access.
- Use `PermissionGuard` for action-level rendering.
- Keep route guards, sidebar menu filtering, table actions, and API affordances consistent.
- Distinguish loading, empty, forbidden, no-scope, expired-membership, disabled-feature, and error states.

```tsx
<PermissionGuard permission="user:edit">
  <Button>แก้ไขผู้ใช้</Button>
</PermissionGuard>
```

## Routes and navigation

- Routes are configured in `src/routes/routes.tsx`.
- Sidebar items are configured in `src/routes/menu.tsx`.
- Route and menu entries must use constants from `src/constants/index.ts`.
- When adding a discoverable page, update both route config and menu config.
- Use `AuthLayout` for `/login`, `/register`, and `/forgot-password`.
- Use `AdminLayout` for authenticated application pages.
- Keep `/403` within the authenticated layout and `/404` standalone as implemented.
- Verify guest, authenticated, forbidden, and direct-URL behavior for every route change.

## UI and accessibility

- Prefer Ant Design components such as Table, Form, Input, Select, Drawer, Modal, Card, Empty, Skeleton, Tag, and Descriptions.
- Use theme tokens or CSS variables instead of hardcoded colors in components.
- Keep the white-first, restrained enterprise visual style.
- Use Noto Sans Thai consistently and write clear Thai validation and status messages.
- Every async screen needs loading, empty, error/retry, and forbidden states.
- Tables must have a usable narrow-screen path: horizontal scroll, reduced columns, or a detail drawer.
- Every form needs labels and every icon-only control needs an accessible name.
- Preserve visible focus states and keyboard operation. Respect `prefers-reduced-motion`.
- Use confirmation for destructive or privilege-changing actions and show the result near the action.

## Component and CSS organization

- Put broadly reusable UI in `src/components/global/`.
- Put reusable components tied to an existing feature in `src/components/features/`.
- Put page-specific orchestration in `src/pages/`.
- Put page CSS in `src/styles/pages/<page-name>.css` and import it through the existing styles entry point.
- Use named exports by default. Default exports are acceptable for the application entry point or an existing third-party convention.
- Use PascalCase for components/files and camelCase for utilities.
- Use `useMemo` and `useCallback` only for expensive calculations, stable references required by memoized children, or necessary effect dependencies.

## Mock and real API modes

Services that support local mock behavior use:

```ts
const USE_MOCK = !import.meta.env.VITE_API_BASE_URL;
```

With no `VITE_API_BASE_URL`, auth, dashboard, and users provide local mock data. When it is set, use the real API through `apiClient`. Audit Logs currently uses the API directly and must not assume mock data exists.

Use `npm`, not `bun`, for the documented workflow:

```bash
npm install
npm run dev
npm run build
npm run lint
npm run preview
```


