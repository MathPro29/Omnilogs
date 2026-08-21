import { useEffect } from "react";
import { Form, Input, Modal, Switch, message } from "antd";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { productService, type ProductOption } from "@/services/product.service";

interface ProductEditModalProps {
  open: boolean;
  product?: ProductOption;
  onClose: () => void;
  onSuccess: () => void;
}

export function ProductEditModal({ open, product, onClose, onSuccess }: ProductEditModalProps) {
  const [form] = Form.useForm();
  const queryClient = useQueryClient();

  useEffect(() => {
    if (open && product) {
      form.setFieldsValue({
        product_name: product.name,
        description: product.description,
        is_active: product.is_active,
      });
    } else {
      form.resetFields();
    }
  }, [open, product, form]);

  const updateMutation = useMutation({
    mutationFn: async (values: {
      product_name?: string;
      description?: string;
      is_active?: boolean;
    }) => {
      if (!product) throw new Error("No product selected");
      return await productService.updateProduct(product.id, values);
    },
    onSuccess: () => {
      message.success(`อัปเดตข้อมูล Product "${product?.name}" สำเร็จ`);
      void queryClient.invalidateQueries({ queryKey: ["products", "management"] });
      onSuccess();
    },
    onError: (err) => {
      message.error("เกิดข้อผิดพลาด ไม่สามารถอัปเดต Product ได้");
      console.error("Failed to update product:", err);
    },
  });

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      updateMutation.mutate(values);
    } catch {
      // Form validation failed
    }
  };

  return (
    <Modal
      title="แก้ไขข้อมูล Product"
      open={open}
      onCancel={onClose}
      onOk={handleSubmit}
      okText="บันทึก"
      cancelText="ยกเลิก"
      confirmLoading={updateMutation.isPending}
      destroyOnHidden
    >
      <Form
        form={form}
        layout="vertical"
        className="mt-4"
        initialValues={{ is_active: true }}
      >
        <Form.Item
          name="product_name"
          label="ชื่อ Product"
          rules={[{ required: true, message: "กรุณาระบุชื่อ Product" }]}
        >
          <Input placeholder="เช่น E-commerce Platform" />
        </Form.Item>
        <Form.Item
          name="description"
          label="รายละเอียด (Description)"
        >
          <Input.TextArea placeholder="คำอธิบายเพิ่มเติมเกี่ยวกับ Product นี้" rows={3} />
        </Form.Item>
        <Form.Item
          name="is_active"
          label="สถานะการใช้งาน"
          valuePropName="checked"
        >
          <Switch checkedChildren="Active" unCheckedChildren="Inactive" />
        </Form.Item>
      </Form>
    </Modal>
  );
}
