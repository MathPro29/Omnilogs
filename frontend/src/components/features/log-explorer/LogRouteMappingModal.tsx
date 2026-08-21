import { useEffect, useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Alert, Button, Form, Input, Modal, Select, message } from "antd";
import {
  productService,
  type CreateProductLogRoutingRuleInput,
} from "@/services/product.service";
import type { SearchRecord } from "@/features/log-explorer/services/log-search.service";

type FormValues = {
  rule_name: string;
  environment_id?: number;
  service?: string;
  request_method?: string;
  request_path: string;
  path_operator: "equals" | "starts_with";
  payload_field?: string;
  payload_operator?: "equals" | "contains" | "starts_with" | "in";
  payload_value?: string;
  priority?: number;
  target_project_id: number;
  target_category_id?: number;
};
const methods = ["GET", "POST", "PUT", "PATCH", "DELETE"].map((value) => ({
  value,
  label: value,
}));
const text = (record: SearchRecord, ...keys: string[]) => {
  for (const key of keys) {
    const value = key
      .split(".")
      .reduce<unknown>(
        (current, part) =>
          current && typeof current === "object"
            ? (current as Record<string, unknown>)[part]
            : undefined,
        record,
      );
    if (typeof value === "string" && value.trim()) return value.trim();
  }
  return "";
};

function extractPayloadCandidates(
  record: SearchRecord,
): Array<{ label: string; value: string; sample: string }> {
  const candidates: Array<{ label: string; value: string; sample: string }> =
    [];
  const visited = new Set<string>();

  const addCandidate = (fieldPath: string, rawVal: unknown) => {
    if (visited.has(fieldPath)) return;
    visited.add(fieldPath);
    if (
      rawVal !== undefined &&
      rawVal !== null &&
      (typeof rawVal === "string" || typeof rawVal === "number")
    ) {
      const sample = String(rawVal).trim();
      if (sample && sample.length < 100) {
        candidates.push({
          label: `${fieldPath} (ค่าปัจจุบัน: "${sample}")`,
          value: fieldPath,
          sample,
        });
      }
    }
  };

  const priorityKeys = [
    "type",
    "bill_type",
    "category",
    "document_type",
    "sub_type",
    "action",
    "event",
  ];
  const objSources: Array<{
    prefix: string;
    obj: Record<string, unknown> | undefined;
  }> = [
    { prefix: "", obj: record as Record<string, unknown> },
    {
      prefix: "payload.",
      obj: record.payload as Record<string, unknown> | undefined,
    },
    {
      prefix: "data.",
      obj: record.data as Record<string, unknown> | undefined,
    },
    {
      prefix: "metadata.",
      obj: record.metadata as Record<string, unknown> | undefined,
    },
  ];

  for (const { prefix, obj } of objSources) {
    if (!obj || typeof obj !== "object") continue;
    for (const key of priorityKeys) {
      if (key in obj) addCandidate(`${prefix}${key}`, obj[key]);
    }
    for (const [k, v] of Object.entries(obj)) {
      if (typeof v === "string" || typeof v === "number") {
        addCandidate(`${prefix}${k}`, v);
      }
    }
  }

  return candidates;
}

export function LogRouteMappingModal({
  record,
  productId,
  open,
  onClose,
}: {
  record: SearchRecord;
  productId?: number;
  open: boolean;
  onClose: () => void;
}) {
  const [form] = Form.useForm<FormValues>();
  const [projectId, setProjectId] = useState<number>();
  const client = useQueryClient();
  const path = text(
    record,
    "path",
    "request_path",
    "payload.request_path",
    "data.request_path",
    "request.path",
  );
  const method = text(
    record,
    "method",
    "request_method",
    "payload.request_method",
    "data.request_method",
    "request.method",
  ).toUpperCase();
  const service = text(
    record,
    "service",
    "service_name",
    "payload.service",
    "data.service",
  );
  const environmentId =
    typeof record.environment_id === "number"
      ? record.environment_id
      : undefined;

  const payloadCandidates = useMemo(
    () => extractPayloadCandidates(record),
    [record],
  );

  const projects = useQuery({
    queryKey: ["log-explorer-routing-projects", productId],
    queryFn: () => productService.listProjects(productId!),
    enabled: open && Boolean(productId),
  });
  const environments = useQuery({
    queryKey: ["log-explorer-routing-environments", productId],
    queryFn: () => productService.listEnvironments(productId!),
    enabled: open && Boolean(productId),
  });
  const features = useQuery({
    queryKey: ["log-explorer-routing-features", productId, projectId],
    queryFn: () => productService.listFeatures(productId!, projectId),
    enabled: open && Boolean(productId) && Boolean(projectId),
  });
  const existingRulesQuery = useQuery({
    queryKey: ["log-routing-rules", productId],
    queryFn: () => productService.listLogRoutingRules(productId!),
    enabled: open && Boolean(productId),
  });

  const currentPath = Form.useWatch("request_path", form);
  const currentMethod = Form.useWatch("request_method", form);
  const currentPayloadField = Form.useWatch("payload_field", form);
  const currentPayloadValue = Form.useWatch("payload_value", form);

  const hasPayloadCondition = Boolean(
    currentPayloadField?.trim() && currentPayloadValue?.trim(),
  );

  const isRouteAlreadyMapped = useMemo(() => {
    const rules = existingRulesQuery.data ?? [];
    if (rules.length === 0 || !currentPath) return false;
    const pathTrimmed = currentPath.trim().toLowerCase();

    return rules.some((rule) =>
      rule.conditions.some(
        (c) =>
          c.field === "request_path" &&
          c.value?.trim().toLowerCase() === pathTrimmed &&
          (!currentMethod ||
            !rule.conditions.some((mc) => mc.field === "request_method") ||
            rule.conditions.some(
              (mc) =>
                mc.field === "request_method" &&
                mc.value?.toUpperCase() === currentMethod.toUpperCase(),
            )),
      ),
    );
  }, [existingRulesQuery.data, currentPath, currentMethod]);

  const create = useMutation({
    mutationFn: (input: CreateProductLogRoutingRuleInput) =>
      productService.createLogRoutingRule(productId!, input),
    onSuccess: () => {
      message.success("สร้าง HTTP route mapping สำเร็จ");
      void client.invalidateQueries({
        queryKey: ["log-routing-rules", productId],
      });
      void client.invalidateQueries({
        queryKey: ["routing-discovery", productId],
      });
      onClose();
    },
    onError: () => message.error("ไม่สามารถสร้าง HTTP route mapping ได้"),
  });

  useEffect(() => {
    if (!open) return;
    // Reset selection when a different log is opened for mapping.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setProjectId(undefined);

    // Try to prefill a suggested payload field if available (e.g. type or bill_type)
    const firstCandidate = payloadCandidates[0];
    const initialField = firstCandidate?.value;
    const initialVal = firstCandidate?.sample;

    form.setFieldsValue({
      rule_name: `Map ${service || path}`,
      environment_id: environmentId,
      service: service || undefined,
      request_method: method || undefined,
      request_path: path,
      path_operator: "equals",
      payload_field: initialField || undefined,
      payload_operator: "equals",
      payload_value: initialVal || undefined,
      priority: initialField && initialVal ? 10 : 0,
      target_project_id: undefined,
      target_category_id: undefined,
    });
  }, [environmentId, form, method, open, path, payloadCandidates, service]);

  const projectOptions = (projects.data ?? []).map((item) => ({
    value: item.project_id,
    label: `${item.project_name} (#${item.project_id})`,
  }));
  const environmentOptions = (environments.data ?? []).map((item) => ({
    value: item.environment_id,
    label: `${item.environment_name} (${item.environment_code})`,
  }));
  const featureOptions = useMemo(() => {
    const list = features.data ?? [];
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
  }, [features.data]);

  const submit = (values: FormValues) => {
    const conditions: CreateProductLogRoutingRuleInput["conditions"] = [
      {
        field: "request_path",
        operator: values.path_operator,
        value: values.request_path.trim(),
      },
    ];
    if (values.service?.trim())
      conditions.unshift({
        field: "service",
        operator: "equals",
        value: values.service.trim(),
      });
    if (values.request_method)
      conditions.push({
        field: "request_method",
        operator: "equals",
        value: values.request_method,
      });

    if (values.payload_field?.trim() && values.payload_value?.trim()) {
      conditions.push({
        field: values.payload_field.trim(),
        operator: values.payload_operator || "equals",
        value: values.payload_value.trim(),
      });
    }

    const priority =
      values.priority !== undefined
        ? Number(values.priority)
        : values.payload_field?.trim() && values.payload_value?.trim()
          ? 10
          : 0;

    create.mutate({
      environment_id: values.environment_id,
      rule_name: values.rule_name.trim(),
      priority,
      conditions,
      target_project_id: values.target_project_id,
      target_category_id: values.target_category_id,
    });
  };

  return (
    <Modal
      title="ระบุเส้นทาง route และ feature"
      open={open}
      onCancel={onClose}
      footer={null}
      destroyOnHidden
    >
      {isRouteAlreadyMapped && !hasPayloadCondition && (
        <Alert
          type="warning"
          showIcon
          message="Route / API Path นี้มี rule อยู่แล้วในระบบ"
          description="หาก API เดียวกันมีหลาย Feature (เช่น บิลค่าอินเตอร์เน็ต vs ค่าขยะ) กรุณาใส่ Payload Condition และตั้ง Priority สูงกว่าเพื่อแยก Feature"
          className="mb-4"
        />
      )}
      {isRouteAlreadyMapped && hasPayloadCondition && (
        <Alert
          type="info"
          showIcon
          message="แยก Feature สำหรับ API เดียวกัน"
          description="Rule นี้จะใช้ Payload Condition ร่วมกับ Priority สูงขึ้นเพื่อเจาะจง Feature ของ Log นี้"
          className="mb-4"
        />
      )}

      <Form form={form} layout="vertical" onFinish={submit}>
        <Form.Item
          name="rule_name"
          label="Rule name"
          rules={[{ required: true }]}
        >
          <Input />
        </Form.Item>
        <Form.Item name="environment_id" label="Environment">
          <Select
            allowClear
            options={environmentOptions}
            placeholder="All environments"
          />
        </Form.Item>

        <Form.Item name="request_method" label="HTTP method">
          <Select allowClear options={methods} />
        </Form.Item>
        <Form.Item
          name="request_path"
          label="API path"
          rules={[{ required: true, message: "API path is required" }]}
        >
          <Input />
        </Form.Item>
        <Form.Item
          name="path_operator"
          label="Path match"
          initialValue="equals"
        >
          <Select
            options={[
              { value: "equals", label: "Exact match" },
              { value: "starts_with", label: "Starts with" },
            ]}
          />
        </Form.Item>

        <Form.Item
          name="target_project_id"
          label="Target project"
          rules={[{ required: true, message: "Select a target project" }]}
        >
          <Select
            options={projectOptions}
            onChange={(value) => {
              setProjectId(value);
              form.setFieldValue("target_category_id", undefined);
            }}
          />
        </Form.Item>
        <Form.Item name="target_category_id" label="Feature / sub-feature">
          <Select
            allowClear
            loading={features.isLoading}
            options={featureOptions}
            placeholder="Optional"
          />
        </Form.Item>
        <Button
          type="primary"
          htmlType="submit"
          loading={create.isPending}
          disabled={!productId}
          className="w-full"
        >
          บันทึก
        </Button>
      </Form>
    </Modal>
  );
}
