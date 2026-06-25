import type { ReactNode } from 'react';
import { ROUTES, PERMISSIONS } from '@/constants';
import {
  DashboardPage,
  UsersPage,
  SettingsPage,
  ForbiddenPage,
} from '@/pages';

/**
 * Route item config
 * เพิ่มหน้าใหม่ = เพิ่ม 1 entry ที่นี่ + สร้าง page component
 */
export interface RouteConfig {
  /** path ของ route */
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
    path: ROUTES.DASHBOARD,
    element: <DashboardPage />,
    requiredPermissions: [PERMISSIONS.DASHBOARD_VIEW],
  },
  {
    path: ROUTES.USERS,
    element: <UsersPage />,
    requiredPermissions: [PERMISSIONS.USER_VIEW],
  },
  // TODO: เพิ่มหน้าในอนาคต
  // {
  //   path: ROUTES.ROLES,
  //   element: <RolesPage />,
  //   requiredPermissions: [PERMISSIONS.ROLE_VIEW],
  // },
  {
    path: ROUTES.SETTINGS,
    element: <SettingsPage />,
    requiredPermissions: [PERMISSIONS.SETTINGS_VIEW],
  },
  {
    path: ROUTES.FORBIDDEN,
    element: <ForbiddenPage />,
    // ไม่ต้องเช็ค permission - แค่ต้อง login
  },
];
