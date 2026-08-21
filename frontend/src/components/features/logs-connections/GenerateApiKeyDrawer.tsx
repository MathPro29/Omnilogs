import { useMemo, useState } from "react";
import {
  Button,
  DatePicker,
  Drawer,
  Form,
  Input,
  InputNumber,
  Select,
  Switch,
} from "antd";
import type {
  GenerateApiKeyInput,
  ProductEnvironment,
} from "@/types/apiKey.types";
import { generateApiKeySchema } from "@/schemas/apiKey.schema";
import type { ProductFeatureOption, ProductProjectOption } from "@/services/product.service";
const scopes = [
  { value: "LOG_INGEST_CREATE", label: "Send logs" },
  { value: "CONNECTION_STATUS_READ", label: "Read connection status" },
  { value: "AUDIT_LOG_CREATE", label: "Send audit logs" },
  { value: "APPLICATION_LOG_CREATE", label: "Send application logs" },
  { value: "CUSTOM_LOG_CREATE", label: "Send custom logs" },
];
const initial: GenerateApiKeyInput = {
  name: "",
  description: "",
  expiration: "90d",
  permissions: ["LOG_INGEST_CREATE"],
  allowedSources: [],
  ipAllowlist: [],
  rateLimit: 1000,
  active: true,
};
interface Props {
  open: boolean;
  productName?: string;
  environments: ProductEnvironment[];
  projects?: ProductProjectOption[];
  features?: ProductFeatureOption[];
  loading: boolean;
  onClose: () => void;
  onSubmit: (value: GenerateApiKeyInput) => void;
}
export function GenerateApiKeyDrawer(props: Props) {
  const [value, setValue] = useState(initial);
  const parsed = useMemo(() => generateApiKeySchema.safeParse(value), [value]);
  const error = (field: string) =>
    parsed.success
      ? undefined
      : parsed.error.issues.find((issue) => issue.path[0] === field)?.message;
  const close = () => {
    setValue(initial);
    props.onClose();
  };
  const selectedProjectFeatureOptions = useMemo(() => {
    const list = (props.features ?? []).filter(
      (feature) => feature.project_id === value.defaultProjectId,
    );
    if (list.length === 0) return [];

    const categories = list.filter(
      (item) => !item.parent_id || item.level === 1,
    );
    const subMap = new Map<number, typeof list>();

    list.forEach((item) => {
      if (item.parent_id || item.level > 1) {
        const parentId = item.parent_id ?? 0;
        if (!subMap.has(parentId)) subMap.set(parentId, []);
        subMap.get(parentId)!.push(item);
      }
    });

    const groups: Array<{
      label: string;
      options: Array<{ value: number; label: string }>;
    }> = [];

    categories.forEach((cat) => {
      const subs = subMap.get(cat.category_id) ?? [];
      groups.push({
        label: `${cat.category_name} (#${cat.category_id})`,
        options: [
          {
            value: cat.category_id,
            label: `${cat.category_name} (Category Main)`,
          },
          ...subs.map((sub) => ({
            value: sub.category_id,
            label: `${sub.category_name} (#${sub.category_id})`,
          })),
        ],
      });
    });

    subMap.forEach((subs, parentId) => {
      if (
        parentId !== 0 &&
        !categories.some((c) => c.category_id === parentId)
      ) {
        groups.push({
          label: `Category #${parentId}`,
          options: subs.map((sub) => ({
            value: sub.category_id,
            label: `${sub.category_name} (#${sub.category_id})`,
          })),
        });
      }
    });

    return groups;
  }, [props.features, value.defaultProjectId]);

  return (
    <Drawer
      title="Generate API key"
      open={props.open}
      size={480}
      onClose={close}
      destroyOnHidden
      footer={
        <div className="drawer-actions text-white">
          <Button onClick={close}>Cancel</Button>
          <Button
            type="primary"
            loading={props.loading}
            disabled={!parsed.success}
            onClick={() => parsed.success && props.onSubmit(parsed.data)}
          >
            Generate API Key
          </Button>
        </div>
      }
    >
      <Form layout="vertical">
        <Form.Item
          label="Key name"
          required
          validateStatus={error("name") ? "error" : undefined}
          help={error("name")}
        >
          <Input
            value={value.name}
            onChange={(event) =>
              setValue({ ...value, name: event.target.value })
            }
            placeholder="Production ingestion"
          />
        </Form.Item>
        <Form.Item label="Product">
          <Input value={props.productName} disabled />
        </Form.Item>
        <Form.Item label="Environment" required validateStatus={error("environmentId") ? "error" : undefined} help={error("environmentId")}>
          <Select
            allowClear
            value={value.environmentId}
            onChange={(environmentId) => setValue({ ...value, environmentId })}
            placeholder="Select the environment that may send logs"
            options={props.environments.map((item) => ({
              value: item.environmentId,
              label: `${item.name} (${item.code})`,
            }))}
          />
        </Form.Item>
        <Form.Item label="Default project" extra="Logs sent with this key are automatically classified into this project.">
          <Select
            allowClear
            value={value.defaultProjectId}
            onChange={(defaultProjectId) =>
              setValue({ ...value, defaultProjectId, defaultCategoryId: undefined })
            }
            placeholder="Leave unclassified at product level"
            options={(props.projects ?? []).map((project) => ({
              value: project.project_id,
              label: `${project.project_name} (#${project.project_id})`,
            }))}
          />
        </Form.Item>
        <Form.Item label="Default feature" extra="Optional. Choose the most specific destination for this source.">
          <Select
            allowClear
            disabled={!value.defaultProjectId}
            value={value.defaultCategoryId}
            onChange={(defaultCategoryId) =>
              setValue({ ...value, defaultCategoryId })
            }
            placeholder={
              value.defaultProjectId
                ? "Leave at project level"
                : "Choose a project first"
            }
            options={selectedProjectFeatureOptions}
          />
        </Form.Item>
        <Form.Item label="Description">
          <Input.TextArea
            value={value.description}
            maxLength={500}
            onChange={(event) =>
              setValue({ ...value, description: event.target.value })
            }
          />
        </Form.Item>
        <Form.Item label="Expiration policy">
          <Select
            value={value.expiration}
            onChange={(expiration) => setValue({ ...value, expiration })}
            options={[
              { value: "30d", label: "30 days" },
              { value: "90d", label: "90 days" },
              { value: "180d", label: "180 days" },
              { value: "1y", label: "1 year" },
              { value: "custom", label: "Custom date" },
              { value: "never", label: "Never expires" },
            ]}
          />
        </Form.Item>
        {value.expiration === "custom" && (
          <Form.Item
            label="Expiration date"
            validateStatus={error("customDate") ? "error" : undefined}
            help={error("customDate")}
          >
            <DatePicker
              className="w-full"
              onChange={(date) =>
                setValue({ ...value, customDate: date?.toISOString() })
              }
            />
          </Form.Item>
        )}
        <Form.Item
          label="Permission scope"
          required
          extra="Least privilege is selected by default. Management permissions are never granted to ingestion keys."
        >
          <Select
            mode="multiple"
            value={value.permissions}
            onChange={(permissions) => setValue({ ...value, permissions })}
            options={scopes}
          />
        </Form.Item>
        <Form.Item label="Allowed log sources">
          <Select
            mode="tags"
            value={value.allowedSources}
            onChange={(allowedSources) =>
              setValue({ ...value, allowedSources })
            }
            placeholder="service-name"
          />
        </Form.Item>
        <Form.Item
          label="IP allowlist"
          validateStatus={error("ipAllowlist") ? "error" : undefined}
          help={error("ipAllowlist")}
        >
          <Select
            mode="tags"
            tokenSeparators={[","]}
            value={value.ipAllowlist}
            onChange={(ipAllowlist) => setValue({ ...value, ipAllowlist })}
            placeholder="10.0.0.1"
          />
        </Form.Item>
        <Form.Item label="Rate limit (requests/minute)">
          <InputNumber
            min={1}
            max={100000}
            value={value.rateLimit}
            onChange={(rateLimit) =>
              setValue({ ...value, rateLimit: rateLimit ?? undefined })
            }
            className="w-full"
          />
        </Form.Item>
        <Form.Item label="Status">
          <Switch
            checked={value.active}
            onChange={(active) => setValue({ ...value, active })}
            checkedChildren="Active"
            unCheckedChildren="Disabled"
          />
        </Form.Item>
      </Form>
    </Drawer>
  );
}
