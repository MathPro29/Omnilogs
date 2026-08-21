// ===== Route Paths =====
export const ROUTES = {
  LOGIN: '/login',
  REGISTER: '/register',
  FORGOT_PASSWORD: '/forgot-password',
  DASHBOARD: '/dashboard',
  LOGS_EXPLORE : '/logs-explorer',
  USERS: '/users',
  USER_CREATE: '/users/create',
  USER_EDIT: '/users/:id/edit',
  USER_DETAIL: '/users/:id',
  ROLES: '/roles',
  SETTINGS: '/settings',
  AUDIT_LOGS: '/audit-logs',
  PRODUCTS: '/products',
  CONNECTIONS: '/connections',
  RETENTION_POLICIES: '/retention-policies',
  FORBIDDEN: '/403',
  NOT_FOUND: '/404',
} as const;

// ===== Permission Keys =====
export const PERMISSIONS = {
  // Dashboard
  DASHBOARD_VIEW: 'dashboard:view',

  // Users
  USER_VIEW: 'user:view',
  USER_CREATE: 'user:create',
  USER_EDIT: 'user:edit',
  USER_DELETE: 'user:delete',
  USER_EXPORT: 'user:export',

  // Roles
  ROLE_VIEW: 'role:view',
  ROLE_CREATE: 'role:create',
  ROLE_EDIT: 'role:edit',
  ROLE_DELETE: 'role:delete',

  // Settings
  SETTINGS_VIEW: 'settings:view',
  SETTINGS_EDIT: 'settings:edit',

  // Connections
  CONNECTION_VIEW: 'connection:view',

  // Approvals
  APPROVE_USER: 'approve:user',
} as const;

// ===== All Permissions (สำหรับ Super Admin) =====
export const ALL_PERMISSIONS = Object.values(PERMISSIONS);

// ===== Default Roles =====
export const DEFAULT_ROLES = {
  SUPER_ADMIN: 'super_admin',
  ADMIN: 'admin',
  MANAGER: 'manager',
  VIEWER: 'viewer',
} as const;

// ===== Role Permission Mapping =====
export const ROLE_PERMISSIONS: Record<string, string[]> = {
  [DEFAULT_ROLES.SUPER_ADMIN]: ALL_PERMISSIONS,
  [DEFAULT_ROLES.ADMIN]: [
    PERMISSIONS.DASHBOARD_VIEW,
    PERMISSIONS.USER_VIEW,
    PERMISSIONS.USER_CREATE,
    PERMISSIONS.USER_EDIT,
    PERMISSIONS.USER_EXPORT,
    PERMISSIONS.ROLE_VIEW,
    PERMISSIONS.SETTINGS_VIEW,
  ],
  [DEFAULT_ROLES.MANAGER]: [
    PERMISSIONS.DASHBOARD_VIEW,
    PERMISSIONS.USER_VIEW,
    PERMISSIONS.USER_EXPORT,
    PERMISSIONS.APPROVE_USER,
  ],
  [DEFAULT_ROLES.VIEWER]: [
    PERMISSIONS.DASHBOARD_VIEW,
    PERMISSIONS.USER_VIEW,
  ],
};

// ===== Menu Keys =====
export const MENU_KEYS = {
  DASHBOARD: 'dashboard',
  USER_MANAGEMENT: 'user-management',
  USERS: 'users',
  ROLES: 'roles',
  SETTINGS: 'settings',
  LOGS_EXPLORE: 'logs-explorer',
  AUDIT_LOGS: 'audit-logs',
  PRODUCTS: 'products',
  CONNECTIONS: 'connections',
  RETENTION_POLICIES: 'retention-policies',
} as const;

// ===== API Endpoints =====
export const API_ENDPOINTS = {
  AUTH: {
    LOGIN: '/v1/auth/login',
    REGISTER: '/v1/auth/register',
    FORGOT_PASSWORD: '/v1/auth/forgot-password',
    RESET_PASSWORD: '/v1/auth/reset-password',
    LOGOUT: '/v1/auth/logout',
    REFRESH: '/v1/auth/refresh-token',
    ME: '/v1/auth/me',
  },
  USERS: {
    LIST: '/v1/admin/show-users',
    DETAIL: (id: string) => `/v1/users/${id}`,
    CREATE: '/v1/users',
    UPDATE: (id: string) => `/v1/users/${id}`,
    DELETE: (id: string) => `/v1/users/${id}`,
  },
  ROLES: {
    LIST: '/roles',
    DETAIL: (id: string) => `/roles/${id}`,
  },
  DASHBOARD: {
    SUMMARY: '/dashboard/summary',
  },
  AUDIT_LOGS: {
    LIST: '/v1/audit-logs',
    ACTIVITIES: '/v1/audit-logs/activities',
    DETAIL: (auditId: string) => `/v1/audit-logs/${auditId}`,
  },
  LOGS_EXPLORE: {
    LIST: '/logs-explorer',
  },
  SETTINGS: {
    GET: '/settings',
    UPDATE: '/settings',
  },
  CONNECTIONS: {
    LIST: '/connections',
  },
  RETENTION_POLICIES: {
    LIST: (productId: number | string) => `/v1/products/${productId}/retention-policies`,
    DETAIL: (productId: number | string, id: number | string) => `/v1/products/${productId}/retention-policies/${id}`,
    CREATE: (productId: number | string) => `/v1/products/${productId}/retention-policies`,
    UPDATE: (productId: number | string, id: number | string) => `/v1/products/${productId}/retention-policies/${id}`,
    DELETE: (productId: number | string, id: number | string) => `/v1/products/${productId}/retention-policies/${id}`,
    TOGGLE: (productId: number | string, id: number | string) => `/v1/products/${productId}/retention-policies/${id}/toggle`,
    RUN_NOW: (productId: number | string, id: number | string) => `/v1/products/${productId}/retention-policies/${id}/run-now`,
    STATS: (productId: number | string) => `/v1/products/${productId}/retention-policies/stats`,
    SIMULATE: (productId: number | string) => `/v1/products/${productId}/retention-policies/simulate`,
    HISTORICAL: (productId: number | string) => `/v1/products/${productId}/retention-policies/historical`,
    PREVIEW: (productId: number | string) => `/v1/products/${productId}/retention-policies/preview`,
  },
  LOG_ARCHIVES: {
    LIST: (productId: number | string) => `/v1/products/${productId}/log-archives`,
    CREATE: (productId: number | string) => `/v1/products/${productId}/log-archives`,
    DOWNLOAD: (productId: number | string) => `/v1/products/${productId}/log-archives/download`,
    IMPORT: (productId: number | string) => `/v1/products/${productId}/log-archives/import`,
    READY_DOWNLOADS: (productId: number | string) => `/v1/products/${productId}/log-archive-downloads`,
    READY_DOWNLOAD: (productId: number | string, downloadId: string) => `/v1/products/${productId}/log-archive-downloads/${downloadId}`,
    VERIFY: (productId: number | string, archiveId: string) =>
      `/v1/products/${productId}/log-archives/${archiveId}/verify`,
    RESTORE: (productId: number | string, archiveId: string) =>
      `/v1/products/${productId}/log-archives/${archiveId}/restore`,
  },
} as const;

// ===== Query Keys =====
export const QUERY_KEYS = {
  DASHBOARD_SUMMARY: ['dashboard', 'summary'],
  USERS: ['users'],
  USER_DETAIL: (id: string) => ['users', id],
  ROLES: ['roles'],
  SETTINGS: ['settings'],
} as const;

// ===== Date Formats =====
export const DATE_FORMAT = {
  DISPLAY: 'DD/MM/YYYY',
  DISPLAY_TIME: 'DD/MM/YYYY HH:mm',
  DISPLAY_FULL: 'DD/MM/YYYY HH:mm:ss',
  API: 'YYYY-MM-DD',
  API_TIME: 'YYYY-MM-DDTHH:mm:ss',
} as const;

// ===== Pagination Defaults =====
export const PAGINATION = {
  DEFAULT_PAGE: 1,
  DEFAULT_PAGE_SIZE: 10,
  PAGE_SIZE_OPTIONS: ['10', '20', '50', '100'],
} as const;

// ===== Status Labels =====
export const STATUS_LABELS: Record<string, string> = {
  active: 'ใช้งานอยู่',
  inactive: 'ไม่ใช้งาน',
  pending: 'รอดำเนินการ',
};

// ===== Status Colors =====
export const STATUS_COLORS: Record<string, string> = {
  active: 'success',
  inactive: 'error',
  pending: 'warning',
};



