import { Navigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '@/store';
import { ROUTES } from '@/constants';
import type { ReactNode } from 'react';

interface ProtectedRouteProps {
  children: ReactNode;
  requiredPermissions?: string[];
  requireAll?: boolean;
}

/**
 * ProtectedRoute - ตรวจสอบ authentication และ authorization
 * - ถ้ายังไม่ login -> redirect ไป /login
 * - ถ้า login แล้วแต่ไม่มี permission -> redirect ไป /403
 * - ถ้า requiredPermissions เป็น undefined หรือ [] -> ผ่านได้หมดทุก role (แค่ login)
 */
export function ProtectedRoute({
  children,
  requiredPermissions,
  requireAll = false,
}: ProtectedRouteProps) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const permissions = useAuthStore((state) => state.permissions);
  const location = useLocation();

  // ยังไม่ login
  if (!isAuthenticated) {
    return <Navigate to={ROUTES.LOGIN} state={{ from: location }} replace />;
  }

  // เช็ค permission (ถ้าไม่ระบุ หรือ [] = ทุก role เข้าได้)
  if (requiredPermissions && requiredPermissions.length > 0) {
    const hasAccess = requireAll
      ? requiredPermissions.every((p) => permissions.includes(p))
      : requiredPermissions.some((p) => permissions.includes(p));

    if (!hasAccess) {
      return <Navigate to={ROUTES.FORBIDDEN} replace />;
    }
  }

  return <>{children}</>;
}

/**
 * GuestRoute - เฉพาะ guest (ยังไม่ login)
 * ถ้า login แล้ว -> redirect ไป dashboard
 */
export function GuestRoute({ children }: { children: ReactNode }) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  if (isAuthenticated) {
    return <Navigate to={ROUTES.DASHBOARD} replace />;
  }

  return <>{children}</>;
}
