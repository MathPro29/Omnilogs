import { useState } from "react";
import { Card, Button, Tag, Alert, Input } from "antd";
import {
  SendOutlined,
  CheckCircleFilled,
  CloseCircleFilled,
  LoadingOutlined,
} from "@ant-design/icons";
import type {
  FeatureNode,
  GeneratedApiKeyInfo,
  ProjectInfo,
  TestLogResult,
} from "@/features/product-setup/types/productSetup.types";
import { productSetupService } from "@/features/product-setup/services/productSetup.service";
import { getSetupErrorMessage } from "@/features/product-setup/utils/setupError";
import { getOmniLogsBaseUrl } from "@/utils/omnilogs-url";

interface ConnectionGuideStepProps {
  productId?: number;
  environmentCode: string;
  projects: ProjectInfo[];
  features: FeatureNode[];
  apiKey: GeneratedApiKeyInfo | null;
  testLogResult: TestLogResult | null;
  apiKeySecret: string;
  onApiKeySecretChange: (secret: string) => void;
  onTestResult: (result: TestLogResult) => void;
}

export function ConnectionGuideStep({
  productId,
  environmentCode,
  projects,
  features,
  apiKey,
  testLogResult,
  apiKeySecret,
  onApiKeySecretChange,
  onTestResult,
}: ConnectionGuideStepProps) {
  const [isSending, setIsSending] = useState(false);
  const hasBoundIngestionKey = Boolean(apiKey?.key_id && apiKey.raw_key);

  const handleSendTestLog = async () => {
    if (!hasBoundIngestionKey) {
      onTestResult({
        status: "failed",
        error_message:
          "การเชื่อมต่อล้มเหลว: กรุณาสร้าง API Key ที่ผูกกับ Environment ในขั้นตอนที่ 5 ก่อนส่ง Test log",
      });
      return;
    }

    const testSecret = apiKey?.raw_key ?? "";

    setIsSending(true);
    onTestResult({ status: "sending" });

    try {
      const targetFeature =
        features.find((feature) => feature.category_id && feature.parent_id) ??
        features.find((feature) => feature.category_id);
      const targetProject = projects.find(
        (project) =>
          project.project_id &&
          String(project.project_id) === String(targetFeature?.project_id),
      );
      if (!targetProject?.project_id || !targetFeature?.category_id) {
        throw new Error(
          "ไม่พบ Project และ Feature hierarchy ที่บันทึกไว้ กรุณากลับไปขั้นตอนที่ 3 และบันทึก Hierarchy อีกครั้ง",
        );
      }
      const now = new Date().toISOString();
      const payload = {
        timestamp: now,
        level: "INFO" as const,
        message: "OmniLogs connection test from Product Setup Wizard",
        project_id: targetProject.project_id,
        category_id: targetFeature.category_id,
        custom_fields: {
          product_id: productId,
          environment_code: environmentCode,
          project_id: targetProject.project_id,
          project_code: targetProject.project_code,
          category_id: targetFeature.category_id,
          category_code: targetFeature.category_code,
          feature_code: targetFeature.category_code,
        },
        service: "setup-test-runner",
        metadata: {
          setup_test: true,
          product_id: productId,
          environment_code: environmentCode,
        },
      };

      const res = await productSetupService.sendTestLog(testSecret, payload);
      const batchId = res.data.batch_id;

      onTestResult({
        status: "received",
        timestamp: now,
        log_id: batchId,
        details: {
          product_id: productId,
          environment_code: environmentCode,
          project_id: targetProject.project_id,
          category_id: targetFeature.category_id,
                message: "OmniLogs connection test from Product Setup Wizard",
          level: "INFO",
        },
      });
    } catch (error: unknown) {
      onTestResult({
        status: "failed",
        error_message: getSetupErrorMessage(
          error,
          "คำขอรับ Log ล้มเหลว",
        ),
      });
    } finally {
      setIsSending(false);
    }
  };

  const baseUrl = getOmniLogsBaseUrl();

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <div>
        <h2 className="text-xl font-bold text-zinc-900">
          ขั้นตอนที่ 6: เชื่อมต่อ Log และส่ง Test log
        </h2>
        <p className="text-sm text-zinc-500 mt-1">
          ตรวจสอบการเชื่อมต่อด้วยการส่ง Test log จริงโดยใช้
          API Key ที่สร้างไว้
        </p>
      </div>
      <Alert
        className="mt-4"
        type="info"
        showIcon
        message="การทดสอบการเชื่อมต่อต้องใช้ Source ที่พร้อมใช้งาน"
        description="หาก Source ยังไม่เชื่อมต่อหรือไม่มี Secret ของ API Key การทดสอบนี้จะแสดงว่าการเชื่อมต่อล้มเหลว"
      />{" "}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Credentials & Setup Info */}
        <Card
          title="การตั้งค่าการรับ Log"
          className="border-zinc-200 shadow-sm rounded-lg"
        >
          <div className="space-y-3 text-sm">
            <div>
              {!apiKey?.raw_key && (
                <div>
                  <span className="text-zinc-500 text-xs block">
                    Secret ของ API Key:
                  </span>
                  <Input.Password
                    value={apiKeySecret}
                    onChange={(event) =>
                      onApiKeySecretChange(event.target.value)
                    }
                    placeholder="วาง Secret จากหน้าสร้าง API Key"
                    className="mt-1"
                  />
                </div>
              )}{" "}
              <span className="text-zinc-500 text-xs block">Product ID:</span>
              <Tag color="default" className="font-mono font-bold">
                {productId ? `#${productId}` : "ระบบจะกำหนดหลังจากสร้าง Product"}
              </Tag>
            </div>
            <div>
              <span className="text-zinc-500 text-xs block">
                รหัส Environment:
              </span>
              <Tag color="blue" className="font-mono font-bold">
                {environmentCode}
              </Tag>
            </div>
            <div>
              <span className="text-zinc-500 text-xs block">
                ปลายทางรับ Log:
              </span>
              <code className="font-mono text-xs bg-zinc-100 p-1 rounded text-zinc-900 block mt-0.5 truncate">
                {baseUrl}/api/v1/ingest/logs
              </code>
            </div>
            <div>
              <span className="text-zinc-500 text-xs block">
                Authorization Header:
              </span>
              <code className="font-mono text-xs bg-zinc-100 p-1 rounded text-zinc-900 block mt-0.5 truncate">
                Bearer{" "}
                {apiKey?.key_prefix
                  ? `${apiKey.key_prefix}...`
                  : "YOUR_API_KEY"}
              </code>
            </div>
          </div>
        </Card>

        {/* Live Test Action Card */}
        <Card
          title="ทดสอบการรับ Log"
          className="border-zinc-200 shadow-sm rounded-lg flex flex-col justify-between"
        >
          <div className="space-y-4">
            <p className="text-xs text-zinc-600">
              กด <strong>ส่ง Test log</strong> เพื่อส่งคำขอ HTTP POST จริง
              ไปยังบริการรับ Log ของ Backend
            </p>

            <Button
              type="primary"
              size="large"
              block
              icon={isSending ? <LoadingOutlined /> : <SendOutlined />}
              onClick={handleSendTestLog}
              loading={isSending}
              className="bg-zinc-900 hover:bg-zinc-800 font-medium"
            >
              {isSending ? "กำลังส่ง Test log..." : "ส่ง Test log"}
            </Button>
          </div>

          {/* Test Log Status Indicator */}
          <div className="mt-4 pt-4 border-t border-zinc-200">
            {testLogResult?.status === "received" && (
              <Alert
                type="success"
                showIcon
                icon={<CheckCircleFilled className="text-emerald-500" />}
                message="เชื่อมต่อสำเร็จ"
                description={
                  <div className="mt-2 space-y-1 text-xs text-zinc-700">
                    <p className="font-medium text-emerald-800">
                      OmniLogs ได้รับ Log รายการแรกของคุณแล้ว
                    </p>
                    <div className="bg-white p-2 rounded border border-emerald-200 font-mono text-[11px] space-y-0.5">
                      <div>
                        <strong className="text-zinc-500">Log ID:</strong>{" "}
                        {testLogResult.log_id}
                      </div>
                      <div>
                        <strong className="text-zinc-500">ระดับ:</strong>{" "}
                        <span className="text-blue-600 font-bold">INFO</span>
                      </div>
                      <div>
                        <strong className="text-zinc-500">ข้อความ:</strong>{" "}
                        {String(testLogResult.details?.message ?? "—")}
                      </div>
                      <div>
                        <strong className="text-zinc-500">เวลา:</strong>{" "}
                        {testLogResult.timestamp}
                      </div>
                    </div>
                  </div>
                }
              />
            )}

            {testLogResult?.status === "failed" && (
              <Alert
                type="error"
                showIcon
                icon={<CloseCircleFilled className="text-red-500" />}
                message="ทดสอบการรับ Log Failed"
                description={
                  <div className="mt-1 text-xs text-red-800">
                    <p>
                      {testLogResult.error_message ||
                        "ไม่สามารถเชื่อมต่อบริการรับ Log ได้"}
                    </p>
                  </div>
                }
              />
            )}

            {(!testLogResult || testLogResult.status === "waiting") && (
              <div className="text-center py-3 text-xs text-zinc-400 font-mono border border-dashed border-zinc-200 rounded">
                สถานะ: รอการดำเนินการส่ง Test log
              </div>
            )}
          </div>
        </Card>
      </div>
    </div>
  );
}

