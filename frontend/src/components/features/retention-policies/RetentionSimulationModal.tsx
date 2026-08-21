import React, { useState } from "react";
import { Modal, Select, InputNumber, Button, Alert, Divider } from "antd";
import { ThunderboltOutlined, ExperimentOutlined } from "@ant-design/icons";
import type { RetentionMode, RetentionPolicy, RetentionUnit, SimulationResult } from "@/features/retention-policies/types/retentionPolicy.types";
import { RetentionFolderVisualizer } from "./RetentionFolderVisualizer";

interface RetentionSimulationModalProps {
  open: boolean;
  onClose: () => void;
  onRunSimulate: (params: {
    retention_mode: RetentionMode;
    retention_unit: RetentionUnit;
    retention_value: number;
  }) => Promise<SimulationResult>;
  policy?: RetentionPolicy | null;
  loading?: boolean;
}

export const RetentionSimulationModal: React.FC<RetentionSimulationModalProps> = ({
  open,
  onClose,
  onRunSimulate,
  policy,
  loading = false,
}) => {
  const [simulationResult, setSimulationResult] = useState<SimulationResult | null>(null);
  const [mode, setMode] = useState<RetentionMode>(policy?.retention_mode || "WEEKLY");
  const [value, setValue] = useState<number>(policy?.retention_value || 1);

  const handleSimulate = async () => {
    const res = await onRunSimulate({
      retention_mode: mode,
      retention_unit: mode === "WEEKLY" ? "WEEKS" : mode === "MONTHLY" ? "MONTHS" : "DAYS",
      retention_value: value,
    });
    setSimulationResult(res);
  };

  return (
    <Modal
      title={
        <div className="flex items-center gap-2 text-emerald-700 font-bold">
          <ExperimentOutlined />
          จำลองผลการทำงาน Retention Policy (Dry-Run Simulator)
        </div>
      }
      open={open}
      onCancel={onClose}
      width={800}
      footer={[
        <Button key="close" onClick={onClose}>
          ปิด
        </Button>,
        <Button
          key="run"
          type="primary"
          icon={<ThunderboltOutlined />}
          onClick={handleSimulate}
          loading={loading}
          className="bg-emerald-600 hover:bg-emerald-700"
        >
          ประมวลผลการจำลอง
        </Button>,
      ]}
    >
      <div className="space-y-4 my-2">
        <Alert
          type="warning"
          showIcon
          message="การจำลองแบบ Dry-Run จะไม่ส่งผลต่อ Log จริงในระบบ"
          description="ระบบจะคำนวณวันหมดอายุ จำนวนโฟลเดอร์ที่จะถูกหมุนเวียนลบ และประมาณการประหยัดเนื้อที่ดิสก์ให้เห็นก่อนนำไปใช้งานจริง"
        />

        <div className="bg-gray-50 p-4 rounded-lg border border-gray-200 grid grid-cols-1 md:grid-cols-3 gap-3">
          <div>
            <label className="block text-xs font-bold text-gray-700 mb-1">โหมด Retention Policy</label>
            <Select
              value={mode}
              onChange={(v) => setMode(v)}
              className="w-full"
              options={[
                { label: "รายสัปดาห์ (7 โฟลเดอร์ย่อย)", value: "WEEKLY" },
                { label: "รายเดือน (4 โฟลเดอร์ย่อย/เดือน)", value: "MONTHLY" },
                { label: "รายวัน (Daily)", value: "DAILY" },
                { label: "กำหนดเอง (Custom)", value: "CUSTOM" },
              ]}
            />
          </div>

          <div>
            <label className="block text-xs font-bold text-gray-700 mb-1">จำนวนระยะเวลา</label>
            <InputNumber
              min={1}
              max={365}
              value={value}
              onChange={(v) => setValue(v || 1)}
              className="w-full"
              addonAfter={mode === "WEEKLY" ? "สัปดาห์" : mode === "MONTHLY" ? "เดือน" : "วัน"}
            />
          </div>

          <div className="flex items-end">
            <Button
              type="primary"
              block
              icon={<ThunderboltOutlined />}
              onClick={handleSimulate}
              loading={loading}
              className="bg-emerald-600"
            >
              ทดสอบจำลอง
            </Button>
          </div>
        </div>

        {simulationResult && (
          <div>
            <Divider children="ผลการจำลอง (Simulation Breakdown)" />
            <div className="grid grid-cols-1 md:grid-cols-4 gap-3 mb-4">
              <div className="bg-blue-50 border border-blue-200 p-3 rounded-lg">
                <div className="text-xs text-blue-700 font-medium">วันหมดอายุของ Log (Cutoff)</div>
                <div className="text-base font-bold text-blue-900 mt-1">{simulationResult.calculated_cutoff_date}</div>
              </div>

              <div className="bg-emerald-50 border border-emerald-200 p-3 rounded-lg">
                <div className="text-xs text-emerald-700 font-medium">โฟลเดอร์ที่เก็บรักษาอยู่</div>
                <div className="text-base font-bold text-emerald-900 mt-1">
                  {simulationResult.active_folders_count} Folders
                </div>
              </div>

              <div className="bg-amber-50 border border-amber-200 p-3 rounded-lg">
                <div className="text-xs text-amber-700 font-medium">โฟลเดอร์ที่จะถูกลบออก</div>
                <div className="text-base font-bold text-amber-900 mt-1">
                  {simulationResult.purge_folders_count} Folders
                </div>
              </div>

              <div className="bg-purple-50 border border-purple-200 p-3 rounded-lg">
                <div className="text-xs text-purple-700 font-medium">พื้นที่ดิสก์ที่จะได้คืน</div>
                <div className="text-base font-bold text-purple-900 mt-1">
                  ~{simulationResult.estimated_savings_mb.toFixed(1)} MB
                </div>
              </div>
            </div>

            <RetentionFolderVisualizer
              mode={simulationResult.retention_mode}
              retentionValue={value}
              treeData={simulationResult.tree_data}
              title="โครงสร้างโฟลเดอร์และสถานะการลบจากการจำลอง"
              subtitle={`ระบบคำนวณ Log อายุเกินวันที่ ${simulationResult.calculated_cutoff_date} ว่าจะถูกหมุนเวียนลบโฟลเดอร์ออก`}
            />
          </div>
        )}
      </div>
    </Modal>
  );
};
