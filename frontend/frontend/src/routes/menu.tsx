import type { ReactNode } from 'react';
import {
  HomeIcon,
  UsersIcon,
  ClipboardDocumentListIcon,
  BeakerIcon,
  Squares2X2Icon,
  ChartBarSquareIcon,
  ArchiveBoxIcon,
  LockClosedIcon,
  AdjustmentsHorizontalIcon,
} from '@heroicons/react/24/outline';
import { ROUTES, PERMISSIONS, MENU_KEYS } from '@/constants';
import type { MenuItem } from '@/types';

/**
 * Sidebar Menu Config
 *
 * ลำดับของ array = ลำดับการแสดงผลใน sidebar
 * ระบบจะ filter ตาม user permissions อัตโนมัติ
 * requiredPermissions เป็น undefined หรือ [] = ทุก role เห็น
 *
 * วิธีเพิ่ม menu ใหม่:
 * 1. เพิ่ม MENU_KEYS ใน src/constants/index.ts
 * 2. เพิ่ม item ที่นี่ ระบุ key, label, icon, path, requiredPermissions
 * 3. ถ้าเป็น submenu ให้ใส่ใน children array
 */
export const menuConfig: MenuItem[] = [
  {
    key: MENU_KEYS.API_TEST,
    label: 'API Test',
    icon: <BeakerIcon className="w-5 h-5" /> as ReactNode,
    path: ROUTES.API_TEST,
  },
  // ===== แดชบอร์ด =====
  {
    key: MENU_KEYS.DASHBOARD,
    label: 'แดชบอร์ด',
    icon: <HomeIcon className="w-5 h-5" /> as ReactNode,
    path: ROUTES.DASHBOARD,
    requiredPermissions: [PERMISSIONS.DASHBOARD_VIEW],
  },
  // ===== Logs Explorer =====
  {
    key: MENU_KEYS.LOGS_EXPLORER,
    label: 'Logs Explorer',
    icon: <ChartBarSquareIcon className="w-5 h-5" /> as ReactNode,
    path: ROUTES.LOGS_EXPLORER,
  },
  {
    key: MENU_KEYS.PRODUCTS,
    label: 'ผลิตภัณฑ์ที่จัดการ',
    icon: <Squares2X2Icon className="w-5 h-5" /> as ReactNode,
    path: ROUTES.PRODUCTS,
  },
  {
    key: MENU_KEYS.RETENTION_TEST,
    label: 'Retention',
    icon: <ArchiveBoxIcon className="w-5 h-5" /> as ReactNode,
    path: ROUTES.RETENTION_TEST,
  },

  // ===== จัดการผู้ใช้ (submenu) =====
  {
    key: MENU_KEYS.USER_MANAGEMENT,
    label: 'จัดการผู้ใช้',
    icon: <UsersIcon className="w-5 h-5" /> as ReactNode,
    requiredPermissions: [PERMISSIONS.USER_VIEW],
    children: [
      {
        key: MENU_KEYS.USERS,
        label: 'ผู้ใช้งาน',
        path: ROUTES.USERS,
        requiredPermissions: [PERMISSIONS.USER_VIEW],
      },
      {
        key: MENU_KEYS.ROLES,
        label: 'บทบาทและสิทธิ์',
        path: ROUTES.ROLES,
        requiredPermissions: [PERMISSIONS.ROLE_VIEW],
      },
    ],
  },

  // ===== บันทึกกิจกรรม (Audits) =====
  {
    key: MENU_KEYS.AUDITS,
    label: 'บันทึกกิจกรรม',
    icon: <ClipboardDocumentListIcon className="w-5 h-5" /> as ReactNode,
    path: ROUTES.AUDITS,
  },
  {
    key: MENU_KEYS.SENSITIVE_ACCESS,
    label: 'Sensitive log access',
    icon: <LockClosedIcon className="w-5 h-5" /> as ReactNode,
    path: ROUTES.SENSITIVE_ACCESS,
  },
  {
    key: MENU_KEYS.CUSTOM_FIELDS,
    label: 'Custom Fields',
    icon: <AdjustmentsHorizontalIcon className="w-5 h-5" /> as ReactNode,
    path: ROUTES.CUSTOM_FIELDS,
  }
];
