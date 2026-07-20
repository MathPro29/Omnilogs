import { Component, useCallback, useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import {
  Alert,
  Badge,
  Button,
  Card,
  Checkbox,
  Col,
  DatePicker,
  Empty,
  Form,
  Input,
  Modal,
  Popconfirm,
  Row,
  Select,
  Space,
  Spin,
  Switch,
  Table,
  Tabs,
  Tag,
  Tooltip,
  Typography,
  message,
} from "antd";
import {
  ArrowPathIcon,
  PencilSquareIcon,
  PlusIcon,
  ShieldCheckIcon,
  TrashIcon,
  UsersIcon,
} from "@heroicons/react/24/outline";
import dayjs from "dayjs";
import { Navigate } from "react-router-dom";
import { PageTransition } from "@/components";
import { PERMISSIONS, ROUTES } from "@/constants";
import { productAdminService } from "@/services/product-admin.service";
import { userService } from "@/services/user.service";
import {
  customFieldService,
  type CustomField,
} from "@/services/custom-field.service";
import { useAuthStore } from "@/store";
import type {
  Product,
  ProductAccessMember,
  ProductAccessOverview,
  ProductAccessRole,
  Project,
  ProjectFeature,
  User,
} from "@/types";

const { Title, Text } = Typography;

const RESOURCE_ACTIONS = [
  { resource: "PRODUCT", actions: ["CREATE", "READ", "UPDATE", "DELETE"] },
  { resource: "PROJECT", actions: ["CREATE", "READ", "UPDATE", "DELETE"] },
  { resource: "FEATURE", actions: ["CREATE", "READ", "UPDATE", "DELETE"] },
  { resource: "CATEGORY", actions: ["CREATE", "READ", "UPDATE", "DELETE"] },
  { resource: "ROLE", actions: ["CREATE", "READ", "UPDATE", "DELETE"] },
  { resource: "ACCESS", actions: ["READ", "GRANT", "REVOKE"] },
  { resource: "USER", actions: ["CREATE", "READ", "UPDATE", "DELETE"] },
  { resource: "API_KEY", actions: ["CREATE", "READ", "UPDATE", "DELETE"] },
  { resource: "ENVIRONMENT", actions: ["CREATE", "READ", "UPDATE", "DELETE"] },
  { resource: "LOG", actions: ["READ", "EXPORT", "VIEW_SENSITIVE"] },
  { resource: "AUDIT_LOG", actions: ["READ"] },
  {
    resource: "ELASTIC_INDEX_POLICY",
    actions: ["CREATE", "READ", "UPDATE", "DELETE"],
  },
];

const CUSTOM_FIELD_ACTIONS = [
  "VISIBLE",
  "SEARCH",
  "FILTER",
  "SORT",
  "AGGREGATE",
];

type MemberForm = {
  userId: number;
  roleId: number;
  scopes: string[];
  expiresAt?: dayjs.Dayjs;
  isActive: boolean;
};
type RoleForm = {
  roleCode: string;
  roleName: string;
  isActive: boolean;
};

const levelColor: Record<string, string> = {
  FULL_ACCESS: "purple",
  ADMIN: "red",
  EDITOR: "blue",
  READ_ONLY: "green",
};
const scopeKey = (scope: {
  scopeLevel: string;
  projectId?: number | null;
  categoryId?: number | null;
}) =>
  scope.scopeLevel === "PRODUCT"
    ? "PRODUCT"
    : scope.scopeLevel === "PROJECT"
      ? `PROJECT:${scope.projectId}`
      : `CATEGORY:${scope.projectId}:${scope.categoryId}`;

function AccessControlContent() {
  const authPermissions = useAuthStore((state) => state.permissions);
  const platformRoles = useAuthStore((state) => state.roles);
  const isAdmin = platformRoles.some((role) =>
    ["god", "owner", "superadmin", "super_admin"].includes(role.toLowerCase()),
  );
  const canView = isAdmin || authPermissions.includes(PERMISSIONS.ROLE_VIEW);
  const canEdit = isAdmin || authPermissions.includes(PERMISSIONS.ROLE_EDIT);

  const [products, setProducts] = useState<Product[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [features, setFeatures] = useState<ProjectFeature[]>([]);
  const [customFields, setCustomFields] = useState<CustomField[]>([]);
  const [selectedPermissions, setSelectedPermissions] = useState<
    Record<string, boolean>
  >({});
  const [productId, setProductId] = useState<number>();
  const [overview, setOverview] = useState<ProductAccessOverview>();
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [search, setSearch] = useState("");
  const [memberOpen, setMemberOpen] = useState(false);
  const [roleOpen, setRoleOpen] = useState(false);
  const [editingMember, setEditingMember] = useState<ProductAccessMember>();
  const [editingRole, setEditingRole] = useState<ProductAccessRole>();
  const [saving, setSaving] = useState(false);
  const [memberForm] = Form.useForm<MemberForm>();
  const [roleForm] = Form.useForm<RoleForm>();

  const loadOverview = useCallback(async (id: number, quiet = false) => {
    if (!quiet) setRefreshing(true);
    try {
      setOverview(await productAdminService.getAccessOverview(id));
    } catch (error: any) {
      if (!quiet)
        message.error(
          error?.response?.data?.error?.message ||
            error?.message ||
            "Failed to load product access.",
        );
    } finally {
      if (!quiet) setRefreshing(false);
    }
  }, []);

  const loadProductContext = useCallback(async (id: number) => {
    setLoading(true);
    try {
      const [nextOverview, nextProjects, nextCustomFields] = await Promise.all([
        productAdminService.getAccessOverview(id),
        productAdminService.listProjects(id),
        customFieldService.list(id),
      ]);
      const featureGroups = await Promise.all(
        nextProjects.map((project) =>
          productAdminService.listFeatures(id, project.projectId),
        ),
      );
      setOverview(nextOverview);
      setProjects(nextProjects);
      setFeatures(featureGroups.flat());
      setCustomFields(nextCustomFields);
    } catch (error: any) {
      message.error(
        error?.response?.data?.error?.message ||
          error?.message ||
          "Failed to load access configuration.",
      );
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    (async () => {
      try {
        const [ps, us] = await Promise.all([
          productAdminService.listProducts(),
          userService.getUsers({ page: 1, pageSize: 1000 }),
        ]);
        setProducts(ps);
        setUsers(us.data);
        if (ps[0]) setProductId(ps[0].productId);
      } catch {
        message.error("Failed to initialize access management.");
      }
    })();
  }, []);
  useEffect(() => {
    if (productId) void loadProductContext(productId);
  }, [productId, loadProductContext]);

  const scopeOptions = useMemo(
    () => [
      { value: "PRODUCT", label: "Entire Product" },
      ...projects.map((p) => ({
        value: `PROJECT:${p.projectId}`,
        label: `Project / ${p.projectName}`,
      })),
      ...features.map((f) => ({
        value: `CATEGORY:${f.projectId}:${f.categoryId}`,
        label: `Feature / ${projects.find((p) => p.projectId === f.projectId)?.projectName || f.projectId} / ${f.categoryName}`,
      })),
    ],
    [projects, features],
  );
  const filteredMembers = useMemo(
    () =>
      (overview?.members || []).filter((member) =>
        `${member.fullName} ${member.email} ${member.roleName}`
          .toLowerCase()
          .includes(search.toLowerCase()),
      ),
    [overview, search],
  );
  const watchedRoleId = Form.useWatch("roleId", memberForm);
  const selectedRole = useMemo(
    () => overview?.roles.find((role) => role.roleId === watchedRoleId),
    [overview?.roles, watchedRoleId],
  );

  if (!canView) return <Navigate to={ROUTES.FORBIDDEN} replace />;

  const openMember = (member?: ProductAccessMember) => {
    setEditingMember(member);
    setMemberOpen(true);
    memberForm.setFieldsValue(
      member
        ? {
            userId: member.userId,
            roleId: member.roleId,
            scopes: member.scopes.map(scopeKey),
            expiresAt: member.expiresAt ? dayjs(member.expiresAt) : undefined,
            isActive: member.isActive,
          }
        : {
            userId: undefined as unknown as number,
            roleId: undefined as unknown as number,
            scopes: ["PRODUCT"],
            expiresAt: undefined,
            isActive: true,
          },
    );
  };
  const autoCompleteExistingAccess = (userId: number) => {
    const member = overview?.members.find((item) => item.userId === userId);
    setEditingMember(member);
    if (member)
      memberForm.setFieldsValue({
        roleId: member.roleId,
        scopes: member.scopes.map(scopeKey),
        expiresAt: member.expiresAt ? dayjs(member.expiresAt) : undefined,
        isActive: member.isActive,
      });
  };
  const saveMember = async (values: MemberForm) => {
    if (!productId) return;
    setSaving(true);
    try {
      const scopes = values.scopes.map((key) => {
        const [scopeLevel, project, category] = key.split(":");
        return {
          scopeLevel: scopeLevel as "PRODUCT" | "PROJECT" | "CATEGORY",
          projectId: project ? Number(project) : null,
          categoryId: category ? Number(category) : null,
        };
      });
      await productAdminService.upsertProductAccess(productId, values.userId, {
        roleId: values.roleId,
        scopes,
        expiresAt: values.expiresAt?.endOf("day").toISOString() || null,
        isActive: values.isActive,
      });
      message.success(
        editingMember
          ? "Existing access was updated."
          : "Member access was added.",
      );
      setMemberOpen(false);
      await loadOverview(productId);
    } catch (error: any) {
      message.error(
        error?.response?.data?.error?.message ||
          error?.message ||
          "Failed to save access.",
      );
    } finally {
      setSaving(false);
    }
  };
  const removeMember = async (member: ProductAccessMember) => {
    if (!productId) return;
    await productAdminService.deleteMembership(productId, member.membershipId);
    message.success("Member removed.");
    await loadOverview(productId);
  };

  const openRole = (role?: ProductAccessRole) => {
    setEditingRole(role);
    setRoleOpen(true);
    const mapped: Record<string, boolean> = {};
    if (role) {
      role.permissions.forEach((p) => {
        mapped[`${p.resourceType}:${p.action}`] = true;
      });
    } else {
      mapped["PRODUCT:READ"] = true;
    }
    setSelectedPermissions(mapped);
    roleForm.setFieldsValue(
      role
        ? {
            roleCode: role.roleCode,
            roleName: role.roleName,
            isActive: role.isActive,
          }
        : {
            roleCode: "",
            roleName: "",
            isActive: true,
          },
    );
  };
  const saveRole = async (values: RoleForm) => {
    if (!productId) return;
    setSaving(true);

    const finalPermissions: { resourceType: string; action: string }[] = [];
    Object.entries(selectedPermissions).forEach(([key, enabled]) => {
      if (enabled) {
        const separator = key.lastIndexOf(":");
        const resourceType = separator >= 0 ? key.slice(0, separator) : key;
        const action = separator >= 0 ? key.slice(separator + 1) : "";
        finalPermissions.push({ resourceType, action });
      }
    });

    // Handle custom fields default values (just like ProductRoleConsolePage does)
    const existingKeys = new Set(
      finalPermissions.map((p) => `${p.resourceType}:${p.action}`),
    );
    customFields.forEach((field) => {
      const resource = `CUSTOM_FIELD:${field.field_definition_id}`;
      const defaults: Record<string, boolean> = {
        VISIBLE: field.is_visible,
        SEARCH: field.is_searchable,
        FILTER: field.is_filterable,
        SORT: field.is_sortable,
        AGGREGATE: field.is_aggregatable,
      };
      CUSTOM_FIELD_ACTIONS.forEach((action) => {
        const key = `${resource}:${action}`;
        if (
          !Object.prototype.hasOwnProperty.call(selectedPermissions, key) &&
          defaults[action]
        ) {
          finalPermissions.push({ resourceType: resource, action });
          existingKeys.add(key);
        }
      });
    });

    try {
      if (editingRole)
        await productAdminService.updateRole(productId, editingRole.roleId, {
          roleName: values.roleName,
          permissions: finalPermissions,
          isActive: values.isActive,
        });
      else
        await productAdminService.createRole(productId, {
          roleCode: values.roleCode.trim().toLowerCase().replace(/\s+/g, "_"),
          roleName: values.roleName,
          permissions: finalPermissions,
        });
      message.success(
        editingRole
          ? "Role updated; effective permissions refreshed."
          : "Role created.",
      );
      setRoleOpen(false);
      await loadOverview(productId);
    } catch (error: any) {
      message.error(
        error?.response?.data?.error?.message ||
          error?.message ||
          "Failed to save role.",
      );
    } finally {
      setSaving(false);
    }
  };
  const removeRole = async (role: ProductAccessRole) => {
    if (!productId) return;
    await productAdminService.deleteRole(productId, role.roleId);
    message.success("Role removed.");
    await loadOverview(productId);
  };

  const memberColumns = [
    {
      title: "Member",
      key: "member",
      render: (_: unknown, member: ProductAccessMember) => (
        <div>
          <div className="font-semibold">{member.fullName}</div>
          <Text type="secondary" className="text-xs">
            {member.email}
          </Text>
        </div>
      ),
    },
    {
      title: "Role",
      key: "role",
      render: (_: unknown, member: ProductAccessMember) => (
        <div>
          <Tag color="blue">{member.roleName}</Tag>
          <Text type="secondary" className="block text-xs mt-1">
            {member.accessLevel.replace("_", " ")}
          </Text>
        </div>
      ),
    },
    {
      title: "Scope",
      key: "scope",
      render: (_: unknown, member: ProductAccessMember) => (
        <Space wrap>
          {member.scopes.map((scope) => (
            <Tag key={scope.scopeId}>
              {scopeKey(scope).replaceAll(":", " / ")}
            </Tag>
          ))}
        </Space>
      ),
    },
    {
      title: "Effective Permissions",
      key: "permissions",
      render: (_: unknown, member: ProductAccessMember) => (
        <Tooltip
          title={member.effectivePermissions
            .map((p) => `${p.resourceType}.${p.action} (${p.source})`)
            .join(", ")}
        >
          <Badge
            count={member.effectivePermissions.length}
            showZero
            color="#6366f1"
          />
          <Text className="ml-2">permissions</Text>
        </Tooltip>
      ),
    },
    {
      title: "Status",
      key: "status",
      render: (_: unknown, member: ProductAccessMember) => (
        <Badge
          status={member.isActive ? "success" : "default"}
          text={member.isActive ? "Active" : "Inactive"}
        />
      ),
    },
    {
      title: "",
      key: "actions",
      render: (_: unknown, member: ProductAccessMember) =>
        canEdit && (
          <Space>
            <Button
              type="text"
              icon={<PencilSquareIcon className="w-4 h-4" />}
              onClick={() => openMember(member)}
            />
            <Popconfirm
              title="Remove this member from the product?"
              onConfirm={() => void removeMember(member)}
            >
              <Button
                type="text"
                danger
                icon={<TrashIcon className="w-4 h-4" />}
              />
            </Popconfirm>
          </Space>
        ),
    },
  ];

  return (
    <PageTransition>
      <div className="space-y-5">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <Title level={3} className="mb-1!">
              <ShieldCheckIcon className="inline w-7 h-7 mr-2 text-indigo-600" />
              Product Access & Role Management
            </Title>
            <Text type="secondary">
              Centralized RBAC, effective permissions, and scope management.
            </Text>
          </div>
          <Space>
            <Button
              icon={<ArrowPathIcon className="w-4 h-4" />}
              loading={refreshing}
              onClick={() => productId && void loadOverview(productId)}
            >
              Refresh
            </Button>
          </Space>
        </div>
        <Card>
          <Row gutter={[16, 16]} align="middle">
            <Col xs={24} md={12}>
              <Text strong>Product</Text>
              <Select
                className="w-full mt-2"
                showSearch
                optionFilterProp="label"
                value={productId}
                onChange={setProductId}
                options={products.map((p) => ({
                  value: p.productId,
                  label: `${p.productName} (${p.productCode})`,
                }))}
              />
            </Col>
            <Col xs={24} md={12}>
              {overview && (
                <div className="flex justify-end gap-6">
                  <div>
                    <Text type="secondary">Members</Text>
                    <div className="text-2xl font-semibold">
                      {overview.members.length}
                    </div>
                  </div>
                  <div>
                    <Text type="secondary">Roles</Text>
                    <div className="text-2xl font-semibold">
                      {overview.roles.length}
                    </div>
                  </div>
                  <div>
                    <Text type="secondary">Last sync</Text>
                    <div>{dayjs(overview.updatedAt).format("HH:mm:ss")}</div>
                  </div>
                </div>
              )}
            </Col>
          </Row>
        </Card>
        {loading ? (
          <Card>
            <div className="py-20 text-center">
              <Spin size="large" />
            </div>
          </Card>
        ) : !overview ? (
          <Empty />
        ) : (
          <Tabs
            items={[
              {
                key: "members",
                label: (
                  <span>
                    <UsersIcon className="inline w-4 h-4 mr-1" />
                    Members
                  </span>
                ),
                children: (
                  <Card
                    title="Product members"
                    extra={
                      canEdit && (
                        <Button
                          type="primary"
                          icon={<PlusIcon className="w-4 h-4" />}
                          onClick={() => openMember()}
                        >
                          Add member
                        </Button>
                      )
                    }
                  >
                    <Input.Search
                      className="mb-4 max-w-md"
                      placeholder="Search member, email, or role"
                      allowClear
                      onChange={(event) => setSearch(event.target.value)}
                    />
                    <Table
                      rowKey="membershipId"
                      dataSource={filteredMembers}
                      columns={memberColumns}
                      pagination={{ pageSize: 10 }}
                      expandable={{
                        expandedRowRender: (member) => {
                          const grouped = member.effectivePermissions.reduce(
                            (acc, p) => {
                              let categoryName = p.resourceType;
                              if (p.resourceType.startsWith("CUSTOM_FIELD:")) {
                                const fieldId = p.resourceType.split(":")[1];
                                const field = customFields.find(
                                  (f) =>
                                    String(f.field_definition_id) === fieldId,
                                );
                                categoryName = field
                                  ? `Custom Field: ${field.display_name || field.field_key}`
                                  : `Custom Field (${fieldId})`;
                              } else if (p.resourceType === "CUSTOM_FIELD") {
                                categoryName = "Custom Fields (Global)";
                              }

                              if (!acc[categoryName]) {
                                acc[categoryName] = [];
                              }
                              acc[categoryName].push(p);
                              return acc;
                            },
                            {} as Record<
                              string,
                              typeof member.effectivePermissions
                            >,
                          );

                          return (
                            <div className="p-2 bg-gray-50/50 rounded-lg">
                              <div className="mb-3 font-semibold text-gray-700 flex items-center gap-2">
                                <ShieldCheckIcon className="w-5 h-5 text-indigo-500" />
                                <span>Effective Permission Details</span>
                              </div>
                              {Object.keys(grouped).length === 0 ? (
                                <Text
                                  type="secondary"
                                  className="italic text-xs"
                                >
                                  No permissions assigned.
                                </Text>
                              ) : (
                                <Row gutter={[12, 12]}>
                                  {Object.entries(grouped).map(
                                    ([category, perms]) => (
                                      <Col
                                        xs={24}
                                        sm={12}
                                        md={8}
                                        lg={6}
                                        key={category}
                                      >
                                        <Card
                                          size="small"
                                          title={
                                            <Text
                                              strong
                                              className="text-xs text-indigo-900"
                                            >
                                              {category}
                                            </Text>
                                          }
                                          className="h-full border-indigo-100 hover:border-indigo-300 transition-colors shadow-sm bg-white"
                                          bodyStyle={{ padding: "8px 12px" }}
                                        >
                                          <Space wrap size={[4, 4]}>
                                            {perms.map((p) => (
                                              <Tag
                                                color={
                                                  p.source === "ROLE"
                                                    ? "blue"
                                                    : "gold"
                                                }
                                                key={`${p.resourceType}:${p.action}`}
                                                className="m-0 text-[11px]"
                                              >
                                                {p.action}{" "}
                                                <span className="text-[9px] opacity-75">
                                                  ({p.source})
                                                </span>
                                              </Tag>
                                            ))}
                                          </Space>
                                        </Card>
                                      </Col>
                                    ),
                                  )}
                                </Row>
                              )}
                            </div>
                          );
                        },
                      }}
                    />
                  </Card>
                ),
              },
              {
                key: "roles",
                label: "Roles & Permission Matrix",
                children: (
                  <Card
                    title="Roles are the source of user configuration"
                    extra={
                      canEdit && (
                        <Button
                          type="primary"
                          icon={<PlusIcon className="w-4 h-4" />}
                          onClick={() => openRole()}
                        >
                          Create role
                        </Button>
                      )
                    }
                  >
                    <Alert
                      className="mb-4"
                      type="info"
                      showIcon
                      title="Changing a role automatically updates effective permissions for every member assigned to it."
                    />
                    <Row gutter={[16, 16]}>
                      {overview.roles.map((role) => {
                        // Group the role's permissions by resource type/custom field
                        const groupedPerms = role.permissions.reduce(
                          (acc, p) => {
                            let categoryName = p.resourceType;
                            if (p.resourceType.startsWith("CUSTOM_FIELD:")) {
                              const fieldId = p.resourceType.split(":")[1];
                              const field = customFields.find(
                                (f) =>
                                  String(f.field_definition_id) === fieldId,
                              );
                              categoryName = field
                                ? `Custom Field: ${field.display_name || field.field_key}`
                                : `Custom Field (${fieldId})`;
                            } else if (p.resourceType === "CUSTOM_FIELD") {
                              categoryName = "Custom Fields (Global)";
                            }

                            if (!acc[categoryName]) {
                              acc[categoryName] = [];
                            }
                            acc[categoryName].push(p);
                            return acc;
                          },
                          {} as Record<string, typeof role.permissions>,
                        );

                        return (
                          <Col xs={24} lg={12} key={role.roleId}>
                            <Card
                              size="small"
                              title={
                                <Space>
                                  <Tag color={levelColor[role.accessLevel]}>
                                    {role.accessLevel}
                                  </Tag>
                                  <Text strong>{role.roleName}</Text>
                                </Space>
                              }
                              className="hover:shadow-md transition-shadow border-indigo-100/60 shadow-sm"
                              extra={
                                canEdit && (
                                  <Space>
                                    <Button
                                      type="text"
                                      icon={
                                        <PencilSquareIcon className="w-4 h-4 text-indigo-600" />
                                      }
                                      onClick={() => openRole(role)}
                                    />
                                    <Popconfirm
                                      title="Delete role?"
                                      disabled={role.memberCount > 0}
                                      onConfirm={() => void removeRole(role)}
                                    >
                                      <Tooltip
                                        title={
                                          role.memberCount
                                            ? "Move members to another role first."
                                            : "Delete role"
                                        }
                                      >
                                        <Button
                                          type="text"
                                          danger
                                          disabled={role.memberCount > 0}
                                          icon={
                                            <TrashIcon className="w-4 h-4" />
                                          }
                                        />
                                      </Tooltip>
                                    </Popconfirm>
                                  </Space>
                                )
                              }
                            >
                              <div className="mb-3 flex justify-between items-center bg-gray-50 p-2 rounded">
                                <Text type="secondary" className="text-xs">
                                  Code:{" "}
                                  <code className="bg-white px-1.5 py-0.5 rounded border text-indigo-600 font-mono text-[11px]">
                                    {role.roleCode}
                                  </code>
                                </Text>
                                <Space>
                                  <Badge
                                    status={
                                      role.isActive ? "success" : "default"
                                    }
                                    text={role.isActive ? "Active" : "Inactive"}
                                    className="text-xs"
                                  />
                                  <Tag color="cyan" className="m-0 text-[11px]">
                                    {role.memberCount} member(s)
                                  </Tag>
                                </Space>
                              </div>

                              <div className="space-y-3 mt-3">
                                <Text
                                  strong
                                  className="text-[10px] text-gray-400 block uppercase tracking-wider"
                                >
                                  Granted Permissions
                                </Text>
                                {Object.keys(groupedPerms).length === 0 ? (
                                  <Text
                                    type="secondary"
                                    className="italic text-xs block"
                                  >
                                    No permissions assigned.
                                  </Text>
                                ) : (
                                  <Row gutter={[8, 8]}>
                                    {Object.entries(groupedPerms).map(
                                      ([category, perms]) => (
                                        <Col xs={24} sm={12} key={category}>
                                          <div className="bg-indigo-50/20 p-2 rounded border border-indigo-50 h-full">
                                            <div className="text-[10px] font-bold text-indigo-900 mb-1 leading-tight">
                                              {category}
                                            </div>
                                            <Space wrap size={[2, 2]}>
                                              {perms.map((p) => (
                                                <Tag
                                                  key={`${p.resourceType}:${p.action}`}
                                                  className="m-0 text-[10px] bg-white border-indigo-100/50 text-indigo-955"
                                                >
                                                  {p.action}
                                                </Tag>
                                              ))}
                                            </Space>
                                          </div>
                                        </Col>
                                      ),
                                    )}
                                  </Row>
                                )}
                              </div>
                            </Card>
                          </Col>
                        );
                      })}
                    </Row>
                  </Card>
                ),
              },
            ]}
          />
        )}

        <Modal
          title={editingMember ? "Update Product Access" : "Add Product Member"}
          open={memberOpen}
          onCancel={() => setMemberOpen(false)}
          onOk={() => memberForm.submit()}
          confirmLoading={saving}
          width={720}
        >
          <Alert
            className="mb-4"
            type="info"
            showIcon
            message="One user has one role per product"
            description="Saving replaces the current role and scopes; it never adds a second role."
          />
          <Form form={memberForm} layout="vertical" onFinish={saveMember}>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  name="userId"
                  label="User"
                  rules={[{ required: true }]}
                >
                  <Select
                    showSearch
                    optionFilterProp="label"
                    disabled={Boolean(editingMember)}
                    onChange={autoCompleteExistingAccess}
                    options={users.map((user) => {
                      const existing = overview?.members.find(
                        (m) => m.userId === Number(user.id),
                      );
                      return {
                        value: Number(user.id),
                        label: `${user.fullName} (${user.email})${existing ? ` — current: ${existing.roleName}` : ""}`,
                      };
                    })}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  name="roleId"
                  label="Role"
                  rules={[{ required: true }]}
                >
                  <Select
                    showSearch
                    optionFilterProp="label"
                    options={(overview?.roles || [])
                      .filter((r) => r.isActive)
                      .map((role) => ({
                        value: role.roleId,
                        label: `${role.roleName} — ${role.accessLevel}${editingMember?.roleId === role.roleId ? " (current)" : ""}`,
                      }))}
                  />
                </Form.Item>
              </Col>
            </Row>
            {selectedRole && (
              <Alert
                className="mb-4"
                type="info"
                title={`${selectedRole.roleName}: ${selectedRole.accessLevel}`}
                description={`${selectedRole.permissions.length} role permissions will be used to calculate effective access.`}
              />
            )}
            <Form.Item
              name="scopes"
              label="Access Scope (Product / Project / Feature / Category)"
              rules={[{ required: true, type: "array", min: 1 }]}
            >
              <Select
                mode="multiple"
                showSearch
                optionFilterProp="label"
                options={scopeOptions}
              />
            </Form.Item>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item name="expiresAt" label="Expires at">
                  <DatePicker className="w-full" />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  name="isActive"
                  label="Active"
                  valuePropName="checked"
                >
                  <Switch />
                </Form.Item>
              </Col>
            </Row>
          </Form>
        </Modal>
        <Modal
          title={
            editingRole ? "Edit Role & Permissions" : "Create Product Role"
          }
          open={roleOpen}
          onCancel={() => setRoleOpen(false)}
          onOk={() => roleForm.submit()}
          confirmLoading={saving}
          width={800}
        >
          <Form form={roleForm} layout="vertical" onFinish={saveRole}>
            <Row gutter={16}>
              <Col span={10}>
                <Form.Item
                  name="roleCode"
                  label="Role code"
                  rules={[{ required: true }]}
                >
                  <Input
                    disabled={Boolean(editingRole)}
                    placeholder="e.g. support_engineer"
                  />
                </Form.Item>
              </Col>
              <Col span={10}>
                <Form.Item
                  name="roleName"
                  label="Role name"
                  rules={[{ required: true }]}
                >
                  <Input />
                </Form.Item>
              </Col>
              <Col span={4}>
                <Form.Item
                  name="isActive"
                  label="Active"
                  valuePropName="checked"
                >
                  <Switch />
                </Form.Item>
              </Col>
            </Row>

            <div className="mt-4">
              <Text strong className="block mb-2">
                Resource Permissions
              </Text>
              <div className="max-h-[300px] overflow-y-auto border border-gray-200 rounded">
                <table className="w-full text-sm">
                  <thead className="bg-gray-50 sticky top-0">
                    <tr>
                      <th className="px-3 py-2 text-left border-b w-1/3">
                        Resource
                      </th>
                      <th className="px-3 py-2 text-left border-b w-2/3">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-100">
                    {RESOURCE_ACTIONS.map((item) => (
                      <tr key={item.resource}>
                        <td className="px-3 py-2 font-medium align-top">
                          {item.resource}
                        </td>
                        <td className="px-3 py-2">
                          <Space wrap size={[16, 8]}>
                            {item.actions.map((action) => {
                              const key = `${item.resource}:${action}`;
                              return (
                                <Checkbox
                                  key={action}
                                  checked={Boolean(selectedPermissions[key])}
                                  onChange={(e) => {
                                    setSelectedPermissions((prev) => ({
                                      ...prev,
                                      [key]: e.target.checked,
                                    }));
                                  }}
                                >
                                  {action}
                                </Checkbox>
                              );
                            })}
                          </Space>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>

            <div className="mt-6 mb-4">
              <Text strong className="block mb-1">
                Custom Field Capabilities
              </Text>
              <Text type="secondary" className="block text-xs mb-2">
                Configure direct visibility and capabilities for custom fields.
              </Text>
              {customFields.length === 0 ? (
                <div className="p-3 text-center text-gray-500 border border-dashed rounded">
                  No custom fields configured for this product.
                </div>
              ) : (
                <div className="max-h-[250px] overflow-y-auto border border-gray-200 rounded">
                  <table className="w-full text-sm text-left">
                    <thead className="bg-gray-50 sticky top-0">
                      <tr>
                        <th className="px-3 py-2 border-b">Field</th>
                        {CUSTOM_FIELD_ACTIONS.map((action) => (
                          <th
                            key={action}
                            className="px-2 py-2 text-center border-b font-medium"
                          >
                            {action}
                          </th>
                        ))}
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-100">
                      {customFields.map((field) => {
                        const resource = `CUSTOM_FIELD:${field.field_definition_id}`;
                        const defaultValues: Record<string, boolean> = {
                          VISIBLE: field.is_visible,
                          SEARCH: field.is_searchable,
                          FILTER: field.is_filterable,
                          SORT: field.is_sortable,
                          AGGREGATE: field.is_aggregatable,
                        };
                        return (
                          <tr key={field.field_definition_id}>
                            <td className="px-3 py-2">
                              <div className="font-medium text-gray-800">
                                {field.display_name || field.field_key}
                              </div>
                              <div className="text-xs text-gray-400">
                                {field.field_path || field.field_key}
                              </div>
                            </td>
                            {CUSTOM_FIELD_ACTIONS.map((action) => {
                              const key = `${resource}:${action}`;
                              const checked =
                                Object.prototype.hasOwnProperty.call(
                                  selectedPermissions,
                                  key,
                                )
                                  ? Boolean(selectedPermissions[key])
                                  : defaultValues[action];
                              return (
                                <td
                                  key={action}
                                  className="px-2 py-2 text-center"
                                >
                                  <Checkbox
                                    checked={checked}
                                    onChange={(e) => {
                                      setSelectedPermissions((prev) => ({
                                        ...prev,
                                        [key]: e.target.checked,
                                      }));
                                    }}
                                  />
                                </td>
                              );
                            })}
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>
              )}
            </div>

            <Space className="mt-4">
              <Button
                size="small"
                onClick={() => {
                  const next = { ...selectedPermissions };
                  RESOURCE_ACTIONS.forEach((item) => {
                    item.actions.forEach((action) => {
                      next[`${item.resource}:${action}`] = true;
                    });
                  });
                  setSelectedPermissions(next);
                }}
              >
                Full standard access
              </Button>
              <Button
                size="small"
                type="text"
                onClick={() => setSelectedPermissions({})}
              >
                Clear
              </Button>
            </Space>
          </Form>
        </Modal>
      </div>
    </PageTransition>
  );
}

export default AccessControlPage;

class AccessControlErrorBoundary extends Component<
  { children: ReactNode },
  { hasError: boolean }
> {
  state = { hasError: false };
  static getDerivedStateFromError() {
    return { hasError: true };
  }
  componentDidCatch(error: unknown) {
    console.error("AccessControlPage render error:", error);
  }
  render() {
    if (this.state.hasError)
      return (
        <Card className="m-6">
          <Alert
            type="error"
            showIcon
            title="Access Control could not be displayed"
            description="Please refresh the page. If the problem continues, check the API response for this product."
          />
        </Card>
      );
    return this.props.children;
  }
}

export function AccessControlPage() {
  return (
    <AccessControlErrorBoundary>
      <AccessControlContent />
    </AccessControlErrorBoundary>
  );
}
