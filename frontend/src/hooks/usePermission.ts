import { useAuthStore } from '@/store';

/**
 * Checks UI permissions. Platform administrators retain the same UI access as
 * the backend's platform-role middleware.
 */
export function usePermission() {
  const permissions = useAuthStore((state) => state.permissions);
  const roles = useAuthStore((state) => state.roles);
  const isPlatformAdmin = roles.some((role) => {
    const normalized = role.trim().toLowerCase().replace(/\s+/g, '_');
    return normalized === 'god' || normalized === 'owner' || normalized === 'superadmin' || normalized === 'super_admin' || normalized === 'admin';
  });

  const can = (permission: string): boolean => {
    return isPlatformAdmin || permissions.includes(permission);
  };

  const canAny = (permissionList: string[]): boolean => {
    return isPlatformAdmin || permissionList.some((permission) => permissions.includes(permission));
  };

  const canAll = (permissionList: string[]): boolean => {
    return isPlatformAdmin || permissionList.every((permission) => permissions.includes(permission));
  };

  return { can, canAny, canAll, permissions, isPlatformAdmin };
}