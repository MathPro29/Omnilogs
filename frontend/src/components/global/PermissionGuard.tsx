import type { ReactNode } from 'react';
import { usePermission } from '@/hooks';

interface PermissionGuardProps {
  /** Permission ที่ต้องมี (ใช้ canAny ถ้าส่งเป็น array) */
  permission?: string;
  permissions?: string[];
  /** ต้องมีทุก permission หรือเปล่า (default: false = canAny) */
  requireAll?: boolean;
  /** Content ที่จะแสดงถ้ามี permission */
  children: ReactNode;
  /** Content ที่จะแสดงถ้าไม่มี permission */
  fallback?: ReactNode;
}

/**
 * Component สำหรับซ่อน/แสดง UI ตาม permission
 * ใช้ครอบ button, section, หรือ component ที่ต้องเช็คสิทธิ์
 */
export function PermissionGuard({
  permission,
  permissions,
  requireAll = false,
  children,
  fallback = null,
}: PermissionGuardProps) {
  const { can, canAny, canAll } = usePermission();

  let hasAccess = false;

  if (permission) {
    hasAccess = can(permission);
  } else if (permissions && permissions.length > 0) {
    hasAccess = requireAll ? canAll(permissions) : canAny(permissions);
  } else {
    hasAccess = true; // ถ้าไม่ระบุ permission = แสดงเสมอ
  }

  if (!hasAccess) {
    return <>{fallback}</>;
  }

  return <>{children}</>;
}
