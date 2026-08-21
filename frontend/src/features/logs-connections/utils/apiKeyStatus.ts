import dayjs from "dayjs";
import type { ApiKeyRecord, ApiKeyStatus } from "@/types/apiKey.types";

export const statusOf = (key: ApiKeyRecord): ApiKeyStatus =>
  key.revokedAt ? "revoked"
    : key.expiresAt && dayjs(key.expiresAt).isBefore(dayjs()) ? "expired"
      : !key.active ? "disabled"
        : key.expiresAt && dayjs(key.expiresAt).diff(dayjs(), "day") <= 14 ? "expiring"
          : !key.lastUsedAt ? "never-used" : "active";
