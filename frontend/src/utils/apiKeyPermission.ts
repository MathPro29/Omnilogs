export type ApiKeyAction = "READ" | "CREATE" | "UPDATE" | "DELETE";

export const hasApiKeyPermission = (roles: string[], permissions: string[], action: ApiKeyAction) => {
  if (roles.some((role) => ["god", "owner", "superadmin", "super_admin", "admin"].includes(role.toLowerCase()))) return true;
  const aliases = action === "READ" ? ["READ", "VIEW"] : [action];
  return aliases.some((candidate) => [
    `API_KEY_${candidate}`,
    `API_KEY:${candidate}`,
    `api_key:${candidate.toLowerCase()}`,
  ].some((permission) => permissions.some((value) => value.toLowerCase() === permission.toLowerCase())));
};