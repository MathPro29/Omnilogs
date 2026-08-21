import { useEffect } from "react";
import { Form, Input, Modal, Select } from "antd";
import type { User, UserStatus } from "@/types";

export interface UserFormValues {
  username?: string;
  fullName: string;
  email: string;
  phone?: string;
  password?: string;
  role: string;
  status: UserStatus;
  permissions?: string[];
}

interface Props {
  open: boolean;
  user?: User;
  loading: boolean;
  onCancel: () => void;
  onSubmit: (values: UserFormValues) => void;
}

const OMNILOGS_FEATURES = [
  { label: "Logs Explorer (ค้นหาและสืบค้น Log)", value: "explore-logs" },
  { label: "Audit Logs (ประวัติการใช้งานระบบ)", value: "audit-logs" },
  { label: "Products (จัดการผลิตภัณฑ์และสภาพแวดล้อม)", value: "products" },
  { label: "User Management (จัดการผู้ใช้งานและสิทธิ์)", value: "users" },
  { label: "Connections (จัดการการเชื่อมต่อข้อมูล)", value: "connections" },
];

const ALL_FEATURE_VALUES = OMNILOGS_FEATURES.map((f) => f.value);

const roleOptions = [
  { label: "GOD — จัดการทั้งแพลตฟอร์ม", value: "god" },
  { label: "Owner — ดูแลผลิตภัณฑ์", value: "owner" },
  { label: "Superadmin — จัดการสิทธิ์ที่ได้รับมอบหมาย", value: "superadmin" },
  { label: "Admin — ใช้งานตาม Product Membership", value: "admin" },
];

export function UserFormModal({
  open,
  user,
  loading,
  onCancel,
  onSubmit,
}: Props) {
  const [form] = Form.useForm<UserFormValues>();

  useEffect(() => {
    if (!open) return;
    if (user) {
      form.setFieldsValue({
        fullName: user.fullName,
        email: user.email,
        phone: user.phone,
        role: user.roles[0]?.name || "admin",
        status: user.status,
        permissions:
          user.permissions && user.permissions.length > 0
            ? user.permissions
            : ALL_FEATURE_VALUES,
      });
    } else {
      form.resetFields();
      form.setFieldsValue({
        role: "admin",
        status: "active",
        permissions: ALL_FEATURE_VALUES,
      });
    }
  }, [form, open, user]);

  return (
    <Modal
      title={user ? "แก้ไขผู้ใช้งาน" : "เพิ่มผู้ใช้งาน"}
      open={open}
      width={640}
      okText="บันทึก"
      cancelText="ยกเลิก"
      confirmLoading={loading}
      onCancel={onCancel}
      onOk={() => form.submit()}
      destroyOnHidden
    >
      <Form<UserFormValues> form={form} layout="vertical" onFinish={onSubmit}>
        {!user && (
          <Form.Item
            name="username"
            label="ชื่อผู้ใช้"
            rules={[
              { required: true, message: "กรุณาระบุชื่อผู้ใช้" },
              { min: 3, message: "ชื่อผู้ใช้ต้องมีอย่างน้อย 3 ตัวอักษร" },
            ]}
          >
            <Input autoComplete="username" placeholder="เช่น somchai" />
          </Form.Item>
        )}
        <Form.Item
          name="fullName"
          label="ชื่อ-นามสกุล"
          rules={[{ required: true, message: "กรุณาระบุชื่อ-นามสกุล" }]}
        >
          <Input placeholder="ชื่อที่แสดงในระบบ" />
        </Form.Item>
        <Form.Item
          name="email"
          label="อีเมล"
          rules={[
            { required: true, message: "กรุณาระบุอีเมล" },
            { type: "email", message: "รูปแบบอีเมลไม่ถูกต้อง" },
          ]}
        >
          <Input
            type="email"
            autoComplete="email"
            placeholder="example@mail.com"
          />
        </Form.Item>
        <Form.Item
          name="phone"
          label="เบอร์โทรศัพท์"
          rules={[
            {
              pattern: /^[0-9+ -]{8,20}$/,
              message: "รูปแบบเบอร์โทรศัพท์ไม่ถูกต้อง",
            },
          ]}
        >
          <Input type="tel" autoComplete="tel" />
        </Form.Item>
        {!user && (
          <Form.Item
            name="password"
            label="รหัสผ่านเริ่มต้น"
            rules={[
              { required: true, message: "กรุณาระบุรหัสผ่าน" },
              { min: 8, message: "รหัสผ่านต้องมีอย่างน้อย 8 ตัวอักษร" },
            ]}
          >
            <Input.Password autoComplete="new-password" />
          </Form.Item>
        )}
        <Form.Item
          name="role"
          label="Platform Role"
          rules={[{ required: true }]}
        >
          <Select options={roleOptions} />
        </Form.Item>
        {user && (
          <Form.Item name="status" label="สถานะ" rules={[{ required: true }]}>
            <Select
              options={[
                { label: "Active", value: "active" },
                { label: "Deactive", value: "inactive" },
              ]}
            />
          </Form.Item>
        )}

        {/* <div style={{ marginTop: 16, marginBottom: 12 }}>
          <label style={{ fontWeight: 600, display: "block", marginBottom: 8 }}>
            สิทธิ์การเข้าถึง Feature (Omnilogs Features Access)
          </label>
          <div
            style={{
              padding: "12px 16px",
              background: "rgba(255, 255, 255, 0.03)",
              borderRadius: 8,
              border: "1px solid rgba(255, 255, 255, 0.08)",
            }}
          >
            <Checkbox
              indeterminate={isIndeterminate}
              onChange={handleCheckAllChange}
              checked={isAllChecked}
              style={{ fontWeight: 600 }}
            >
              เลือกทั้งหมด (Select All)
            </Checkbox>
            <Divider style={{ margin: "10px 0" }} />
            <Form.Item name="permissions" noStyle>
              <Checkbox.Group
                options={OMNILOGS_FEATURES}
                style={{ display: "flex", flexDirection: "column", gap: 10 }}
              />
            </Form.Item>
          </div>
        </div> */}
      </Form>
    </Modal>
  );
}
