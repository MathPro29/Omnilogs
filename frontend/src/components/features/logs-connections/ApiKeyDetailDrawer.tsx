import { Descriptions, Drawer, Empty, Tabs } from "antd";
import dayjs from "dayjs";
import type { ApiKeyRecord, ProductEnvironment } from "@/types/apiKey.types";
import type { ProductOption } from "@/services/product.service";
import { ApiKeyStatusBadge } from "./ApiKeyStatusBadge";
import { ConnectionGuide } from "./ConnectionGuide";
import { maskApiKey } from "@/features/logs-connections/utils/apiKeyMask";
export function ApiKeyDetailDrawer({
  value,
  product,
  environment,
  onClose,
}: {
  value?: ApiKeyRecord;
  product?: ProductOption;
  environment?: ProductEnvironment;
  onClose: () => void;
}) {
  if (!value) return null;
  const overview = (
    <Descriptions
      column={1}
      size="small"
      items={[
        {
          key: "status",
          label: "Status",
          children: <ApiKeyStatusBadge value={value} />,
        },
        {
          key: "prefix",
          label: "Key",
          children: <code>{maskApiKey(value.prefix)}</code>,
        },
        {
          key: "product",
          label: "Product",
          children: product?.name ?? value.productId,
        },
        {
          key: "environment",
          label: "Environment",
          children: environment?.name ?? "All",
        },
        {
          key: "created",
          label: "Created",
          children: value.createdAt
            ? dayjs(value.createdAt).format("DD MMM YYYY, HH:mm")
            : "—",
        },
        {
          key: "last",
          label: "Last used",
          children: value.lastUsedAt
            ? dayjs(value.lastUsedAt).format("DD MMM YYYY, HH:mm")
            : "Never",
        },
        {
          key: "expires",
          label: "Expiration",
          children: value.expiresAt
            ? dayjs(value.expiresAt).format("DD MMM YYYY, HH:mm")
            : "Never",
        },
      ]}
    />
  );
  const unavailable = (description: string) => (
    <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={description} />
  );
  return (
    <Drawer open title={value.name} size={560} onClose={onClose}>
      <Tabs
        items={[
          { key: "overview", label: "Overview", children: overview },
          {
            key: "permissions",
            label: "Permissions",
            children: (
              <div className="permission-list">
                {value.permissions.map((permission) => (
                  <span key={permission}>{permission}</span>
                ))}
              </div>
            ),
          },
          {
            key: "usage",
            label: "Usage",
            children: unavailable(
              "Usage metrics are not available from the current API.",
            ),
          },
          {
            key: "guide",
            label: "Connection Guide",
            children: (
              <ConnectionGuide
                productId={product?.id}
                environmentCode={environment?.code}
              />
            ),
          },
          { key: "security", label: "Security", children: overview },
          {
            key: "audit",
            label: "Audit History",
            children: unavailable(
              "API key audit history endpoint is not available yet.",
            ),
          },
        ]}
      />
    </Drawer>
  );
}
