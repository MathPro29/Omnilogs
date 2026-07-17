// ===== Route Paths =====
// เพิ่มลด ROUTE ของ Frontend จะมีผลต่อ route ของ backend
export const ROUTES = {
  LOGIN: '/login',
  REGISTER: '/register',
  API_TEST: '/api-test',
  DASHBOARD: '/dashboard',
  PRODUCTS: '/products',
  USERS: '/users',
  USER_CREATE: '/users/create',
  USER_EDIT: '/users/:id/edit',
  USER_DETAIL: '/users/:id',
  ROLES: '/access-control',
  SETTINGS: '/settings',
  AUDITS: '/audits',
  LOGS_EXPLORER: '/logs-explorer',
  RETENTION_TEST: '/retention-test',
  SENSITIVE_ACCESS: '/sensitive-access',
  CUSTOM_FIELDS: '/custom-fields',
  PRODUCT_ROLES: '/product-roles',
  FORBIDDEN: '/403',
  NOT_FOUND: '/404',
  FORGOT_PASSWORD: '/forgot-password',
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

  // Approvals
  APPROVE_USER: 'approve:user',

  // Features / Pages Access (for custom user permissions)
  FEATURE_API_TEST: 'feature:api_test',
  FEATURE_DASHBOARD: 'feature:dashboard',
  FEATURE_PRODUCTS: 'feature:products',
  FEATURE_RETENTION_TEST: 'feature:retention_test',
  FEATURE_USERS: 'feature:users',
  FEATURE_ROLES: 'feature:roles',
  FEATURE_AUDITS: 'feature:audits',
  FEATURE_LOGS_EXPLORER: 'feature:logs_explorer',
  FEATURE_SENSITIVE_ACCESS: 'feature:sensitive_access',
  FEATURE_CUSTOM_FIELDS: 'feature:custom_fields',
  FEATURE_PRODUCT_ROLES: 'feature:product_roles',
} as const;

// ===== All Permissions (สำหรับ Super Admin) =====
export const ALL_PERMISSIONS = Object.values(PERMISSIONS);

// ===== Default Admin Features =====
export const DEFAULT_ADMIN_FEATURES = [
  PERMISSIONS.FEATURE_DASHBOARD,
  PERMISSIONS.FEATURE_LOGS_EXPLORER,
  PERMISSIONS.FEATURE_AUDITS,
  PERMISSIONS.FEATURE_SENSITIVE_ACCESS,
  PERMISSIONS.FEATURE_CUSTOM_FIELDS,
  PERMISSIONS.FEATURE_PRODUCT_ROLES,
  PERMISSIONS.FEATURE_API_TEST,
];

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
  API_TEST: 'api-test',
  DASHBOARD: 'dashboard',
  PRODUCTS: 'products',
  USER_MANAGEMENT: 'user-management',
  USERS: 'users',
  ROLES: 'access-control',
  SETTINGS: 'settings',
  AUDITS: 'audits',
  LOGS_EXPLORER: 'logs-explorer',
  RETENTION_TEST: 'retention-test',
  SENSITIVE_ACCESS: 'sensitive-access',
  CUSTOM_FIELDS: 'custom-fields',
  PRODUCT_ROLES: 'product-roles',
} as const;

// ===== API Endpoints =====
export const API_ENDPOINTS = {
  AUTH: {
    LOGIN: '/auth/login',
    REGISTER: '/auth/register',
    LOGOUT: '/auth/logout',
    REFRESH: '/auth/refresh',
    ME: '/auth/me',
  },
  USERS: {
    LIST: '/users',
    DETAIL: (id: string) => `/users/${id}`,
    CREATE: '/users',
    UPDATE: (id: string) => `/users/${id}`,
    DELETE: (id: string) => `/users/${id}`,
  },
  ROLES: {
    LIST: '/roles',
    DETAIL: (id: string) => `/roles/${id}`,
  },
  DASHBOARD: {
    SUMMARY: '/dashboard/summary',
  },
  SETTINGS: {
    GET: '/settings',
    UPDATE: '/settings',
  },
  ADMIN: {
    USERS: '/admin/show-users',
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
  active: 'ใช้งาน',
  inactive: 'ไม่ใช้งาน',
  pending: 'รอดำเนินการ',
};

// ===== Status Colors =====
export const STATUS_COLORS: Record<string, string> = {
  active: 'success',
  inactive: 'error',
  pending: 'warning',
};
