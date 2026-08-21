import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  ApartmentOutlined,
  CheckCircleOutlined,
  DeleteOutlined,
  DownOutlined,
  EditOutlined,
  InfoCircleOutlined,
  PlusOutlined,
  SettingOutlined,
  TeamOutlined,
  UserAddOutlined,
  CrownOutlined,
} from "@ant-design/icons";
import {
  Button,
  Card,
  Checkbox,
  Descriptions,
  Divider,
  Drawer,
  Empty,
  Form,
  Input,
  List,
  Modal,
  Select,
  Space,
  Tag,
  message,
} from "antd";
import {
  productService,
  type ProductOption,
  type ProductAccessRole,
} from "@/services/product.service";
import { userService } from "@/services/user.service";
import {
  userAccessService,
  type AccessScopeLevel,
  type CreateMembershipScopeInput,
} from "@/services/user-access.service";
import type { GeneratedApiKey, ProductEnvironment } from "@/types/apiKey.types";
import { FeatureScopeTree } from "@/components/features/users/FeatureScopeTree";
import { ProductAccessPanel } from "./ProductAccessPanel";
import { ProductPayloadRoutingPanel } from "./ProductPayloadRoutingPanel";

interface ProductDetailsDrawerProps {
  open: boolean;
  product?: ProductOption;
  canManage: boolean;
  onClose: () => void;
  onAfterOpenChange: (open: boolean) => void;
  onContinueSetup: (product: ProductOption) => void;
  onEdit?: (product: ProductOption) => void;
  onDelete?: (product: ProductOption) => void;
  onDeleteProduct?: (productId: number) => Promise<void>;
  onGenerated: (
    value: GeneratedApiKey,
    environment?: ProductEnvironment,
  ) => void;
}

export function ProductDetailsDrawer(props: ProductDetailsDrawerProps) {
  const product = props.product;
  const queryClient = useQueryClient();
  const [isAddUserModalOpen, setIsAddUserModalOpen] = useState(false);
  const [selectedMembershipIds, setSelectedMembershipIds] = useState<number[]>(
    [],
  );
  const [addUserForm] = Form.useForm<{
    userId: number;
    roleId: number;
    scopeLevel: AccessScopeLevel;
    projectIds?: number[];
    featureIds?: number[];
  }>();
  const selectedScopeLevel = Form.useWatch("scopeLevel", addUserForm);
  const [isCreateRoleModalOpen, setIsCreateRoleModalOpen] = useState(false);
  const [roleForm] = Form.useForm<{ roleName: string }>();
  const [editingRole, setEditingRole] = useState<ProductAccessRole | null>(
    null,
  );
  const [isEditRoleModalOpen, setIsEditRoleModalOpen] = useState(false);
  const [editRoleForm] = Form.useForm<{ roleName: string }>();
  const [openInfo, setOpenInfo] = useState(true);
  const [openUsers, setOpenUsers] = useState(true);
  const [openEnv, setOpenEnv] = useState(true);
  const [openAccess, setOpenAccess] = useState(true);

  const members = useQuery({
    queryKey: ["products", product?.id, "access-overview"],
    queryFn: () => productService.getAccessOverview(product!.id),
    enabled: Boolean(product && props.canManage),
    staleTime: 15_000,
  });

  const users = useQuery({
    queryKey: ["users", "product-members-picker"],
    queryFn: () => userService.getUsers({ page: 1, pageSize: 100 }),
    enabled: Boolean(product && props.canManage && isAddUserModalOpen),
    staleTime: 60_000,
  });

  const projects = useQuery({
    queryKey: ["products", product?.id, "projects"],
    queryFn: () => productService.listProjects(product!.id),
    enabled: Boolean(product && props.canManage && isAddUserModalOpen),
    staleTime: 60_000,
  });

  const features = useQuery({
    queryKey: ["products", product?.id, "features"],
    queryFn: () => productService.listFeatures(product!.id),
    enabled: Boolean(product && props.canManage && isAddUserModalOpen),
    staleTime: 60_000,
  });

  const createMembers = useMutation({
    mutationFn: async (values: {
      userId: number;
      roleId: number;
      scopeLevel: AccessScopeLevel;
      projectIds?: number[];
      featureIds?: number[];
    }) => {
      const membership = await productService.createMembership(product!.id, {
        user_id: values.userId,
        role_id: values.roleId,
      });
      let scopes: CreateMembershipScopeInput[];
      if (values.scopeLevel === "PRODUCT") {
        scopes = [{ scope_level: "PRODUCT" }];
      } else if (values.scopeLevel === "PROJECT") {
        scopes = (values.projectIds ?? []).map((projectId) => ({
          scope_level: "PROJECT",
          project_id: projectId,
        }));
      } else {
        scopes = (values.featureIds ?? [])
          .map((categoryId) =>
            (features.data ?? []).find(
              (feature) => feature.category_id === categoryId,
            ),
          )
          .filter((feature) => Boolean(feature))
          .map((feature) => ({
            scope_level: "CATEGORY" as const,
            project_id: feature!.project_id,
            category_id: feature!.category_id,
          }));
      }
      await userAccessService.replaceScopes(
        product!.id,
        membership.membership_id,
        scopes,
      );
      return membership;
    },
    onSuccess: () => {
      message.success("User added to product");
      setIsAddUserModalOpen(false);
      addUserForm.resetFields();
      void queryClient.invalidateQueries({
        queryKey: ["products", product?.id, "access-overview"],
      });
    },
    onError: () => {
      message.error("Failed to add user");
      void queryClient.invalidateQueries({
        queryKey: ["products", product?.id, "access-overview"],
      });
    },
  });

  const createRole = useMutation({
    mutationFn: (values: { roleName: string }) =>
      productService.createRole(product!.id, {
        role_name: values.roleName.trim(),
      }),
    onSuccess: () => {
      message.success("Product role created");
      setIsCreateRoleModalOpen(false);
      roleForm.resetFields();
      void queryClient.invalidateQueries({
        queryKey: ["products", product?.id, "access-overview"],
      });
    },
    onError: () => message.error("Failed to create product role"),
  });

  const updateRole = useMutation({
    mutationFn: (values: { roleName: string }) =>
      productService.updateRole(product!.id, editingRole!.role_id, {
        role_name: values.roleName.trim(),
      }),
    onSuccess: () => {
      message.success("Product role updated");
      setIsEditRoleModalOpen(false);
      setEditingRole(null);
      editRoleForm.resetFields();
      void queryClient.invalidateQueries({
        queryKey: ["products", product?.id, "access-overview"],
      });
    },
    onError: () => message.error("Failed to update product role"),
  });

  const handleOpenEditRole = (role: ProductAccessRole) => {
    setEditingRole(role);
    editRoleForm.setFieldsValue({ roleName: role.role_name });
    setIsEditRoleModalOpen(true);
  };

  const deleteRoleMutation = useMutation({
    mutationFn: (roleId: number) =>
      productService.deleteRole(product!.id, roleId),
    onSuccess: () => {
      message.success("Product role deleted");
      void queryClient.invalidateQueries({
        queryKey: ["products", product?.id, "access-overview"],
      });
    },
    onError: () => message.error("Failed to delete product role"),
  });

  const confirmDeleteRole = (role: { role_id: number; role_name: string }) => {
    Modal.confirm({
      title: `Delete role "${role.role_name}"?`,
      content: "Are you sure you want to delete this role from the product?",
      okText: "Delete role",
      okButtonProps: { danger: true },
      onOk: () => deleteRoleMutation.mutate(role.role_id),
    });
  };

  const deleteProductMutation = useMutation({
    mutationFn: async () => {
      if (!product) return;
      if (props.onDeleteProduct) {
        await props.onDeleteProduct(product.id);
      } else {
        await productService.deleteProduct(product.id);
      }
    },
    onSuccess: () => {
      message.success(`Product "${product?.name}" deleted`);
      props.onClose();
      if (product) {
        props.onDelete?.(product);
      }
      void queryClient.invalidateQueries({ queryKey: ["products"] });
      void queryClient.invalidateQueries({ queryKey: ["user-products"] });
    },
    onError: () => message.error("Failed to delete product"),
  });

  const confirmDeleteProduct = () => {
    if (!product) return;
    Modal.confirm({
      title: `Delete product "${product.name}"?`,
      content: (
        <div>
          <p>
            This action cannot be undone. All environments, projects, and access
            configurations under this product will be permanently removed.
          </p>
          <p className="text-xs text-red-500 font-semibold mt-1">
            Product ID: {product.id}
          </p>
        </div>
      ),
      okText: "Delete Product",
      okButtonProps: { danger: true },
      cancelText: "Cancel",
      onOk: () => deleteProductMutation.mutateAsync(),
    });
  };

  const deleteMembers = useMutation({
    mutationFn: async (membershipIds: number[]) => {
      const promises = membershipIds.map((id) =>
        productService.deleteMembership(product!.id, id),
      );
      await Promise.all(promises);
    },
    onSuccess: (_, membershipIds) => {
      message.success(`Successfully removed ${membershipIds.length} user(s)`);
      setSelectedMembershipIds([]);
      void queryClient.invalidateQueries({
        queryKey: ["products", product?.id, "access-overview"],
      });
    },
    onError: () => {
      message.error("Failed to remove selected users");
      void queryClient.invalidateQueries({
        queryKey: ["products", product?.id, "access-overview"],
      });
    },
  });

  const confirmBulkRemove = () => {
    if (selectedMembershipIds.length === 0) return;
    Modal.confirm({
      title: `Remove ${selectedMembershipIds.length} user(s)?`,
      content: "These users will lose access to this product.",
      okText: "Remove All Selected",
      okButtonProps: { danger: true },
      onOk: () => deleteMembers.mutate(selectedMembershipIds),
    });
  };

  const confirmRemove = (member: {
    membership_id: number;
    full_name?: string;
    email: string;
  }) => {
    Modal.confirm({
      title: `Remove ${member.full_name || member.email}?`,
      content: "This user will lose access to this product.",
      okText: "Remove",
      okButtonProps: { danger: true },
      onOk: () => deleteMembers.mutate([member.membership_id]),
    });
  };

  const roles = (members.data?.roles ?? []).filter(
    (r) => r.is_active !== false,
  );
  const environments = product?.product_environments ?? [];
  const activeEnvironmentCount = environments.filter(
    (environment) => environment.is_active !== false,
  ).length;
  const inactiveEnvironmentCount = environments.length - activeEnvironmentCount;
  const memberUserIds = new Set(
    (members.data?.members ?? []).map((m) => m.user_id),
  );
  const availableUsers = (users.data?.data ?? []).filter(
    (u) => !memberUserIds.has(Number(u.id)),
  );

  const removableMembers = (members.data?.members ?? []).filter(
    (u) => u.role_code?.toLowerCase() !== "owner",
  );
  const isAllSelected =
    removableMembers.length > 0 &&
    removableMembers.every((u) =>
      selectedMembershipIds.includes(u.membership_id),
    );
  const isIndeterminate = selectedMembershipIds.length > 0 && !isAllSelected;

  const toggleSelectAll = () => {
    if (isAllSelected) {
      setSelectedMembershipIds([]);
    } else {
      setSelectedMembershipIds(removableMembers.map((u) => u.membership_id));
    }
  };

  const toggleSelectUser = (membershipId: number) => {
    setSelectedMembershipIds((prev) =>
      prev.includes(membershipId)
        ? prev.filter((id) => id !== membershipId)
        : [...prev, membershipId],
    );
  };

  return (
    <Drawer
      open={props.open}
      title={null}
      size={620}
      className="product-details-drawer"
      onClose={props.onClose}
      afterOpenChange={props.onAfterOpenChange}
    >
      {product && (
        <>
          <div className="product-details-hero">
            <div className="product-details-hero-icon" aria-hidden="true">
              <ApartmentOutlined />
            </div>
            <div className="product-details-hero-copy">
              <div className="flex items-center justify-between gap-2">
                <div className="product-details-eyebrow">Product overview</div>
                {props.canManage && (
                  <Space>
                    <Button
                      size="small"
                      icon={<EditOutlined />}
                      onClick={() => props.onContinueSetup(product)}
                    >
                      Edit
                    </Button>
                    <Button
                      size="small"
                      danger
                      icon={<DeleteOutlined />}
                      loading={deleteProductMutation.isPending}
                      onClick={confirmDeleteProduct}
                    >
                      Delete
                    </Button>
                  </Space>
                )}
              </div>
              <h2>{product.name}</h2>
              <code>Product ID: {product.id}</code>
              <p>
                {product.description || "No description has been added yet."}
              </p>
            </div>
          </div>
          <div className="product-details-status-row">
            <Tag
              icon={
                product.is_active !== false ? (
                  <CheckCircleOutlined />
                ) : undefined
              }
              color={product.is_active === false ? "default" : "success"}
            >
              {product.is_active === false ? "Inactive" : "Active"}
            </Tag>
            <Tag color={product.setup_status === "ACTIVE" ? "blue" : "default"}>
              Setup: {product.setup_status || "DRAFT"}
            </Tag>
            <span className="product-details-role">
              {product.is_owner
                ? "Product owner"
                : product.membership_role || "Member"}
            </span>
          </div>
          <div className="product-details-stats" aria-label="Product summary">
            <div>
              <ApartmentOutlined />
              <strong>{product.product_environments?.length ?? 0}</strong>
              <span>Environments</span>
            </div>
            <div>
              <TeamOutlined />
              <strong>{product.is_owner ? "Owner" : "Member"}</strong>
              <span>Access level</span>
            </div>
            <div>
              <SettingOutlined />
              <strong>
                {product.setup_status === "ACTIVE" ? "Ready" : "Pending"}
              </strong>
              <span>Configuration</span>
            </div>
          </div>

          {/* Collapsible Product Information */}
          <Card
            className="product-details-section"
            title={
              <div
                onClick={() => setOpenInfo(!openInfo)}
                className="flex items-center justify-between cursor-pointer select-none w-full py-1 text-zinc-800 hover:text-zinc-950 transition-colors"
              >
                <Space>
                  <InfoCircleOutlined className="text-[#1F8457]" />
                  <span className="font-semibold">Product information</span>
                </Space>
                <DownOutlined
                  className={`transition-transform duration-200 text-xs text-zinc-400 ${openInfo ? "rotate-0" : "-rotate-90"}`}
                />
              </div>
            }
          >
            {openInfo && (
              <Descriptions
                column={{ xs: 1, sm: 2 }}
                size="small"
                items={[
                  {
                    key: "id",
                    label: "Product ID",
                    children: <code>{product.id}</code>,
                  },
                  {
                    key: "code",
                    label: "Product code",
                    children: <code>{product.product_code || "—"}</code>,
                  },
                  {
                    key: "status",
                    label: "Current status",
                    children:
                      product.is_active === false ? "Inactive" : "Active",
                  },
                  {
                    key: "setup",
                    label: "Setup status",
                    children: product.setup_status || "DRAFT",
                  },
                  {
                    key: "role",
                    label: "Your access",
                    children: product.is_owner
                      ? "Owner"
                      : product.membership_role || "Member",
                  },
                ]}
              />
            )}
          </Card>

          {/* Collapsible Environments */}
          <section
            className="product-details-section product-details-environments"
            aria-labelledby="product-environments-heading"
          >
            <button
              type="button"
              onClick={() => setOpenEnv(!openEnv)}
              className="product-details-section-heading product-details-section-heading-button"
              aria-expanded={openEnv}
            >
              <span className="flex items-center gap-3 text-left">
                <span
                  className="product-environment-heading-icon"
                  aria-hidden="true"
                >
                  <ApartmentOutlined />
                </span>
                <span>
                  <span
                    id="product-environments-heading"
                    className="product-details-section-heading-title"
                  >
                    Environments
                  </span>
                  <span className="product-details-section-heading-subtitle">
                    {environments.length}
                    {environments.length === 1
                      ? " environment"
                      : " environments"}
                    where this product receives and searches logs.
                  </span>
                </span>
              </span>
              <DownOutlined
                className={`transition-transform duration-200 text-xs text-zinc-400 ${openEnv ? "rotate-0" : "-rotate-90"}`}
                aria-hidden="true"
              />
            </button>
            {openEnv && (
              <div className="product-environment-content">
                {environments.length ? (
                  <>
                    <div
                      className="product-environment-summary"
                      aria-label="Environment summary"
                    >
                      <div>
                        <span>Total</span>
                        <strong>{environments.length}</strong>
                      </div>
                      <div>
                        <span>Active</span>
                        <strong className="is-active">
                          {activeEnvironmentCount}
                        </strong>
                      </div>
                      <div>
                        <span>Inactive</span>
                        <strong className="is-inactive">
                          {inactiveEnvironmentCount}
                        </strong>
                      </div>
                    </div>
                    <div className="product-environment-list">
                      {environments.map((environment) => {
                        const isActive = environment.is_active !== false;
                        const environmentCode =
                          environment.environment_code || "ENV";
                        const environmentName =
                          environment.environment_name || "Unnamed environment";

                        return (
                          <div
                            key={environment.environment_id}
                            className={`product-environment-item${isActive ? "" : " is-inactive"}`}
                          >
                            <span
                              className="product-environment-icon"
                              aria-hidden="true"
                            >
                              {environmentCode.slice(0, 3).toUpperCase()}
                            </span>
                            <span className="product-environment-copy">
                              <span className="product-environment-name-row">
                                <strong>{environmentName}</strong>
                                <Tag color={isActive ? "success" : "default"}>
                                  {isActive ? "Active" : "Inactive"}
                                </Tag>
                              </span>
                              <code>{environmentCode}</code>
                            </span>
                          </div>
                        );
                      })}
                    </div>
                  </>
                ) : (
                  <Empty
                    image={Empty.PRESENTED_IMAGE_SIMPLE}
                    description="No environments configured"
                  />
                )}
              </div>
            )}
          </section>
          {/* Collapsible User in responsible */}
          <Card
            className="product-details-section"
            title={
              <div
                onClick={() => setOpenUsers(!openUsers)}
                className="flex items-center justify-between cursor-pointer select-none w-full py-1 text-zinc-800 hover:text-zinc-950 transition-colors"
              >
                <Space>
                  <TeamOutlined className="text-[#1F8457]" />
                  <span className="font-semibold">User in responsible</span>
                  <Tag color="default">
                    {members.data?.members?.length ?? 0}
                  </Tag>
                </Space>
                <DownOutlined
                  className={`transition-transform duration-200 text-xs text-zinc-400 ${openUsers ? "rotate-0" : "-rotate-90"}`}
                />
              </div>
            }
          >
            {openUsers && (
              <div className="flex flex-col gap-3">
                {props.canManage && (
                  <div className="flex items-center justify-between flex-wrap gap-2 pb-2 border-b border-zinc-100">
                    <div>
                      {removableMembers.length > 0 && (
                        <Checkbox
                          indeterminate={isIndeterminate}
                          checked={isAllSelected}
                          onChange={toggleSelectAll}
                        >
                          Select All ({removableMembers.length})
                        </Checkbox>
                      )}
                    </div>
                    <Space>
                      {selectedMembershipIds.length > 0 && (
                        <Button
                          danger
                          icon={<DeleteOutlined />}
                          loading={deleteMembers.isPending}
                          onClick={confirmBulkRemove}
                        >
                          Remove Selected ({selectedMembershipIds.length})
                        </Button>
                      )}
                      <Button
                        type="primary"
                        icon={<UserAddOutlined />}
                        onClick={() => setIsAddUserModalOpen(true)}
                      >
                        Add user
                      </Button>
                    </Space>
                  </div>
                )}
                <List
                  itemLayout="horizontal"
                  loading={members.isLoading || deleteMembers.isPending}
                  dataSource={members.data?.members ?? []}
                  locale={{ emptyText: "ยังไม่มีผู้ใช้ใน Product นี้" }}
                  renderItem={(user) => {
                    const isOwner = user.role_code?.toLowerCase() === "owner";
                    const isSelected = selectedMembershipIds.includes(
                      user.membership_id,
                    );
                    return (
                      <List.Item
                        actions={
                          props.canManage
                            ? [
                                <Button
                                  key="remove"
                                  type="text"
                                  danger
                                  size="small"
                                  icon={<DeleteOutlined />}
                                  onClick={() => confirmRemove(user)}
                                >
                                  Remove
                                </Button>,
                              ]
                            : undefined
                        }
                      >
                        <List.Item.Meta
                          avatar={
                            <div className="flex items-center gap-2">
                              {props.canManage && (
                                <Checkbox
                                  checked={isSelected}
                                  onChange={() =>
                                    toggleSelectUser(user.membership_id)
                                  }
                                />
                              )}
                            </div>
                          }
                          title={user.full_name || user.email}
                          description={
                            <Space size="small">
                              <span>{user.email}</span>
                              <Tag color={isOwner ? "gold" : "blue"}>
                                {user.role_name || user.role_code}
                              </Tag>
                            </Space>
                          }
                        />
                      </List.Item>
                    );
                  }}
                />
              </div>
            )}
          </Card>

          {props.canManage && (
            <Card
              className="product-details-section"
              title={
                <Space>
                  <CrownOutlined className="text-[#1F8457]" />
                  <span className="font-semibold">Product roles</span>
                  <Tag color="default">{members.data?.roles?.length ?? 0}</Tag>
                </Space>
              }
              extra={
                <Button
                  type="primary"
                  icon={<PlusOutlined />}
                  onClick={() => setIsCreateRoleModalOpen(true)}
                >
                  Create role
                </Button>
              }
            >
              <List
                size="small"
                loading={members.isLoading}
                dataSource={members.data?.roles ?? []}
                locale={{ emptyText: "No product roles configured" }}
                renderItem={(role) => (
                  <List.Item>
                    <Space>
                      <Tag
                        color={
                          role.role_code.toLowerCase() === "owner"
                            ? "gold"
                            : "blue"
                        }
                      >
                        {role.role_name}
                      </Tag>
                      <span>{role.access_level}</span>
                      <span className="text-zinc-500">
                        {role.member_count} member(s)
                      </span>
                    </Space>
                    <Space>
                      <Button
                        size="small"
                        icon={<EditOutlined />}
                        onClick={() => handleOpenEditRole(role)}
                      >
                        Edit
                      </Button>
                      <Button
                        size="small"
                        icon={<DeleteOutlined />}
                        onClick={() => confirmDeleteRole(role)}
                        loading={
                          deleteRoleMutation.isPending &&
                          deleteRoleMutation.variables === role.role_id
                        }
                        danger
                      >
                        Delete
                      </Button>
                    </Space>
                  </List.Item>
                )}
              />
            </Card>
          )}

          {props.canManage && (
            <Card className="product-details-setup-card">
              <div>
                <strong>
                  {product.setup_status === "ACTIVE"
                    ? "Edit product hierarchy"
                    : "Finish product setup"}
                </strong>
                <p>
                  {product.setup_status === "ACTIVE"
                    ? "แก้ไข Project, Feature และ Sub-feature ของ Product นี้"
                    : "Complete environments, access, and ingestion settings to make this product ready."}
                </p>
              </div>
              <Button
                type="primary"
                onClick={() => props.onContinueSetup(product)}
              >
                {product.setup_status === "ACTIVE"
                  ? "Edit hierarchy"
                  : "Continue setup"}
              </Button>
            </Card>
          )}

          {props.canManage && (
            <>
              <ProductPayloadRoutingPanel product={product} />
            </>
          )}
          {props.canManage && (
            <Card
              className="product-details-section border-red-200 bg-red-50/20"
              style={{ borderColor: "#ffccc7" }}
            >
              <div className="flex items-center justify-between flex-wrap gap-2">
                <div>
                  <strong
                    className="text-red-600 font-semibold"
                    style={{ color: "#ff4d4f" }}
                  >
                    Danger Zone
                  </strong>
                  <p className="m-0 text-xs text-zinc-500">
                    Permanently delete this product and all associated
                    configuration.
                  </p>
                </div>
                <Button
                  danger
                  type="primary"
                  icon={<DeleteOutlined />}
                  loading={deleteProductMutation.isPending}
                  onClick={confirmDeleteProduct}
                >
                  Delete product
                </Button>
              </div>
            </Card>
          )}

          <Divider className="product-details-divider" />

          {/* Collapsible Access & API keys */}
          <div
            onClick={() => setOpenAccess(!openAccess)}
            className="product-details-access-heading cursor-pointer select-none flex items-center justify-between py-1 mb-3"
          >
            <div>
              <h3 className="m-0">Access &amp; API keys</h3>
              <p className="m-0 text-xs text-zinc-500">
                Manage the people and credentials connected to this product.
              </p>
            </div>
            <DownOutlined
              className={`transition-transform duration-200 text-xs text-zinc-400 ${openAccess ? "rotate-0" : "-rotate-90"}`}
            />
          </div>
          {openAccess && (
            <ProductAccessPanel
              product={product}
              canManage={props.canManage}
              summaryOnly
              onGenerated={props.onGenerated}
            />
          )}

          <Modal
            open={isCreateRoleModalOpen}
            title="Create product role"
            okText="Create role"
            confirmLoading={createRole.isPending}
            onCancel={() => {
              setIsCreateRoleModalOpen(false);
              roleForm.resetFields();
            }}
            onOk={() => void roleForm.submit()}
          >
            <Form
              form={roleForm}
              layout="vertical"
              onFinish={(values) => createRole.mutate(values)}
            >
              <Form.Item
                name="roleName"
                label="Role name"
                rules={[{ required: true, min: 2 }]}
              >
                <Input placeholder="e.g. Log viewer" />
              </Form.Item>
            </Form>
          </Modal>

          <Modal
            open={isEditRoleModalOpen}
            title="Edit product role"
            okText="Save changes"
            confirmLoading={updateRole.isPending}
            onCancel={() => {
              setIsEditRoleModalOpen(false);
              setEditingRole(null);
              editRoleForm.resetFields();
            }}
            onOk={() => void editRoleForm.submit()}
          >
            <Form
              form={editRoleForm}
              layout="vertical"
              onFinish={(values) => updateRole.mutate(values)}
            >
              <Form.Item
                name="roleName"
                label="Role name"
                rules={[{ required: true, min: 2 }]}
              >
                <Input placeholder="e.g. Log viewer" />
              </Form.Item>
            </Form>
          </Modal>

          <Modal
            open={isAddUserModalOpen}
            title="Add user to product"
            okText="Add user"
            confirmLoading={createMembers.isPending}
            onCancel={() => {
              setIsAddUserModalOpen(false);
              addUserForm.resetFields();
            }}
            onOk={() => void addUserForm.submit()}
          >
            <Form
              form={addUserForm}
              layout="vertical"
              initialValues={{ scopeLevel: "PRODUCT" }}
              onFinish={(values) => createMembers.mutate(values)}
            >
              <Form.Item
                name="userId"
                label="User"
                rules={[{ required: true, message: "Please select a user" }]}
              >
                <Select
                  allowClear
                  showSearch
                  placeholder="Select a user"
                  loading={users.isLoading}
                  optionFilterProp="label"
                  options={availableUsers.map((u) => ({
                    value: Number(u.id),
                    label: `${u.fullName} (${u.email})`,
                  }))}
                />
              </Form.Item>
              <Form.Item
                name="roleId"
                label="Product role"
                rules={[{ required: true, message: "Please select a role" }]}
              >
                <Select
                  placeholder="Select role"
                  options={roles.map((r) => ({
                    value: r.role_id,
                    label: `${r.role_name} (${r.access_level})`,
                  }))}
                />
              </Form.Item>
              <Form.Item
                name="scopeLevel"
                label="Access scope"
                rules={[{ required: true, message: "Please select a scope" }]}
              >
                <Select
                  options={[
                    { value: "PRODUCT", label: "Entire product" },
                    { value: "PROJECT", label: "Selected projects" },
                    { value: "CATEGORY", label: "Selected features" },
                  ]}
                />
              </Form.Item>
              {selectedScopeLevel === "PROJECT" && (
                <Form.Item
                  name="projectIds"
                  label="Projects"
                  rules={[{ required: true, type: "array", min: 1 }]}
                >
                  <Select
                    mode="multiple"
                    loading={projects.isLoading}
                    options={(projects.data ?? []).map((project) => ({
                      value: project.project_id,
                      label: project.project_name,
                    }))}
                  />
                </Form.Item>
              )}
              {selectedScopeLevel === "CATEGORY" && (
                <Form.Item
                  name="featureIds"
                  label="Features"
                  rules={[{ required: true, type: "array", min: 1 }]}
                >
                  <FeatureScopeTree
                    projects={projects.data ?? []}
                    features={features.data ?? []}
                    value={addUserForm.getFieldValue("featureIds") ?? []}
                    onChange={(featureIds) =>
                      addUserForm.setFieldValue("featureIds", featureIds)
                    }
                    disabled={features.isLoading}
                  />
                </Form.Item>
              )}
            </Form>
          </Modal>
        </>
      )}
    </Drawer>
  );
}
