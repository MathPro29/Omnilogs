import { apiClient } from "@/api";

interface Envelope<T> {
  data: T;
}

export type AccessScopeLevel = "PRODUCT" | "PROJECT" | "CATEGORY";

export interface MembershipScope {
  scope_id: number;
  membership_id: number;
  product_id: number;
  project_id?: number;
  category_id?: number;
  scope_level: AccessScopeLevel;
  is_active: boolean;
}

export interface CreateMembershipScopeInput {
  scope_level: AccessScopeLevel;
  project_id?: number;
  category_id?: number;
}

function scopeKey(scope: Pick<MembershipScope, "scope_level" | "project_id" | "category_id">): string {
  return `${scope.scope_level}:${scope.project_id ?? ""}:${scope.category_id ?? ""}`;
}

export const userAccessService = {
  listScopes: async (productId: number, membershipId: number): Promise<MembershipScope[]> => {
    const response = await apiClient.get<Envelope<MembershipScope[]>>(
      `/v1/products/${productId}/memberships/${membershipId}/scopes`,
    );
    return response.data.data ?? [];
  },

  createScope: async (
    productId: number,
    membershipId: number,
    input: CreateMembershipScopeInput,
  ): Promise<MembershipScope> => {
    const response = await apiClient.post<Envelope<MembershipScope>>(
      `/v1/products/${productId}/memberships/${membershipId}/scopes`,
      input,
    );
    return response.data.data;
  },

  deleteScope: async (productId: number, membershipId: number, scopeId: number): Promise<void> => {
    await apiClient.delete(
      `/v1/products/${productId}/memberships/${membershipId}/scopes/${scopeId}`,
    );
  },

  replaceScopes: async (
    productId: number,
    membershipId: number,
    desired: CreateMembershipScopeInput[],
  ): Promise<void> => {
    const current = await userAccessService.listScopes(productId, membershipId);
    const desiredKeys = new Set(desired.map(scopeKey));
    const currentKeys = new Set(current.map(scopeKey));
    await Promise.all([
      ...current
        .filter((scope) => !desiredKeys.has(scopeKey(scope)))
        .map((scope) => userAccessService.deleteScope(productId, membershipId, scope.scope_id)),
      ...desired
        .filter((scope) => !currentKeys.has(scopeKey(scope)))
        .map((scope) => userAccessService.createScope(productId, membershipId, scope)),
    ]);
  },
};
