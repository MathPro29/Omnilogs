import type { ReactNode } from 'react';
import { lazy } from 'react';
import { ROUTES, PERMISSIONS } from '@/constants';

const DashboardPage = lazy(() => import('@/pages/DashboardPage').then(module => ({ default: module.DashboardPage })));
const UsersPage = lazy(() => import('@/pages/UsersPage').then(module => ({ default: module.UsersPage })));
const ForbiddenPage = lazy(() => import('@/pages/ForbiddenPage').then(module => ({ default: module.ForbiddenPage })));
const AuditsPage = lazy(() => import('@/pages/Audits').then(module => ({ default: module.default })));
const ProductCatalogPage = lazy(() => import('@/pages/ProductCatalogPage').then(module => ({ default: module.ProductCatalogPage })));
const AccessControlPage = lazy(() => import('@/pages/AccessControlPage').then(module => ({ default: module.AccessControlPage })));
const ApiTestPage = lazy(() => import('@/pages/ApiTestPage').then(module => ({ default: module.ApiTestPage })));
const LogsExplorerPage = lazy(() => import('@/pages/LogsExplorerPage').then(module => ({ default: module.LogsExplorerPage })));
const RetentionTestPage = lazy(() => import('@/pages/RetentionTestPage').then(module => ({ default: module.RetentionTestPage })));
const SensitiveAccessPage = lazy(() => import('@/pages/SensitiveAccessPage').then(module => ({ default: module.SensitiveAccessPage })));
const CustomFieldsPage = lazy(() => import('@/pages/CustomFieldsPage').then(module => ({ default: module.CustomFieldsPage })));
const ProductRoleConsolePage = lazy(() => import('@/pages/ProductRoleConsolePage').then(module => ({ default: module.ProductRoleConsolePage })));

/**
 * Route item config
 * เพิ่มหน้าใหม่ = เพิ่ม 1 entry ที่นี่ + สร้าง page component
 */
export interface RouteConfig {
  path: string;
  /** page component ที่จะ render */
  element: ReactNode;
  /** permissions ที่ต้องมี (undefined หรือ [] = ทุก role เข้าได้ แค่ login) */
  requiredPermissions?: string[];
  /** ต้องมี permission ทุกตัว (default: false) */
  requireAll?: boolean;
}

/**
 * Admin Routes - หน้าที่ใช้ AdminLayout (sidebar + header)
 *
 * วิธีเพิ่มหน้าใหม่:
 * 1. สร้าง page component ใน src/pages/
 * 2. เพิ่ม route config ที่นี่
 * 3. เพิ่ม menu config ที่ src/routes/menu.tsx (ถ้าต้องการแสดงใน sidebar)
 * 4. เพิ่ม permission key ที่ src/constants/index.ts
 */
export const adminRoutes: RouteConfig[] = [
  {
    path: ROUTES.API_TEST,
    element: <ApiTestPage />,
    requiredPermissions: [PERMISSIONS.FEATURE_API_TEST],
  },
  {
    path: ROUTES.DASHBOARD,
    element: <DashboardPage />,
    requiredPermissions: [PERMISSIONS.FEATURE_DASHBOARD],
  },
  {
    path: ROUTES.PRODUCTS,
    element: <ProductCatalogPage />,
    requiredPermissions: [PERMISSIONS.FEATURE_PRODUCTS],
  },
  {
    path: ROUTES.RETENTION_TEST,
    element: <RetentionTestPage />,
    requiredPermissions: [PERMISSIONS.FEATURE_RETENTION_TEST],
  },
  {
    path: ROUTES.USERS,
    element: <UsersPage />,
    requiredPermissions: [PERMISSIONS.FEATURE_USERS],
  },
  {
    path: ROUTES.ROLES,
    element: <AccessControlPage />,
    requiredPermissions: [PERMISSIONS.FEATURE_ROLES],
  },
  {
    path: ROUTES.AUDITS,
    element: <AuditsPage />,
    requiredPermissions: [PERMISSIONS.FEATURE_AUDITS],
  },
  {
    path: ROUTES.LOGS_EXPLORER,
    element: <LogsExplorerPage />,
    requiredPermissions: [PERMISSIONS.FEATURE_LOGS_EXPLORER],
  },
  {
    path: ROUTES.SENSITIVE_ACCESS,
    element: <SensitiveAccessPage />,
    requiredPermissions: [PERMISSIONS.FEATURE_SENSITIVE_ACCESS],
  },
  {
    path: ROUTES.CUSTOM_FIELDS,
    element: <CustomFieldsPage />,
    requiredPermissions: [PERMISSIONS.FEATURE_CUSTOM_FIELDS],
  },
  {
    path: ROUTES.PRODUCT_ROLES,
    element: <ProductRoleConsolePage />,
    requiredPermissions: [PERMISSIONS.FEATURE_PRODUCT_ROLES],
  },
  {
    path: ROUTES.FORBIDDEN,
    element: <ForbiddenPage />,
    // ไม่ต้องเช็ค permission - แค่ต้อง login
  },
];
