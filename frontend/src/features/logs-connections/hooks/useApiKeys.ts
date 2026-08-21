import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { message } from "antd";
import { apiKeyService } from "@/services/apiKey.service";
import type { GenerateApiKeyInput } from "@/types/apiKey.types";
export function useApiKeys(productId?: number) {
  const client = useQueryClient(); const queryKey = ["products", productId, "api-keys"];
  const query = useQuery({ queryKey, queryFn: () => apiKeyService.list(productId!), enabled: Boolean(productId), staleTime: 15_000 });
  const invalidate = () => void client.invalidateQueries({ queryKey });
  const create = useMutation({ mutationFn: (input: GenerateApiKeyInput) => apiKeyService.create(productId!, input), onSuccess: invalidate, onError: () => message.error("Could not generate API key") });
  const update = useMutation({ mutationFn: ({ keyId, input }: { keyId: number; input: { keyName?: string; permissions?: string[]; active?: boolean; expiresAt?: string } }) => apiKeyService.update(productId!, keyId, input), onSuccess: () => { message.success("API key updated"); invalidate(); }, onError: () => message.error("Could not update API key") });
  const setActive = useMutation({ mutationFn: ({ keyId, active }: { keyId: number; active: boolean }) => apiKeyService.setActive(productId!, keyId, active), onSuccess: (_, value) => { message.success(value.active ? "API key enabled" : "API key disabled"); invalidate(); }, onError: () => message.error("Could not update API key") });
  const revoke = useMutation({ mutationFn: (keyId: number) => apiKeyService.revoke(productId!, keyId), onSuccess: () => { message.success("API key revoked"); invalidate(); }, onError: () => message.error("Could not revoke API key") });
  return { query, create, update, setActive, revoke };
}
