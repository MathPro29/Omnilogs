import { apiClient } from "@/api";

export interface ProductEnvironmentOption {
  environment_id: number;
  environment_code: string;
  environment_name: string;
  is_active?: boolean;
}
export interface ProductProjectOption {
  product_id?: number;
  project_id: number;
  project_code: string;
  project_name: string;
  is_active: boolean;
}
export interface ProductOption {
  id: number;
  name: string;
  product_code?: string;
  description?: string;
  setup_status?: string;
  is_active?: boolean;
  product_environments?: ProductEnvironmentOption[];
  membership_id?: number;
  membership_role?: string;
  is_owner?: boolean;
}
interface ProductResponse {
  product_id?: number;
  productId?: number;
  id?: number;
  product_name?: string;
  productName?: string;
  name?: string;
  product_code?: string;
  productCode?: string;
  description?: string;
  setup_status?: string;
  setupStatus?: string;
  is_active?: boolean;
  isActive?: boolean;
  product_environments?: ProductEnvironmentOption[];
  productEnvironments?: ProductEnvironmentOption[];
  membership_id?: number;
  membership_role?: string;
  is_owner?: boolean;
}
export interface ProductLogRoute {
  route_id: number;
  product_id: number;
  environment_id: number;
  route_key: string;
  project_id: number;
  category_id?: number | null;
  priority: number;
  is_active: boolean;
}
export interface CreateProductLogRouteInput {
  environment_id: number;
  route_key: string;
  project_id: number;
  category_id?: number | null;
  priority?: number;
}
export type LogRoutingOperator =
  | "equals"
  | "not_equals"
  | "starts_with"
  | "ends_with"
  | "contains"
  | "in"
  | "not_in"
  | "regex"
  | "range"
  | "greater_than"
  | "less_than"
  | "greater_than_or_equal"
  | "less_than_or_equal";

export interface LogRoutingCondition {
  field: string;
  operator: LogRoutingOperator;
  value?: string;
  min?: unknown;
  max?: unknown;
}
export interface ProductLogRoutingRule {
  rule_id: number;
  product_id: number;
  environment_id?: number | null;
  rule_name: string;
  priority: number;
  conditions: LogRoutingCondition[];
  target_project_id: number;
  target_category_id?: number | null;
  is_active: boolean;
}
export interface CreateProductLogRoutingRuleInput {
  environment_id?: number | null;
  rule_name: string;
  priority?: number;
  conditions: LogRoutingCondition[];
  target_project_id: number;
  target_category_id?: number | null;
}
export interface RoutingDiscoveryItem {
  environment_id: number;
  service_name: string;
  request_method: string;
  request_path: string;
  route_pattern?: string;
  routing_status: string;
  routing_method?: string;
  total_logs: number;
  last_seen_at: string;
}

export interface ProductFeatureOption {
  category_id: number;
  product_id: number;
  project_id: number;
  parent_id?: number | null;
  category_type?: string;
  category_code: string;
  category_name: string;
  level: number;
  full_path?: string;
  path_ids?: string;
  is_active: boolean;
}
export interface ProductRolePermission {
  resource_type: string;
  action: string;
}
export interface ProductAccessRole {
  role_id: number;
  role_code: string;
  role_name: string;
  permissions?: ProductRolePermission[];
  access_level: string;
  member_count: number;
  is_active: boolean;
}
export interface CreateProductRoleInput {
  role_name: string;
}
export interface ProductAccessMember {
  membership_id: number;
  user_id: number;
  username?: string;
  full_name: string;
  email: string;
  role_id: number;
  role_code: string;
  role_name: string;
  access_level: string;
  expires_at?: string;
  is_active: boolean;
}
export interface ProductAccessOverview {
  product_id: number;
  roles: ProductAccessRole[];
  members: ProductAccessMember[];
  updated_at: string;
}
export interface ProductMembershipRecord {
  membership_id: number;
  user_id: number;
  product_id: number;
  role_id: number;
  is_active: boolean;
}
interface Envelope<T> {
  data: T;
}
const normalizeProduct = (product: ProductResponse): ProductOption => ({
  id: product.product_id ?? product.productId ?? product.id ?? 0,
  name:
    product.product_name ??
    product.productName ??
    product.name ??
    "Unknown product",
  product_code: product.product_code ?? product.productCode ?? "",
  description: product.description,
  setup_status: product.setup_status ?? product.setupStatus,
  is_active: product.is_active ?? product.isActive,
  product_environments:
    product.product_environments ?? product.productEnvironments ?? [],
  membership_id: product.membership_id,
  membership_role: product.membership_role,
  is_owner: product.is_owner,
});
export const productService = {
  listOptions: async (): Promise<ProductOption[]> => {
    const response =
      await apiClient.get<Envelope<ProductResponse[]>>("/v1/products");
    return (response.data.data ?? [])
      .map(normalizeProduct)
      .filter((product) => product.id > 0);
  },
  getAccessOverview: async (
    productId: number,
  ): Promise<ProductAccessOverview> => {
    const response = await apiClient.get<Envelope<ProductAccessOverview>>(
      `/v1/products/${productId}/access-overview`,
    );
    return response.data.data;
  },
  createRole: async (
    productId: number,
    input: CreateProductRoleInput,
  ): Promise<ProductAccessRole> => {
    const response = await apiClient.post<Envelope<ProductAccessRole>>(
      `/v1/products/${productId}/roles`,
      input,
    );
    return response.data.data;
  },
  createMembership: async (
    productId: number,
    input: { user_id: number; role_id: number },
  ): Promise<ProductMembershipRecord> => {
    const response = await apiClient.post<Envelope<ProductMembershipRecord>>(
      `/v1/products/${productId}/memberships`,
      { ...input, product_id: productId },
    );
    return response.data.data;
  },
  createMemberships: async (
    productId: number,
    userIds: number[],
    roleId: number,
  ): Promise<ProductMembershipRecord[]> => {
    const response = await apiClient.post<Envelope<ProductMembershipRecord[]>>(
      `/v1/products/${productId}/memberships/bulk`,
      {
        memberships: userIds.map((userId) => ({
          user_id: userId,
          product_id: productId,
          role_id: roleId,
        })),
      },
    );
    return response.data.data ?? [];
  },
  updateMembership: async (
    productId: number,
    membershipId: number,
    input: { role_id?: number; is_active?: boolean },
  ) => {
    const response = await apiClient.patch<Envelope<unknown>>(
      `/v1/products/${productId}/memberships/${membershipId}`,
      input,
    );
    return response.data.data;
  },
  updateProduct: async (
    productId: number,
    input: {
      product_name?: string;
      product_code?: string;
      description?: string;
      is_active?: boolean;
    },
  ) => {
    const response = await apiClient.patch<Envelope<ProductResponse>>(
      `/v1/products/${productId}`,
      input,
    );
    return response.data.data;
  },
  deleteMembership: async (productId: number, membershipId: number) => {
    await apiClient.delete(
      `/v1/products/${productId}/memberships/${membershipId}`,
    );
  },
  deleteRole: async (productId: number, roleId: number): Promise<void> => {
    await apiClient.delete(`/v1/products/${productId}/roles/${roleId}`);
  },
  updateRole: async (
    productId: number,
    roleId: number,
    input: { role_name?: string; description?: string; is_active?: boolean },
  ): Promise<ProductAccessRole> => {
    const response = await apiClient.patch<Envelope<ProductAccessRole>>(
      `/v1/products/${productId}/roles/${roleId}`,
      input,
    );
    return response.data.data;
  },
  deleteProduct: async (productId: number): Promise<void> => {
    await apiClient.delete(`/v1/products/${productId}`);
  },
  bulkDeleteProducts: async (productIds: number[]): Promise<void> => {
    await apiClient.delete("/v1/products/product/bulk-delete", {
      data: { product_ids: productIds },
    });
  },
  listProjects: async (productId: number): Promise<ProductProjectOption[]> => {
    const response = await apiClient.get<Envelope<ProductProjectOption[]>>(
      `/v1/products/${productId}/projects`,
    );
    return (response.data.data ?? []).filter(
      (project) => project.project_id > 0 && project.is_active !== false,
    );
  },
  listEnvironments: async (
    productId: number,
  ): Promise<ProductEnvironmentOption[]> => {
    const response = await apiClient.get<Envelope<ProductEnvironmentOption[]>>(
      `/v1/products/${productId}/environments`,
    );
    return (response.data.data ?? []).filter(
      (environment) =>
        environment.environment_id > 0 && environment.is_active !== false,
    );
  },
  listFeatures: async (
    productId: number,
    projectId?: number,
  ): Promise<ProductFeatureOption[]> => {
    if (projectId) {
      const response = await apiClient.get<Envelope<ProductFeatureOption[]>>(
        `/v1/products/${productId}/projects/${projectId}/features`,
      );
      return (response.data.data ?? []).filter(
        (feature) => feature.category_id > 0 && feature.is_active !== false,
      );
    }
    const projects = await productService.listProjects(productId);
    const featureResults = await Promise.all(
      projects.map((p) =>
        apiClient
          .get<Envelope<ProductFeatureOption[]>>(
            `/v1/products/${productId}/projects/${p.project_id}/features`,
          )
          .then((res) => res.data.data ?? [])
          .catch(() => []),
      ),
    );
    return featureResults
      .flat()
      .filter(
        (feature) => feature.category_id > 0 && feature.is_active !== false,
      );
  },
  listLogRoutes: async (productId: number): Promise<ProductLogRoute[]> => {
    const response = await apiClient.get<Envelope<ProductLogRoute[]>>(
      `/v1/products/${productId}/log-routes`,
    );
    return response.data.data ?? [];
  },
  createLogRoute: async (
    productId: number,
    input: CreateProductLogRouteInput,
  ): Promise<ProductLogRoute> => {
    const response = await apiClient.post<Envelope<ProductLogRoute>>(
      `/v1/products/${productId}/log-routes`,
      input,
    );
    return response.data.data;
  },
  deleteLogRoute: async (productId: number, routeId: number): Promise<void> => {
    await apiClient.delete(`/v1/products/${productId}/log-routes/${routeId}`);
  },
  listLogRoutingRules: async (
    productId: number,
  ): Promise<ProductLogRoutingRule[]> => {
    const response = await apiClient.get<Envelope<ProductLogRoutingRule[]>>(
      `/v1/products/${productId}/log-routing-rules`,
    );
    return response.data.data ?? [];
  },
  listRoutingDiscovery: async (
    productId: number,
  ): Promise<RoutingDiscoveryItem[]> => {
    const response = await apiClient.get<Envelope<RoutingDiscoveryItem[]>>(
      `/v1/products/${productId}/routing-discovery`,
    );
    return response.data.data ?? [];
  },
  createLogRoutingRule: async (
    productId: number,
    input: CreateProductLogRoutingRuleInput,
  ): Promise<ProductLogRoutingRule> => {
    const response = await apiClient.post<Envelope<ProductLogRoutingRule>>(
      `/v1/products/${productId}/log-routing-rules`,
      input,
    );
    return response.data.data;
  },
  updateLogRoutingRule: async (
    productId: number,
    ruleId: number,
    input: CreateProductLogRoutingRuleInput,
  ): Promise<ProductLogRoutingRule> => {
    const response = await apiClient.put<Envelope<ProductLogRoutingRule>>(
      `/v1/products/${productId}/log-routing-rules/${ruleId}`,
      input,
    );
    return response.data.data;
  },  deleteLogRoutingRule: async (
    productId: number,
    ruleId: number,
  ): Promise<void> => {
    await apiClient.delete(
      `/v1/products/${productId}/log-routing-rules/${ruleId}`,
    );
  },
};
