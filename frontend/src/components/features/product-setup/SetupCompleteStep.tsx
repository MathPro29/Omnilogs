import { useNavigate } from "react-router-dom";
import { Card, Button, Tag, Result } from "antd";
import {
  CheckCircleFilled,
  RocketOutlined,
  ArrowRightOutlined,
} from "@ant-design/icons";
import { ROUTES } from "@/constants";
import type {
  ProductSetupDraft,
  ProjectInfo,
  FeatureNode,
} from "@/features/product-setup/types/productSetup.types";

interface SetupCompleteStepProps {
  productId?: number;
  draft: ProductSetupDraft;
  projects: ProjectInfo[];
  features: FeatureNode[];
  onComplete: () => Promise<void>;
}

export function SetupCompleteStep({
  productId,
  draft,
  projects,
  features,
  onComplete,
}: SetupCompleteStepProps) {
  const navigate = useNavigate();

  const getEnvTagColor = (codeOrType?: string) => {
    const code = (codeOrType || "").toUpperCase();
    switch (code) {
      case "DEV":
      case "DEVELOPMENT":
        return "blue";
      case "UAT":
        return "orange";
      case "PROD":
      case "PRODUCTION":
        return "green";
      case "STAGING":
        return "purple";
      case "QA":
      case "TEST":
        return "cyan";
      default:
        return "default";
    }
  };

  const env_tag = draft.environments?.map((env, index) => {
    const code = env.environment_code || env.environment_type;
    return (
      <Tag
        color={getEnvTagColor(code)}
        key={env.environment_id || code || index}
        className="font-mono"
      >
        {env.environment_name || code}
      </Tag>
    );
  });

  const handleOpenLogExplorer = async () => {
    await onComplete();
    navigate(
      `${ROUTES.LOGS_EXPLORE}?product=${productId ?? ""}&environment=DEV`,
    );
  };

  return (
    <div className="max-w-3xl mx-auto space-y-6 py-4">
      <Card className="border-zinc-200 shadow-md rounded-xl text-center p-6">
        <Result
          icon={<CheckCircleFilled className="text-6xl text-emerald-500" />}
          title={
            <h2 className="text-2xl font-extrabold text-zinc-900">
              ตั้งค่า Product เสร็จสมบูรณ์
            </h2>
          }
          subTitle={
            <p className="text-sm text-zinc-600 max-w-md mx-auto mt-1">
              Product ของคุณ{" "}
              <strong className="text-zinc-900">{draft.product_name}</strong>{" "}
              ได้รับการตั้งค่าครบแล้วและพร้อมรับ Log
            </p>
          }
        />

        {/* Success Summary Table */}
        <div className="max-w-xl mx-auto bg-zinc-50 border border-zinc-200 rounded-lg p-4 my-6 text-left space-y-2 text-sm">
          <div className="flex items-center justify-between border-b border-zinc-200/80 pb-2">
            <span className="text-zinc-500 font-medium">Product:</span>
            <span className="font-bold text-zinc-900">
              {draft.product_name}{" "}
            </span>
          </div>

          <div className="flex items-center justify-between border-b border-zinc-200/80 pb-2">
            <span className="text-zinc-500 font-medium">
              Projects ที่ตั้งค่าแล้ว:
            </span>
            <span className="font-semibold text-zinc-900">
              {projects.length} Projects
            </span>
          </div>

          <div className="flex items-center justify-between border-b border-zinc-200/80 pb-2">
            <span className="text-zi nc-500 font-medium">
              Features และ Sub-features:
            </span>
            <span className="font-semibold text-zinc-900">
              {features.length} รายการ
            </span>
          </div>

          <div className="flex items-center justify-between border-b border-zinc-200/80 pb-2">
            <span className="text-zinc-500 font-medium">Environments:</span>
            <div className="flex gap-1 flex-wrap">{env_tag}</div>
          </div>

          <div className="flex items-center justify-between border-b border-zinc-200/80 pb-2">
            <span className="text-zinc-500 font-medium">
              สถานะการเชื่อมต่อ:
            </span>
            <span className="inline-flex items-center gap-1 font-semibold text-emerald-600">
              <CheckCircleFilled className="text-xs" /> ใช้งานอยู่
            </span>
          </div>

          <div className="flex items-center justify-between pt-1">
            <span className="text-zinc-500 font-medium">ตรวจสอบ Log แรก:</span>
            <span className="font-semibold text-emerald-700">
              ได้รับและจัดทำดัชนีแล้ว
            </span>
          </div>
        </div>

        {/* Primary Action Button */}
        <div className="space-y-3">
          <Button
            type="primary"
            size="large"
            icon={<RocketOutlined />}
            onClick={handleOpenLogExplorer}
            className="w-full sm:w-auto px-8 h-12 bg-zinc-900 hover:bg-zinc-800 font-semibold text-base rounded-lg shadow-sm"
          >
            เปิด Log Explorer <ArrowRightOutlined />
          </Button>
        </div>
      </Card>
    </div>
  );
}
