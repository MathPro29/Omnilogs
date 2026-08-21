import { useQueries } from "@tanstack/react-query";
import { useAuthStore } from "@/store";
import { apiKeyService } from "@/services/apiKey.service";
import { hasApiKeyPermission, type ApiKeyAction } from "@/utils/apiKeyPermission";

const actions: ApiKeyAction[] = ["READ", "CREATE", "UPDATE", "DELETE"];

export function useApiKeyPermissions(productId?: number) {
  const roles = useAuthStore((state) => state.roles) ?? [];
  const permissions = useAuthStore((state) => state.permissions) ?? [];
  const safeRoles = roles.filter((role): role is string => typeof role === "string");
  const safePermissions = permissions.filter((permission): permission is string => typeof permission === "string");
  const isPlatformAdmin = safeRoles.some((role) => ["god", "owner", "superadmin", "super_admin", "admin"].includes(role.toLowerCase()));
  const checks = useQueries({
    queries: actions.map((action) => ({
      queryKey: ["authorization", "api-key", productId, action],
      queryFn: () => apiKeyService.can(productId!, action),
      enabled: Boolean(productId) && !isPlatformAdmin,
      staleTime: 30_000,
      retry: false,
    })),
  });
  const allowed = (action: ApiKeyAction) => {
    if (isPlatformAdmin || hasApiKeyPermission(safeRoles, safePermissions, action)) return true;
    return checks[actions.indexOf(action)]?.data === true;
  };
  return {
    canView: allowed("READ"),
    canCreate: allowed("CREATE"),
    canUpdate: allowed("UPDATE"),
    canDelete: allowed("DELETE"),
    isPlatformAdmin,
    isChecking: checks.some((query) => query.isPending),
  };
}
