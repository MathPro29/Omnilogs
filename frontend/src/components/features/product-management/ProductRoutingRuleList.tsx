import { DeleteOutlined, EditOutlined } from "@ant-design/icons";
import { Button, List, Popconfirm, Space, Tag } from "antd";
import type { ProductLogRoutingRule } from "@/services/product.service";
import { describeConditions } from "./product-routing.utils";

interface ProductRoutingRuleListProps {
  rules: ProductLogRoutingRule[];
  loading: boolean;
  environmentOptions: Array<{ value: number; label: string }>;
  projectName: (id: number) => string;
  featureName: (id?: number | null) => string;
  onEdit: (rule: ProductLogRoutingRule) => void;
  onDelete: (ruleId: number) => void;
}

export function ProductRoutingRuleList({
  rules,
  loading,
  environmentOptions,
  projectName,
  featureName,
  onEdit,
  onDelete,
}: ProductRoutingRuleListProps) {
  return (
    <List
      loading={loading}
      dataSource={rules}
      locale={{ emptyText: "No HTTP routing rules" }}
      style={{ marginTop: 18 }}
      renderItem={(rule) => (
        <List.Item
          actions={[
            <Button
              key="edit"
              type="text"
              icon={<EditOutlined />}
              onClick={() => onEdit(rule)}
              aria-label="Edit routing rule"
            />,
            <Popconfirm
              key="delete"
              title="Delete this routing rule?"
              onConfirm={() => onDelete(rule.rule_id)}
            >
              <Button
                type="text"
                danger
                aria-label="Delete routing rule"
                icon={<DeleteOutlined />}
              />
            </Popconfirm>,
          ]}
        >
          <List.Item.Meta
            title={
              <Space>
                <strong>{rule.rule_name}</strong>
                {rule.environment_id ? (
                  <Tag>
                    {environmentOptions.find(
                      (item) => item.value === rule.environment_id,
                    )?.label ?? rule.environment_id}
                  </Tag>
                ) : (
                  <Tag>All environments</Tag>
                )}
              </Space>
            }
            description={
              <div>
                <div>{describeConditions(rule)}</div>
                <div>
                  {projectName(rule.target_project_id)} /{" "}
                  {featureName(rule.target_category_id)}
                </div>
              </div>
            }
          />
        </List.Item>
      )}
    />
  );
}
