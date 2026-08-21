import React, { useEffect } from "react";
import {
  Drawer,
  Form,
  Input,
  Select,
  Radio,
  InputNumber,
  Switch,
  Button,
  Space,
  Divider,
  Alert,
} from "antd";
import type {
  RetentionPolicy,
  CreateRetentionPolicyPayload,
  RetentionMode,
  RetentionUnit,
} from "@/features/retention-policies/types/retentionPolicy.types";
import { RetentionFolderVisualizer } from "./RetentionFolderVisualizer";

interface RetentionPolicyDrawerProps {
  open: boolean;
  onClose: () => void;
  onSave: (payload: CreateRetentionPolicyPayload) => Promise<void>;
  initialData?: RetentionPolicy | null;
  loading?: boolean;
}

type RetentionPolicyFormValues = Omit<CreateRetentionPolicyPayload, "product_id">;

export const RetentionPolicyDrawer: React.FC<RetentionPolicyDrawerProps> = ({
  open,
  onClose,
  onSave,
  initialData,
  loading = false,
}) => {
  const [form] = Form.useForm<RetentionPolicyFormValues>();
  const selectedMode: RetentionMode = Form.useWatch("retention_mode", form) || "WEEKLY";

  useEffect(() => {
    if (open) {
      if (initialData) {
        form.setFieldsValue({
          name: initialData.name,
          description: initialData.description,
          environment_id: initialData.environment_id,
          retention_mode: initialData.retention_mode,
          retention_unit: initialData.retention_unit,
          retention_value: initialData.retention_value,
          auto_purge_action: initialData.auto_purge_action,
          storage_provider: initialData.storage_provider,
          bucket_name: initialData.bucket_name,
          cron_schedule: initialData.cron_schedule,
          is_active: initialData.is_active,
        });
      } else {
        form.resetFields();
        form.setFieldsValue({
          retention_mode: "WEEKLY",
          retention_unit: "WEEKS",
          retention_value: 1,
          auto_purge_action: "DELETE",
          storage_provider: "LOCAL",
          cron_schedule: "0 0 * * *",
          is_active: true,
        });
      }
    }
  }, [open, initialData, form]);

  const handleFinish = async (values: RetentionPolicyFormValues) => {
    let unit: RetentionUnit = values.retention_unit;
    if (values.retention_mode === "WEEKLY") unit = "WEEKS";
    else if (values.retention_mode === "MONTHLY") unit = "MONTHS";
    else if (values.retention_mode === "DAILY") unit = "DAYS";

    const payload: CreateRetentionPolicyPayload = {
      product_id: 1,
      name: values.name,
      description: values.description || "",
      environment_id: values.environment_id,
      retention_mode: values.retention_mode,
      retention_unit: unit,
      retention_value: values.retention_value || 1,
      folder_structure:
        values.retention_mode === "WEEKLY"
          ? "7_DAILY_FOLDERS"
          : values.retention_mode === "MONTHLY"
          ? "4_WEEKLY_SUBFOLDERS"
          : "CUSTOM",
      auto_purge_action: values.auto_purge_action,
      storage_provider: values.storage_provider,
      bucket_name: values.bucket_name,
      cron_schedule: values.cron_schedule,
      is_active: values.is_active,
    };

    await onSave(payload);
    onClose();
  };

  const handleModeChange = (mode: RetentionMode) => {
    if (mode === "WEEKLY") {
      form.setFieldsValue({ retention_unit: "WEEKS", retention_value: 1 });
    } else if (mode === "MONTHLY") {
      form.setFieldsValue({ retention_unit: "MONTHS", retention_value: 1 });
    } else if (mode === "DAILY") {
      form.setFieldsValue({ retention_unit: "DAYS", retention_value: 7 });
    }
  };

  return (
    <Drawer
      title={initialData ? "แก้ไข Retention Policy" : "สร้าง Retention Policy ใหม่"}
      width={720}
      open={open}
      onClose={onClose}
      destroyOnClose
      extra={
        <Space>
          <Button onClick={onClose}>ยกเลิก</Button>
          <Button
            type="primary"
            onClick={() => form.submit()}
            loading={loading}
            className="bg-[#1F8457]"
          >
            บันทึกนโยบาย
          </Button>
        </Space>
      }
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={handleFinish}
        initialValues={{
          retention_mode: "WEEKLY",
          retention_unit: "WEEKS",
          retention_value: 1,
          auto_purge_action: "DELETE",
          storage_provider: "LOCAL",
          cron_schedule: "0 0 * * *",
          is_active: true,
        }}
      >
        <Form.Item
          name="name"
          label="ชื่อนโยบาย (Policy Name)"
          rules={[{ required: true, message: "กรุณาระบุชื่อนโยบาย" }]}
        >
          <Input placeholder="เช่น Standard Weekly Log Rotation" />
        </Form.Item>

        <Form.Item name="description" label="คำอธิบาย (Description)">
          <Input.TextArea rows={2} placeholder="เช่น หมุนเวียนลบ Log ทุก 7 วันโดยแบ่ง 7 โฟลเดอร์ย่อย" />
        </Form.Item>

        <Divider children="การเลือกรูปแบบช่วงเวลาเก็บรักษา (Retention Strategy)" />

        <Form.Item
          name="retention_mode"
          label="รูปแบบการกำหนดอายุ Log"
          rules={[{ required: true }]}
        >
          <Radio.Group
            onChange={(e) => handleModeChange(e.target.value)}
            optionType="button"
            buttonStyle="solid"
            className="w-full grid grid-cols-2 md:grid-cols-4 gap-2 text-center"
          >
            <Radio.Button value="WEEKLY" className="text-center">ตามสัปดาห์ (7 Folders)</Radio.Button>
            <Radio.Button value="MONTHLY" className="text-center">ตามเดือน (4 Subfolders)</Radio.Button>
            <Radio.Button value="DAILY" className="text-center">ตามวัน (Daily)</Radio.Button>
            <Radio.Button value="CUSTOM" className="text-center">กำหนดเอง (Custom)</Radio.Button>
          </Radio.Group>
        </Form.Item>

        {selectedMode === "WEEKLY" && (
          <Alert
            type="success"
            showIcon
            message="โหมดตามสัปดาห์ (Weekly Mode): 7 โฟลเดอร์ย่อย"
            description="ระบบจะสร้าง 7 โฟลเดอร์ย่อยตามรายวัน (Day 1 - Day 7) เมื่อครบรอบ Log วันเก่าจะถูกหมุนเวียนลบอัตโนมัติ"
            className="mb-4 text-xs"
          />
        )}

        {selectedMode === "MONTHLY" && (
          <Alert
            type="info"
            showIcon
            message="โหมดตามเดือน (Monthly Mode): 4 โฟลเดอร์ย่อยรายสัปดาห์"
            description="ระบบจะสร้าง 4 โฟลเดอร์ย่อยรายสัปดาห์ (Week 1 - Week 4) และภายในแต่ละสัปดาห์มี 7 โฟลเดอร์ย่อยรายวัน"
            className="mb-4 text-xs"
          />
        )}

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Form.Item
            name="retention_value"
            label="จำนวนระยะเวลาเก็บรักษา"
            rules={[{ required: true, message: "กรุณาระบุจำนวน" }]}
          >
            <InputNumber min={1} max={365} className="w-full" addonAfter={selectedMode === "WEEKLY" ? "สัปดาห์" : selectedMode === "MONTHLY" ? "เดือน" : "วัน"} />
          </Form.Item>

          {selectedMode === "CUSTOM" && (
            <Form.Item name="retention_unit" label="หน่วยเวลา">
              <Select options={[
                { label: "วัน (Days)", value: "DAYS" },
                { label: "สัปดาห์ (Weeks)", value: "WEEKS" },
                { label: "เดือน (Months)", value: "MONTHS" },
                { label: "ปี (Years)", value: "YEARS" },
              ]} />
            </Form.Item>
          )}
        </div>

        {/* Live Visualizer Preview Component inside Drawer */}
        <div className="my-4">
          <div className="text-xs font-bold text-gray-700 mb-2">
            ตัวอย่างโครงสร้างโฟลเดอร์แบบเรียลไทม์ (Live Folder Preview)
          </div>
          <RetentionFolderVisualizer />
        </div>

        <Divider children="การตั้งค่าการลบและคลังเก็บข้อมูล (Storage & Auto-Purge)" />

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Form.Item name="auto_purge_action" label="การดำเนินการเมื่อ Log หมดอายุ">
            <Select options={[
              { label: "ลบถาวร (Permanent Delete)", value: "DELETE" },
              { label: "ย้ายไป Cold Storage (Archive)", value: "ARCHIVE_COLD" },
              { label: "บีบอัดไฟล์ (GZIP Compress)", value: "COMPRESS_GZIP" },
            ]} />
          </Form.Item>

          <Form.Item name="storage_provider" label="ผู้ให้บริการ Storage">
            <Select options={[
              { label: "Local Disk / Server", value: "LOCAL" },
              { label: "Amazon S3", value: "S3" },
              { label: "Google Cloud Storage (GCS)", value: "GCS" },
              { label: "MinIO S3 Compatible", value: "MINIO" },
            ]} />
          </Form.Item>
        </div>

        <Form.Item name="bucket_name" label="ชื่อ Bucket / Path ปลายทาง">
          <Input placeholder="เช่น omnilogs-daily-archive-bucket" />
        </Form.Item>

        <Form.Item name="cron_schedule" label="รอบการทำงาน (Cron Schedule)">
          <Input placeholder="0 0 * * * (ทุกวัน เวลา 00:00 UTC)" />
        </Form.Item>

        <Form.Item name="is_active" valuePropName="checked" label="เปิดใช้งานทันที">
          <Switch checkedChildren="เปิด" unCheckedChildren="ปิด" />
        </Form.Item>
      </Form>
    </Drawer>
  );
};
