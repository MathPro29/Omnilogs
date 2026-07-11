import { menuConfig } from './menu';
import type { MenuItem } from '@/types';

/**
 * สร้าง flat list ของ menu items (รวม children) สำหรับหา breadcrumbs
 */
export function getFlatMenuItems(items: MenuItem[]): MenuItem[] {
  const result: MenuItem[] = [];
  items.forEach((item) => {
    result.push(item);
    if (item.children) {
      result.push(...getFlatMenuItems(item.children));
    }
  });
  return result;
}

/**
 * หา menu item จาก path
 */
export function findMenuByPath(path: string): MenuItem | undefined {
  return getFlatMenuItems(menuConfig).find((item) => item.path === path);
}
