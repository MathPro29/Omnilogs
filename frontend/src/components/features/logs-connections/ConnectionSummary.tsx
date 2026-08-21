import type { ApiKeyRecord } from "@/types/apiKey.types";
import { statusOf } from "@/features/logs-connections/utils/apiKeyStatus";
import dayjs from "dayjs";
export function ConnectionSummary({ keys }: { keys: ApiKeyRecord[] }) {
  const statuses = keys.map(statusOf);
  const last = keys
    .map((key) => key.lastUsedAt)
    .filter(Boolean)
    .sort()
    .at(-1);
  const values = [
    ["Total API Keys", keys.length],
    [
      "Active Keys",
      statuses.filter(
        (status) => status === "active" || status === "never-used",
      ).length,
    ],
    [
      "Expiring Soon",
      statuses.filter((status) => status === "expiring").length,
    ],
    ["Revoked Keys", statuses.filter((status) => status === "revoked").length],
    [
      "Last Log Received",
      last ? dayjs(last).format("DD MMM, HH:mm") : "No activity",
    ],
    [
      "Active Connections",
      keys.filter((key) => key.active && !key.revokedAt).length,
    ],
  ];
  return (
    <section className="connections-summary">
      {values.map(([label, value]) => (
        <div key={label}>
          <span>{label}</span>
          <strong>{value}</strong>
        </div>
      ))}
    </section>
  );
}
