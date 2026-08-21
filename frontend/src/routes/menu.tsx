import type { ReactNode } from "react";
import {
  HomeIcon,
  UsersIcon,
  MagnifyingGlassIcon,
  ClipboardDocumentCheckIcon,
  ClockIcon,
} from "@heroicons/react/24/outline";
import { ROUTES, PERMISSIONS, MENU_KEYS } from "@/constants";
import type { MenuItem } from "@/types";

/**
 * Sidebar Menu Config
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

  {
    key: MENU_KEYS.AUDIT_LOGS,
    label: "Audit Logs",
    icon: (<ClipboardDocumentCheckIcon className="w-5 h-5" />) as ReactNode,
    path: ROUTES.AUDIT_LOGS,
  },

  {
    key: MENU_KEYS.LOGS_EXPLORE,
    label: "Logs Explorer",
    icon: (<MagnifyingGlassIcon className="w-5 h-5" />) as ReactNode,
    path: ROUTES.LOGS_EXPLORE,
  },
  {
    key: MENU_KEYS.RETENTION_POLICIES,
    label: "Retention Policy",
    icon: (<ClockIcon className="w-5 h-5" />) as ReactNode,
    path: ROUTES.RETENTION_POLICIES,
  },
  {
    key: MENU_KEYS.CONNECTIONS,
    label: "Connections",
    icon: (<MagnifyingGlassIcon className="w-5 h-5" />) as ReactNode,
    path: ROUTES.CONNECTIONS,
  },
];
