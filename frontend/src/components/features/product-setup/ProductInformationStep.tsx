import { useQuery } from "@tanstack/react-query";
import { Form, Input, Switch, Select, Alert, Card, Button } from "antd";
import {
  CheckCircleOutlined,
  InfoCircleOutlined,
  UserOutlined,
  PlusOutlined,
  CheckOutlined,
} from "@ant-design/icons";
import { userService } from "@/services/user.service";
import type { User } from "@/types";
import type { EnvironmentInfo } from "@/features/product-setup/types/productSetup.types";

const PRESET_ENVIRONMENTS: EnvironmentInfo[] = [
  {
    environment_code: "DEV",
    environment_name: "Development",
    environment_type: "DEVELOPMENT",
    is_default: false,
    is_active: true,
  },
  {
    environment_code: "PROD",
    environment_name: "Production",
    environment_type: "PRODUCTION",
    is_default: false,
    is_active: true,
  },
  {
    environment_code: "UAT",
    environment_name: "UAT",
    environment_type: "UAT",
    is_default: false,
    is_active: true,
  },
  {
    environment_code: "STAGING",
    environment_name: "Staging",
    environment_type: "STAGING",
    is_default: false,
    is_active: true,
  },
  {
    environment_code: "QA",
    environment_name: "QA",
    environment_type: "QA",
    is_default: false,
    is_active: true,
  },
  {
    environment_code: "TEST",
    environment_name: "Testing",
    environment_type: "TEST",
    is_default: false,
    is_active: true,
  },
];

interface ProductInformationValues {
  product_name: string;
  product_code: string;
  description: string;
  owner_id?: number;
  is_active: boolean;
  environments: EnvironmentInfo[];
}

interface ProductInformationStepProps {
  initialValues: {
    product_name: string;
    product_code: string;
    description: string;
    owner_id?: number;
    is_active: boolean;
    environments: EnvironmentInfo[];
  };
  onChange: (values: {
    product_name: string;
    product_code: string;
    description: string;
    owner_id?: number;
    is_active: boolean;
    environments: EnvironmentInfo[];
  }) => void;
  isCreated?: boolean;
  productNameError?: string;
  onProductNameBlur?: () => void;
  navigation?: React.ReactNode;
}

export function ProductInformationStep({
  initialValues,
  onChange,
  isCreated = false,
  productNameError,
  onProductNameBlur,
  navigation,
}: ProductInformationStepProps) {
  const [form] = Form.useForm<ProductInformationValues>();
  const watchedEnvs = Form.useWatch("environments", form) || [];
  const selectedCodes = (watchedEnvs as EnvironmentInfo[]).map(
    (e) => e.environment_code,
  );

  const usersQuery = useQuery({
    queryKey: ["users", "product-owner-options"],
    queryFn: () => userService.getUsers({ page: 1, pageSize: 50 }),
    staleTime: 30_000,
  });
  const users: User[] = usersQuery.data?.data ?? [];
  const loadingUsers = usersQuery.isLoading;

  const handleValuesChange = (
    _changedValues: Partial<ProductInformationValues>,
    allValues: ProductInformationValues,
  ) => {

    onChange({
      product_name: allValues.product_name || "",
      product_code: allValues.product_code || "",
      description: allValues.description || "",
      owner_id: allValues.owner_id,
      is_active: allValues.is_active ?? true,
      environments: allValues.environments || [],
    });
  };

  const handleTogglePresetEnv = (preset: EnvironmentInfo) => {
    const currentValues = form.getFieldsValue();
    const currentEnvs = currentValues.environments || [];
    const exists = currentEnvs.some(
      (e) => e.environment_code === preset.environment_code,
    );

    let updatedEnvs: EnvironmentInfo[];
    if (exists) {
      updatedEnvs = currentEnvs.filter(
        (e) => e.environment_code !== preset.environment_code,
      );
    } else {
      // A product is configured against one active environment at a time.
      updatedEnvs = [preset];
    }

    form.setFieldsValue({ environments: updatedEnvs });
    onChange({
      product_name: currentValues.product_name || "",
      product_code: currentValues.product_code || "",
      description: currentValues.description || "",
      owner_id: currentValues.owner_id,
      is_active: currentValues.is_active ?? true,
      environments: updatedEnvs,
    });
  };

  return (
    <div className="max-w-3xl mx-auto space-y-6">
      <div>
        <h2 className="text-xl font-bold text-zinc-900">
          ขั้นตอนที่ 1: ข้อมูล Product
        </h2>
        <p className="text-sm text-zinc-500 mt-1">
          กรอกรายละเอียดพื้นฐานเพื่อสร้าง Product แล้วเลือก Environment เฉพาะที่ต้องการใช้งาน
        </p>
      </div>

      {isCreated && (
        <Alert
          type="success"
          showIcon
          icon={<CheckCircleOutlined />}
          message="สร้าง Product สำเร็จ"
          description={
            <div>
              <p className="text-sm text-zinc-700 font-medium">
                Product นี้ยังไม่มี Environment อัตโนมัติ
              </p>
            </div>
          }
        />
      )}

      <Card className="border-zinc-200 shadow-sm rounded-lg">
        <Form
          form={form}
          layout="vertical"
          initialValues={{
            product_name: initialValues.product_name,
            product_code: initialValues.product_code || "",
            description: initialValues.description,
            owner_id: initialValues.owner_id,
            is_active: initialValues.is_active ?? true,
            environments: initialValues.environments,
          }}
          onValuesChange={handleValuesChange}
          requiredMark={false}
        >
          <div className="w-full">
            <Form.Item
              name="product_name"
              label={
                <span className="font-medium text-zinc-700">
                  ชื่อ Product <span className="text-red-500">*</span>
                </span>
              }
              rules={[{ required: true, message: "กรุณากรอกชื่อ Product" }]}
              validateStatus={productNameError ? "error" : undefined}
              help={productNameError}
            >
              <Input
                placeholder="กรอกชื่อ Product"
                size="large"
                className="rounded-md"
                onBlur={onProductNameBlur}
              />
            </Form.Item>

            <Form.Item
              name="environments"
              label={
                <div className="flex items-center gap-2">
                  <span className="font-medium text-zinc-700">
                    Environments
                  </span>
                </div>
              }
            >
              <div className="space-y-3">
                <div>
                  <span className="text-xs text-zinc-500 font-medium block mb-1.5">
                    เลือก Environment 1 รายการ:
                  </span>
                  <div className="flex flex-wrap items-center gap-2">
                    {PRESET_ENVIRONMENTS.map((preset) => {
                      const isSelected = selectedCodes.includes(
                        preset.environment_code,
                      );
                      return (
                        <Button
                          key={preset.environment_code}
                          size="small"
                          type={isSelected ? "primary" : "default"}
                          ghost={isSelected}
                          icon={
                            isSelected ? <CheckOutlined /> : <PlusOutlined />
                          }
                          onClick={() => handleTogglePresetEnv(preset)}
                          className="font-mono text-xs rounded-md flex items-center"
                        >
                          {preset.environment_code} ({preset.environment_name})
                        </Button>
                      );
                    })}
                  </div>
                </div>
              </div>
            </Form.Item>
          </div>

          <Form.Item
            name="description"
            label={
              <span className="font-medium text-zinc-700">คำอธิบาย</span>
            }
          >
            <Input.TextArea
              placeholder="แพลตฟอร์มจัดการกำลังคนสำหรับตารางงานและการลงเวลา"
              rows={3}
              className="rounded-md"
            />
          </Form.Item>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <Form.Item
              name="owner_id"
              label={
                <span className="font-medium text-zinc-700">เจ้าของ Product</span>
              }
            >
              <Select
                placeholder="เลือกเจ้าของ Product"
                size="large"
                loading={loadingUsers}
                allowClear
                options={users.map((u) => ({
                  value: Number(u.id) || u.id,
                  label: (
                    <span className="flex items-center gap-2">
                      <UserOutlined className="text-zinc-400" />
                      <span>{u.fullName || u.username}</span>
                      {u.email && (
                        <span className="text-xs text-zinc-400">
                          ({u.email})
                        </span>
                      )}
                    </span>
                  ),
                }))}
              />
            </Form.Item>

            <Form.Item
              name="is_active"
              label={<span className="font-medium text-zinc-700">สถานะ</span>}
              valuePropName="checked"
            >
              <div className="flex items-center gap-3 pt-1">
                <Switch defaultChecked />
                <span className="text-sm font-medium text-zinc-700">
                  เปิดใช้งาน (พร้อมรับ Log)
                </span>
              </div>
            </Form.Item>
          </div>

          <div className="p-3.5 bg-zinc-50 rounded-lg border border-zinc-200/80 text-xs text-zinc-600 flex items-start gap-2 mt-2">
            <InfoCircleOutlined className="text-zinc-400 mt-0.5" />
            <div>
              <span className="font-semibold text-zinc-800">
                Environment เริ่มต้น:{" "}
              </span>
              OmniLogs จะเพิ่ม Environment{" "}
              <code className="font-mono bg-zinc-200 px-1 py-0.5 rounded text-zinc-900">
                DEV
              </code>{" "}
              and{" "}
              <code className="font-mono bg-zinc-200 px-1 py-0.5 rounded text-zinc-900">
                PROD
              </code>{" "}
              ให้ Product นี้โดยอัตโนมัติเมื่อสร้าง Product
            </div>
          </div>
        </Form>
        {navigation}
      </Card>
    </div>
  );
}



