import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  ApiOutlined,
  EditOutlined,
  DownOutlined,
  PlusOutlined,
} from "@ant-design/icons";
import {
  Alert,
  Button,
  Card,
  Form,
  Input,
  Select,
  Space,
  message,
} from "antd";
import {
  productService,
  type CreateProductLogRoutingRuleInput,
  type ProductLogRoutingRule,
  type ProductOption,
  type RoutingDiscoveryItem,
} from "@/services/product.service";
import {
  buildFeatureOptions,
  buildRoutingInput,
  methodOptions,
  toRoutingFormValues,
  type RoutingFormValues,
} from "./product-routing.utils";
import { ProductRoutingDiscoveryList } from "./ProductRoutingDiscoveryList";
import { ProductRoutingRuleList } from "./ProductRoutingRuleList";

export function ProductPayloadRoutingPanel({
  product,
}: {
  product: ProductOption;
}) {
  const [isOpen, setIsOpen] = useState(true);
  const [form] = Form.useForm<RoutingFormValues>();
  const [projectId, setProjectId] = useState<number>();
  const [editingRule, setEditingRule] = useState<ProductLogRoutingRule | null>(null);
  const client = useQueryClient();

  const projects = useQuery({
    queryKey: ["payload-routing-projects", product.id],
    queryFn: () => productService.listProjects(product.id),
  });

  const features = useQuery({
    queryKey: ["payload-routing-features", product.id, projectId],
    queryFn: () => productService.listFeatures(product.id, projectId),
    enabled: Boolean(projectId),
  });

  const environments = useQuery({
    queryKey: ["payload-routing-environments", product.id],
    queryFn: () => productService.listEnvironments(product.id),
  });

  const rules = useQuery({
    queryKey: ["log-routing-rules", product.id],
    queryFn: () => productService.listLogRoutingRules(product.id),
  });

  const discovery = useQuery({
    queryKey: ["routing-discovery", product.id],
    queryFn: () => productService.listRoutingDiscovery(product.id),
    staleTime: 15_000,
  });

  const create = useMutation({
    mutationFn: (input: CreateProductLogRoutingRuleInput) =>
      productService.createLogRoutingRule(product.id, input),
    onSuccess: () => {
      message.success("เพิ่ม HTTP Routing Rule สำเร็จ");
      form.resetFields();
      setProjectId(undefined);
      void client.invalidateQueries({
        queryKey: ["log-routing-rules", product.id],
      });
      void client.invalidateQueries({
        queryKey: ["routing-discovery", product.id],
      });
    },
    onError: () => message.error("ไม่สามารถเพิ่ม HTTP Routing Rule ได้"),
  });

  const update = useMutation({
    mutationFn: ({ ruleId, input }: { ruleId: number; input: CreateProductLogRoutingRuleInput }) =>
      productService.updateLogRoutingRule(product.id, ruleId, input),
    onSuccess: () => {
      message.success("อัปเดต HTTP Routing Rule สำเร็จ");
      form.resetFields();
      setProjectId(undefined);
      setEditingRule(null);
      void client.invalidateQueries({ queryKey: ["log-routing-rules", product.id] });
      void client.invalidateQueries({ queryKey: ["routing-discovery", product.id] });
    },
    onError: () => message.error("ไม่สามารถอัปเดต HTTP Routing Rule ได้"),
  });

  const remove = useMutation({
    mutationFn: (ruleId: number) =>
      productService.deleteLogRoutingRule(product.id, ruleId),
    onSuccess: () => {
      message.success("ลบ HTTP Routing Rule สำเร็จ");
      void client.invalidateQueries({
        queryKey: ["log-routing-rules", product.id],
      });
      void client.invalidateQueries({
        queryKey: ["routing-discovery", product.id],
      });
    },
  });

  const projectOptions = (projects.data ?? []).map((item) => ({
    value: item.project_id,
    label: `${item.project_name} (#${item.project_id})`,
  }));

  const featureOptions = buildFeatureOptions(features.data ?? []);

  const environmentOptions = (environments.data ?? []).map((item) => ({
    value: item.environment_id,
    label: `${item.environment_name} (${item.environment_code})`,
  }));

  const projectName = (id: number) =>
    projects.data?.find((item) => item.project_id === id)?.project_name ??
    String(id);

  const featureName = (id?: number | null) =>
    features.data?.find((item) => item.category_id === id)?.category_name ??
    (id ? String(id) : "ทั้ง Project");

  const submit = (values: RoutingFormValues) => {
    const input = buildRoutingInput(values);
    if (editingRule) update.mutate({ ruleId: editingRule.rule_id, input });
    else create.mutate(input);
  };

  const editRule = (rule: ProductLogRoutingRule) => {
    setEditingRule(rule);
    setProjectId(rule.target_project_id);
    form.setFieldsValue(toRoutingFormValues(rule));
  };

  const useSuggestion = (item: RoutingDiscoveryItem) => {
    setProjectId(undefined);
    form.setFieldsValue({
      rule_name: `Map ${item.service_name || item.request_path}`,
      environment_id: item.environment_id,
      service: item.service_name || undefined,
      request_method: item.request_method || undefined,
      request_path: item.request_path,
      path_operator: "equals",
      target_project_id: undefined,
      target_category_id: undefined,
    });
    message.info("ตรวจสอบปลายทางแล้ว กดเพิ่ม HTTP Rule เพื่อบันทึก mapping");
  };

  return (
    <Card
      className="product-details-section"
      title={
        <div
          onClick={() => setIsOpen(!isOpen)}
          className="flex items-center justify-between cursor-pointer select-none w-full py-1 text-zinc-800 hover:text-zinc-950 transition-colors"
        >
          <Space>
            <ApiOutlined className="text-[#1F8457]" />
            <span className="font-semibold">HTTP Route Mapping</span>
          </Space>
          <DownOutlined
            className={`transition-transform duration-200 text-xs text-zinc-400 ${isOpen ? "rotate-0" : "-rotate-90"}`}
          />
        </div>
      }
    >
      {isOpen && (
        <>
          <Alert
            type="info"
            showIcon
            message="Map เงื่อนไขการจัดเส้นทาง request และกำหนด connector ที่เกี่ยวข้อง"
            description="กำหนด service, request path และ method เพื่อจัดเส้นทาง log เข้าไปยัง Project/Feature ปลายทางที่ต้องการตาม route_key ที่ตั้งไว้ใน .env"
          />
          {discovery.data?.length ? (
            <div style={{ marginTop: 16 }}>
              <div className="mb-2 text-sm text-zinc-600">
                API paths detected from developer traffic. Choose a path to
                prefill a new mapping.
              </div>
              <Alert
                type="warning"
                showIcon
                message="Unmapped traffic discovered"
                description="เลือก pattern ที่พบจริงเพื่อเติมแบบฟอร์ม แล้วเลือก Project/Feature ปลายทางก่อนบันทึก"
              />
              <ProductRoutingDiscoveryList
                items={discovery.data ?? []}
                rules={rules.data ?? []}
                loading={discovery.isLoading}
                onUseSuggestion={useSuggestion}
              />
            </div>
          ) : null}
          <Form
            form={form}
            layout="vertical"
            onFinish={submit}
            style={{ marginTop: 16 }}
          >
            <div className="grid grid-cols-1 md:grid-cols-2 gap-x-4">
              <Form.Item
                name="rule_name"
                label="ชื่อ Rule"
                rules={[{ required: true, message: "กรุณาระบุชื่อ Rule" }]}
              >
                <Input placeholder="ASW members update" />
              </Form.Item>
              <Form.Item name="environment_id" label="Environment">
                <Select
                  allowClear
                  options={environmentOptions}
                  placeholder="ทุก Environment"
                />
              </Form.Item>
              <Form.Item name="service" label="Service">
                <Input placeholder="เช่น asw-web (เว้นว่างได้)" />
              </Form.Item>
              <Form.Item name="source_project_id" label="Source Project ID">
                <Input type="number" placeholder="เช่น projectsId จากระบบต้นทาง" />
              </Form.Item>
              <Form.Item name="request_method" label="HTTP Method">
                <Select
                  allowClear
                  options={methodOptions}
                  placeholder="ทุก Method"
                />
              </Form.Item>
              <Form.Item
                name="request_path"
                label="Request path"
                dependencies={["source_project_id"]}
                rules={[
                  ({ getFieldValue }) => ({
                    validator: (_, value) =>
                      value?.trim() || getFieldValue("source_project_id")
                        ? Promise.resolve()
                        : Promise.reject(new Error("Provide a request path or source project ID")),
                  }),
                ]}
              >
                <Input placeholder="เช่น /api/members/update" />
              </Form.Item>
              <Form.Item
                name="path_operator"
                label="รูปแบบการจับคู่"
                initialValue="equals"
              >
                <Select
                  options={[
                    { value: "equals", label: "ตรงกันทั้งหมด" },
                    { value: "starts_with", label: "ขึ้นต้นด้วย" },
                  ]}
                />
              </Form.Item>
              <Form.Item
                name="target_project_id"
                label="Project ปลายทาง"
                rules={[{ required: true, message: "กรุณาเลือก Project" }]}
              >
                <Select
                  options={projectOptions}
                  onChange={(value) => {
                    setProjectId(value);
                    form.setFieldValue("target_category_id", undefined);
                  }}
                />
              </Form.Item>
              <Form.Item
                name="target_category_id"
                label="Feature / Sub-feature"
              >
                <Select
                  allowClear
                  loading={features.isLoading}
                  options={featureOptions}
                  placeholder="เว้นว่างเพื่อจัดเข้าทั้ง Project"
                />
              </Form.Item>
            </div>
            <Form.Item name="priority" label="Priority" initialValue={0}>
              <Input type="number" />
            </Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                icon={editingRule ? <EditOutlined /> : <PlusOutlined />}
                loading={create.isPending || update.isPending}
              >
                {editingRule ? "บันทึกการแก้ไข" : "เพิ่ม HTTP Rule"}
              </Button>
              {editingRule ? (
                <Button onClick={() => { form.resetFields(); setProjectId(undefined); setEditingRule(null); }}>
                  ยกเลิก
                </Button>
              ) : null}
            </Space>
          </Form>
          <ProductRoutingRuleList
            rules={rules.data ?? []}
            loading={rules.isLoading}
            environmentOptions={environmentOptions}
            projectName={projectName}
            featureName={featureName}
            onEdit={editRule}
            onDelete={(ruleId) => remove.mutate(ruleId)}
          />
        </>
      )}
    </Card>
  );
}
