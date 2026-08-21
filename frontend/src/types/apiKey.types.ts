export type ApiKeyStatus = "active" | "disabled" | "expiring" | "expired" | "revoked" | "never-used";
export interface ProductEnvironment { environmentId: number; productId: number; code: string; name: string; }
export interface ApiKeyRecord { keyId: number; productId: number; environmentId?: number; name: string; prefix: string; permissions: string[]; active: boolean; createdAt?: string; updatedAt?: string; expiresAt?: string; lastUsedAt?: string; revokedAt?: string; }
export interface GeneratedApiKey extends ApiKeyRecord { secret: string; }
export interface GenerateApiKeyInput { name: string; environmentId?: number; defaultProjectId?: number; defaultCategoryId?: number; description: string; expiration: "30d" | "90d" | "180d" | "1y" | "custom" | "never"; customDate?: string; permissions: string[]; allowedSources: string[]; ipAllowlist: string[]; rateLimit?: number; active: boolean; }
