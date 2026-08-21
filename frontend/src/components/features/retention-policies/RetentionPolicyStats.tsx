import React from "react";
import { Card, Statistic, Row, Col, Tooltip } from "antd";
import {
  ClockCircleOutlined,
  FolderOpenOutlined,
  HddOutlined,
  CheckCircleOutlined,
  SyncOutlined,
} from "@ant-design/icons";
import type { PolicyStats } from "@/features/retention-policies/types/retentionPolicy.types";

interface RetentionPolicyStatsProps {
  stats?: PolicyStats;
  loading?: boolean;
}

export const RetentionPolicyStats: React.FC<RetentionPolicyStatsProps> = ({
  stats,
  loading = false,
}) => {
  const total = stats?.total_policies ?? 0;
  const active = stats?.active_policies ?? 0;
  const storage = stats?.total_storage_gb ?? 0;
  const purges = stats?.scheduled_purges_24h ?? 0;
  const folders = stats?.total_folders_count ?? 0;

  return (
    <div className="mb-6">
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} md={8} lg={4.8} className="w-full sm:w-1/2 md:w-1/3 lg:w-1/5">
          <Card size="small" loading={loading} className="hover:shadow-md transition-shadow border-gray-200">
            <Statistic
              title={
                <span className="text-xs text-gray-500 font-medium flex items-center gap-1.5">
                  <ClockCircleOutlined className="text-emerald-600" />
                  Policy ทั้งหมด
                </span>
              }
              value={total}
              suffix="Rules"
              valueStyle={{ color: "#111827", fontWeight: 700, fontSize: "1.5rem" }}
            />
            <div className="text-[11px] text-gray-400 mt-1">นโยบายการเก็บ Logs ในระบบ</div>
          </Card>
        </Col>

        <Col xs={24} sm={12} md={8} lg={4.8} className="w-full sm:w-1/2 md:w-1/3 lg:w-1/5">
          <Card size="small" loading={loading} className="hover:shadow-md transition-shadow border-emerald-200 bg-emerald-50/30">
            <Statistic
              title={
                <span className="text-xs text-emerald-700 font-medium flex items-center gap-1.5">
                  <CheckCircleOutlined className="text-emerald-600" />
                  เปิดใช้งานอยู่ (Active)
                </span>
              }
              value={active}
              suffix={`/ ${total}`}
              valueStyle={{ color: "#065f46", fontWeight: 700, fontSize: "1.5rem" }}
            />
            <div className="text-[11px] text-emerald-600 mt-1">พร้อมทำงานตามรอบ Cron Schedule</div>
          </Card>
        </Col>

        <Col xs={24} sm={12} md={8} lg={4.8} className="w-full sm:w-1/2 md:w-1/3 lg:w-1/5">
          <Card size="small" loading={loading} className="hover:shadow-md transition-shadow border-blue-200 bg-blue-50/30">
            <Statistic
              title={
                <span className="text-xs text-blue-700 font-medium flex items-center gap-1.5">
                  <FolderOpenOutlined className="text-blue-600" />
                  โฟลเดอร์ย่อยทั้งหมด
                </span>
              }
              value={folders}
              suffix="Folders"
              valueStyle={{ color: "#1e40af", fontWeight: 700, fontSize: "1.5rem" }}
            />
            <div className="text-[11px] text-blue-600 mt-1">7 Folders/สัปดาห์ & 4 Subfolders/เดือน</div>
          </Card>
        </Col>

        <Col xs={24} sm={12} md={8} lg={4.8} className="w-full sm:w-1/2 md:w-1/3 lg:w-1/5">
          <Card size="small" loading={loading} className="hover:shadow-md transition-shadow border-purple-200 bg-purple-50/30">
            <Statistic
              title={
                <span className="text-xs text-purple-700 font-medium flex items-center gap-1.5">
                  <HddOutlined className="text-purple-600" />
                  พื้นที่บริหารจัดการ
                </span>
              }
              value={storage}
              precision={1}
              suffix="GB"
              valueStyle={{ color: "#5b21b6", fontWeight: 700, fontSize: "1.5rem" }}
            />
            <div className="text-[11px] text-purple-600 mt-1">ขนาด Log Archives ที่ถูกดูแล</div>
          </Card>
        </Col>

        <Col xs={24} sm={12} md={8} lg={4.8} className="w-full sm:w-1/2 md:w-1/3 lg:w-1/5">
          <Card size="small" loading={loading} className="hover:shadow-md transition-shadow border-amber-200 bg-amber-50/30">
            <Statistic
              title={
                <span className="text-xs text-amber-700 font-medium flex items-center gap-1.5">
                  <SyncOutlined className="text-amber-600" />
                  รอบการลบถัดไป
                </span>
              }
              value={purges}
              suffix="Jobs (24h)"
              valueStyle={{ color: "#92400e", fontWeight: 700, fontSize: "1.5rem" }}
            />
            <Tooltip title="งานหมุนเวียนลบโฟลเดอร์เก่าที่จะทำงานภายใน 24 ชั่วโมงข้างหน้า">
              <div className="text-[11px] text-amber-600 mt-1">กำหนดทำงานเวลา 00:00 UTC</div>
            </Tooltip>
          </Card>
        </Col>
      </Row>
    </div>
  );
};
