import { Button, Form, Input, Modal, message } from "antd";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { productSetupService } from "@/features/product-setup/services/productSetup.service";

interface AddProductModalProps {
  open: boolean;
  onClose: () => void;
  onSuccess?: (productId: number) => void;
}

interface FormValues {
  product_name: string;
  product_code?: string;
  description?: string;
}

export function AddProductModal({ open, onClose, onSuccess }: AddProductModalProps) {
  const [form] = Form.useForm<FormValues>();
  const queryClient = useQueryClient();

  const createMutation = useMutation({
    mutationFn: (values: FormValues) =>
      productSetupService.createProductSetup({
        product_name: values.product_name.trim(),
        product_code: values.product_code?.trim() || undefined,
        description: values.description?.trim() || undefined,
        environments: [
          { environment_code: "UAT", environment_name: "UAT", environment_type: "STAGING" },
          { environment_code: "DEV", environment_name: "Development", environment_type: "DEVELOPMENT" },
          { environment_code: "PROD", environment_name: "Production", environment_type: "PRODUCTION" },
        ],
      }),
    onSuccess: (result) => {
      message.success("สร้าง Product ใหม่สำเร็จ");
      void queryClient.invalidateQueries({ queryKey: ["products"] });
      form.resetFields();
      onClose();
      const id = "product_id" in result ? result.product_id : result.id;
      if (id && onSuccess) {
        onSuccess(id);
      }
    },
    onError: (err: Error) => {
      message.error(err.message || "ไม่สามารถสร้าง Product ได้");
    },
  });

  const handleSubmit = (values: FormValues) => {
    createMutation.mutate(values);
  };

  return (
    <Modal
      title="เพิ่ม Product ใหม่"
      open={open}
      onCancel={onClose}
      footer={null}
      destroyOnClose
    >
      <Form form={form} layout="vertical" onFinish={handleSubmit} className="mt-4">
        <Form.Item
          name="product_name"
          label="ชื่อ Product"
          rules={[{ required: true, message: "กรุณาระบุชื่อ Product" }]}
        >
          <Input placeholder="เช่น Assetwise Core App" />
        </Form.Item>
        <Form.Item
          name="product_code"
          label="Product Code (รหัสสำหรับการจับคู่)"
          tooltip="ถ้าว่างไว้ ระบบจะสร้างให้อัตโนมัติ สามารถใช้รหัสตรงกับ custom_fields.product_code ได้"
        >
          <Input placeholder="เช่น ASSETWISE หรือ ASW" />
        </Form.Item>
        <Form.Item name="description" label="รายละเอียด (Optional)">
          <Input.TextArea rows={3} placeholder="คำอธิบายสั้นๆ เกี่ยวกับ Product นี้" />
        </Form.Item>
        <div className="flex justify-end gap-2 mt-6">
          <Button onClick={onClose}>ยกเลิก</Button>
          <Button type="primary" htmlType="submit" loading={createMutation.isPending}>
            บันทึก Product
          </Button>
        </div>
      </Form>
    </Modal>
  );
}
