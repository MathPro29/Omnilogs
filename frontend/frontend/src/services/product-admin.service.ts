import axios from 'axios';
import { apiClient } from '@/api';
import type {
  ApiResponse,
  LogQueueBatch,
  LogQueueItem,
  MainLog,
  MembershipScope,
  PermissionCheckResult,
  PermissionRule,
  Product,
  ProductApiKey,
  ProductMembership,
  RolePermissionAssignment,
  ProductRoleDefinition,
  Project,
  ProjectFeature,
} from '@/types';

interface CreateProductPayload {
  productName: string;
  productCode?: string;
  environments?: Array<{ environmentCode: string; environmentName: string }>;
}

interface CreateProjectPayload {
  projectCode?: string;
  projectName: string;
}

interface CreateFeaturePayload {
  categoryCode?: string;
  categoryName: string;
  parentId?: number | null;
  categoryType?: string | null;
}

interface CreateRolePayload {
  roleCode: string;
  roleName: string;
  permissions: RolePermissionAssignment[];
}

interface CreateMembershipPayload {
  userId: number;
  roleId: number;
  expiresAt?: string | null;
}

interface CreateScopePayload {
  projectId?: number | null;
  categoryId?: number | null;
  scopeLevel: 'PRODUCT' | 'PROJECT' | 'CATEGORY';
}

interface CreatePermissionRulePayload {
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

interface SearchLogsParams {
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

interface SearchLogsMultiParams {
  productId: number;
  environmentId?: number;
  projectIds?: number[];
  categoryIds?: number[];
  levels?: string[];
  keyword?: string;
  customFieldPath?: string;
  customFieldValue?: string;
  page?: number;
  perPage?: number;
}

interface DashboardLogStatsParams {
  productId: number;
  environmentId?: number;
  projectIds?: number[];
  categoryIds?: number[];
}

interface ImportLogsPayload {
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
}

function shouldUseMock(): boolean {
  return import.meta.env.VITE_USE_MOCK === 'true';
}

let mockProducts: Product[] = [
  {
    productId: 1,
    productName: 'OmniLogs',
    productCode: 'OMNILOGS',
    isActive: true,
    createdAt: '2026-06-01T09:00:00Z',
    updatedAt: '2026-06-01T09:00:00Z',
  },
];

let mockProjects: Project[] = [
  { projectId: 1, productId: 1, projectCode: 'PROJA', projectName: 'Project A', isActive: true },
  { projectId: 2, productId: 1, projectCode: 'PROJB', projectName: 'Project B', isActive: true },
];

let mockFeatures: ProjectFeature[] = [
  { categoryId: 1, productId: 1, projectId: 1, categoryCode: 'FEA', categoryName: 'Feature A', level: 1, isActive: true, pathIds: '1', fullPath: 'FEA' },
  { categoryId: 2, productId: 1, projectId: 1, parentId: 1, categoryCode: 'FEA11', categoryName: 'Feature A11', level: 2, isActive: true, pathIds: '1,2', fullPath: 'FEA/FEA11' },
  { categoryId: 3, productId: 1, projectId: 1, parentId: 1, categoryCode: 'FEA12', categoryName: 'Feature A12', level: 2, isActive: true, pathIds: '1,3', fullPath: 'FEA/FEA12' },
  { categoryId: 4, productId: 1, projectId: 2, categoryCode: 'FEB', categoryName: 'Feature B', level: 1, isActive: true, pathIds: '4', fullPath: 'FEB' },
  { categoryId: 5, productId: 1, projectId: 2, parentId: 4, categoryCode: 'FEB21', categoryName: 'Feature B21', level: 2, isActive: true, pathIds: '4,5', fullPath: 'FEB/FEB21' },
  { categoryId: 6, productId: 1, projectId: 2, parentId: 4, categoryCode: 'FEB22', categoryName: 'Feature B22', level: 2, isActive: true, pathIds: '4,6', fullPath: 'FEB/FEB22' },
];

let mockRoles: ProductRoleDefinition[] = [
  {
    roleId: 1,
    productId: 1,
    roleCode: 'product_admin',
    roleName: 'Product Admin',
    permissions: [
      { resourceType: 'PROJECT', action: 'CREATE' },
      { resourceType: 'PROJECT', action: 'READ' },
      { resourceType: 'PROJECT', action: 'UPDATE' },
      { resourceType: 'FEATURE', action: 'CREATE' },
      { resourceType: 'FEATURE', action: 'READ' },
      { resourceType: 'FEATURE', action: 'UPDATE' },
      { resourceType: 'FEATURE', action: 'DELETE' },
      { resourceType: 'ACCESS', action: 'READ' },
      { resourceType: 'ACCESS', action: 'GRANT' },
      { resourceType: 'ROLE', action: 'READ' },
      { resourceType: 'ROLE', action: 'CREATE' },
      { resourceType: 'ROLE', action: 'UPDATE' },
    ],
    isActive: true,
  },
];

let mockMemberships: ProductMembership[] = [
  { membershipId: 1, userId: 1, productId: 1, roleId: 1, isActive: true },
];

let mockScopes: MembershipScope[] = [
  { scopeId: 1, membershipId: 1, productId: 1, scopeLevel: 'PRODUCT', isActive: true },
];

let mockRules: PermissionRule[] = [
  {
    permissionRuleId: 1,
    userId: 1,
    productId: 1,
    resourceType: 'FEATURE',
    action: 'READ',
    effect: 'ALLOW',
    scopeLevel: 'PRODUCT',
    isActive: true,
  },
];

function nextId<T>(values: T[], pick: (value: T) => number): number {
  return values.reduce((max, value) => Math.max(max, pick(value)), 0) + 1;
}

function delay(ms = 250) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function normalizeProduct(value: any): Product {
  return {
    productId: value.product_id ?? value.productId,
    productName: value.product_name ?? value.productName,
    productCode: value.product_code ?? value.productCode,
    productEnvironments: (value.environments ?? value.product_environments ?? value.productEnvironments)?.map((env: any) => ({
      environmentId: env.environment_id ?? env.environmentId,
      productId: env.product_id ?? env.productId,
      environmentCode: env.environment_code ?? env.environmentCode,
      environmentName: env.environment_name ?? env.environmentName,
      createdAt: env.created_at ?? env.createdAt,
    })) || [],
    isActive: value.is_active ?? value.isActive ?? true,
    createdAt: value.created_at ?? value.createdAt,
    updatedAt: value.updated_at ?? value.updatedAt,
  };
}

function normalizeProject(value: any): Project {
  return {
    projectId: value.project_id ?? value.projectId,
    productId: value.product_id ?? value.productId,
    projectCode: value.project_code ?? value.projectCode,
    projectName: value.project_name ?? value.projectName,
    isActive: value.is_active ?? value.isActive ?? true,
    createdAt: value.created_at ?? value.createdAt,
    updatedAt: value.updated_at ?? value.updatedAt,
  };
}

function normalizeFeature(value: any): ProjectFeature {
  return {
    categoryId: value.category_id ?? value.categoryId,
    productId: value.product_id ?? value.productId,
    projectId: value.project_id ?? value.projectId,
    parentId: value.parent_id ?? value.parentId ?? null,
    categoryType: value.category_type ?? value.categoryType ?? null,
    categoryCode: value.category_code ?? value.categoryCode,
    categoryName: value.category_name ?? value.categoryName,
    fullPath: value.full_path ?? value.fullPath ?? null,
    pathIds: value.path_ids ?? value.pathIds ?? null,
    level: value.level ?? 1,
    isActive: value.is_active ?? value.isActive ?? true,
    createdAt: value.created_at ?? value.createdAt,
    updatedAt: value.updated_at ?? value.updatedAt,
  };
}

function normalizeRole(value: any): ProductRoleDefinition {
  return {
    roleId: value.role_id ?? value.roleId,
    productId: value.product_id ?? value.productId,
    roleCode: value.role_code ?? value.roleCode,
    roleName: value.role_name ?? value.roleName,
    permissions: (value.permissions ?? []).map((permission: any) => ({
      resourceType: permission.resource_type ?? permission.resourceType,
      action: permission.action,
    })),
    isActive: value.is_active ?? value.isActive ?? true,
    createdAt: value.created_at ?? value.createdAt,
    updatedAt: value.updated_at ?? value.updatedAt,
  };
}

function normalizeMembership(value: any): ProductMembership {
  return {
    membershipId: value.membership_id ?? value.membershipId,
    userId: value.user_id ?? value.userId,
    productId: value.product_id ?? value.productId,
    roleId: value.role_id ?? value.roleId,
    expiresAt: value.expires_at ?? value.expiresAt ?? null,
    isActive: value.is_active ?? value.isActive ?? true,
    createdAt: value.created_at ?? value.createdAt,
    updatedAt: value.updated_at ?? value.updatedAt,
  };
}

function normalizeScope(value: any): MembershipScope {
  return {
    scopeId: value.scope_id ?? value.scopeId,
    membershipId: value.membership_id ?? value.membershipId,
    productId: value.product_id ?? value.productId,
    projectId: value.project_id ?? value.projectId ?? null,
    categoryId: value.category_id ?? value.categoryId ?? null,
    scopeLevel: value.scope_level ?? value.scopeLevel,
    isActive: value.is_active ?? value.isActive ?? true,
    createdAt: value.created_at ?? value.createdAt,
    updatedAt: value.updated_at ?? value.updatedAt,
  };
}

function normalizeRule(value: any): PermissionRule {
  return {
    permissionRuleId: value.permission_rule_id ?? value.permissionRuleId,
    userId: value.user_id ?? value.userId,
    productId: value.product_id ?? value.productId ?? null,
    roleId: value.role_id ?? value.roleId ?? null,
    projectId: value.project_id ?? value.projectId ?? null,
    categoryId: value.category_id ?? value.categoryId ?? null,
    resourceType: value.resource_type ?? value.resourceType,
    action: value.action,
    effect: value.effect,
    scopeLevel: value.scope_level ?? value.scopeLevel,
    grantedBy: value.granted_by ?? value.grantedBy ?? null,
    isActive: value.is_active ?? value.isActive ?? true,
    expiresAt: value.expires_at ?? value.expiresAt ?? null,
    createdAt: value.created_at ?? value.createdAt,
    updatedAt: value.updated_at ?? value.updatedAt,
  };
}

function normalizeApiKey(value: any): ProductApiKey {
  return {
    keyId: value.key_id ?? value.keyId,
    productId: value.product_id ?? value.productId,
    environmentId: value.environment_id ?? value.environmentId ?? null,
    keyName: value.key_name ?? value.keyName ?? null,
    keyPrefix: value.key_prefix ?? value.keyPrefix,
    permissions: value.permissions,
    isActive: value.is_active ?? value.isActive ?? true,
    expiresAt: value.expires_at ?? value.expiresAt ?? null,
    createdAt: value.created_at ?? value.createdAt,
    updatedAt: value.updated_at ?? value.updatedAt,
  };
}

// Helper function to normalize API response into MainLog interface
export function normalizeMainLog(value: any): MainLog {
  return {
    logId: value.log_id ?? value.logId,
    productId: value.product_id ?? value.productId,
    environmentId: value.environment_id ?? value.environmentId ?? null,
    sourceId: value.source_id ?? value.sourceId ?? null,
    timestamp: value.timestamp ?? null,
    level: value.level ?? null,
    logType: value.log_type ?? value.logType ?? null,
    message: value.message ?? null,
    requestId: value.request_id ?? value.requestId ?? null,
    traceId: value.trace_id ?? value.traceId ?? null,
    method: value.method ?? null,
    path: value.path ?? null,
    url: value.url ?? null,
    statusCode: value.status_code ?? value.statusCode ?? null,
    latencyMs: value.latency_ms ?? value.latencyMs ?? null,
    requestHeaders: value.request_headers ?? value.requestHeaders,
    responseHeaders: value.response_headers ?? value.responseHeaders,
    requestPayload: value.request_payload ?? value.requestPayload,
    responsePayload: value.response_payload ?? value.responsePayload,
    errorCode: value.error_code ?? value.errorCode ?? null,
    errorMessage: value.error_message ?? value.errorMessage ?? null,
    stackTrace: value.stack_trace ?? value.stackTrace ?? null,
    customFields: value.custom_fields ?? value.customFields,
    raw: value.raw ?? {},
  };
}

function normalizeQueueBatch(value: any): LogQueueBatch {
  return {
    batchId: value.batch_id ?? value.batchId,
    productId: value.product_id ?? value.productId ?? null,
    sourceId: value.source_id ?? value.sourceId ?? null,
    environmentId: value.environment_id ?? value.environmentId ?? null,
    queueKey: value.queue_key ?? value.queueKey,
    sourceType: value.source_type ?? value.sourceType,
    sourcePlatform: value.source_platform ?? value.sourcePlatform,
    workerId: value.worker_id ?? value.workerId ?? null,
    processingAttempts: value.processing_attempts ?? value.processingAttempts ?? 0,
    totalLogs: value.total_logs ?? value.totalLogs ?? 0,
    status: value.status,
    priority: value.priority ?? 0,
    receivedAt: value.received_at ?? value.receivedAt ?? null,
    processingStartedAt: value.processing_started_at ?? value.processingStartedAt ?? null,
    processedAt: value.processed_at ?? value.processedAt ?? null,
    errorMessage: value.error_message ?? value.errorMessage ?? null,
  };
}

function normalizeQueueItem(value: any): LogQueueItem {
  return {
    queueItemId: value.queue_item_id ?? value.queueItemId,
    batchId: value.batch_id ?? value.batchId,
    sequenceNo: value.sequence_no ?? value.sequenceNo,
    sourceType: value.source_type ?? value.sourceType,
    sourcePlatform: value.source_platform ?? value.sourcePlatform,
    status: value.status,
    retryCount: value.retry_count ?? value.retryCount ?? 0,
    maxRetryCount: value.max_retry_count ?? value.maxRetryCount ?? 0,
    nextRetryAt: value.next_retry_at ?? value.nextRetryAt ?? null,
    lastRetryAt: value.last_retry_at ?? value.lastRetryAt ?? null,
    processedAt: value.processed_at ?? value.processedAt ?? null,
    errorMessage: value.error_message ?? value.errorMessage ?? null,
  };
}

function rebuildMockFeaturePaths(projectId: number) {
  const values = mockFeatures.filter((item) => item.projectId === projectId);
  const byId = new Map(values.map((item) => [item.categoryId, item]));

  const visit = (feature: ProjectFeature): ProjectFeature => {
    if (!feature.parentId) {
      feature.level = 1;
      feature.fullPath = feature.categoryCode;
      feature.pathIds = `${feature.categoryId}`;
      return feature;
    }

    const parent = byId.get(feature.parentId);
    if (!parent) {
      feature.level = 1;
      feature.fullPath = feature.categoryCode;
      feature.pathIds = `${feature.categoryId}`;
      return feature;
    }

    const parentNode = visit(parent);
    feature.level = parentNode.level + 1;
    feature.fullPath = `${parentNode.fullPath}/${feature.categoryCode}`;
    feature.pathIds = `${parentNode.pathIds},${feature.categoryId}`;
    return feature;
  };

  values.forEach(visit);
}

export const productAdminService = {
  listProducts: async (): Promise<Product[]> => {
    if (shouldUseMock()) {
      await delay();
      return [...mockProducts];
    }
    const response = await apiClient.get<ApiResponse<any[]>>('/products');
    return response.data.data.map(normalizeProduct);
  },

  createProduct: async (payload: CreateProductPayload): Promise<Product> => {
    if (shouldUseMock()) {
      await delay();
      const productCode = (payload.productCode || payload.productName.replaceAll(' ', '_')).toUpperCase();
      const productId = nextId(mockProducts, (item) => item.productId);
      const product: Product = {
        productId,
        productName: payload.productName,
        productCode: productCode,
        productEnvironments: payload.environments?.map((env, idx) => ({
          environmentId: idx + 1,
          productId,
          environmentCode: env.environmentCode,
          environmentName: env.environmentName,
        })) || [],
        isActive: true,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };
      mockProducts.push(product);
      return product;
    }

    const response = await apiClient.post<ApiResponse<any>>('/products', {
      product_name: payload.productName,
      product_code: payload.productCode,
      environments: payload.environments?.map((env) => ({
        environment_code: env.environmentCode,
        environment_name: env.environmentName,
      })),
    });
    return normalizeProduct(response.data.data);
  },

  updateProduct: async (productId: number, payload: Partial<CreateProductPayload> & { isActive?: boolean }): Promise<Product> => {
    if (shouldUseMock()) {
      await delay();
      const current = mockProducts.find((item) => item.productId === productId);
      if (!current) {
        throw new Error('Product not found');
      }
      Object.assign(current, payload, { updatedAt: new Date().toISOString() });
      return current;
    }

    const response = await apiClient.patch<ApiResponse<any>>(`/products/${productId}`, {
      product_name: payload.productName,
      is_active: payload.isActive,
    });
    return normalizeProduct(response.data.data);
  },

  deleteProduct: async (productId: number): Promise<void> => {
    if (shouldUseMock()) {
      await delay();
      mockProducts = mockProducts.filter((item) => item.productId !== productId);
      return;
    }
    await apiClient.delete(`/products/${productId}`);
  },

  bulkDeleteProducts: async (productIds: number[]): Promise<void> => {
    if (shouldUseMock()) {
      await delay();
      mockProducts = mockProducts.filter((item) => !productIds.includes(item.productId));
      return;
    }
    await apiClient.delete('/products/product/bulk-delete', { data: { product_ids: productIds } });
  },

  listApiKeys: async (productId: number): Promise<ProductApiKey[]> => {
    if (shouldUseMock()) {
      await delay();
      return [
        {
          keyId: 1,
          productId,
          environmentId: 1,
          keyName: 'Default ingest key',
          keyPrefix: 'omni_mock',
          permissions: [{ resource: 'LOG', action: 'INGEST' }],
          isActive: true,
        },
      ];
    }
    const response = await apiClient.get<ApiResponse<any[]>>(`/products/${productId}/api-keys`);
    return response.data.data.map(normalizeApiKey);
  },

  listProjects: async (productId: number): Promise<Project[]> => {
    if (shouldUseMock()) {
      await delay();
      return mockProjects.filter((item) => item.productId === productId);
    }
    const response = await apiClient.get<ApiResponse<any[]>>(`/products/${productId}/projects`);
    return response.data.data.map(normalizeProject);
  },

  createProject: async (productId: number, payload: CreateProjectPayload): Promise<Project> => {
    if (shouldUseMock()) {
      await delay();
      const projectCode = (payload.projectCode || payload.projectName.replaceAll(' ', '_')).toUpperCase();
      const project: Project = {
        projectId: nextId(mockProjects, (item) => item.projectId),
        productId,
        projectCode: projectCode,
        projectName: payload.projectName,
        isActive: true,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };
      mockProjects.push(project);
      return project;
    }
    const response = await apiClient.post<ApiResponse<any>>(`/products/${productId}/projects`, {
      project_code: payload.projectCode,
      project_name: payload.projectName,
    });
    return normalizeProject(response.data.data);
  },

  updateProject: async (productId: number, projectId: number, payload: Partial<CreateProjectPayload> & { isActive?: boolean }): Promise<Project> => {
    if (shouldUseMock()) {
      await delay();
      const current = mockProjects.find((item) => item.projectId === projectId);
      if (!current) {
        throw new Error('Project not found');
      }
      Object.assign(current, {
        projectName: payload.projectName ?? current.projectName,
        projectCode: payload.projectCode ? payload.projectCode.toUpperCase() : current.projectCode,
        isActive: payload.isActive ?? current.isActive,
        updatedAt: new Date().toISOString(),
      });
      return current;
    }
    const response = await apiClient.patch<ApiResponse<any>>(`/products/${productId}/projects/${projectId}`, {
      project_name: payload.projectName,
      is_active: payload.isActive,
    });
    return normalizeProject(response.data.data);
  },

  listFeatures: async (productId: number, projectId: number): Promise<ProjectFeature[]> => {
    if (shouldUseMock()) {
      await delay();
      return mockFeatures
        .filter((item) => item.productId === productId && item.projectId === projectId)
        .sort((a, b) => a.level - b.level || a.categoryId - b.categoryId);
    }
    const response = await apiClient.get<ApiResponse<any[]>>(`/products/${productId}/projects/${projectId}/features`);
    return response.data.data.map(normalizeFeature);
  },

  createFeature: async (productId: number, projectId: number, payload: CreateFeaturePayload): Promise<ProjectFeature> => {
    if (shouldUseMock()) {
      await delay();
      const categoryCode = (payload.categoryCode || payload.categoryName.replaceAll(' ', '_')).toUpperCase();
      const feature: ProjectFeature = {
        categoryId: nextId(mockFeatures, (item) => item.categoryId),
        productId,
        projectId,
        parentId: payload.parentId ?? null,
        categoryType: payload.categoryType ?? null,
        categoryCode: categoryCode,
        categoryName: payload.categoryName,
        level: 1,
        isActive: true,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };
      mockFeatures.push(feature);
      rebuildMockFeaturePaths(projectId);
      return feature;
    }
    const response = await apiClient.post<ApiResponse<any>>(`/products/${productId}/projects/${projectId}/features`, {
      parent_id: payload.parentId ?? undefined,
      category_type: payload.categoryType ?? undefined,
      category_code: payload.categoryCode,
      category_name: payload.categoryName,
    });
    return normalizeFeature(response.data.data);
  },

  updateFeature: async (
    productId: number,
    projectId: number,
    featureId: number,
    payload: Partial<CreateFeaturePayload> & { isActive?: boolean }
  ): Promise<ProjectFeature> => {
    if (shouldUseMock()) {
      await delay();
      const feature = mockFeatures.find((item) => item.categoryId === featureId);
      if (!feature) {
        throw new Error('Feature not found');
      }
      Object.assign(feature, payload, { updatedAt: new Date().toISOString() });
      rebuildMockFeaturePaths(projectId);
      return feature;
    }
    const response = await apiClient.patch<ApiResponse<any>>(`/products/${productId}/projects/${projectId}/features/${featureId}`, {
      parent_id: payload.parentId,
      category_type: payload.categoryType,
      category_name: payload.categoryName,
      is_active: payload.isActive,
    });
    return normalizeFeature(response.data.data);
  },

  deleteFeature: async (productId: number, projectId: number, featureId: number): Promise<void> => {
    if (shouldUseMock()) {
      await delay();
      mockFeatures = mockFeatures.filter((item) => item.categoryId !== featureId);
      rebuildMockFeaturePaths(projectId);
      return;
    }
    await apiClient.delete(`/products/${productId}/projects/${projectId}/features/${featureId}`);
  },

  listRoles: async (productId: number): Promise<ProductRoleDefinition[]> => {
    if (shouldUseMock()) {
      await delay();
      return mockRoles.filter((item) => item.productId === productId);
    }
    const response = await apiClient.get<ApiResponse<any[]>>(`/products/${productId}/roles`);
    return response.data.data.map(normalizeRole);
  },

  createRole: async (productId: number, payload: CreateRolePayload): Promise<ProductRoleDefinition> => {
    if (shouldUseMock()) {
      await delay();
      const role: ProductRoleDefinition = {
        roleId: nextId(mockRoles, (item) => item.roleId),
        productId,
        roleCode: payload.roleCode,
        roleName: payload.roleName,
        permissions: payload.permissions,
        isActive: true,
      };
      mockRoles.push(role);
      return role;
    }
    const response = await apiClient.post<ApiResponse<any>>(`/products/${productId}/roles`, {
      role_code: payload.roleCode,
      role_name: payload.roleName,
      permissions: payload.permissions.map((permission) => ({
        resource_type: permission.resourceType,
        action: permission.action,
      })),
    });
    return normalizeRole(response.data.data);
  },

  updateRole: async (productId: number, roleId: number, payload: Partial<CreateRolePayload> & { isActive?: boolean }): Promise<ProductRoleDefinition> => {
    if (shouldUseMock()) {
      await delay();
      const role = mockRoles.find((item) => item.roleId === roleId);
      if (!role) {
        throw new Error('Role not found');
      }
      Object.assign(role, payload);
      return role;
    }
    const request: Record<string, unknown> = {};
    if (payload.roleName) request.role_name = payload.roleName;
    if (payload.permissions) {
      request.permissions = payload.permissions.map((permission) => ({
        resource_type: permission.resourceType,
        action: permission.action,
      }));
    }
    if (payload.isActive !== undefined) request.is_active = payload.isActive;
    const response = await apiClient.patch<ApiResponse<any>>(`/products/${productId}/roles/${roleId}`, request);
    return normalizeRole(response.data.data);
  },

  deleteRole: async (productId: number, roleId: number): Promise<void> => {
    if (shouldUseMock()) {
      await delay();
      mockRoles = mockRoles.filter((item) => item.roleId !== roleId);
      mockMemberships = mockMemberships.filter((item) => item.roleId !== roleId);
      return;
    }
    await apiClient.delete(`/products/${productId}/roles/${roleId}`);
  },

  listMemberships: async (productId: number): Promise<ProductMembership[]> => {
    if (shouldUseMock()) {
      await delay();
      return mockMemberships.filter((item) => item.productId === productId);
    }
    const response = await apiClient.get<ApiResponse<any[]>>(`/products/${productId}/memberships`);
    return response.data.data.map(normalizeMembership);
  },

  createMembership: async (productId: number, payload: CreateMembershipPayload): Promise<ProductMembership> => {
    if (shouldUseMock()) {
      await delay();
      const membership: ProductMembership = {
        membershipId: nextId(mockMemberships, (item) => item.membershipId),
        productId,
        userId: payload.userId,
        roleId: payload.roleId,
        expiresAt: payload.expiresAt ?? null,
        isActive: true,
      };
      mockMemberships.push(membership);
      mockScopes.push({
        scopeId: nextId(mockScopes, (item) => item.scopeId),
        membershipId: membership.membershipId,
        productId,
        scopeLevel: 'PRODUCT',
        isActive: true,
      });
      return membership;
    }
    const response = await apiClient.post<ApiResponse<any>>(`/products/${productId}/memberships`, {
      user_id: payload.userId,
      product_id: productId,
      role_id: payload.roleId,
      expires_at: payload.expiresAt ?? undefined,
    });
    return normalizeMembership(response.data.data);
  },

  updateMembership: async (productId: number, membershipId: number, payload: { roleId?: number; isActive?: boolean; expiresAt?: string | null }): Promise<ProductMembership> => {
    if (shouldUseMock()) {
      await delay();
      const membership = mockMemberships.find((item) => item.membershipId === membershipId);
      if (!membership) {
        throw new Error('Membership not found');
      }
      Object.assign(membership, payload);
      return membership;
    }
    const response = await apiClient.patch<ApiResponse<any>>(`/products/${productId}/memberships/${membershipId}`, {
      role_id: payload.roleId,
      is_active: payload.isActive,
      expires_at: payload.expiresAt ?? undefined,
    });
    return normalizeMembership(response.data.data);
  },

  listScopes: async (productId: number, membershipId: number): Promise<MembershipScope[]> => {
    if (shouldUseMock()) {
      await delay();
      return mockScopes.filter((item) => item.productId === productId && item.membershipId === membershipId);
    }
    const response = await apiClient.get<ApiResponse<any[]>>(`/products/${productId}/memberships/${membershipId}/scopes`);
    return response.data.data.map(normalizeScope);
  },

  createScope: async (productId: number, membershipId: number, payload: CreateScopePayload): Promise<MembershipScope> => {
    if (shouldUseMock()) {
      await delay();
      const scope: MembershipScope = {
        scopeId: nextId(mockScopes, (item) => item.scopeId),
        membershipId,
        productId,
        projectId: payload.projectId ?? null,
        categoryId: payload.categoryId ?? null,
        scopeLevel: payload.scopeLevel,
        isActive: true,
      };
      mockScopes.push(scope);
      return scope;
    }
    const response = await apiClient.post<ApiResponse<any>>(`/products/${productId}/memberships/${membershipId}/scopes`, {
      project_id: payload.projectId ?? undefined,
      category_id: payload.categoryId ?? undefined,
      scope_level: payload.scopeLevel,
    });
    return normalizeScope(response.data.data);
  },

  deleteScope: async (productId: number, membershipId: number, scopeId: number): Promise<void> => {
    if (shouldUseMock()) {
      await delay();
      mockScopes = mockScopes.filter((item) => item.scopeId !== scopeId);
      return;
    }
    await apiClient.delete(`/products/${productId}/memberships/${membershipId}/scopes/${scopeId}`);
  },

  listPermissionRules: async (productId: number): Promise<PermissionRule[]> => {
    if (shouldUseMock()) {
      await delay();
      return mockRules.filter((item) => item.productId === productId);
    }
    const response = await apiClient.get<ApiResponse<any[]>>(`/products/${productId}/permission-rules`);
    return response.data.data.map(normalizeRule);
  },

  createPermissionRule: async (payload: CreatePermissionRulePayload): Promise<PermissionRule> => {
    if (shouldUseMock()) {
      await delay();
      const rule: PermissionRule = {
        permissionRuleId: nextId(mockRules, (item) => item.permissionRuleId),
        userId: payload.userId,
        productId: payload.productId,
        roleId: payload.roleId ?? null,
        projectId: payload.projectId ?? null,
        categoryId: payload.categoryId ?? null,
        resourceType: payload.resourceType,
        action: payload.action,
        effect: payload.effect,
        scopeLevel: payload.scopeLevel,
        expiresAt: payload.expiresAt ?? null,
        isActive: true,
      };
      mockRules.push(rule);
      return rule;
    }
    const response = await apiClient.post<ApiResponse<any>>(`/products/${payload.productId}/permission-rules`, {
      user_id: payload.userId,
      product_id: payload.productId,
      role_id: payload.roleId ?? undefined,
      project_id: payload.projectId ?? undefined,
      category_id: payload.categoryId ?? undefined,
      resource_type: payload.resourceType,
      action: payload.action,
      effect: payload.effect,
      scope_level: payload.scopeLevel,
      expires_at: payload.expiresAt ?? undefined,
    });
    return normalizeRule(response.data.data);
  },

  updatePermissionRule: async (productId: number, ruleId: number, payload: { effect?: 'ALLOW' | 'DENY'; isActive?: boolean; expiresAt?: string | null }): Promise<PermissionRule> => {
    if (shouldUseMock()) {
      await delay();
      const rule = mockRules.find((item) => item.permissionRuleId === ruleId);
      if (!rule) {
        throw new Error('Permission rule not found');
      }
      Object.assign(rule, payload);
      return rule;
    }
    const response = await apiClient.patch<ApiResponse<any>>(`/products/${productId}/permission-rules/${ruleId}`, {
      effect: payload.effect,
      is_active: payload.isActive,
      expires_at: payload.expiresAt ?? undefined,
    });
    return normalizeRule(response.data.data);
  },

  checkPermission: async (productId: number, resourceType: string, action: string, projectId?: number, categoryId?: number): Promise<PermissionCheckResult> => {
    if (shouldUseMock()) {
      await delay(150);
      return { allowed: true, reason: 'mock permission evaluation' };
    }
    const response = await apiClient.post<ApiResponse<any>>('/authorization/check', {
      product_id: productId,
      project_id: projectId,
      category_id: categoryId,
      resource_type: resourceType,
      action: action,
    });
    return {
      allowed: response.data.data.allowed,
      matchedRuleId: response.data.data.matched_rule_id,
      reason: response.data.data.reason,
    };
  },

  searchLogs: async (params: SearchLogsParams): Promise<{ data: MainLog[]; total: number; page: number; perPage: number }> => {
    if (shouldUseMock()) {
      await delay();
      const now = new Date().toISOString();
      return {
        data: [
          {
            logId: 'mock-log-1',
            productId: params.productId,
            environmentId: params.environmentId ?? 1,
            timestamp: now,
            level: params.level || 'INFO',
            logType: params.logType || 'MANUAL_IMPORT',
            message: `Mock log for product ${params.productId}`,
            path: '/api/mock',
            raw: {
              payload: {
                project_id: params.projectId,
                category_id: params.categoryId,
                feature_path_ids: params.categoryId ? `${params.categoryId}` : undefined,
              },
            },
          },
        ],
        total: 1,
        page: params.page ?? 1,
        perPage: params.perPage ?? 10,
      };
    }

    const response = await apiClient.get<ApiResponse<any[]>>('/logs', {
      params: {
        product_id: params.productId,
        environment_id: params.environmentId,
        project_id: params.projectId,
        category_id: params.categoryId,
        level: params.level,
        log_type: params.logType,
        keyword: params.keyword,
        page: params.page ?? 1,
        per_page: params.perPage ?? 10,
      },
    });

    return {
      data: response.data.data.map(normalizeMainLog),
      total: Number((response.data as any).meta?.total ?? response.data.data.length),
      page: Number((response.data as any).meta?.page ?? params.page ?? 1),
      perPage: Number((response.data as any).meta?.per_page ?? params.perPage ?? 10),
    };
  },

  searchLogsMulti: async (params: SearchLogsMultiParams): Promise<{ data: MainLog[]; total: number; page: number; perPage: number }> => {
    if (shouldUseMock()) {
      await delay();
      const now = new Date().toISOString();
      const levels = params.levels?.length ? params.levels : ['INFO', 'WARN', 'ERROR', 'DEBUG'];
      return {
        data: levels.map((lvl, i) => ({
          logId: `mock-multi-${i}`,
          productId: params.productId,
          environmentId: params.environmentId ?? 1,
          timestamp: new Date(Date.now() - i * 60000).toISOString(),
          level: lvl,
          logType: 'MANUAL_IMPORT',
          message: `Mock multi-log ${lvl} (${now})`,
          raw: {},
        })),
        total: levels.length,
        page: params.page ?? 1,
        perPage: params.perPage ?? 20,
      };
    }

    const response = await apiClient.get<ApiResponse<any[]>>('/logs', {
      params: {
        product_id: params.productId,
        environment_id: params.environmentId,
        project_ids: params.projectIds?.join(',') || undefined,
        category_ids: params.categoryIds?.join(',') || undefined,
        level: params.levels?.join(',') || undefined,
        keyword: params.keyword,
        custom_field_path: params.customFieldPath,
        custom_field_value: params.customFieldValue,
        page: params.page ?? 1,
        per_page: params.perPage ?? 20,
      },
    });

    return {
      data: response.data.data.map(normalizeMainLog),
      total: Number((response.data as any).meta?.total ?? response.data.data.length),
      page: Number((response.data as any).meta?.page ?? params.page ?? 1),
      perPage: Number((response.data as any).meta?.per_page ?? params.perPage ?? 20),
    };
  },

  getDashboardLogStats: async (params: DashboardLogStatsParams): Promise<any> => {
    if (shouldUseMock()) {
      await delay();
      // Generate mock stats data
      const now = Date.now();
      const buckets = Array.from({ length: 24 }, (_, i) => {
        const time = new Date(now - (23 - i) * 3600000).toISOString();
        return {
          key_as_string: time,
          doc_count: Math.floor(Math.random() * 50),
          by_level: {
            buckets: [
              { key: 'INFO', doc_count: Math.floor(Math.random() * 30) },
              { key: 'WARN', doc_count: Math.floor(Math.random() * 10) },
              { key: 'ERROR', doc_count: Math.floor(Math.random() * 5) },
              { key: 'DEBUG', doc_count: Math.floor(Math.random() * 15) },
            ],
          },
        };
      });
      return {
        log_levels: {
          buckets: [
            { key: 'INFO', doc_count: 120 },
            { key: 'WARN', doc_count: 35 },
            { key: 'ERROR', doc_count: 12 },
            { key: 'DEBUG', doc_count: 80 },
          ],
        },
        logs_over_time: { buckets },
      };
    }

    const response = await apiClient.get<any>('/dashboard/stats', {
      params: {
        product_id: params.productId,
        environment_id: params.environmentId,
        project_ids: params.projectIds?.join(',') || undefined,
        category_ids: params.categoryIds?.join(',') || undefined,
      },
    });

    return response.data;
  },

  getBatchItems: async (batchId: string): Promise<LogQueueItem[]> => {
    if (shouldUseMock()) {
      await delay();
      return [
        {
          queueItemId: 1,
          batchId,
          sequenceNo: 1,
          sourceType: 'application',
          sourcePlatform: 'api-key-import',
          status: 'PROCESSED',
          retryCount: 0,
          maxRetryCount: 3,
          processedAt: new Date().toISOString(),
        },
      ];
    }

    const response = await apiClient.get<ApiResponse<any[]>>(`/queues/batches/${batchId}/items`);
    return response.data.data.map(normalizeQueueItem);
  },

  importLogs: async (payload: ImportLogsPayload): Promise<LogQueueBatch> => {
    const body = {
      product_id: payload.productId,
      environment_id: payload.environmentId,
      queue_key: 'default',
      source_type: 'application',
      source_platform: 'api-key-import',
      priority: 1,
      logs: [
        {
          sequence_no: 1,
          source_type: 'application',
          source_platform: 'api-key-import',
          input_payload: {
            log_level: payload.logLevel,
            event_type: payload.eventType || 'MANUAL_IMPORT',
            message: payload.message,
            timestamp: new Date().toISOString(),
            ...(payload.projectId ? { project_id: payload.projectId } : {}),
            ...(payload.categoryId ? { category_id: payload.categoryId } : {}),
            ...(payload.featureFullPath ? { feature_full_path: payload.featureFullPath } : {}),
            ...(payload.featurePathIds ? { feature_path_ids: payload.featurePathIds } : {}),
            ...(payload.customFields && Object.keys(payload.customFields).length > 0 ? { custom_fields: payload.customFields } : {}),
          },
        },
      ],
    };

    if (shouldUseMock()) {
      await delay();
      return normalizeQueueBatch({
        batch_id: `mock-batch-${Date.now()}`,
        product_id: payload.productId,
        environment_id: payload.environmentId,
        queue_key: 'default',
        source_type: 'application',
        source_platform: 'api-key-import',
        processing_attempts: 1,
        total_logs: 1,
        status: 'COMPLETED',
        priority: 1,
        received_at: new Date().toISOString(),
      });
    }

    const enqueueResponse = await apiClient.post<ApiResponse<any>>('/queues', body);
    try {
      await apiClient.post<ApiResponse<any>>('/queues/consume', {});
    } catch (e) {
      console.warn('Manual consume trigger ignored:', e);
    }
    return normalizeQueueBatch(enqueueResponse.data.data);
  },

  checkElasticOnline: async (): Promise<{ success: boolean; data: { status: string } }> => {
    if (shouldUseMock()) {
      await delay(100);
      return { success: true, data: { status: 'online' } };
    }
    try {
      const baseURLObj = new URL(apiClient.defaults.baseURL || window.location.origin, window.location.origin);
      const healthURL = `${baseURLObj.origin}/health/ready`;
      const response = await axios.get(healthURL);
      if (response.data?.success && response.data?.data?.status === 'ready') {
        return { success: true, data: { status: 'online' } };
      }
      return { success: false, data: { status: 'offline' } };
    } catch (error) {
      return { success: false, data: { status: 'offline' } };
    }
  },

  getFailedBatches: async (): Promise<ApiResponse<any[]>> => {
    if (shouldUseMock()) {
      return { success: true, data: [], message: 'Mock failed batches retrieved' };
    }
    const response = await apiClient.get<ApiResponse<any[]>>('/queues/failed');
    return response.data;
  },

  retryFailedBatches: async (): Promise<ApiResponse<any>> => {
    return {
      success: false,
      data: { status: 'retry_disabled' },
      message: 'Retry has been removed from the system',
    };
  },
};
