import { useEffect, useMemo } from "react";
import { Alert, Button, Form, Input, Modal, Select, message } from "antd";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { productService } from "@/services/product.service";
import { productSetupService } from "@/features/product-setup/services/productSetup.service";

interface AddFeatureModalProps {
  productId?: number;
  projectId?: number;
  initialCode?: string;
  initialName?: string;
  open: boolean;
  onClose: () => void;
  onSuccess?: () => void;
}

interface FormValues {
  project_id: number;
  category_name: string;
  category_code?: string;
  parent_id?: number | null;
  description?: string;
}

export function AddFeatureModal({
  productId,
  projectId: initialProjectId,
  initialCode,
  initialName,
  open,
  onClose,
  onSuccess,
}: AddFeatureModalProps) {
  const [form] = Form.useForm<FormValues>();
  const queryClient = useQueryClient();

  const selectedProjectId = Form.useWatch("project_id", form) ?? initialProjectId;
  const categoryName = Form.useWatch("category_name", form);
  const categoryCode = Form.useWatch("category_code", form);

  const projectsQuery = useQuery({
    queryKey: ["log-explorer-add-feature-projects", productId],
    queryFn: () => productService.listProjects(productId!),
    enabled: open && Boolean(productId),
  });

  const parentFeaturesQuery = useQuery({
    queryKey: ["log-explorer-parent-features", productId, selectedProjectId],
    queryFn: () => productService.listFeatures(productId!, selectedProjectId),
    enabled: open && Boolean(productId) && Boolean(selectedProjectId),
  });

  const isDuplicate = useMemo(() => {
    const list = parentFeaturesQuery.data ?? [];
    if (list.length === 0) return false;
    const nameTrimmed = (categoryName || "").trim().toLowerCase();
    const codeTrimmed = (categoryCode || "").trim().toLowerCase();
    if (!nameTrimmed && !codeTrimmed) return false;

    return list.some((item) => {
      const itemName = (item.category_name || "").trim().toLowerCase();
      const itemCode = (item.category_code || "").trim().toLowerCase();
      const isNameMatch = nameTrimmed !== "" && itemName === nameTrimmed;
      const isCodeMatch = codeTrimmed !== "" && itemCode === codeTrimmed;
      return isNameMatch || isCodeMatch;
    });
  }, [parentFeaturesQuery.data, categoryName, categoryCode]);

  useEffect(() => {
    if (open) {
      form.setFieldsValue({
        project_id: initialProjectId,
        category_name: initialName || initialCode || "",
        category_code: initialCode || "",
      });
    }
  }, [form, initialCode, initialName, initialProjectId, open]);

  const createMutation = useMutation({
    mutationFn: (values: FormValues) =>
      productSetupService.createFeature(productId!, values.project_id, {
        category_name: values.category_name.trim(),
        category_code: values.category_code?.trim() || undefined,
        parent_id: values.parent_id || null,
        description: values.description?.trim() || undefined,
      }),
    onSuccess: () => {
      message.success("เพิ่ม Feature / Category สำเร็จ");
      void queryClient.invalidateQueries({ queryKey: ["products"] });
      void queryClient.invalidateQueries({ queryKey: ["features"] });
      void queryClient.invalidateQueries({ queryKey: ["log-routing-features"] });
      void queryClient.invalidateQueries({ queryKey: ["log-explorer-parent-features"] });
      void queryClient.invalidateQueries({ queryKey: ["log-explorer-routing-features"] });
      void queryClient.invalidateQueries({ queryKey: ["features-options"] });
      form.resetFields();
      onClose();
      if (onSuccess) onSuccess();
    },
    onError: (err: Error) => {
      message.error(err.message || "ไม่สามารถเพิ่ม Feature ได้");
    },
  });

  const handleSubmit = (values: FormValues) => {
    if (!productId) {
      message.warning("กรุณาเลือก Product ก่อนสร้าง Feature");
      return;
    }
    createMutation.mutate(values);
  };

  const projectOptions = (projectsQuery.data ?? []).map((p) => ({
    value: p.project_id,
    label: `${p.project_name} (#${p.project_id})`,
  }));

  const parentOptions = useMemo(() => {
    const list = parentFeaturesQuery.data ?? [];
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
        label: `${cat.category_name} (${cat.category_code})`,
        options: [
          {
            value: cat.category_id,
            label: `${cat.category_name} (${cat.category_code}) [Category หลัก]`,
          },
          ...subs.map((sub) => ({
            value: sub.category_id,
            label: `${sub.category_name} (${sub.category_code})`,
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
            label: `${sub.category_name} (${sub.category_code})`,
          })),
        });
      }
    });

    return groups;
  }, [parentFeaturesQuery.data]);

  return (
    <Modal
      title="เพิ่ม Feature / Category ใหม่"
      open={open}
      onCancel={onClose}
      footer={null}
      destroyOnClose
    >
      <Form form={form} layout="vertical" onFinish={handleSubmit} className="mt-4">
        {isDuplicate && (
          <Alert
            type="warning"
            showIcon
            message="Feature หรือ Category นี้ถูกเพิ่มไปแล้วใน Project นี้"
            className="mb-4"
          />
        )}
        <Form.Item
          name="project_id"
          label="Project ที่สังกัด"
          rules={[{ required: true, message: "กรุณาเลือก Project" }]}
        >
          <Select
            options={projectOptions}
            placeholder="เลือก Project"
            loading={projectsQuery.isLoading}
          />
        </Form.Item>
        <Form.Item
          name="category_name"
          label="ชื่อ Feature / Category"
          rules={[{ required: true, message: "กรุณาระบุชื่อ Feature" }]}
        >
          <Input placeholder="เช่น การจัดการใบเสร็จรับเงิน หรือ CREATE" />
        </Form.Item>
        <Form.Item
          name="category_code"
          label="Feature Code (รหัสจับคู่ custom_fields)"
          tooltip="ตรงกับ custom_fields.feature_code หรือ sub_feature_code"
        >
          <Input placeholder="เช่น RECEIPTS หรือ CREATE" />
        </Form.Item>
        <Form.Item
          name="parent_id"
          label="Category แม่ (กรณีเป็น Sub-feature)"
        >
          <Select
            allowClear
            showSearch
            optionFilterProp="label"
            options={parentOptions}
            placeholder="ไม่มี (เป็น Feature ระดับแรก)"
            loading={parentFeaturesQuery.isLoading}
          />
        </Form.Item>
        <Form.Item name="description" label="รายละเอียดเพิ่มเติม">
          <Input.TextArea rows={2} placeholder="คำอธิบายเพิ่มเติม" />
        </Form.Item>
        <div className="flex justify-end gap-2 mt-6">
          <Button onClick={onClose}>ยกเลิก</Button>
          <Button
            type="primary"
            htmlType="submit"
            loading={createMutation.isPending}
            disabled={!productId || isDuplicate}
          >
            {isDuplicate ? "ถูกเพิ่มไปแล้ว" : "บันทึก Feature"}
          </Button>
        </div>
      </Form>
    </Modal>
  );
}
