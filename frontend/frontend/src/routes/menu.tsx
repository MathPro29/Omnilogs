import type { ReactNode } from "react";
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
  ShieldCheckIcon,
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
  {
    key: MENU_KEYS.API_TEST,
    label: "API Test",
    icon: (<BeakerIcon className="w-5 h-5" />) as ReactNode,
    path: ROUTES.API_TEST,
    requiredPermissions: [PERMISSIONS.FEATURE_API_TEST],
  },
  // ===== แดชบอร์ด =====
  {
    key: MENU_KEYS.DASHBOARD,
    label: "แดชบอร์ด",
    icon: (<HomeIcon className="w-5 h-5" />) as ReactNode,
    path: ROUTES.DASHBOARD,
    requiredPermissions: [PERMISSIONS.FEATURE_DASHBOARD],
  },
  // ===== Logs Explorer =====
  {
    key: MENU_KEYS.LOGS_EXPLORER,
    label: "Logs Explorer",
    icon: (<ChartBarSquareIcon className="w-5 h-5" />) as ReactNode,
    path: ROUTES.LOGS_EXPLORER,
    requiredPermissions: [PERMISSIONS.FEATURE_LOGS_EXPLORER],
  },
  {
    key: MENU_KEYS.PRODUCTS,
    label: "ผลิตภัณฑ์ที่จัดการ",
    icon: (<Squares2X2Icon className="w-5 h-5" />) as ReactNode,
    path: ROUTES.PRODUCTS,
    requiredPermissions: [PERMISSIONS.FEATURE_PRODUCTS],
  },
  {
    key: MENU_KEYS.RETENTION_TEST,
    label: "Retention",
    icon: (<ArchiveBoxIcon className="w-5 h-5" />) as ReactNode,
    path: ROUTES.RETENTION_TEST,
    requiredPermissions: [PERMISSIONS.FEATURE_RETENTION_TEST],
  },

  // ===== จัดการผู้ใช้ (submenu) =====
  {
    key: MENU_KEYS.USER_MANAGEMENT,
    label: "จัดการผู้ใช้",
    icon: (<UsersIcon className="w-5 h-5" />) as ReactNode,
    requiredPermissions: [PERMISSIONS.FEATURE_USERS],
    children: [
      {
        key: MENU_KEYS.USERS,
        label: "ผู้ใช้งาน",
        path: ROUTES.USERS,
        requiredPermissions: [PERMISSIONS.FEATURE_USERS],
      },
      {
        key: MENU_KEYS.ROLES,
        label: "บทบาทและสิทธิ์",
        path: ROUTES.ROLES,
        requiredPermissions: [PERMISSIONS.FEATURE_ROLES],
      },
    ],
  },

  // ===== บันทึกกิจกรรม (Audits) =====
  {
    key: MENU_KEYS.AUDITS,
    label: "Audit Logs",
    icon: (<ClipboardDocumentListIcon className="w-5 h-5" />) as ReactNode,
    path: ROUTES.AUDITS,
    requiredPermissions: [PERMISSIONS.FEATURE_AUDITS],
  },
  {
    key: MENU_KEYS.SENSITIVE_ACCESS,
    label: "Sensitive log access",
    icon: (<LockClosedIcon className="w-5 h-5" />) as ReactNode,
    path: ROUTES.SENSITIVE_ACCESS,
    requiredPermissions: [PERMISSIONS.FEATURE_SENSITIVE_ACCESS],
  },
  {
    key: MENU_KEYS.CUSTOM_FIELDS,
    label: "Custom Fields",
    icon: (<AdjustmentsHorizontalIcon className="w-5 h-5" />) as ReactNode,
    path: ROUTES.CUSTOM_FIELDS,
    requiredPermissions: [PERMISSIONS.FEATURE_CUSTOM_FIELDS],
  },
  {
    key: MENU_KEYS.PRODUCT_ROLES,
    label: "Product Roles",
    icon: (<ShieldCheckIcon className="w-5 h-5" />) as ReactNode,
    path: ROUTES.PRODUCT_ROLES,
    requiredPermissions: [PERMISSIONS.FEATURE_PRODUCT_ROLES],
  },
];
