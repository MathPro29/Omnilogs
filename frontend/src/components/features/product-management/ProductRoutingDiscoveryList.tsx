import { Button, List, Space, Tag } from "antd";
import type {
  ProductLogRoutingRule,
  RoutingDiscoveryItem,
} from "@/services/product.service";
import { isDiscoveryItemMapped } from "./product-routing.utils";

interface ProductRoutingDiscoveryListProps {
  items: RoutingDiscoveryItem[];
  rules: ProductLogRoutingRule[];
  loading: boolean;
  onUseSuggestion: (item: RoutingDiscoveryItem) => void;
}

export function ProductRoutingDiscoveryList({
  items,
  rules,
  loading,
  onUseSuggestion,
}: ProductRoutingDiscoveryListProps) {
  return (
    <List
      size="small"
      loading={loading}
      dataSource={items}
      renderItem={(item) => {
        const isMapped = isDiscoveryItemMapped(item, rules);

        return (
          <List.Item
            actions={[
              <Button
                key="use"
                size="small"
                disabled={isMapped}
                onClick={() => onUseSuggestion(item)}
              >
                {isMapped ? "Already mapped" : "Add mapping"}
              </Button>,
            ]}
          >
            <List.Item.Meta
              title={
                <Space wrap>
                  {item.service_name ? (
                    <Tag color="blue">{item.service_name}</Tag>
                  ) : null}
                  {item.request_method ? <Tag>{item.request_method}</Tag> : null}
                  <code>{item.request_path}</code>
                  <Tag color={isMapped ? "success" : "default"}>
                    {isMapped ? "Already mapped" : item.routing_status}
                  </Tag>
                </Space>
              }
              description={
                <span>
                  {item.route_pattern && item.route_pattern !== item.request_path ? (
                    <>
                      <code>{item.route_pattern}</code> ·{" "}
                    </>
                  ) : null}
                  {item.total_logs.toLocaleString()} logs
                  {item.routing_method ? ` · ${item.routing_method}` : ""}
                </span>
              }
            />
          </List.Item>
        );
      }}
    />
  );
}
