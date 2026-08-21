import type { User } from "@/types";

type UnknownRecord = Record<string, unknown>;

function asRecord(value: unknown): UnknownRecord {
  return typeof value === "object" && value !== null
    ? (value as UnknownRecord)
    : {};
}

function firstText(...values: unknown[]): string {
  const value = values.find(
    (item): item is string => typeof item === "string" && item.length > 0,
  );
  return value ?? "";
}

export function buildUserDisplayMap(
  currentUser: User | null,
  response: unknown,
): Record<number, string> {
  const map: Record<number, string> = {};
  if (currentUser) {
    const currentId = Number(currentUser.id);
    if (Number.isFinite(currentId) && currentId > 0) {
      map[currentId] = firstText(
        currentUser.fullName,
        currentUser.username,
        currentUser.email,
      );
    }
  }

  const envelope = asRecord(response);
  const rawUsers = Array.isArray(response)
    ? response
    : Array.isArray(envelope.data)
      ? envelope.data
      : [];
  rawUsers.forEach((value) => {
    const user = asRecord(value);
    const id = Number(user.id ?? user.user_id ?? user.ID);
    if (!Number.isFinite(id) || id <= 0) return;
    map[id] = firstText(
      user.fullName,
      user.full_name,
      user.username,
      user.name,
      user.email,
    );
  });
  return map;
}
