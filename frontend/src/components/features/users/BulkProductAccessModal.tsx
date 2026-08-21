import { useEffect, useMemo, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { Alert, Button, Form, Modal, Select, Space, Spin, Typography, message } from "antd";
import { SaveOutlined, TeamOutlined } from "@ant-design/icons";
import { FeatureScopeTree } from "@/components/features/users/FeatureScopeTree";
import type { ProductAccessData } from "@/components/features/users/UserAccessDrawer";
import { productService } from "@/services/product.service";
import { userAccessService, type CreateMembershipScopeInput } from "@/services/user-access.service";
import { userService } from "@/services/user.service";

const { Text } = Typography;

interface Props {
  open: boolean;
  products: ProductAccessData[];
  onClose: () => void;
  onSaved: () => void;
}

interface FormValues {
  productId: number;
  userIds: number[];
  roleId: number;
}

export function BulkProductAccessModal({ open, products, onClose, onSaved }: Props) {
  const [form] = Form.useForm<FormValues>();
  const productId = Form.useWatch("productId", form);
  const roleId = Form.useWatch("roleId", form);
  const [projectIds, setProjectIds] = useState<number[]>([]);
  const [featureIds, setFeatureIds] = useState<number[]>([]);
  const productAccess = products.find((item) => item.product.id === productId);
  const selectedRole = productAccess?.overview?.roles.find((role) => role.role_id === roleId);
  const isOwner = selectedRole?.role_code.toLowerCase() === "owner";
  const users = useQuery({
    queryKey: ["users", "bulk-product-picker"],
    queryFn: () => userService.getUsers({ page: 1, pageSize: 100 }),
    enabled: open,
    staleTime: 30_000,
  });
  const projects = useQuery({
    queryKey: ["products", productId, "projects"],
    queryFn: () => productService.listProjects(productId),
    enabled: Boolean(productId),
    staleTime: 60_000,
  });
  const features = useQuery({
    queryKey: ["products", productId, "features"],
    queryFn: () => productService.listFeatures(productId),
    enabled: Boolean(productId),
    staleTime: 60_000,
  });
  const existingMemberIds = useMemo(
    () => new Set((productAccess?.overview?.members ?? []).map((member) => member.user_id)),
    [productAccess?.overview?.members],
  );

  useEffect(() => {
    if (!open) return;
    form.resetFields();
    // Opening the modal starts a fresh bulk assignment draft.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setProjectIds([]);
    setFeatureIds([]);
  }, [form, open]);

  const save = useMutation({
    mutationFn: async (values: FormValues) => {
      const memberships = await productService.createMemberships(values.productId, values.userIds, values.roleId);
      const desired: CreateMembershipScopeInput[] = isOwner
        ? [{ scope_level: "PRODUCT" }]
        : [
            ...projectIds.map((projectId) => ({ scope_level: "PROJECT" as const, project_id: projectId })),
            ...featureIds
              .map((categoryId) => (features.data ?? []).find((feature) => feature.category_id === categoryId))
              .filter((feature) => feature && !projectIds.includes(feature.project_id))
              .map((feature) => ({
                scope_level: "CATEGORY" as const,
                project_id: feature!.project_id,
                category_id: feature!.category_id,
              })),
          ];
      await Promise.all(
        memberships.map((membership) =>
          userAccessService.replaceScopes(values.productId, membership.membership_id, desired),
        ),
      );
      return memberships.length;
    },
    onSuccess: (count) => {
      message.success(`บันทึก Product Access ให้ ${count} User แล้ว`);
      onSaved();
      onClose();
    },
    onError: () => message.error("บันทึก Product Access ไม่สำเร็จ กรุณาตรวจสอบสิทธิ์แล้วลองอีกครั้ง"),
  });

  const submit = (values: FormValues) => {
    if (!isOwner && projectIds.length === 0 && featureIds.length === 0) {
      message.warning("กรุณาเลือก Project หรือ Feature อย่างน้อย 1 รายการ");
      return;
    }
    save.mutate(values);
  };

  return (
    <Modal
      title={<Space><TeamOutlined />เพิ่มหลาย User เข้า Product</Space>}
      open={open}
      width={780}
      onCancel={onClose}
      destroyOnHidden
      footer={[
        <Button key="cancel" onClick={onClose} disabled={save.isPending}>ยกเลิก</Button>,
        <Button key="save" type="primary" icon={<SaveOutlined />} loading={save.isPending} onClick={() => form.submit()}>
          บันทึกการมอบหมายและสิทธิ์
        </Button>,
      ]}
    >
      <Form<FormValues> form={form} layout="vertical" onFinish={submit}>
        <Form.Item name="productId" label="1. เลือก Product" rules={[{ required: true, message: "กรุณาเลือก Product" }]}>
          <Select
            showSearch
            optionFilterProp="label"
            placeholder="เลือก Product หนึ่งรายการ"
            options={products.map(({ product }) => ({ label: product.name, value: product.id }))}
            onChange={() => {
              form.setFieldsValue({ roleId: undefined, userIds: [] });
              setProjectIds([]);
              setFeatureIds([]);
            }}
          />
        </Form.Item>
        <Form.Item name="userIds" label="2. เลือก User ได้หลายคน" rules={[{ required: true, message: "กรุณาเลือก User อย่างน้อย 1 คน" }]}>
          <Select
            mode="multiple"
            showSearch
            optionFilterProp="label"
            loading={users.isLoading}
            disabled={!productId}
            placeholder="ค้นหาและเลือก User"
            maxTagCount="responsive"
            options={(users.data?.data ?? []).map((user) => ({
              label: `${user.fullName} · ${user.email}${existingMemberIds.has(Number(user.id)) ? " · อยู่ใน Product แล้ว" : ""}`,
              value: Number(user.id),
            }))}
          />
        </Form.Item>
        <Form.Item name="roleId" label="3. เลือก Product Role" rules={[{ required: true, message: "กรุณาเลือก Role" }]}>
          <Select
            disabled={!productId}
            loading={productAccess?.loading}
            placeholder="เลือก Role สำหรับ User ที่เลือกทั้งหมด"
            options={(productAccess?.overview?.roles ?? []).filter((role) => role.is_active).map((role) => ({
              label: `${role.role_name} (${role.access_level})`,
              value: role.role_id,
            }))}
            onChange={() => {
              setProjectIds([]);
              setFeatureIds([]);
            }}
          />
        </Form.Item>

        {roleId && isOwner && (
          <Alert
            type="success"
            showIcon
            message="Owner ได้ Full Access อัตโนมัติ"
            description="ระบบจะให้สิทธิ์ทั้ง Product โดยไม่ต้องเลือก Project หรือ Feature เพิ่ม"
          />
        )}
        {roleId && !isOwner && (
          <div className="bulk-access-scope">
            <Text strong>4. กำหนดขอบเขตให้ User ที่เลือกทั้งหมด</Text>
            <Form.Item label="เข้าถึงทั้ง Project (เลือกได้หลายรายการ)">
              <Select
                mode="multiple"
                loading={projects.isLoading}
                value={projectIds}
                onChange={setProjectIds}
                placeholder="ไม่บังคับ หากต้องการเลือกเฉพาะ Feature"
                options={(projects.data ?? []).map((project) => ({ label: project.project_name, value: project.project_id }))}
              />
            </Form.Item>
            <div className="bulk-feature-label">
              <Text strong>หรือเลือก Feature หลายรายการ</Text>
              <Text type="secondary">เลือก Feature ลูกแล้วระบบจะแสดง parent ที่จำเป็นให้อัตโนมัติ</Text>
            </div>
            {projects.isLoading || features.isLoading ? (
              <div className="bulk-scope-loading"><Spin /></div>
            ) : (
              <FeatureScopeTree
                projects={projects.data ?? []}
                features={features.data ?? []}
                value={featureIds}
                onChange={setFeatureIds}
              />
            )}
          </div>
        )}
      </Form>
    </Modal>
  );
}
