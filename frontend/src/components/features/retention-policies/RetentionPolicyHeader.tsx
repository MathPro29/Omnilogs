import React from "react";
import { Button, Space, Breadcrumb } from "antd";
import {
  PlusOutlined,
  ExperimentOutlined,
  ReloadOutlined,
  ClockCircleOutlined,
} from "@ant-design/icons";

interface RetentionPolicyHeaderProps {
  onOpenCreate: () => void;
  onOpenSimulate: () => void;
  onRefresh: () => void;
  isLoading?: boolean;
}

export const RetentionPolicyHeader: React.FC<RetentionPolicyHeaderProps> = ({
  onOpenCreate,
  onOpenSimulate,
  onRefresh,
  isLoading,
}) => {
  return (
    <div className="mb-6">
      <Breadcrumb
        items={[
          { title: "หน้าหลัก" },
          { title: "การตั้งค่าระบบ" },
          { title: "Retention Policy" },
        ]}
        className="mb-2 text-xs"
      />
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2.5">
            <ClockCircleOutlined className="text-emerald-600" />
            การจัดการ Retention Policy (Log Expiration Rules)
          </h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
            กำหนดระยะเวลาการเก็บรักษา Log (ตามวัน, ตามสัปดาห์ 7 โฟลเดอร์, ตามเดือน 4 โฟลเดอร์ย่อย) และตั้งค่าลบ/ย้ายข้อมูลอัตโนมัติ
          </p>
        </div>

        <Space wrap>
          <Button
            icon={<ReloadOutlined spin={isLoading} />}
            onClick={onRefresh}
            disabled={isLoading}
          >
            รีเฟรชข้อมูล
          </Button>
          <Button
            icon={<ExperimentOutlined />}
            onClick={onOpenSimulate}
            className="border-emerald-600 text-emerald-700 hover:bg-emerald-50"
          >
            จำลองการลบ Log (Dry-Run)
          </Button>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={onOpenCreate}
            className="bg-[#1F8457] hover:bg-[#186a45]"
          >
            สร้าง Retention Policy ใหม่
          </Button>
        </Space>
      </div>
    </div>
  );
};
