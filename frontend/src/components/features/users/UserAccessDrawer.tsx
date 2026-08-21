import { useEffect, useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Alert,
  Button,
  Card,
  Collapse,
  Drawer,
  Empty,
  Popconfirm,
  Select,
  Space,
  Switch,
  Tag,
  Typography,
  message,
} from "antd";
import { SaveOutlined, SafetyCertificateOutlined } from "@ant-design/icons";
import { FeatureScopeTree } from "@/components/features/users/FeatureScopeTree";
import { productService } from "@/services/product.service";
import type {
  ProductAccessMember,
  ProductAccessOverview,
  ProductOption,
} from "@/services/product.service";
import { userAccessService, type CreateMembershipScopeInput } from "@/services/user-access.service";
import type { User } from "@/types";

const { Text, Title } = Typography;

export interface ProductAccessData {
  product: ProductOption;
  overview?: ProductAccessOverview;
  loading: boolean;
  error: boolean;
}

interface Props {
  open: boolean;
  user?: User;
  products: ProductAccessData[];
  onClose: () => void;
  onChanged: () => void;
}

function ScopeEditor({ product, member }: { product: ProductOption; member: ProductAccessMember }) {
  const client = useQueryClient();
  const isOwner = member.role_code.toLowerCase() === "owner";
  const scopeKey = ["users", member.user_id, "products", product.id, "scopes"];
  const [projectIds, setProjectIds] = useState<number[]>([]);
  const [featureIds, setFeatureIds] = useState<number[]>([]);
  const scopes = useQuery({
    queryKey: scopeKey,
    queryFn: () => userAccessService.listScopes(product.id, member.membership_id),
  });
  const projects = useQuery({
    queryKey: ["products", product.id, "projects"],
    queryFn: () => productService.listProjects(product.id),
    enabled: !isOwner,
    staleTime: 60_000,
  });
  const features = useQuery({
    queryKey: ["products", product.id, "features"],
    queryFn: () => productService.listFeatures(product.id),
    enabled: !isOwner,
    staleTime: 60_000,
  });

  useEffect(() => {
    if (!scopes.data) return;
    // Server scopes initialize the editable local draft after the query resolves.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setProjectIds(scopes.data.filter((scope) => scope.scope_level === "PROJECT" && scope.project_id).map((scope) => scope.project_id!));
    setFeatureIds(scopes.data.filter((scope) => scope.scope_level === "CATEGORY" && scope.category_id).map((scope) => scope.category_id!));
  }, [scopes.data]);

  const desiredScopes = useMemo<CreateMembershipScopeInput[]>(() => {
    if (isOwner) return [{ scope_level: "PRODUCT" }];
    return [
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
  }, [featureIds, features.data, isOwner, projectIds]);
  const isDirty = useMemo(() => {
    if (!scopes.data) return false;
    const keyOf = (scope: { scope_level: string; project_id?: number; category_id?: number }) =>
      `${scope.scope_level}:${scope.project_id ?? ""}:${scope.category_id ?? ""}`;
    const current = scopes.data.map(keyOf).sort();
    const desired = desiredScopes.map(keyOf).sort();
    return current.length !== desired.length || current.some((key, index) => key !== desired[index]);
  }, [desiredScopes, scopes.data]);

  const save = useMutation({
    mutationFn: () => userAccessService.replaceScopes(product.id, member.membership_id, desiredScopes),
    onSuccess: () => {
      message.success(`บันทึกขอบเขตของ ${product.name} แล้ว`);
      void client.invalidateQueries({ queryKey: scopeKey });
    },
    onError: () => message.error("บันทึกขอบเขตไม่สำเร็จ"),
  });

  if (isOwner) {
    return (
      <Alert
        type="success"
        showIcon
        message="Owner · Full Product Access"
        description="Owner เข้าถึงทุก Project และ Feature ใน Product นี้โดยอัตโนมัติ ไม่ต้องเลือก Scope เพิ่ม"
      />
    );
  }

  return (
    <div className="user-scope-editor">
      {scopes.isError ? <Alert type="error" showIcon message="โหลดขอบเขตไม่สำเร็จ" /> : null}
      <div>
        <Text strong>เข้าถึงทั้ง Project</Text>
        <Select
          mode="multiple"
          className="w-full"
          loading={projects.isLoading || scopes.isLoading}
          value={projectIds}
          onChange={setProjectIds}
          placeholder="เลือก Project ได้หลายรายการ"
          options={(projects.data ?? []).map((project) => ({ label: project.project_name, value: project.project_id }))}
        />
      </div>
      <div className="bulk-feature-label">
        <Text strong>หรือเลือก Feature หลายรายการ</Text>
        <Text type="secondary">Feature ด้านบนในเส้นทางเดียวกันจะถูกเลือกให้อัตโนมัติ</Text>
      </div>
      <FeatureScopeTree
        projects={projects.data ?? []}
        features={features.data ?? []}
        value={featureIds}
        onChange={setFeatureIds}
        disabled={scopes.isLoading || features.isLoading}
      />
      {isDirty && (
        <Alert
          type="warning"
          showIcon
          message="การเปลี่ยนแปลงยังไม่ถูกบันทึก"
          description="ตรวจสอบ Project และ Feature ที่เลือก แล้วกดบันทึกด้านล่างเพื่อใช้งานสิทธิ์ใหม่"
        />
      )}
      <Button
        type="primary"
        icon={<SaveOutlined />}
        loading={save.isPending}
        disabled={!isDirty || scopes.isLoading}
        onClick={() => save.mutate()}
      >
        บันทึกขอบเขตการเข้าถึง
      </Button>
    </div>
  );
}

export function UserAccessDrawer({ open, user, products, onClose, onChanged }: Props) {
  const client = useQueryClient();
  const assigned = useMemo(
    () => products.flatMap(({ product, overview }) => {
      const member = overview?.members.find((item) => String(item.user_id) === user?.id);
      return member ? [{ product, overview: overview!, member }] : [];
    }),
    [products, user?.id],
  );
  const updateMembership = useMutation({
    mutationFn: ({ productId, membershipId, roleId, active }: { productId: number; membershipId: number; roleId?: number; active?: boolean }) =>
      productService.updateMembership(productId, membershipId, { role_id: roleId, is_active: active }),
    onSuccess: (_, variables) => {
      message.success("อัปเดต Product Role แล้ว");
      void client.invalidateQueries({
        queryKey: ["users", Number(user?.id), "products", variables.productId, "scopes"],
      });
      onChanged();
    },
    onError: () => message.error("ไม่สามารถอัปเดต Product Role ได้"),
  });
  const removeMembership = useMutation({
    mutationFn: ({ productId, membershipId }: { productId: number; membershipId: number }) =>
      productService.deleteMembership(productId, membershipId),
    onSuccess: () => {
      message.success("นำผู้ใช้ออกจาก Product แล้ว");
      onChanged();
    },
    onError: () => message.error("ไม่สามารถนำผู้ใช้ออกจาก Product ได้"),
  });

  return (
    <Drawer title="Product & Feature Access" open={open} onClose={onClose} width={760}>
      {user && (
        <div className="user-access-heading">
          <SafetyCertificateOutlined />
          <div><Title level={5}>{user.fullName}</Title><Text type="secondary">{user.email}</Text></div>
        </div>
      )}
      <Alert
        type="info"
        showIcon
        message="ต้องการเพิ่มหลาย User?"
        description="ใช้ปุ่ม “เพิ่ม User เข้า Product” บนหน้าหลัก เพื่อเลือก Product หนึ่งรายการและเพิ่ม User หลายคนพร้อมกัน"
        className="mb-4"
      />
      {products.some((item) => item.error) && <Alert type="warning" showIcon message="ข้อมูลสิทธิ์บาง Product โหลดไม่สำเร็จ" className="mb-4" />}
      {!assigned.length ? (
        <Card><Empty description="ผู้ใช้นี้ยังไม่ได้รับมอบหมาย Product" /></Card>
      ) : (
        <Collapse
          className="user-access-products"
          items={assigned.map(({ product, overview, member }) => ({
            key: member.membership_id,
            label: (
              <div className="user-product-label">
                <strong>{product.name}</strong>
                <Tag color={member.is_active ? "success" : "default"}>{member.is_active ? "Active" : "Deactive"}</Tag>
                <Tag color={member.role_code.toLowerCase() === "owner" ? "gold" : "blue"}>{member.role_name}</Tag>
              </div>
            ),
            children: (
              <Space direction="vertical" size="large" className="w-full">
                <div className="user-membership-controls">
                  <Select
                    aria-label={`Role สำหรับ ${product.name}`}
                    value={member.role_id}
                    options={overview.roles.filter((role) => role.is_active).map((role) => ({ label: `${role.role_name} (${role.access_level})`, value: role.role_id }))}
                    onChange={(roleId) => updateMembership.mutate({ productId: product.id, membershipId: member.membership_id, roleId })}
                  />
                  <Space><Text>Active</Text><Switch checked={member.is_active} onChange={(active) => updateMembership.mutate({ productId: product.id, membershipId: member.membership_id, active })} /></Space>
                  <Popconfirm title={`นำ ${user?.fullName} ออกจาก ${product.name}?`} okText="นำออก" cancelText="ยกเลิก" onConfirm={() => removeMembership.mutate({ productId: product.id, membershipId: member.membership_id })}>
                    <Button danger>นำออกจาก Product</Button>
                  </Popconfirm>
                </div>
                <ScopeEditor product={product} member={member} />
              </Space>
            ),
          }))}
        />
      )}
    </Drawer>
  );
}
