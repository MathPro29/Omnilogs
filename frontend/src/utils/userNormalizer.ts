import type { User } from "@/types";

type UnknownRecord = Record<string, unknown>;

function asRecord(value: unknown): UnknownRecord {
  return typeof value === "object" && value !== null
    ? (value as UnknownRecord)
    : {};
}

function text(value: unknown): string {
  return typeof value === "string" ? value : "";
}

export function normalizeUsersResponse(response: unknown): User[] {
  const root = asRecord(response);
  const data = root.data;
  const dataEnvelope = asRecord(data);
  const list = Array.isArray(data)
    ? data
    : Array.isArray(dataEnvelope.data)
      ? dataEnvelope.data
      : Array.isArray(response)
        ? response
        : [];

  return list.map((value) => {
    const user = asRecord(value);
    const username = text(user.username);
    const email = text(user.email);
    const roleName = text(user.role) || text(user.platform_role_name);
    const isActive =
      typeof user.is_active === "boolean" ? user.is_active : undefined;
    return {
      id: String(user.id ?? user.user_id ?? user.ID ?? ""),
      username,
      email,
      fullName:
        text(user.full_name) ||
        text(user.fullName) ||
        username ||
        email,
      phone: text(user.phone) || text(user.phone_number),
      department: text(user.department),
      position: text(user.position),
      status: (text(user.status) || (isActive === false ? "inactive" : "active")) as User["status"],
      roles: Array.isArray(user.roles)
        ? (user.roles as User["roles"])
        : roleName
          ? [{ id: String(user.platform_role_id ?? roleName), name: roleName, permissions: [] }]
          : [],
      permissions: Array.isArray(user.permissions)
        ? (user.permissions as User["permissions"])
        : [],
      createdAt: text(user.created_at) || text(user.createdAt),
      updatedAt: text(user.updated_at) || text(user.updatedAt),
    };
  });
}
