import React from "react";
import { Table, Tag, Switch, Button, Popconfirm, Tooltip, Space } from "antd";
import {
  EditOutlined,
  DeleteOutlined,
  FolderOpenOutlined,
  ClockCircleOutlined,
  CloudUploadOutlined,
  CompressOutlined,
  ThunderboltOutlined,
} from "@ant-design/icons";
import type { ColumnsType } from "antd/es/table";
import type { RetentionPolicy, RetentionMode } from "@/features/retention-policies/types/retentionPolicy.types";

interface RetentionPolicyTableProps {
  policies: RetentionPolicy[];
  loading?: boolean;
  onEdit: (policy: RetentionPolicy) => void;
  onDelete: (policyId: number) => void;
  onToggle: (policyId: number, isActive: boolean) => void;
  onSelectPolicyForVisualizer: (policy: RetentionPolicy) => void;
  onSimulate: (policy: RetentionPolicy) => void;
}

export const RetentionPolicyTable: React.FC<RetentionPolicyTableProps> = ({
  policies,
  loading = false,
  onEdit,
  onDelete,
  onToggle,
  onSelectPolicyForVisualizer,
  onSimulate,
}) => {
  const getModeTag = (mode: RetentionMode) => {
    switch (mode) {
      case "WEEKLY":
        return <Tag color="blue" icon={<ClockCircleOutlined />}>รายสัปดาห์ (7 Folders)</Tag>;
      case "MONTHLY":
        return <Tag color="purple" icon={<FolderOpenOutlined />}>รายเดือน (4 Subfolders)</Tag>;
      case "DAILY":
        return <Tag color="cyan">รายวัน (Daily)</Tag>;
      case "CUSTOM":
        return <Tag color="orange">กำหนดเอง (Custom)</Tag>;
      default:
        return <Tag>{mode}</Tag>;
    }
  };

  const getActionTag = (action: string) => {
    switch (action) {
      case "DELETE":
        return <Tag color="error" icon={<DeleteOutlined />}>Permanent Delete</Tag>;
      case "ARCHIVE_COLD":
        return <Tag color="warning" icon={<CloudUploadOutlined />}>Cold Storage Archive</Tag>;
      case "COMPRESS_GZIP":
        return <Tag color="magenta" icon={<CompressOutlined />}>GZIP Compress</Tag>;
      default:
        return <Tag>{action}</Tag>;
    }
  };

  const columns: ColumnsType<RetentionPolicy> = [
    {
      title: "ชื่อ นโยบาย (Policy Name)",
      dataIndex: "name",
      key: "name",
      render: (text, record) => (
        <div>
          <div className="font-bold text-gray-900 flex items-center gap-2">
            <span>{text}</span>
            {record.is_active ? (
              <Tag color="success" className="m-0 text-[10px]">Active</Tag>
            ) : (
              <Tag color="default" className="m-0 text-[10px]">Inactive</Tag>
            )}
          </div>
          <div className="text-xs text-gray-500 line-clamp-1">{record.description}</div>
        </div>
      ),
    },
    {
      title: "โหมด Retention Period",
      dataIndex: "retention_mode",
      key: "retention_mode",
      render: (mode: RetentionMode, record) => (
        <div>
          {getModeTag(mode)}
          <div className="text-xs font-semibold text-gray-700 mt-1">
            เก็บ {record.retention_value} {record.retention_unit} ({record.total_days} วัน)
          </div>
        </div>
      ),
    },
    {
      title: "โครงสร้างโฟลเดอร์",
      dataIndex: "folder_structure",
      key: "folder_structure",
      render: (_, record) => {
        if (record.retention_mode === "WEEKLY") {
          return (
            <div className="text-xs">
              <span className="font-semibold text-emerald-700">7 โฟลเดอร์ย่อยรายวัน</span>
              <div className="text-[11px] text-gray-500">Day 1 ถึง Day 7</div>
            </div>
          );
        }
        if (record.retention_mode === "MONTHLY") {
          return (
            <div className="text-xs">
              <span className="font-semibold text-purple-700">4 โฟลเดอร์ย่อยรายสัปดาห์</span>
              <div className="text-[11px] text-gray-500">Week 1 ถึง Week 4 (28-31 Days)</div>
            </div>
          );
        }
        return <span className="text-xs text-gray-600 font-mono">{record.folder_structure}</span>;
      },
    },
    {
      title: "การจัดการเมื่อครบกำหนด",
      dataIndex: "auto_purge_action",
      key: "auto_purge_action",
      render: (action) => getActionTag(action),
    },
    {
      title: "ที่เก็บข้อมูล (Provider)",
      dataIndex: "storage_provider",
      key: "storage_provider",
      render: (provider, record) => (
        <div className="text-xs">
          <Tag color="geekblue">{provider}</Tag>
          {record.bucket_name && (
            <div className="text-[11px] text-gray-400 font-mono mt-0.5 truncate max-w-[120px]">
              {record.bucket_name}
            </div>
          )}
        </div>
      ),
    },
    {
      title: "เปิดใช้งาน",
      dataIndex: "is_active",
      key: "is_active",
      width: 100,
      render: (isActive: boolean, record) => (
        <Switch
          checked={isActive}
          onChange={(checked) => onToggle(record.policy_id, checked)}
          size="small"
        />
      ),
    },
    {
      title: "จัดการ",
      key: "actions",
      width: 160,
      render: (_, record) => (
        <Space size="small">
          <Tooltip title="ดูโครงสร้างโฟลเดอร์ย่อย">
            <Button
              type="text"
              icon={<FolderOpenOutlined className="text-emerald-600" />}
              onClick={() => onSelectPolicyForVisualizer(record)}
            />
          </Tooltip>
          <Tooltip title="จำลองการลบ (Simulate)">
            <Button
              type="text"
              icon={<ThunderboltOutlined className="text-amber-600" />}
              onClick={() => onSimulate(record)}
            />
          </Tooltip>
          <Tooltip title="แก้ไขนโยบาย">
            <Button
              type="text"
              icon={<EditOutlined className="text-blue-600" />}
              onClick={() => onEdit(record)}
            />
          </Tooltip>
          <Popconfirm
            title="ลบนโยบายนี้?"
            description="คุณแน่ใจหรือไม่ว่าต้องการลบ Retention Policy นี้?"
            onConfirm={() => onDelete(record.policy_id)}
            okText="ลบ"
            cancelText="ยกเลิก"
            okButtonProps={{ danger: true }}
          >
            <Tooltip title="ลบ">
              <Button type="text" danger icon={<DeleteOutlined />} />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Table
      columns={columns}
      dataSource={policies}
      rowKey="policy_id"
      loading={loading}
      pagination={{ pageSize: 10, showSizeChanger: true }}
      className="bg-white rounded-xl shadow-xs border border-gray-200"
    />
  );
};
