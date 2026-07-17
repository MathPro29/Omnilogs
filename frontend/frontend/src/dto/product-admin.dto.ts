import type { RolePermissionAssignment } from '@/types';

export interface CreateProductPayload {
  productName: string;
  productCode?: string;
  environments?: Array<{ environmentCode: string; environmentName: string }>;
}

export interface CreateProjectPayload {
  projectCode?: string;
  projectName: string;
}

export interface CreateFeaturePayload {
  categoryCode?: string;
  categoryName: string;
  parentId?: number | null;
  categoryType?: string | null;
}

export interface CreateRolePayload {
  roleCode: string;
  roleName: string;
  permissions: RolePermissionAssignment[];
}

export interface CreateMembershipPayload {
  userId: number;
  roleId: number;
  expiresAt?: string | null;
}

export interface CreateScopePayload {
  projectId?: number | null;
  categoryId?: number | null;
  scopeLevel: 'PRODUCT' | 'PROJECT' | 'CATEGORY';
}

export interface CreatePermissionRulePayload {
  userId: number;
  productId: number;
  roleId?: number | null;
  projectId?: number | null;
  categoryId?: number | null;
  resourceType: string;
  action: string;
  effect: 'ALLOW' | 'DENY';
  scopeLevel: 'GLOBAL' | 'PRODUCT' | 'PROJECT' | 'CATEGORY';
  expiresAt?: string | null;
}

export interface SearchLogsParams {
  productId: number;
  environmentId?: number;
  projectId?: number;
  categoryId?: number;
  level?: string;
  logType?: string;
  keyword?: string;
  page?: number;
  perPage?: number;
}

export interface SearchLogsMultiParams {
  productId: number;
  environmentId?: number;
  projectIds?: number[];
  categoryIds?: number[];
  levels?: string[];
  keyword?: string;
  customFieldPath?: string;
  customFieldValue?: string;
  sortField?: string;
  sortOrder?: 'asc' | 'desc';
  page?: number;
  perPage?: number;
}

export interface DashboardLogStatsParams {
  productId: number;
  environmentId?: number;
  projectIds?: number[];
  categoryIds?: number[];
}

export interface ImportLogsPayload {
  productId: number;
  environmentId: number;
  projectId?: number;
  categoryId?: number;
  featureFullPath?: string | null;
  featurePathIds?: string | null;
  logLevel: string;
  message: string;
  eventType?: string;
  customFields?: Record<string, unknown>;
  rawData?: Record<string, unknown>;
}
