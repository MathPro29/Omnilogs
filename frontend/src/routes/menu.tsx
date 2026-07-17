import type { ReactNode } from "react";
import {
  HomeIcon,
  UsersIcon,
  BoltIcon,
  MagnifyingGlassIcon,
  ClipboardDocumentCheckIcon,
} from "@heroicons/react/24/outline";
import { ROUTES, PERMISSIONS, MENU_KEYS } from "@/constants";
import type { MenuItem } from "@/types";

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
  // ===== แดชบอร์ด =====
  {
    key: MENU_KEYS.DASHBOARD,
    label: "แดชบอร์ด",
    icon: (<HomeIcon className="w-5 h-5" />) as ReactNode,
    path: ROUTES.DASHBOARD,
    requiredPermissions: [PERMISSIONS.DASHBOARD_VIEW],
  },

  // ===== จัดการผู้ใช้ (submenu) =====
  {
    key: MENU_KEYS.USER_MANAGEMENT,
    label: "จัดการผู้ใช้",
    icon: (<UsersIcon className="w-5 h-5" />) as ReactNode,
    requiredPermissions: [PERMISSIONS.USER_VIEW],
    children: [
      {
        key: MENU_KEYS.USERS,
        label: "ผู้ใช้งาน",
        path: ROUTES.USERS,
        requiredPermissions: [PERMISSIONS.USER_VIEW],
      },
      {
        key: MENU_KEYS.ROLES,
        label: "บทบาทและสิทธิ์",
        path: ROUTES.ROLES,
        requiredPermissions: [PERMISSIONS.ROLE_VIEW],
      },
    ],
  },

  // ===== เพิ่ม menu ใหม่ที่นี่ =====
  // {
  //   key: MENU_KEYS.LEAVES,
  //   label: 'ลาหยุด',
  //   icon: <CalendarIcon className="w-5 h-5" /> as ReactNode,
  //   path: ROUTES.LEAVES,
  //   requiredPermissions: [PERMISSIONS.LEAVE_VIEW],
  // },
  {
    key: MENU_KEYS.AUDIT_LOGS,
    label: "Audit Logs",
    icon: (<ClipboardDocumentCheckIcon className="w-5 h-5" />) as ReactNode,
    path: ROUTES.AUDIT_LOGS,
  },

  {
    key: MENU_KEYS.LOGS_LIVE,
    label: "Live Tail",
    icon: (<BoltIcon className="w-5 h-5" />) as ReactNode,
    path: ROUTES.LOGS_LIVE,
  },

  {
    key: "logs-explorer",
    label: "Logs Explorer",
    icon: (<MagnifyingGlassIcon className="w-5 h-5" />) as ReactNode,
    path: ROUTES.LOGS_EXPLORE,
    requiredPermissions: [PERMISSIONS.DASHBOARD_VIEW],
  },
];
