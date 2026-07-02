// ===== User Types =====
export interface User {
  id: string;
  username: string;
  email: string;
  fullName: string;
  avatar?: string;
  phone?: string;
  department?: string;
  position?: string;
  status: UserStatus;
  roles: Role[];
  permissions: string[];
  createdAt: string;
  updatedAt: string;
}

export type UserStatus = 'active' | 'inactive' | 'pending';

// ===== Role Types =====
export interface Role {
  id: string;
  name: string;
  description?: string;
  permissions: string[];
}

// ===== Auth Types =====
export interface LoginRequest {
  identifier: string;
  password: string;
}

export interface LoginResponse {
  accessToken: string;
  refreshToken?: string;
  user: User;
}

export interface ProductEnvironment {
  environmentId: number;
  productId: number;
  environmentCode: string;
  environmentName: string;
  createdAt?: string;
}

export interface Product {
  productId: number;
  productName: string;
  productCode: string;
  productEnvironments?: ProductEnvironment[];
  isActive: boolean;
  createdAt?: string;
  updatedAt?: string;
}

export interface Project {
  projectId: number;
  productId: number;
  projectCode: string;
  projectName: string;
  isActive: boolean;
  createdAt?: string;
  updatedAt?: string;
}

export interface ProjectFeature {
  categoryId: number;
  productId: number;
  projectId: number;
  parentId?: number | null;
  categoryType?: string | null;
  categoryCode: string;
  categoryName: string;
  fullPath?: string | null;
  pathIds?: string | null;
  level: number;
  isActive: boolean;
  createdAt?: string;
  updatedAt?: string;
}

export interface ProductRoleDefinition {
  roleId: number;
  productId: number;
  roleCode: string;
  roleName: string;
  permissions: RolePermissionAssignment[];
  isActive: boolean;
  createdAt?: string;
  updatedAt?: string;
}

export interface RolePermissionAssignment {
  resourceType: string;
  action: string;
}

export interface ProductMembership {
  membershipId: number;
  userId: number;
  productId: number;
  roleId: number;
  expiresAt?: string | null;
  isActive: boolean;
  createdAt?: string;
  updatedAt?: string;
}

export interface MembershipScope {
  scopeId: number;
  membershipId: number;
  productId: number;
  projectId?: number | null;
  categoryId?: number | null;
  scopeLevel: 'PRODUCT' | 'PROJECT' | 'CATEGORY';
  isActive: boolean;
  createdAt?: string;
  updatedAt?: string;
}

export interface PermissionRule {
  permissionRuleId: number;
  userId: number;
  productId?: number | null;
  roleId?: number | null;
  projectId?: number | null;
  categoryId?: number | null;
  resourceType: string;
  action: string;
  effect: 'ALLOW' | 'DENY';
  scopeLevel: 'GLOBAL' | 'PRODUCT' | 'PROJECT' | 'CATEGORY';
  grantedBy?: number | null;
  isActive: boolean;
  expiresAt?: string | null;
  createdAt?: string;
  updatedAt?: string;
}

export interface PermissionCheckResult {
  allowed: boolean;
  matchedRuleId?: number | null;
  reason?: string;
}

export interface ProductApiKey {
  keyId: number;
  productId: number;
  environmentId?: number | null;
  keyName?: string | null;
  keyPrefix: string;
  permissions: unknown;
  isActive: boolean;
  expiresAt?: string | null;
  createdAt?: string;
  updatedAt?: string;
}

export interface MainLog {
  logId: string;
  productId: number;
  environmentId?: number | null;
  sourceId?: number | null;
  timestamp?: string | null;
  level?: string | null;
  logType?: string | null;
  message?: string | null;
  requestId?: string | null;
  traceId?: string | null;
  method?: string | null;
  path?: string | null;
  url?: string | null;
  statusCode?: number | null;
  latencyMs?: number | null;
  requestHeaders?: Record<string, unknown>;
  responseHeaders?: Record<string, unknown>;
  requestPayload?: unknown;
  responsePayload?: unknown;
  errorCode?: string | null;
  errorMessage?: string | null;
  stackTrace?: string | null;
  customFields?: Record<string, unknown>;
  raw: Record<string, unknown>;
}

export interface MainLogSearchResponse {
  data: MainLog[];
  total: number;
  page: number;
  perPage: number;
}

export interface LogQueueBatch {
  batchId: string;
  productId?: number | null;
  sourceId?: number | null;
  environmentId?: number | null;
  queueKey: string;
  sourceType: string;
  sourcePlatform: string;
  workerId?: string | null;
  processingAttempts: number;
  totalLogs: number;
  status: string;
  priority: number;
  receivedAt?: string | null;
  processingStartedAt?: string | null;
  processedAt?: string | null;
  errorMessage?: string | null;
}

export interface LogQueueItem {
  queueItemId: number;
  batchId: string;
  sequenceNo: number;
  sourceType: string;
  sourcePlatform: string;
  status: string;
  retryCount: number;
  maxRetryCount: number;
  nextRetryAt?: string | null;
  lastRetryAt?: string | null;
  processedAt?: string | null;
  errorMessage?: string | null;
}

export interface AuthState {
  currentUser: User | null;
  roles: string[];
  permissions: string[];
  isAuthenticated: boolean;
  accessToken: string | null;
  setAuth: (data: LoginResponse) => void;
  clearAuth: () => void;
  updatePermissions: (permissions: string[]) => void;
}

// ===== API Types =====
export interface ApiResponse<T> {
  data: T;
  message: string;
  success: boolean;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface PaginationParams {
  page: number;
  pageSize: number;
}

export interface SortParams {
  sortBy?: string;
  sortOrder?: 'ascend' | 'descend';
}

// ===== Filter Types =====
export interface UserFilterParams extends PaginationParams, SortParams {
  search?: string;
  status?: UserStatus;
  department?: string;
  role?: string;
}

// ===== Dashboard Types =====
export interface DashboardSummary {
  totalUsers: number;
  activeUsers: number;
  newUsersThisMonth: number;
  departments: number;
  recentActivities: Activity[];
  usersByDepartment: ChartData[];
  userGrowth: ChartData[];
}

export interface Activity {
  id: string;
  action: string;
  user: string;
  target: string;
  timestamp: string;
}

export interface ChartData {
  label: string;
  value: number;
}

// ===== Menu Types =====
export interface MenuItem {
  key: string;
  label: string;
  icon?: React.ReactNode;
  path?: string;
  requiredPermissions?: string[];
  children?: MenuItem[];
}

// ===== Settings Types =====
export interface SystemSettings {
  siteName: string;
  siteDescription: string;
  maintenanceMode: boolean;
  allowRegistration: boolean;
  defaultRole: string;
  sessionTimeout: number;
  maxLoginAttempts: number;
}

// ===== UI State Types =====
export interface UIState {
  sidebarCollapsed: boolean;
  toggleSidebar: () => void;
  setSidebarCollapsed: (collapsed: boolean) => void;
}

export interface AppState {
  pageTitle: string;
  setPageTitle: (title: string) => void;
  breadcrumbs: BreadcrumbItem[];
  setBreadcrumbs: (items: BreadcrumbItem[]) => void;
}

export interface BreadcrumbItem {
  title: string;
  path?: string;
}
