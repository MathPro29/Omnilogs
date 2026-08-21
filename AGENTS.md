# AGENTS.md

## Scope and project overview

Omnilogs is a log-management platform. This guide is authoritative for the Go backend. It records only a high-level UI plan; it is not a frontend implementation guide.

The backend uses PostgreSQL for configuration, authorization, queue metadata, and audit data; Elasticsearch for searchable logs; and NATS JetStream to signal queued ingestion work.

Important paths:

- `backend/api/` — Go module (`omnilogs-api`), API, worker, migrations, seed command, and OpenAPI tooling.
- `backend/compose.yml` — local PostgreSQL, Elasticsearch, NATS, migration, API, and worker stack.
- `BACKEND_STRUCTURE_AND_FEATURE_FLOW.md` — detailed domain and request-flow reference. Confirm details against current code because it may lag implementation.
- `backend/api/docs/` — worker architecture and query-performance notes.

## Working rules

- Preserve unrelated working-tree changes. Inspect `git status --short` before modifying files.
- Make the smallest cohesive change. Do not refactor adjacent modules without a concrete need.
- Never commit `.env`, credentials, JWT secrets, encryption keys, API keys, access tokens, or decrypted sensitive log values.
- Do not use programs under `backend/api/scratch/` as normal application commands. Several inspect, reset, clear, or dump real data.
- Avoid editing generated OpenAPI copies manually. Update the source, run the OpenAPI tool, and commit synchronized output.
- Keep API, authorization, audit, masking, retention, and worker behavior backward-compatible unless the requested change explicitly changes the contract.

## Backend architecture

Run backend commands from `backend/api/`.

```text
routes -> module wiring -> handler -> usecase -> repository -> PostgreSQL/Elasticsearch
                                      |
                                      +-> audit/authorization/domain services

API ingestion -> PostgreSQL queue + NATS signal
              -> worker -> validate/transform/mask -> Elasticsearch
```

Key locations:

- `cmd/api`, `cmd/worker`, `cmd/migrate`, `cmd/seed`, `cmd/openapi` — supported entry points.
- `internal/bootstrap/` — dependency construction, middleware, and server/worker lifecycle.
- `routes/` — Gin route grouping and middleware attachment; keep handlers out of this layer.
- `internal/<domain>/handler/` — HTTP parsing, validation, and status/response mapping.
- `internal/<domain>/usecase/` — business rules, authorization decisions, and orchestration.
- `internal/<domain>/repository/` — persistence and Elasticsearch queries.
- `internal/<domain>/module/` — dependency wiring where present.
- `dto/` — request/response transport shapes; `models/` — GORM persistence models.
- `middleware/audit/` — system audit logging; `configs/` — environment and external clients.
- `migrate/` — schema and performance-index migrations.
- `utils/` and `responses/` — shared helpers and the common response envelope.

Backend conventions:

- Preserve layer boundaries. A handler should not contain SQL/Elasticsearch queries, and a repository should not decide HTTP responses.
- Pass `context.Context` through slow or external operations and honor configured timeouts.
- Wrap errors with useful context, but do not expose secrets or raw internal errors to clients.
- Use `log/slog` structured fields. Never log credentials, tokens, encryption material, raw sensitive fields, or unmasked payloads.
- Keep authorization checks in the established middleware/usecase path. UI permission checks are not a security boundary.
- Any sensitive-data access or mutation that is currently audited must remain audited.
- When changing ingestion, consider validation, JSON-path extraction, transformations, masking, batching, retries, acknowledgement behavior, and failure recording together.
- Add focused `*_test.go` files beside the affected package. Prefer table-driven tests for validation and query-building cases.
- Format all changed Go files with `gofmt`.

## Backend authorization model

Platform roles seeded by the backend are:

- `god` — full platform access (`{"all": true}`).
- `owner` — access to owned/related products (`{"own_product_access": true}`).
- `superadmin` — platform role intended for restricted role administration (`{"assign_role_only": true}`).
- `user` — no platform-wide privilege by default.
- `admin` — the default role for self-registered accounts (`{"default_menu_access": true}`). It is non-privileged: it is not admitted by `HasAdminPlatformRole` and receives no product membership or scope automatically.

`HasAdminPlatformRole` currently admits `god`, `owner`, and `superadmin` to platform administration routes. This admission does not mean the three roles may perform every action. Preserve the finer usecase checks; for example, GOD may assign any platform role, Owner cannot assign GOD and may manage only related users, and Superadmin must not be assumed to have GOD authority.

Product access is separate from platform roles. `product_memberships`, `product_roles`, membership scopes, explicit allow/deny rules, and feature availability determine effective product access. Product role capabilities map roughly to `READ_ONLY`, `EDITOR`, `ADMIN`, and `FULL_ACCESS`, but authorization must use resource/action/scope permissions rather than the display level alone.

## UI plan: one design, three experiences

Use one design system, application shell, component library, and route set. Do not build three separate applications or copy pages per role. After login, load one access context containing:

- platform role;
- accessible products and current product;
- product membership, product role, active scope, and effective permissions;
- feature availability for the selected product/project/environment.

The intended experiences are:

| Experience               | Backend identity                                                                         | Default scope                        | Primary purpose                                                                               |
| ------------------------ | ---------------------------------------------------------------------------------------- | ------------------------------------ | --------------------------------------------------------------------------------------------- |
| GOD / Superadmin console | Platform role `god` or `superadmin`                                                      | Platform-wide                        | Platform governance, users, products, global audit, health, and permitted role administration |
| Owner workspace          | Platform role `owner`                                                                    | Owned/assigned products              | Product setup, projects/environments, members, roles, policies, retention, and product audit  |
| Admin workspace          | Admin-capable product membership; normally platform role `admin` after self-registration | Assigned product/project/environment | Daily log operations and only product actions granted by effective permissions                |

Important distinction: `admin` is a non-privileged platform role used as the default for self-registration; it does not grant platform administration or product access. The Admin UI persona still requires a product membership, scope, and effective resource/action permissions. Do not infer access from the role name alone.

### Shared application shell

All experiences share:

- left navigation, top bar, breadcrumbs, global search, notifications, profile menu, and common loading/error/empty states;
- a product/project/environment scope selector in a consistent location;
- the same dashboard, list/detail, explorer, settings, and access-management templates;
- the same typography, spacing, colors, tables, filters, drawers, dialogs, confirmations, and audit-reason prompts;
- a visible role/scope badge showing whether the actor is platform-wide or product-scoped.

Variation comes from configuration and access context, not duplicated pages. Shared templates receive scope, capabilities, labels, summary cards, filters, and allowed actions.

### GOD / Superadmin console

- Platform overview: service health, ingestion/queue failures, product/user totals, storage/search health, and recent global audit activity.
- Platform modules: products, platform users, platform-role assignment, feature availability, system audit, and operational monitoring.
- GOD-only destructive or privilege-changing actions remain hidden/disabled for Superadmin and must still be rejected by the backend.
- Product impersonation or scope switching must be explicit, visibly indicated, and audited; never use it as silent privilege escalation.

### Owner workspace

- Product overview: log volume, projects/environments, queue failures, retention status, members, and recent product audit.
- Product setup: metadata/status, projects, features, environments, API keys, ingestion/masking/index/retention policies, product roles, memberships, and scopes as authorized.
- User and role pickers include only users the Owner may manage.
- Platform-global settings and unrelated products do not appear in navigation, search, or totals.

### Admin workspace

- Operational overview for assigned scopes only.
- Primary tools: log explorer, live tail, dashboards, audit lookup, sensitive-access requests, and permitted custom-field/policy operations.
- Show create/update/grant/revoke/delete actions only when present in effective permissions.
- Never expose platform user management, platform roles, unrelated products, or Owner-only configuration.

### Authorization-driven rendering

- Backend authorization is the security boundary. Hidden menus and disabled buttons are usability only.
- Prefer `can(resource, action, scope)` over scattered checks such as `role === 'owner'`.
- Route guards, navigation, widgets, columns, bulk actions, and API calls use the same access context.
- Distinguish forbidden, no-scope, expired-membership, disabled-feature, and empty-data states.
- On scope change, clear/invalidate scope-bound cached data before rendering the new context.
- Privileged mutations show impact, request confirmation/reason where appropriate, and surface the audit reference.

### Suggested UI delivery order

1. Define the access-context API and permission matrix from existing roles, memberships, scopes, and feature flags.
2. Build the shared shell, scope selector, route metadata, and capability guard.
3. Build shared Dashboard and Logs Explorer templates using scoped data.
4. Add GOD/Superadmin platform modules while preserving GOD-only restrictions.
5. Add Owner product configuration and access-management modules.
6. Add Admin operational modules driven by effective permissions.
7. Test the same routes with GOD, Superadmin, Owner, Product Admin, read-only, expired, and no-scope accounts, including direct URLs and denied APIs.

## Commands

### Full local stack

From `backend/api/`:

```bash
make dev
make dev-down
```

Equivalent repository-root commands:

```bash
docker compose -f backend/compose.yml up --build
docker compose -f backend/compose.yml down
```

Compose exposes API `2910`, PostgreSQL `5499`, Elasticsearch `9200`, NATS `4222`, and NATS monitoring `8222`.

### Backend

```bash
cd backend/api
cp .env.example .env       # first local setup only; adjust locally
go run ./cmd/migrate
go run ./cmd/api
go run ./cmd/worker
go run ./cmd/seed -email you@example.com -password 'strong-password'
go test ./...
go build ./...
go run ./cmd/openapi validate
go run ./cmd/openapi sync  # when the API contract changes
```

The backend CI contract is: `gofmt` clean, `go build ./...`, `go test ./...`, and OpenAPI validation.

## Verification checklist

- Backend changes: `gofmt` changed files, run focused package tests, then `go test ./...` and `go build ./...`.
- Route/DTO/contract changes: run OpenAPI validation and synchronize generated specs.
- UI-facing API changes: verify fields, status codes, auth headers, effective permissions, feature flags, and scope isolation.
- Migration/model changes: verify both a clean migration and compatibility with existing rows.
- Worker/search changes: add focused tests and verify queue-to-index behavior against PostgreSQL, NATS, and Elasticsearch when feasible.

If infrastructure is unavailable, run isolated checks and state exactly which integration verification remains.

## Frontend architecture and conventions

The frontend lives in `frontend/`. It is a React 19 + Vite + TypeScript application using Ant Design 6, Tailwind CSS 4, TanStack Query 5, Zustand 5, Zod 4, Axios, Day.js, Motion, and React Router 7.

Run frontend commands from `frontend/`:

```bash
npm run dev
npm run build
npm run lint
npm run preview
```

Do not read, commit, print, or copy `.env` files. Use only the variable names documented here or in `.env.example` when one exists.

### Current frontend structure

```text
frontend/src/
  api/client.ts              # Axios instance and 401/403 handling
  app/theme.ts               # Ant Design theme and Thai locale
  components/global/         # Shared shell components and permission UI
  constants/index.ts         # Routes, API paths, permissions, query keys
  layouts/                   # AuthLayout and AdminLayout
  pages/                     # Route-level screens
  routes/                    # Public/protected routes and route/menu metadata
  schemas/                   # Zod request/form schemas
  services/                  # HTTP-facing domain services
  store/                     # Zustand auth, UI, and app state
  styles/layout/             # Application shell and navigation CSS`n  styles/pages/              # Page-scoped CSS additions
  types/                     # Shared frontend types
  utils/                     # Formatters and permission helpers
```

`App.tsx` owns the Query Client, Ant Design `ConfigProvider`, and `BrowserRouter`. `AppRoutes.tsx` owns public routes, protected layout routes, loading progress, root redirect, and the standalone 404 route.

### Routing and navigation

- Public auth routes use `AuthLayout`: `/login`, `/register`, and `/forgot-password`.
- Authenticated application routes use `AdminLayout` and must be registered in `src/routes/routes.tsx`.
- A new screen normally needs: a route constant, page export, route configuration, and—if discoverable—a navigation item.
- `src/routes/menu.tsx` and `src/components/global/navigation.ts` currently serve different navigation surfaces. Keep both aligned with the same route constants until they are consolidated; do not add a hard-coded route to only one surface.
- Protect routes with `ProtectedRoute`; use `GuestRoute` for pages that should not be shown to signed-in users.
- The frontend only improves usability. Backend authorization remains the security boundary, so direct URLs and API errors must still be handled correctly.

### API, data, and authorization

- Components must not call Axios directly. Use `src/services/<domain>.service.ts`, which delegates to `apiClient`.
- Add endpoint paths to `API_ENDPOINTS` and query keys to `QUERY_KEYS` when introducing a reusable API operation.
- The API response envelope is generally `{ success, code?, message, data, meta? }`. Paginated backend responses use `meta.page`, `meta.perPage`, `meta.total`, and `meta.totalPage`; do not assume a different shape without checking the handler/OpenAPI contract.
- `api/client.ts` adds the bearer token and redirects on 401/403. Do not bypass its interceptors or manually duplicate token handling.
- Use TanStack Query for server state: include all filters/scope values in query keys, invalidate relevant queries after mutations, and show loading, error, empty, and forbidden states.
- Use Zustand only for client/session UI state. Do not copy query data into a store.
- Use Zod for new form/request validation. Add response parsing when a response is untrusted, complex, or security-sensitive; do not claim existing APIs are all runtime-validated unless that is implemented.
- Some existing services support local mock behavior, while newer services may call the backend directly. Preserve the established behavior of the touched domain; do not silently introduce mock data into operational or security screens.
- Never send tokens, passwords, secrets, decrypted values, raw sensitive log payloads, or audit-secret values to the console, analytics, error messages, or client-side storage.

### Access-context and scope rules

- Render with effective permissions and active product/project/environment scope, not platform-role name checks such as `role === 'owner'`.
- Use the existing permission helper/hook and `PermissionGuard` for component actions. Keep route, navigation, table actions, and API affordances consistent.
- Distinguish no scope, forbidden, expired membership, disabled feature, empty data, and load failure.
- On scope changes, include the scope in query keys and invalidate/reset scope-bound data before showing the new context.
- Privileged changes must make impact clear, require confirmation/reason where the backend expects it, and display the returned audit reference when available.

### UI and accessibility

- Ant Design is the primary component library. Prefer `Table`, `Form`, `Input`, `Select`, `Drawer`, `Modal`, `Card`, `Empty`, `Skeleton`, `Tag`, and `Descriptions` over custom replacements.
- Use Tailwind for layout, spacing, responsive behavior, and small composition utilities. Put reusable visual tokens in `src/app/theme.ts` or CSS variables in `src/index.css`; avoid ad-hoc inline colors and one-off design systems.
- Match the current minimal, white-first UI: clear hierarchy, restrained color, modest radius/shadows, and subtle motion only when it improves feedback.
- Use responsive Ant Design props and Tailwind breakpoints (`sm`, `md`, `lg`, `xl`, `2xl`). Tables must support a usable narrow-screen path, typically horizontal scroll, reduced columns, or a detail drawer.
- Every async screen needs loading, empty, error/retry, and permission-denied behavior. Use keyboard-accessible controls, visible focus states, semantic labels, and `aria-label` for icon-only actions.
- Use the configured Thai locale and `Noto Sans Thai` style consistently. Keep user-facing validation and error copy clear and consistent with the surrounding screen.

### Apple-inspired human interface principles

Design for the qualities associated with Apple Human Interface Guidelines—clarity, deference, and depth—without copying Apple branding, proprietary visuals, or platform-specific controls where Ant Design already provides a consistent web control.

#### Clarity: make the next action obvious

- Each page has one clear primary task and at most one visually dominant primary action.
- Use familiar labels and plain language. Prefer “บันทึกการเปลี่ยนแปลง” over vague labels such as “ดำเนินการ”.
- Make hierarchy visible through spacing, grouping, typography, and progressive disclosure—not decorative boxes or color alone.
- Keep tables scannable: prioritize the few columns needed to decide, move secondary data into a detail drawer, and preserve filters/search near the data they affect.
- Use system status and destructive states deliberately: distinguish success, warning, error, unavailable, and no-permission with text, icon, and color.

#### Deference: let content lead

- Interface chrome must stay quiet. Use white/light surfaces, restrained borders, and sparse shadows; avoid gradients, ornamental illustrations, excessive badges, or competing accent colors.
- Motion is feedback, not decoration. Prefer short opacity/transform transitions; respect `prefers-reduced-motion`; never block an action behind an animation.
- Use meaningful empty states that explain why there is no data and provide the next safe action when one exists.
- Put irreversible or high-impact actions behind explicit confirmation. Describe the affected scope and audit reason before submitting.

#### Depth and continuity: make state understandable

- Preserve user context across list → detail → back navigation. Use a Drawer for contextual inspection, a dedicated route for a deep task, and a Modal only for a focused decision.
- Update the interface immediately when an action is accepted: show pending/loading state, prevent duplicate submissions, then surface success or a recoverable error near the action.
- Do not silently change scope, permissions, filters, or selected product. Show the active scope and explain changes that invalidate visible data.
- Maintain predictable placement: primary navigation, scope selector, page title, filters, row actions, and confirmations should appear consistently across features.

#### Responsive, touch, and input behavior

- Design from the smallest useful viewport upward. Verify 320px, 375px, tablet, laptop, and ultrawide layouts rather than only shrinking a desktop screen.
- Support mouse, keyboard, and touch equally. Icon-only buttons require an accessible name and a target large enough for touch (aim for approximately 44 × 44 CSS px when space allows).
- Do not rely on hover for essential content or actions. Touch users must be able to reveal the same controls through tap/focus.
- Avoid horizontal page overflow. For dense data, use an intentionally scrollable table region, hide nonessential columns at narrow widths, or offer a detail view—never compress text until unreadable.
- Keep forms single-column on narrow screens, retain labels above inputs, use the appropriate input type, and preserve entered values after validation or recoverable network failure.

#### Accessibility and inclusive defaults

- Meet WCAG-aware defaults: sufficient contrast, text labels in addition to color, visible focus, semantic headings, and keyboard-operable dialogs/drawers.
- Announce async success/error with Ant Design feedback components and keep focus sensible after opening or closing a Drawer/Modal.
- Respect user font scaling and browser zoom. Never lock text size, suppress zoom, or encode essential information solely in animation, color, or position.
- Treat loading as a state, not a blank screen: preserve layout with Skeletons where helpful and never imply data is absent while a request is pending.

#### Review questions for every UI change

- Can a first-time user identify the page purpose and primary action within a few seconds?
- Is the current scope, status, and consequence of a destructive action clear before confirmation?
- Does the flow work with keyboard only, touch only, reduced motion, and a narrow viewport?
- Is every transition either immediate or accompanied by clear feedback?
- Does the screen remove visual noise without hiding information the user needs to decide?
`r`n### Page and component conventions

- Pages are route-level orchestration; move shared or feature-specific presentation into `components/` or `features/` once reuse is real.
- Prefer named exports except the application entry point or a third-party convention that requires a default export.
- Keep TypeScript strict. Avoid `any`; model service inputs and responses explicitly.
- Keep components focused. Extract helpers, types, column definitions, and large drawers/forms when a page becomes difficult to review (roughly 300–400 lines is a signal, not a hard limit).
- Use `useMemo` and `useCallback` only for measurable computation, stable dependencies, or memoized children—not by default.
- Debounce free-text searches that issue remote requests. Use backend pagination/filtering for audit logs, log explorers, and other large datasets; do not fetch all records into the browser.
- Lazy loading and route-level code splitting are desirable for heavy future modules, but only add them with a proper loading fallback and after checking the existing route setup.

### Frontend verification checklist

- Run `npm run build` after frontend changes; it includes TypeScript checking and production bundling.
- Run `npm run lint` when changing JSX/TypeScript broadly or adding new patterns.
- For new or changed routes, verify guest, authenticated, forbidden, and direct-URL behavior.
- For API work, verify request parameters, response envelope, auth header behavior, pagination, errors, and scope isolation against the backend handler or OpenAPI source.
- For a table/detail screen, verify filters reset pagination, loading/empty/error states, row identity, keyboard operation, narrow viewport behavior, and that the detail view does not reveal restricted fields.
- If the backend API contract changes, update its source contract and run the backend OpenAPI validation/synchronization; do not hand-edit generated OpenAPI copies.
