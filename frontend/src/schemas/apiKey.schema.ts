import { z } from "zod/v4";

const rawKeySchema = z.object({
  key_id: z.number(), product_id: z.number(), environment_id: z.number().optional(), key_name: z.string().nullish(), key_prefix: z.string(),
  permissions: z.union([z.array(z.string()), z.record(z.string(), z.unknown()), z.string()]), is_active: z.boolean(),
  created_at: z.string().nullish(), updated_at: z.string().nullish(), expires_at: z.string().nullish(), last_used_at: z.string().nullish(), revoked_at: z.string().nullish(), api_key: z.string().optional(),
});
export const apiKeyListSchema = z.array(rawKeySchema);
export const generatedApiKeySchema = rawKeySchema.extend({ api_key: z.string().min(20) });
export const generateApiKeySchema = z.object({
  name: z.string().trim().min(2, "Key name is required").max(100), environmentId: z.number().positive("Environment is required"), defaultProjectId: z.number().optional(), defaultCategoryId: z.number().optional(), description: z.string().max(500),
  expiration: z.enum(["30d", "90d", "180d", "1y", "custom", "never"]), customDate: z.string().optional(),
  permissions: z.array(z.literal("LOG_INGEST_CREATE")).length(1),
  allowedSources: z.array(z.string()), ipAllowlist: z.array(z.string()).default([]), rateLimit: z.number().int().min(1).max(100000).optional(), active: z.boolean(),
}).refine((value) => value.expiration !== "custom" || Boolean(value.customDate), { path: ["customDate"], message: "Choose an expiration date" }).refine((value) => !value.defaultCategoryId || Boolean(value.defaultProjectId), { path: ["defaultProjectId"], message: "Choose a project before choosing a feature" });

export type RawApiKey = z.infer<typeof rawKeySchema>;
