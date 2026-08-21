import dayjs from "dayjs";
import { apiClient } from "@/api";
import { apiKeyListSchema, generatedApiKeySchema, type RawApiKey } from "../schemas/apiKey.schema";
import type { ApiKeyRecord, GenerateApiKeyInput, GeneratedApiKey, ProductEnvironment } from "../types/apiKey.types";
import type { ApiKeyAction } from "@/utils/apiKeyPermission";

interface ApiEnvelope<T> { data: T }
interface EnvironmentPayload {
  environment_id?: number;
  environmentId?: number;
  id?: number;
  product_id?: number;
  productId?: number;
  environment_code?: string;
  environmentCode?: string;
  code?: string;
  environment_name?: string;
  environmentName?: string;
  name?: string;
}
const permissions = (value: RawApiKey["permissions"]): string[] => { if (Array.isArray(value)) return value; if (typeof value === "string") { try { return permissions(JSON.parse(value)); } catch { return [value]; } } return Object.entries(value).filter(([, allowed]) => Boolean(allowed)).map(([key]) => key); };
const mapKey = (value: RawApiKey): ApiKeyRecord => ({ keyId: value.key_id, productId: value.product_id, environmentId: value.environment_id, name: value.key_name ?? "Unnamed key", prefix: value.key_prefix, permissions: permissions(value.permissions), active: value.is_active, createdAt: value.created_at ?? undefined, updatedAt: value.updated_at ?? undefined, expiresAt: value.expires_at ?? undefined, lastUsedAt: value.last_used_at ?? undefined, revokedAt: value.revoked_at ?? undefined });
const expiresAt = (input: GenerateApiKeyInput) => {
  if (input.expiration === "never") return undefined;
  if (input.expiration === "custom") return dayjs(input.customDate).endOf("day").toISOString();
  if (input.expiration === "1y") return dayjs().add(1, "year").toISOString();
  return dayjs().add({ "30d": 30, "90d": 90, "180d": 180 }[input.expiration], "day").toISOString();
};
export const apiKeyService = {
  can: async (productId: number, action: ApiKeyAction) => {
    const response = await apiClient.post<ApiEnvelope<{ allowed: boolean }>>("/v1/authorization/check", { resource_type: "API_KEY", action, product_id: productId });
    return response.data.data.allowed;
  },
  list: async (productId: number) => { const response = await apiClient.get<ApiEnvelope<unknown>>(`/v1/products/${productId}/api-keys`); return apiKeyListSchema.parse(response.data.data).map(mapKey); },
  environments: async (productId: number): Promise<ProductEnvironment[]> => {
    const response = await apiClient.get<ApiEnvelope<EnvironmentPayload[]>>(`/v1/products/${productId}/environments`);
    const list = response.data.data ?? [];
    return list
      .map((item) => ({
        environmentId: item.environment_id ?? item.environmentId ?? item.id ?? 0,
        productId: item.product_id ?? item.productId ?? productId,
        code: item.environment_code ?? item.environmentCode ?? item.code ?? "",
        name: item.environment_name ?? item.environmentName ?? item.name ?? "",
      }))
      .filter((item) => item.environmentId > 0);
  },
  create: async (productId: number, input: GenerateApiKeyInput): Promise<GeneratedApiKey> => { const response = await apiClient.post<ApiEnvelope<unknown>>(`/v1/products/${productId}/api-keys`, { product_id: productId, environment_id: input.environmentId, default_project_id: input.defaultProjectId, default_category_id: input.defaultCategoryId, key_name: input.name, permissions: input.permissions, expires_at: expiresAt(input) }); const parsed = generatedApiKeySchema.parse(response.data.data); return { ...mapKey(parsed), secret: parsed.api_key }; },
  update: async (productId: number, keyId: number, input: { keyName?: string; permissions?: string[]; active?: boolean; expiresAt?: string }) => {
    const response = await apiClient.patch<ApiEnvelope<unknown>>(`/v1/products/${productId}/api-keys/${keyId}`, { key_name: input.keyName, permissions: input.permissions, is_active: input.active, expires_at: input.expiresAt });
    return mapKey(generatedApiKeySchema.omit({ api_key: true }).parse(response.data.data));
  },  setActive: async (productId: number, keyId: number, active: boolean) => { const response = await apiClient.patch<ApiEnvelope<unknown>>(`/v1/products/${productId}/api-keys/${keyId}`, { is_active: active }); return mapKey(generatedApiKeySchema.omit({ api_key: true }).parse(response.data.data)); },
  revoke: async (productId: number, keyId: number) => { await apiClient.delete(`/v1/products/${productId}/api-keys/${keyId}`); },
};

