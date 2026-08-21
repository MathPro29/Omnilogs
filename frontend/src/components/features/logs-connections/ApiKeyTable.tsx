import { Button, Dropdown, Empty, Table } from "antd";
import { MoreOutlined } from "@ant-design/icons";
import dayjs from "dayjs";
import type { ColumnsType } from "antd/es/table";
import type { ApiKeyRecord, ProductEnvironment } from "@/types/apiKey.types";
import type { ProductOption } from "@/services/product.service";
import { ApiKeyStatusBadge } from "./ApiKeyStatusBadge";
import { maskApiKey } from "@/features/logs-connections/utils/apiKeyMask";
interface Props {
  keys: ApiKeyRecord[];
  products: ProductOption[];
  environments: ProductEnvironment[];
  loading: boolean;
  canUpdate: boolean;
  canDelete: boolean;
  onOpen: (key: ApiKeyRecord) => void;
  onEdit?: (key: ApiKeyRecord) => void;
  onActive: (key: ApiKeyRecord, active: boolean) => void;
  onRevoke: (key: ApiKeyRecord) => void;
}
export function ApiKeyTable(props: Props) {
  const productName = (id: number) =>
    props.products.find((item) => item.id === id)?.name ?? `Product ${id}`;
  const environmentName = (id?: number) =>
    props.environments.find((item) => item.environmentId === id)?.name ??
    "All environments";
  const menu = (key: ApiKeyRecord) => ({
    items: [
      { key: "view", label: "View details", onClick: () => props.onOpen(key) },
      ...(props.canUpdate && props.onEdit && !key.revokedAt
        ? [{ key: "edit", label: "Edit", onClick: () => props.onEdit?.(key) }]
        : []),
      ...(props.canUpdate && !key.revokedAt
        ? [
            {
              key: "active",
              label: key.active ? "Disable" : "Enable",
              onClick: () => props.onActive(key, !key.active),
            },
          ]
        : []),
      ...(props.canDelete && !key.revokedAt
        ? [
            {
              key: "revoke",
              danger: true,
              label: "Revoke",
              onClick: () => props.onRevoke(key),
            },
          ]
        : []),
    ],
  });
  const columns: ColumnsType<ApiKeyRecord> = [
    {
      title: "Key name",
      dataIndex: "name",
      fixed: "left",
      width: 170,
      sorter: (a, b) => a.name.localeCompare(b.name),
      render: (name, key) => (
        <button
          className="connection-key-link"
          onClick={() => props.onOpen(key)}
        >
          {name}
        </button>
      ),
    },
    {
      title: "Key prefix",
      dataIndex: "prefix",
      width: 190,
      render: (value) => <code>{maskApiKey(value)}</code>,
    },
    {
      title: "Product",
      dataIndex: "productId",
      width: 150,
      render: productName,
    },
    {
      title: "Environment",
      dataIndex: "environmentId",
      width: 140,
      render: environmentName,
    },
    {
      title: "Status",
      width: 120,
      render: (_, key) => <ApiKeyStatusBadge value={key} />,
    },
    {
      title: "Permission scope",
      dataIndex: "permissions",
      width: 200,
      render: (value: string[]) => value.join(", "),
    },
    {
      title: "Created",
      dataIndex: "createdAt",
      width: 130,
      render: (value) => (value ? dayjs(value).format("DD MMM YYYY") : "—"),
    },
    {
      title: "Last used",
      dataIndex: "lastUsedAt",
      width: 140,
      render: (value) =>
        value ? dayjs(value).format("DD MMM, HH:mm") : "Never",
    },
    {
      title: "Expiration",
      dataIndex: "expiresAt",
      width: 140,
      render: (value) => (value ? dayjs(value).format("DD MMM YYYY") : "Never"),
    },
    {
      title: "Actions",
      fixed: "right",
      width: 72,
      render: (_, key) => (
        <Dropdown trigger={["click"]} menu={menu(key)}>
          <Button
            type="text"
            icon={<MoreOutlined />}
            aria-label={`Actions for ${key.name}`}
          />
        </Dropdown>
      ),
    },
  ];
  return (
    <div className="connections-results">
      <Table
        rowKey="keyId"
        className="connections-table"
        columns={columns}
        dataSource={props.keys}
        loading={props.loading}
        scroll={{ x: 1450 }}
        pagination={{ pageSize: 20, showSizeChanger: false }}
        locale={{
          emptyText: (
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              description="No API keys found for this scope"
            />
          ),
        }}
      />
      <div className="connections-mobile-list">
        {props.keys.length ? (
          props.keys.map((key) => (
            <article key={key.keyId} onClick={() => props.onOpen(key)}>
              <header>
                <div>
                  <strong>{key.name}</strong>
                  <code>{maskApiKey(key.prefix)}</code>
                </div>
                <Dropdown trigger={["click"]} menu={menu(key)}>
                  <Button
                    type="text"
                    icon={<MoreOutlined />}
                    onClick={(event) => event.stopPropagation()}
                    aria-label={`Actions for ${key.name}`}
                  />
                </Dropdown>
              </header>
              <div>
                <ApiKeyStatusBadge value={key} />
                <span>{environmentName(key.environmentId)}</span>
              </div>
              <footer>
                <span>{productName(key.productId)}</span>
                <span>
                  {key.lastUsedAt
                    ? `Used ${dayjs(key.lastUsedAt).format("DD MMM")}`
                    : "Never used"}
                </span>
              </footer>
            </article>
          ))
        ) : (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description="No API keys found for this scope"
          />
        )}
      </div>
    </div>
  );
}
