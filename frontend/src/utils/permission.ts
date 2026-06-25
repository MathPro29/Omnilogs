import { useAuthStore } from '@/store';

/**
 * ตรวจว่า user มี permission ที่กำหนดหรือไม่
 */
export function can(permission: string): boolean {
  const { permissions } = useAuthStore.getState();
  return permissions.includes(permission);
}

/**
 * ตรวจว่า user มี permission ใด permission หนึ่งในรายการหรือไม่
 */
export function canAny(permissionList: string[]): boolean {
  const { permissions } = useAuthStore.getState();
  return permissionList.some((p) => permissions.includes(p));
}

/**
 * ตรวจว่า user มี permission ทุกตัวในรายการหรือไม่
 */
export function canAll(permissionList: string[]): boolean {
  const { permissions } = useAuthStore.getState();
  return permissionList.every((p) => permissions.includes(p));
}

/**
 * Filter menu items ตาม permissions ของ user
 */
export function filterMenuByPermissions<T extends { requiredPermissions?: string[]; children?: T[] }>(
  items: T[]
): T[] {
  return items
    .filter((item) => {
      if (!item.requiredPermissions || item.requiredPermissions.length === 0) {
        return true;
      }
      return canAny(item.requiredPermissions);
    })
    .map((item) => {
      if (item.children && item.children.length > 0) {
        return {
          ...item,
          children: filterMenuByPermissions(item.children),
        };
      }
      return item;
    })
    .filter((item) => {
      // ถ้ามี children แต่ filter แล้วไม่เหลือ ให้ซ่อน parent ด้วย
      if (item.children && item.children.length === 0) {
        return false;
      }
      return true;
    });
}
