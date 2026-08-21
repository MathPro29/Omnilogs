import { Card, Tag, Button, Alert, Statistic } from "antd";
import {
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  EditOutlined,
  FolderOutlined,
  AppstoreOutlined,
} from "@ant-design/icons";
import type {
  FeatureNode,
  ProductSetupDraft,
  ProjectInfo,
} from "@/features/product-setup/types/productSetup.types";

interface ProductReviewStepProps {
  draft: ProductSetupDraft;
  projects: ProjectInfo[];
  features: FeatureNode[];
  validation: {
    valid: boolean;
    errors: string[];
  };
  onJumpToStep: (stepIndex: number) => void;
}

export function ProductReviewStep({
  draft,
  projects,
  features,
  validation,
  onJumpToStep,
}: ProductReviewStepProps) {
  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <div>
        <h2 className="text-xl font-bold text-zinc-900">
          ขั้นตอนที่ 4: ตรวจสอบโครงสร้าง Product
        </h2>
        <p className="text-sm text-zinc-500 mt-1">
          ตรวจสอบ Product, Projects, Environments และ Feature hierarchy
          ก่อนสร้าง API Key
        </p>
      </div>

      {validation.valid ? (
        <Alert
          type="success"
          showIcon
          icon={<CheckCircleOutlined />}
          message="ตรวจสอบโครงสร้างผ่าน"
          description="โครงสร้าง Product ถูกต้องและพร้อมสำหรับการสร้าง API Key"
        />
      ) : (
        <Alert
          type="warning"
          showIcon
          icon={<ExclamationCircleOutlined />}
          message="พบประเด็นที่ต้องแก้ไข"
          description={
            <ul className="list-disc list-inside text-xs space-y-1 mt-1">
              {validation.errors.map((err, idx) => (
                <li key={idx} className="text-amber-800">
                  {err}
                </li>
              ))}
            </ul>
          }
        />
      )}

      {/* Summary Cards Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <Card className="border-zinc-200 shadow-sm text-center">
          <Statistic
            title="Environments"
            value={draft.environments.length}
            valueStyle={{
              fontSize: "16px",
              fontWeight: "bold",
              color: "#10b981",
            }}
          />
        </Card>
        <Card className="border-zinc-200 shadow-sm text-center">
          <Statistic title="Projects" value={projects.length} />
        </Card>
        <Card className="border-zinc-200 shadow-sm text-center">
          <Statistic title="Features ทั้งหมด" value={features.length} />
        </Card>
      </div>

      {/* Product และ Environments Card */}
      <Card
        title={
          <div className="flex items-center justify-between">
            <span className="font-bold text-zinc-900">
              Product และ Environments
            </span>
            <Button
              size="small"
              icon={<EditOutlined />}
              onClick={() => onJumpToStep(0)}
            >
              แก้ไข Product
            </Button>
          </div>
        }
        className="border-zinc-200 shadow-sm rounded-lg"
      >
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-zinc-500 block text-xs">ชื่อ Product:</span>
            <span className="font-semibold text-zinc-900">
              {draft.product_name}
            </span>
          </div>
          <div>
            <span className="text-zinc-500 block text-xs">Description:</span>
            <span className="text-zinc-700">
              {draft.description || "ไม่ได้ระบุคำอธิบาย"}
            </span>
          </div>
          <div>
            <span className="text-zinc-500 block text-xs">Environments:</span>
            <div className="flex flex-wrap items-center gap-2 mt-1">
              {draft.environments.map((environment) => (
                <Tag
                  key={environment.environment_code}
                  color={
                    environment.environment_code.toUpperCase() === "PROD"
                      ? "green"
                      : "blue"
                  }
                  className="font-mono"
                >
                  {environment.environment_code} ({environment.environment_name})
                </Tag>
              ))}
            </div>
          </div>
        </div>
      </Card>

      {/* Projects & Hierarchy Tree Preview */}
      <Card
        title={
          <div className="flex items-center justify-between">
            <span className="font-bold text-zinc-900">
              ตัวอย่าง Project และ Feature hierarchy
            </span>
            <div className="flex items-center gap-2">
              <Button
                size="small"
                icon={<EditOutlined />}
                onClick={() => onJumpToStep(1)}
              >
                แก้ไข Projects
              </Button>
              <Button
                size="small"
                icon={<EditOutlined />}
                onClick={() => onJumpToStep(2)}
              >
                แก้ไข Features
              </Button>
            </div>
          </div>
        }
        className="border-zinc-200 shadow-sm rounded-lg"
      >
        <div className="space-y-4">
          <div className="font-mono text-xs font-bold text-zinc-900 flex items-center gap-2 p-2 bg-zinc-100 rounded">
            <FolderOutlined className="text-zinc-700" />
            <span>{draft.product_name || "Product"}</span>
          </div>

          {projects.map((proj) => {
            const projFeatures = features.filter(
              (f) =>
                String(f.project_id) === String(proj.project_id ?? proj.id),
            );

            return (
              <div
                key={proj.project_id || proj.id}
                className="pl-4 border-l-2 border-zinc-200 space-y-2"
              >
                <div className="flex items-center gap-2">
                  <span className="text-zinc-400 font-mono">├──</span>
                  <AppstoreOutlined className="text-zinc-800" />
                  <span className="font-bold text-zinc-900">
                    {proj.project_name}
                  </span>
                </div>

                {projFeatures.length > 0 ? (
                  <div className="pl-6 space-y-1">
                    {projFeatures
                      .filter(
                        (f) =>
                          f.parent_id === null || f.parent_id === undefined,
                      )
                      .map((rootFeat) => {
                        const childNodes = projFeatures.filter(
                          (child) =>
                            String(child.parent_id) ===
                            String(rootFeat.category_id ?? rootFeat.id),
                        );

                        return (
                          <div
                            key={rootFeat.category_id || rootFeat.id}
                            className="space-y-1"
                          >
                            <div className="flex items-center gap-2 text-sm">
                              <span className="text-zinc-400 font-mono">
                                ├──
                              </span>
                              <span className="font-medium text-zinc-800">
                                {rootFeat.category_name}
                              </span>
                            </div>

                            {childNodes.length > 0 && (
                              <div className="pl-6 space-y-1">
                                {childNodes.map((c) => (
                                  <div
                                    key={c.category_id || c.id}
                                    className="flex items-center gap-2 text-xs text-zinc-600"
                                  >
                                    <span className="text-zinc-300 font-mono">
                                      └──
                                    </span>
                                    <span>{c.category_name}</span>
                                  </div>
                                ))}
                              </div>
                            )}
                          </div>
                        );
                      })}
                  </div>
                ) : (
                  <div className="pl-6 text-xs text-amber-600 italic">
                    ยังไม่มี Feature ใน Project นี้
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </Card>
    </div>
  );
}

