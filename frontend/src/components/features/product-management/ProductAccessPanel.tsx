import { useEffect, useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Alert,
  Button,
  Card,
  Drawer,
  Empty,
  Form,
  Input,
  Modal,
  Select,
  Space,
  Switch,
  Table,
  Tag,
  message,
} from "antd";
import {
  EditOutlined,
  KeyOutlined,
  PlusOutlined,
  UserAddOutlined,
} from "@ant-design/icons";
import type { ProductFeatureOption, ProductOption, ProductProjectOption } from "@/services/product.service";
import {
  productService,
  type ProductAccessMember,
} from "@/services/product.service";
import { userService } from "@/services/user.service";
import { useApiKeyPermissions } from "@/features/logs-connections/hooks/useApiKeyPermissions";
import { useApiKeys } from "@/features/logs-connections/hooks/useApiKeys";
import { ApiKeyTable } from "@/components/features/logs-connections/ApiKeyTable";
import { ApiKeyDetailDrawer } from "@/components/features/logs-connections/ApiKeyDetailDrawer";
import { GenerateApiKeyDrawer } from "@/components/features/logs-connections/GenerateApiKeyDrawer";

import type {
  ApiKeyRecord,
  GeneratedApiKey,
  GenerateApiKeyInput,
  ProductEnvironment,
} from "@/types/apiKey.types";
interface Props {
  product: ProductOption;
  canManage: boolean;
  summaryOnly?: boolean;
  onGenerated: (
    value: GeneratedApiKey,
    environment?: ProductEnvironment,
  ) => void;
}
interface MemberFormValues {
  userId: number;
  roleId: number;
  active: boolean;
}
interface KeyFormValues {
  name: string;
  active: boolean;
}
export function ProductAccessPanel({
  product,
  canManage,
  summaryOnly = false,
  onGenerated,
}: Props) {
  const client = useQueryClient();
  const [keyDetail, setKeyDetail] = useState<ApiKeyRecord>();
  const [keyEditor, setKeyEditor] = useState<ApiKeyRecord>();
  const [generateOpen, setGenerateOpen] = useState(false);
  const [memberEditor, setMemberEditor] = useState<
    ProductAccessMember | "new"
  >();
  const [memberForm] = Form.useForm<MemberFormValues>();
  const [keyForm] = Form.useForm<KeyFormValues>();
  const access = useApiKeyPermissions(canManage ? product.id : undefined);
  const keys = useApiKeys(canManage && access.canView ? product.id : undefined);
  const overview = useQuery({
    queryKey: ["products", product.id, "access-overview"],
    queryFn: () => productService.getAccessOverview(product.id),
    enabled: canManage,
    staleTime: 15_000,
  });
  const users = useQuery({
    queryKey: ["users", "product-members-picker"],
    queryFn: () => userService.getUsers({ page: 1, pageSize: 100 }),
    enabled: canManage,
    staleTime: 60_000,
  });
  const environments = useQuery({
    queryKey: ["products", product.id, "environments"],
    queryFn: () => productService.listEnvironments(product.id),
    enabled: canManage,
    staleTime: 60_000,
  });
  const projects = useQuery({
    queryKey: ["products", product.id, "projects"],
    queryFn: () => productService.listProjects(product.id),
    enabled: canManage && generateOpen,
    staleTime: 60_000,
  });
  const features = useQuery({
    queryKey: ["products", product.id, "features"],
    queryFn: () => productService.listFeatures(product.id),
    enabled: canManage && generateOpen,
    staleTime: 60_000,
  });
  const envs: ProductEnvironment[] = useMemo(
    () =>
      (environments.data ?? []).map((item) => ({
        environmentId: item.environment_id,
        productId: product.id,
        code: item.environment_code,
        name: item.environment_name,
      })),
    [environments.data, product.id],
  );
  const roles = overview.data?.roles.filter((role) => role.is_active) ?? [];
  const memberIds = new Set(
    (overview.data?.members ?? []).map((member) => member.user_id),
  );
  const availableUsers = (users.data?.data ?? []).filter(
    (user) =>
      !memberIds.has(Number(user.id)) ||
      (typeof memberEditor !== "string" &&
        memberEditor?.user_id === Number(user.id)),
  );
  const invalidateAccess = () =>
    void client.invalidateQueries({
      queryKey: ["products", product.id, "access-overview"],
    });
  const createMember = useMutation({
    mutationFn: (value: MemberFormValues) =>
      productService.createMembership(product.id, {
        user_id: value.userId,
        role_id: value.roleId,
      }),
    onSuccess: () => {
      message.success("Member added");
      setMemberEditor(undefined);
      invalidateAccess();
    },
    onError: () => message.error("Could not add member"),
  });
  const updateMember = useMutation({
    mutationFn: (value: MemberFormValues & { membershipId: number }) =>
      productService.updateMembership(product.id, value.membershipId, {
        role_id: value.roleId,
        is_active: value.active,
      }),
    onSuccess: () => {
      message.success("Membership updated");
      setMemberEditor(undefined);
      invalidateAccess();
    },
    onError: () => message.error("Could not update membership"),
  });
  const deleteMember = useMutation({
    mutationFn: (membershipId: number) =>
      productService.deleteMembership(product.id, membershipId),
    onSuccess: () => {
      message.success("Member removed");
      invalidateAccess();
    },
    onError: () => message.error("Could not remove member"),
  });
  useEffect(() => {
    if (memberEditor === "new") memberForm.resetFields();
    else if (memberEditor)
      memberForm.setFieldsValue({
        userId: memberEditor.user_id,
        roleId: memberEditor.role_id,
        active: memberEditor.is_active,
      });
  }, [memberEditor, memberForm]);
  useEffect(() => {
    if (keyEditor)
      keyForm.setFieldsValue({
        name: keyEditor.name,
        active: keyEditor.active,
      });
  }, [keyEditor, keyForm]);
  const productEnvironment = (id?: number) =>
    envs.find((environment) => environment.environmentId === id);
  const submitMember = (value: MemberFormValues) => {
    if (memberEditor === "new") createMember.mutate(value);
    else if (memberEditor)
      updateMember.mutate({
        ...value,
        membershipId: memberEditor.membership_id,
      });
  };
  const submitKey = (value: KeyFormValues) => {
    if (keyEditor)
      keys.update.mutate(
        {
          keyId: keyEditor.keyId,
          input: { keyName: value.name, active: value.active },
        },
        { onSuccess: () => setKeyEditor(undefined) },
      );
  };
  const confirmRemove = (member: ProductAccessMember) =>
    Modal.confirm({
      title: `Remove ${member.full_name || member.email}?`,
      content: "This user will lose access to the product.",
      okText: "Remove",
      okButtonProps: { danger: true },
      onOk: () => deleteMember.mutate(member.membership_id),
    });
  if (!canManage)
    return (
      <Alert
        type="info"
        showIcon
        message="Member access"
        description="You can see this product in your scope, but API keys and owner/member management are available only to the product Owner."
      />
    );
  if (summaryOnly) {
    const members = overview.data?.members ?? [];
    const owner = members.find(
      (member) => member.role_code.toLowerCase() === "owner",
    );
    const memberCount = members.filter(
      (member) => member.role_code.toLowerCase() !== "owner",
    ).length;
    return (
      <div className="product-access-summary">
        <Card className="product-access-summary-card">
          <div className="product-access-summary-icon">
            <KeyOutlined />
          </div>
          <div className="product-access-summary-content">
            <span className="product-access-summary-label">API keys</span>
            <strong>
              {access.canView
                ? keys.query.isLoading
                  ? "—"
                  : (keys.query.data?.length ?? 0)
                : "Restricted"}
            </strong>
            <small>
              {access.canView
                ? "credentials connected to this product"
                : "access is not available for your role"}
            </small>
          </div>
        </Card>
        <Card className="product-access-summary-card">
          <div className="product-access-summary-icon">
            <UserAddOutlined />
          </div>
          <div className="product-access-summary-content">
            <span className="product-access-summary-label">
              Owner &amp; members
            </span>
            <strong>{overview.isLoading ? "—" : members.length}</strong>
            <small>
              {owner
                ? "Owner: " + (owner.full_name || owner.email)
                : memberCount + " members configured"}
            </small>
          </div>
        </Card>
      </div>
    );
  }
  return (
    <div className="product-access-panel">
      <Card
        title="API Keys"
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            disabled={!access.canCreate}
            onClick={() => setGenerateOpen(true)}
          >
            Create API key
          </Button>
        }
      >
        {access.canView ? (
          <ApiKeyTable
            keys={keys.query.data ?? []}
            products={[product]}
            environments={envs}
            loading={keys.query.isLoading}
            canUpdate={access.canUpdate}
            canDelete={access.canDelete}
            onOpen={setKeyDetail}
            onEdit={setKeyEditor}
            onActive={(key, active) =>
              keys.setActive.mutate({ keyId: key.keyId, active })
            }
            onRevoke={(key) =>
              Modal.confirm({
                title: `Revoke ${key.name}?`,
                content: "Revocation is permanent.",
                okText: "Revoke key",
                okButtonProps: { danger: true },
                onOk: () => keys.revoke.mutate(key.keyId),
              })
            }
          />
        ) : (
          <Alert
            type="warning"
            showIcon
            message="API key access denied"
            description="Your current product permissions do not allow API key access."
          />
        )}
      </Card>
      <Card
        title="Owner and members"
        extra={
          <Button
            icon={<UserAddOutlined />}
            onClick={() => setMemberEditor("new")}
            disabled={!roles.length}
          >
            Add member
          </Button>
        }
      >
        {overview.isError ? (
          <Alert
            type="error"
            showIcon
            message="Could not load product access"
            description="Try refreshing the product details."
          />
        ) : (
          <Table<ProductAccessMember>
            rowKey="membership_id"
            loading={overview.isLoading}
            dataSource={overview.data?.members ?? []}
            pagination={false}
            scroll={{ x: 760 }}
            locale={{
              emptyText: (
                <Empty
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                  description="No members configured"
                />
              ),
            }}
            columns={[
              {
                title: "User",
                render: (_, member) => (
                  <div>
                    <strong>{member.full_name || member.email}</strong>
                    <div>{member.email}</div>
                  </div>
                ),
              },
              {
                title: "Role",
                dataIndex: "role_name",
                render: (value, member) => (
                  <Tag
                    color={
                      member.role_code.toLowerCase() === "owner"
                        ? "gold"
                        : "blue"
                    }
                  >
                    {value}
                  </Tag>
                ),
              },
              { title: "Access", dataIndex: "access_level" },
              {
                title: "Status",
                dataIndex: "is_active",
                render: (value) => (
                  <Tag color={value ? "success" : "default"}>
                    {value ? "Active" : "Inactive"}
                  </Tag>
                ),
              },
              {
                title: "Actions",
                width: 130,
                render: (_, member) => (
                  <Space>
                    <Button
                      type="text"
                      icon={<EditOutlined />}
                      aria-label={`Edit ${member.full_name || member.email}`}
                      onClick={() => setMemberEditor(member)}
                    />
                    <Button
                      type="link"
                      danger
                      onClick={() => confirmRemove(member)}
                    >
                      Remove
                    </Button>
                  </Space>
                ),
              },
            ]}
          />
        )}
      </Card>
      <GenerateApiKeyDrawer
        open={generateOpen}
        productName={product.name}
        environments={envs}
        projects={(projects.data ?? []) as ProductProjectOption[]}
        features={(features.data ?? []) as ProductFeatureOption[]}
        loading={keys.create.isPending}
        onClose={() => setGenerateOpen(false)}
        onSubmit={(input: GenerateApiKeyInput) =>
          keys.create.mutate(input, {
            onSuccess: (value) => {
              setGenerateOpen(false);
              onGenerated(value, productEnvironment(value.environmentId));
            },
          })
        }
      />
      <ApiKeyDetailDrawer
        value={keyDetail}
        product={product}
        environment={productEnvironment(keyDetail?.environmentId)}
        onClose={() => setKeyDetail(undefined)}
      />
      <Modal
        open={Boolean(keyEditor)}
        title="Edit API key"
        confirmLoading={keys.update.isPending}
        onCancel={() => setKeyEditor(undefined)}
        onOk={() => void keyForm.submit()}
      >
        <Form form={keyForm} layout="vertical" onFinish={submitKey}>
          <Form.Item
            name="name"
            label="Key name"
            rules={[{ required: true, min: 2 }]}
          >
            <Input />
          </Form.Item>
          <Form.Item name="active" label="Status" valuePropName="checked">
            <Switch checkedChildren="Active" unCheckedChildren="Disabled" />
          </Form.Item>
        </Form>
      </Modal>
      <Drawer
        open={Boolean(memberEditor)}
        title={memberEditor === "new" ? "Add member" : "Edit membership"}
        size={440}
        destroyOnHidden
        onClose={() => setMemberEditor(undefined)}
        extra={
          <Button
            type="primary"
            loading={createMember.isPending || updateMember.isPending}
            onClick={() => void memberForm.submit()}
          >
            Save
          </Button>
        }
      >
        <Form form={memberForm} layout="vertical" onFinish={submitMember}>
          <Form.Item name="userId" label="User" rules={[{ required: true }]}>
            <Select
              disabled={memberEditor !== "new"}
              showSearch
              optionFilterProp="label"
              options={availableUsers.map((user) => ({
                value: Number(user.id),
                label: `${user.fullName} (${user.email})`,
              }))}
            />
          </Form.Item>
          <Form.Item name="roleId" label="Role" rules={[{ required: true }]}>
            <Select
              options={roles.map((role) => ({
                value: role.role_id,
                label: `${role.role_name} (${role.access_level})`,
              }))}
            />
          </Form.Item>
          {memberEditor !== "new" && (
            <Form.Item name="active" label="Status" valuePropName="checked">
              <Switch checkedChildren="Active" unCheckedChildren="Inactive" />
            </Form.Item>
          )}
        </Form>
      </Drawer>
    </div>
  );
}
