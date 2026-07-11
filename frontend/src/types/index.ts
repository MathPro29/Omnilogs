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
  username: string;
  password: string;
}

export interface LoginResponse {
  accessToken: string;
  refreshToken?: string;
  user: User;
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
