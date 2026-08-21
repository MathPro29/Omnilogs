import type { ApiKeyRecord, ApiKeyStatus } from "@/types/apiKey.types";
import { statusOf } from "@/features/logs-connections/utils/apiKeyStatus";
export function ApiKeyStatusBadge({ value }: { value: ApiKeyRecord }) {
  const status = statusOf(value);
  const label: Record<ApiKeyStatus, string> = {
    active: "Active",
    disabled: "Disabled",
    expiring: "Expiring soon",
    expired: "Expired",
    revoked: "Revoked",
    "never-used": "Never used",
  };
  return (
    <span className={`connection-status connection-status--${status}`}>
      {label[status]}
    </span>
  );
}
