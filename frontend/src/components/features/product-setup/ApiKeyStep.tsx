import { useMemo, useState } from "react";
import {
  Alert,
  Button,
  Card,
  Form,
  Input,
  Popconfirm,
  Select,
  Tag,
  message,
} from "antd";
import {
  CheckOutlined,
  CopyOutlined,
  DeleteOutlined,
  KeyOutlined,
  SafetyCertificateOutlined,
  WarningOutlined,
} from "@ant-design/icons";
import type {
  EnvironmentInfo,
  ExistingApiKeyInfo,
  FeatureNode,
  GeneratedApiKeyInfo,
  ProjectInfo,
} from "@/features/product-setup/types/productSetup.types";
import { getSetupErrorMessage } from "@/features/product-setup/utils/setupError";
import { getOmniLogsBaseUrl } from "@/utils/omnilogs-url";

interface ApiKeyFormValues {
  key_name: string;
  environment_code: string;
  expiration: string;
  description?: string;
  default_project_id?: number;
  default_category_id?: number;
}

interface ApiKeyStepProps {
  productName: string;
  productCode: string;
  generatedKey: GeneratedApiKeyInfo | null;
  existingApiKeys: ExistingApiKeyInfo[];
  onDeleteExistingApiKey: (keyId: number) => Promise<void>;
  projects: ProjectInfo[];
  features: FeatureNode[];
  environments: EnvironmentInfo[];
  onGenerate: (values: ApiKeyFormValues) => Promise<void>;
  isGenerating?: boolean;
}

export function ApiKeyStep({
  productName,
  productCode,
  generatedKey,
  existingApiKeys,
  onDeleteExistingApiKey,
  onGenerate,
  projects,
  features,
  environments,
  isGenerating = false,
}: ApiKeyStepProps) {
  const [form] = Form.useForm<ApiKeyFormValues>();
  const [copied, setCopied] = useState<
    "key" | "env" | "curl" | "connection" | null
  >(null);
  const defaultProjectId = Form.useWatch("default_project_id", form);
  const defaultEnvironment =
    environments.find((environment) => environment.is_default) ??
    environments[0];

  const availableFeatureOptions = useMemo(() => {
    const available = features.filter(
      (feature) =>
        Number(feature.project_id) === Number(defaultProjectId) &&
        Boolean(feature.category_id),
    );
    const roots = available.filter(
      (feature) =>
        feature.parent_id === null || feature.parent_id === undefined,
    );

    return roots.map((root) => {
      const children = available.filter(
        (feature) =>
          String(feature.parent_id) === String(root.category_id ?? root.id),
      );
      return {
        label: `${root.category_name} (#${root.category_id})`,
        options: [
          {
            value: root.category_id!,
            label: `${root.category_name} (Root feature)`,
          },
          ...children
            .filter((child): child is FeatureNode & { category_id: number } =>
              Boolean(child.category_id),
            )
            .map((child) => ({
              value: child.category_id,
              label: `${child.category_name} (#${child.category_id})`,
            })),
        ],
      };
    });
  }, [defaultProjectId, features]);

  const copyValue = async (
    value: string,
    type: "key" | "env" | "curl" | "connection",
  ) => {
    await navigator.clipboard.writeText(value);
    setCopied(type);
    window.setTimeout(() => setCopied(null), 2000);
    message.success("คัดลอกไปยังคลิปบอร์ดแล้ว");
  };

  const handleSubmit = async (values: ApiKeyFormValues) => {
    try {
      await onGenerate(values);
      message.success("สร้าง API Key สำเร็จ");
    } catch (error: unknown) {
      message.error(getSetupErrorMessage(error, "สร้าง API Key ไม่สำเร็จ"));
    }
  };

  const baseUrl = getOmniLogsBaseUrl();
  const connectionSlug =
    productCode
      .trim()
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/^-|-$/g, "") || "product";
  const queueKey = `${connectionSlug}-web-logs`;
  const serviceName = `${connectionSlug}-web`;
  const envSnippet = `OMNILOGS_BASE_URL=${baseUrl}
OMNILOGS_API_KEY=${generatedKey?.raw_key || "YOUR_API_KEY"}
OMNILOGS_PRODUCT_CODE=${productCode}
OMNILOGS_ENVIRONMENT_CODE=${generatedKey?.environment_code || defaultEnvironment?.environment_code || "YOUR_ENVIRONMENT_CODE"}
OMNILOGS_QUEUE_KEY=${queueKey}
OMNILOGS_SERVICE=${serviceName}
# OmniLogs จะ map Project / Feature จาก service, request_path และ request_method
# ไม่ต้องกำหนด OMNILOGS_ROUTE_KEY หรือ OMNILOGS_EVENT_NAME เพิ่มเอง`;

  const curlSnippet = `curl --request POST \\  --url ${baseUrl}/api/v1/ingest/logs \\  --header "X-API-Key: ${generatedKey?.raw_key || "YOUR_API_KEY"}" \\  --header "Content-Type: application/json" \\  --data '{
    "queue_key": "${queueKey}",
    "source_type": "SERVICE",
    "source_platform": "HTTP",
    "logs": [{
      "sequence_no": 1,
      "source_type": "SERVICE",
      "source_platform": "HTTP",
      "data": {
        "timestamp": "${new Date().toISOString()}",
        "level": "INFO",
        "service": "${serviceName}",
        "request_path": "/health",
        "request_method": "GET",
        "message": "OmniLogs connection test"
      }
    }]
  }'`;

  return (
    <div className="max-w-3xl mx-auto space-y-6">
      <div>
        <h2 className="text-xl font-bold text-zinc-900">
          ขั้นตอนที่ 5: สร้าง API Key
        </h2>
        <p className="text-sm text-zinc-500 mt-1">
          สร้าง Ingestion API Key เพื่อส่ง Log สำหรับ{" "}
          <strong className="text-zinc-900">{productName}</strong> โดย
          Development (DEV) จะถูกเลือกเป็นค่าเริ่มต้น
        </p>
      </div>

      {!generatedKey ? (
        <>
          {existingApiKeys.length > 0 && (
            <Alert
              type="info"
              showIcon
              className="mb-4"
              message={`พบ API Key ที่มีอยู่ ${existingApiKeys.length} รายการ`}
              description="Secret ของ Key เดิมจะถูกซ่อนและกู้คืนไม่ได้ คุณสามารถใช้ข้อมูล Key แบบปิดบังด้านล่าง หรือสร้าง Key ใหม่สำหรับการเชื่อมต่อแหล่งข้อมูลใหม่"
            />
          )}

          {existingApiKeys.length > 0 && (
            <Card className="mb-6 border-zinc-200 shadow-sm rounded-lg">
              <div className="space-y-3">
                <h3 className="font-semibold text-zinc-900">
                  API Key ที่มีอยู่
                </h3>
                {existingApiKeys.map((key) => (
                  <div
                    key={key.key_id}
                    className="flex items-center justify-between gap-3 rounded-md border border-zinc-200 bg-zinc-50 px-3 py-2"
                  >
                    <div className="min-w-0">
                      <div className="font-medium text-zinc-800 truncate">
                        {key.key_name}
                      </div>
                      <code className="text-xs text-zinc-500">
                        {key.key_prefix}••••••••
                      </code>
                    </div>
                    <Tag color={key.is_active ? "success" : "default"}>
                      {key.is_active ? "ใช้งานอยู่" : "ไม่ได้ใช้งาน"}
                    </Tag>
                    <Popconfirm
                      title="ลบ API Key นี้หรือไม่?"
                      description="Key นี้จะถูกเพิกถอนและไม่สามารถใช้งานได้อีก"
                      okText="ลบ"
                      cancelText="ยกเลิก"
                      okButtonProps={{ danger: true }}
                      onConfirm={() => onDeleteExistingApiKey(key.key_id)}
                    >
                      <Button
                        type="text"
                        danger
                        size="small"
                        icon={<DeleteOutlined />}
                        aria-label={`ลบ ${key.key_name}`}
                      >
                        ลบ
                      </Button>
                    </Popconfirm>
                  </div>
                ))}
              </div>
            </Card>
          )}

          <Card className="border-zinc-200 shadow-sm rounded-lg">
            <Form
              form={form}
              layout="vertical"
              initialValues={{
                key_name: `${productName} Key`,
                environment_code: defaultEnvironment?.environment_code,
                expiration: "365d",
                description: "",
              }}
              onFinish={handleSubmit}
              requiredMark={false}
            >
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <Form.Item
                  name="key_name"
                  label={
                    <span className="font-medium text-zinc-700">
                      ชื่อ API Key *
                    </span>
                  }
                  rules={[{ required: true, message: "กรุณากรอกชื่อ API Key" }]}
                >
                  <Input placeholder="เช่น Development Key" size="large" />
                </Form.Item>

                <Form.Item
                  name="environment_code"
                  label={
                    <span className="font-medium text-zinc-700">
                      Environment *
                    </span>
                  }
                  rules={[
                    { required: true, message: "กรุณาเลือก Environment" },
                  ]}
                >
                  <Select
                    size="large"
                    options={environments.map((environment) => ({
                      value: environment.environment_code,
                      label: `${environment.environment_code} - ${environment.environment_name}`,
                    }))}
                  />
                </Form.Item>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <Form.Item
                  name="default_project_id"
                  label={
                    <span className="font-medium text-zinc-700">
                      Project เริ่มต้น
                    </span>
                  }
                  extra="[ ไม่บังคับ ] ใช้จัดกลุ่ม Log ทุกรายการที่ส่งด้วย Key นี้"
                >
                  <Select
                    allowClear
                    size="large"
                    placeholder="Log ระดับ Product"
                    onChange={() =>
                      form.setFieldValue("default_category_id", undefined)
                    }
                    options={projects
                      .filter((project) => project.project_id)
                      .map((project) => ({
                        value: project.project_id,
                        label: `${project.project_name} (Project ID: ${project.project_id})`,
                      }))}
                  />
                </Form.Item>

                <Form.Item
                  name="default_category_id"
                  label={
                    <span className="font-medium text-zinc-700">
                      Feature เริ่มต้น
                    </span>
                  }
                  extra="[ ไม่บังคับ ] เว้นว่างเพื่อจัดกลุ่มในระดับ Project"
                >
                  <Select
                    allowClear
                    disabled={!defaultProjectId}
                    size="large"
                    placeholder={
                      defaultProjectId
                        ? "เลือก Feature"
                        : "กรุณาเลือก Project ก่อน"
                    }
                    options={availableFeatureOptions}
                  />
                </Form.Item>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <Form.Item
                  name="expiration"
                  label={
                    <span className="font-medium text-zinc-700">
                      วันหมดอายุของ API Key
                    </span>
                  }
                >
                  <Select size="large">
                    <Select.Option value="30d">30 วัน</Select.Option>
                    <Select.Option value="90d">90 วัน</Select.Option>
                    <Select.Option value="365d">
                      1 ปี (ค่าเริ่มต้น)
                    </Select.Option>
                    <Select.Option value="never">ไม่มีวันหมดอายุ</Select.Option>
                  </Select>
                </Form.Item>

                <div>
                  <label className="font-medium text-zinc-700 block mb-2">
                    Scope / Permission เริ่มต้น
                  </label>
                  <div className="p-2.5 bg-zinc-50 border border-zinc-200 rounded-md">
                    <Tag
                      color="default"
                      className="font-mono text-xs bg-[#F6F6F6] text-[#1F8457] border-[#D9CAB3]"
                    >
                      LOG_INGEST_CREATE
                    </Tag>
                    <span className="text-xs text-zinc-500 block mt-1">
                      รับเฉพาะ Application และ Error Log เท่านั้น
                      ไม่มีสิทธิ์เข้าถึงการตั้งค่า Platform หรือการจัดการ User
                    </span>
                  </div>
                </div>
              </div>

              <Form.Item
                name="description"
                label={
                  <span className="font-medium text-zinc-700">คำอธิบาย</span>
                }
              >
                <Input.TextArea
                  placeholder="[ ไม่บังคับ ] คำอธิบายของ API Key"
                  rows={2}
                />
              </Form.Item>

              <Button
                type="primary"
                htmlType="submit"
                icon={<KeyOutlined />}
                loading={isGenerating}
                size="large"
                block
                className="bg-zinc-900 hover:bg-zinc-800 font-medium"
              >
                สร้าง API Key
              </Button>
            </Form>
          </Card>
        </>
      ) : (
        <>
          <Form form={form} style={{ display: "none" }} />
          <div className="space-y-6">
            <Alert
              type="warning"
              showIcon
              icon={<WarningOutlined />}
              message="จัดเก็บ API Key อย่างปลอดภัย"
              description="Secret API Key นี้จะแสดงเพียงครั้งเดียวเท่านั้น กรุณาคัดลอกและจัดเก็บไว้ในไฟล์ Environment ของแอปพลิเคชัน (.env) อย่างปลอดภัย"
            />

            <Card
              title="ค่าการเชื่อมต่อ"
              className="border-zinc-200 shadow-sm rounded-lg"
            >
              <p className="text-xs text-zinc-500 mb-3">
                ใช้ค่าเหล่านี้ใน testproduct หรือแอปพลิเคชัน
                เพื่อให้การจัดเส้นทางและ Filter ใน Log Explorer สอดคล้องกัน
              </p>
              <pre className="p-3 bg-zinc-900 text-zinc-100 font-mono text-xs rounded-md overflow-x-auto">{`QUEUE_KEY=${queueKey}
SERVICE=${serviceName}
ROUTING_FIELDS=service,request_path,request_method`}</pre>
              <Button
                type="link"
                size="small"
                icon={
                  copied === "connection" ? <CheckOutlined /> : <CopyOutlined />
                }
                onClick={() =>
                  void copyValue(
                    `QUEUE_KEY=${queueKey}
SERVICE=${serviceName}
ROUTING_FIELDS=service,request_path,request_method`,
                    "connection",
                  )
                }
              >
                {copied === "connection"
                  ? "คัดลอกแล้ว"
                  : "คัดลอกค่าการเชื่อมต่อ"}
              </Button>
            </Card>

            <Card className="border-2 border-zinc-900 shadow-md rounded-lg p-2">
              <div className="space-y-4">
                <div className="flex items-center justify-between border-b border-zinc-200 pb-3">
                  <div>
                    <h3 className="font-bold text-zinc-900 text-base">
                      {generatedKey.key_name}
                    </h3>
                    <div className="flex items-center gap-2 mt-1">
                      <Tag color="blue" className="font-mono">
                        {generatedKey.environment_code}
                      </Tag>
                      <Tag color="default" className="font-mono">
                        {productCode}
                      </Tag>
                    </div>
                  </div>
                  <SafetyCertificateOutlined className="text-2xl text-emerald-600" />
                </div>

                <div>
                  <label className="text-xs font-semibold text-zinc-700 block mb-1">
                    Secret API Key ที่สร้าง:
                  </label>
                  <div className="flex items-center gap-2">
                    <Input.Password
                      value={generatedKey.raw_key}
                      readOnly
                      size="large"
                      className="font-mono text-sm bg-zinc-50"
                    />
                    <Button
                      type="primary"
                      size="large"
                      icon={
                        copied === "key" ? <CheckOutlined /> : <CopyOutlined />
                      }
                      onClick={() =>
                        void copyValue(generatedKey.raw_key, "key")
                      }
                      className={
                        copied === "key"
                          ? "bg-emerald-600 hover:bg-emerald-500"
                          : "bg-zinc-900 hover:bg-zinc-800"
                      }
                    >
                      {copied === "key" ? "คัดลอกแล้ว" : "คัดลอก Key"}
                    </Button>
                  </div>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div>
                    <div className="flex items-center justify-between mb-1">
                      <label className="text-xs font-semibold text-zinc-700">
                        .env Snippet:
                      </label>
                      <Button
                        type="link"
                        size="small"
                        icon={<CopyOutlined />}
                        onClick={() => void copyValue(envSnippet, "env")}
                      >
                        {copied === "env" ? "คัดลอกแล้ว" : "คัดลอก .env"}
                      </Button>
                    </div>
                    <pre className="p-3 bg-zinc-900 text-zinc-100 font-mono text-xs rounded-md overflow-x-auto">
                      {envSnippet}
                    </pre>
                  </div>

                  <div>
                    <div className="flex items-center justify-between mb-1">
                      <label className="text-xs font-semibold text-zinc-700">
                        cURL Test Snippet:
                      </label>
                      <Button
                        type="link"
                        size="small"
                        icon={<CopyOutlined />}
                        onClick={() => void copyValue(curlSnippet, "curl")}
                      >
                        {copied === "curl" ? "คัดลอกแล้ว" : "คัดลอก cURL"}
                      </Button>
                    </div>
                    <pre className="p-3 bg-zinc-900 text-zinc-100 font-mono text-xs rounded-md overflow-x-auto max-h-32">
                      {curlSnippet}
                    </pre>
                  </div>
                </div>
              </div>
            </Card>
          </div>
        </>
      )}
    </div>
  );
}
