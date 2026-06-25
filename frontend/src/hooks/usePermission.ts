import { useAuthStore } from '@/store';

/**
 * Hook สำหรับเช็ค permission ใน component
 */
export function usePermission() {
  const permissions = useAuthStore((state) => state.permissions);

  const can = (permission: string): boolean => {
    return permissions.includes(permission);
  };

  const canAny = (permissionList: string[]): boolean => {
    return permissionList.some((p) => permissions.includes(p));
  };

  const canAll = (permissionList: string[]): boolean => {
    return permissionList.every((p) => permissions.includes(p));
  };

  return { can, canAny, canAll, permissions };
}
