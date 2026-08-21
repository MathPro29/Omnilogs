import { Button, Popconfirm, Tooltip } from "antd";
import {
  DeleteOutlined,
  EditOutlined,
  FileTextOutlined,
  FolderOutlined,
  PlusOutlined,
} from "@ant-design/icons";
import type { FeatureNode } from "@/features/product-setup/types/productSetup.types";

interface FeatureTreeNodeTitleProps {
  node: FeatureNode;
  nodeId: string | number;
  hasChildren: boolean;
  onAdd: (parentId: string | number) => void;
  onEdit: (node: FeatureNode) => void;
  onDelete: (node: FeatureNode) => void;
}

export function FeatureTreeNodeTitle(props: FeatureTreeNodeTitleProps) {
  return (
    <div className="flex items-center justify-between group py-1 border-b border-zinc-100/50">
      <div className="flex items-center gap-2">
        {props.node.parent_id ? (
          <FileTextOutlined className="text-zinc-400 text-xs" />
        ) : (
          <FolderOutlined className="text-zinc-700" />
        )}
        <span className="font-medium text-zinc-900 text-sm">
          {props.node.category_name}
        </span>
      </div>
      <div className="opacity-0 group-hover:opacity-100 transition-opacity flex items-center gap-1">
        <Tooltip title="เพิ่ม Child feature">
          <Button
            type="text"
            size="small"
            icon={<PlusOutlined />}
            onClick={(event) => {
              event.stopPropagation();
              props.onAdd(props.nodeId);
            }}
            className="text-zinc-500 hover:text-zinc-900"
          />
        </Tooltip>
        <Tooltip title="แก้ไข Feature">
          <Button
            type="text"
            size="small"
            icon={<EditOutlined />}
            onClick={(event) => {
              event.stopPropagation();
              props.onEdit(props.node);
            }}
            className="text-zinc-500 hover:text-zinc-900"
          />
        </Tooltip>
        <Popconfirm
          title="ลบ Feature หรือไม่?"
          description={
            props.hasChildren
              ? "Feature นี้มี Sub-feature อยู่ การลบจะลบ Sub-feature ทั้งหมดด้วย"
              : "ต้องการลบ Feature นี้หรือไม่?"
          }
          onConfirm={(event) => {
            event?.stopPropagation();
            props.onDelete(props.node);
          }}
          okText="ลบ"
          cancelText="ยกเลิก"
        >
          <Button
            type="text"
            danger
            size="small"
            icon={<DeleteOutlined />}
            onClick={(event) => event.stopPropagation()}
          />
        </Popconfirm>
      </div>
    </div>
  );
}


