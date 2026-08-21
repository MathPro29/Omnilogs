import React from "react";
import { Card, Empty, Tag, Tree } from "antd";
import {
  CheckCircleOutlined,
  DeleteOutlined,
  FileTextOutlined,
  FolderOutlined,
  HddOutlined,
} from "@ant-design/icons";
import type { DataNode } from "antd/es/tree";
import type { FolderNode, RetentionMode } from "@/features/retention-policies/types/retentionPolicy.types";

interface RetentionFolderVisualizerProps {
  mode?: RetentionMode;
  retentionValue?: number;
  treeData?: FolderNode;
  title?: string;
  subtitle?: string;
}

const toTreeNode = (node: FolderNode): DataNode => {
  const icon = node.type === "root" ? <HddOutlined /> : node.type === "day" ? <FileTextOutlined /> : <FolderOutlined />;
  const status = node.status === "TO_BE_PURGED" ? (
    <Tag color="error" icon={<DeleteOutlined />}>
      To be purged
    </Tag>
  ) : node.status === "ARCHIVED" ? (
    <Tag color="warning">Archived</Tag>
  ) : (
    <Tag color="success" icon={<CheckCircleOutlined />}>
      Active
    </Tag>
  );

  return {
    key: node.key,
    icon,
    title: (
      <span>
        {node.title} {status}
      </span>
    ),
    children: node.children?.map(toTreeNode),
  };
};

export const RetentionFolderVisualizer: React.FC<RetentionFolderVisualizerProps> = ({
  treeData,
  title = "Retention folder structure",
  subtitle = "Structure returned by the retention simulation API",
}) => {
  if (!treeData) {
    return (
      <Card title={title} size="small">
        <Empty description="Run a simulation to view the folder structure" />
      </Card>
    );
  }

  return (
    <Card title={title} size="small">
      <p className="mb-3 text-xs text-gray-500">{subtitle}</p>
      <Tree showIcon defaultExpandAll treeData={[toTreeNode(treeData)]} />
    </Card>
  );
};
