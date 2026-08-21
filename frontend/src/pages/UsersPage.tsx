import { useEffect, useState } from "react";
import {
  useMutation,
  useQueries,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import {
  Alert,
  Avatar,
  Button,
  Card,
  Input,
  Modal,
  Select,
  Space,
  Switch,
  Table,
  Tag,
  Tooltip,
  Typography,
  message,
} from "antd";
import type { ColumnsType, TablePaginationConfig } from "antd/es/table";
import {
  DeleteOutlined,
  EditOutlined,
  KeyOutlined,
  PlusOutlined,
  ReloadOutlined,
  SearchOutlined,
  UserAddOutlined,
} from "@ant-design/icons";
import { PageTransition, PermissionGuard, TableSkeleton } from "@/components";
import {
  UserAccessDrawer,
  type ProductAccessData,
} from "@/components/features/users/UserAccessDrawer";
import { BulkProductAccessModal } from "@/components/features/users/BulkProductAccessModal";
import {
  UserFormModal,
  type UserFormValues,
} from "@/components/features/users/UserFormModal";
import {
  PAGINATION,
  PERMISSIONS,
  QUERY_KEYS,
  ROUTES,
  STATUS_COLORS,
  STATUS_LABELS,
} from "@/constants";
import { productService } from "@/services/product.service";
import { userService } from "@/services/user.service";
import { useAppStore } from "@/store";
import type { User, UserFilterParams, UserStatus } from "@/types";
import { getInitials } from "@/utils";

const { Text } = Typography;

function errorMessage(error: unknown, fallback: string): string {
  if (typeof error === "object" && error !== null) {
    const value = error as {
      message?: string;
      response?: { data?: { message?: string } };
    };
    return value.response?.data?.message || value.message || fallback;
  }
  return fallback;
}

export function UsersPage() {
  const client = useQueryClient();
  const setBreadcrumbs = useAppStore((state) => state.setBreadcrumbs);
  const [filters, setFilters] = useState<UserFilterParams>({
    page: 1,
    pageSize: 10,
  });
  const [search, setSearch] = useState("");
  const [formOpen, setFormOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<User>();
  const [accessUser, setAccessUser] = useState<User>();
  const [bulkAccessOpen, setBulkAccessOpen] = useState(false);

  useEffect(() => {
    setBreadcrumbs([{ title: "จัดการผู้ใช้งาน", path: ROUTES.USERS }]);
  }, [setBreadcrumbs]);

  useEffect(() => {
    const timer = window.setTimeout(
      () =>
        setFilters((current) => ({
          ...current,
          search: search.trim(),
          page: 1,
        })),
      350,
    );
    return () => window.clearTimeout(timer);
  }, [search]);

  const users = useQuery({
    queryKey: [...QUERY_KEYS.USERS, filters],
    queryFn: () => userService.getUsers(filters),
  });
  const products = useQuery({
    queryKey: ["products", "user-management"],
    queryFn: productService.listOptions,
    staleTime: 30_000,
  });
  const accessQueries = useQueries({
    queries: (products.data ?? []).map((product) => ({
      queryKey: ["products", product.id, "access-overview"],
      queryFn: () => productService.getAccessOverview(product.id),
      staleTime: 15_000,
      retry: 1,
    })),
  });
  const productAccess: ProductAccessData[] = (products.data ?? []).map(
    (product, index) => ({
      product,
      overview: accessQueries[index]?.data,
      loading: accessQueries[index]?.isLoading ?? false,
      error: accessQueries[index]?.isError ?? false,
    }),
  );
  const assignmentsFor = (userId: string) =>
    productAccess.flatMap(({ product, overview }) => {
      const member = overview?.members.find(
        (item) => String(item.user_id) === userId,
      );
      return member ? [{ product, member }] : [];
    });
  const refreshAccess = () => {
    (products.data ?? []).forEach((product) => {
      void client.invalidateQueries({
        queryKey: ["products", product.id, "access-overview"],
      });
    });
  };

  const createUser = useMutation({
    mutationFn: (values: UserFormValues) =>
      userService.createUser({
        username: values.username || "",
        fullName: values.fullName,
        email: values.email,
        phone: values.phone,
        status: values.status,
        role: values.role,
        password: values.password,
        permissions: values.permissions,
      }),
    onSuccess: () => {
      message.success("เพิ่มผู้ใช้งานแล้ว");
      setFormOpen(false);
      void client.invalidateQueries({ queryKey: QUERY_KEYS.USERS });
    },
    onError: (error) =>
      message.error(errorMessage(error, "ไม่สามารถเพิ่มผู้ใช้งานได้")),
  });
  const updateUser = useMutation({
    mutationFn: ({
      user,
      values,
    }: {
      user: User;
      values: Partial<UserFormValues>;
    }) =>
      userService.updateUser(user.id, {
        fullName: values.fullName,
        email: values.email,
        phone: values.phone,
        status: values.status,
        role: values.role,
        permissions: values.permissions,
      }),
    onSuccess: () => {
      message.success("บันทึกข้อมูลผู้ใช้งานแล้ว");
      setFormOpen(false);
      setEditingUser(undefined);
      void client.invalidateQueries({ queryKey: QUERY_KEYS.USERS });
    },
    onError: (error) =>
      message.error(errorMessage(error, "ไม่สามารถแก้ไขผู้ใช้งานได้")),
  });
  const deleteUser = useMutation({
    mutationFn: userService.deleteUser,
    onSuccess: () => {
      message.success("ลบผู้ใช้งานแล้ว");
      void client.invalidateQueries({ queryKey: QUERY_KEYS.USERS });
      refreshAccess();
    },
    onError: (error) =>
      message.error(errorMessage(error, "ไม่สามารถลบผู้ใช้งานได้")),
  });
  const setUserStatus = useMutation({
    mutationFn: ({ user, status }: { user: User; status: UserStatus }) =>
      userService.updateUser(user.id, { status }),
    onSuccess: (_, variables) => {
      message.success(
        variables.status === "active"
          ? "เปิดใช้งาน User แล้ว"
          : "ปิดใช้งาน User แล้ว",
      );
      void client.invalidateQueries({ queryKey: QUERY_KEYS.USERS });
    },
    onError: (error) =>
      message.error(errorMessage(error, "ไม่สามารถเปลี่ยนสถานะได้")),
  });

  const openEditor = (user?: User) => {
    setEditingUser(user);
    setFormOpen(true);
  };
  const submitUser = (values: UserFormValues) => {
    if (editingUser) updateUser.mutate({ user: editingUser, values });
    else createUser.mutate(values);
  };
  const confirmDelete = (user: User) => {
    Modal.confirm({
      title: `ลบ ${user.fullName}?`,
      content:
        "บัญชีและการเข้าถึง Product ของผู้ใช้นี้จะถูกนำออก การดำเนินการนี้ไม่สามารถย้อนกลับได้",
      okText: "ลบผู้ใช้งาน",
      cancelText: "ยกเลิก",
      okButtonProps: { danger: true },
      onOk: () => deleteUser.mutateAsync(user.id),
    });
  };

  const columns: ColumnsType<User> = [
      {
        title: "ผู้ใช้งาน",
        key: "user",
        width: 270,
        render: (_, user) => (
          <div className="user-identity">
            <Avatar>{getInitials(user.fullName)}</Avatar>
            <div>
              <strong>{user.fullName}</strong>
              <Text type="secondary">{user.email}</Text>
            </div>
          </div>
        ),
      },
      {
        title: "Platform Role",
        key: "role",
        width: 145,
        render: (_, user) => (
          <Tag color="purple">{user.roles[0]?.name || "admin"}</Tag>
        ),
      },
      {
        title: "ผลิตภัณฑ์ที่ดูแล",
        key: "products",
        width: 280,
        render: (_, user) => {
          const assignments = assignmentsFor(user.id);
          if (!assignments.length)
            return <Text type="secondary">ยังไม่ได้มอบหมาย</Text>;
          return (
            <div className="user-product-tags">
              {assignments.slice(0, 2).map(({ product, member }) => (
                <Tag
                  key={product.id}
                  color={member.is_active ? "blue" : "default"}
                >
                  {product.name} · {member.role_name}
                </Tag>
              ))}
              {assignments.length > 2 && <Tag>+{assignments.length - 2}</Tag>}
            </div>
          );
        },
      },
      {
        title: "สถานะ",
        key: "status",
        width: 145,
        render: (_, user) => (
          <Space>
            <Switch
              checked={user.status === "active"}
              loading={
                setUserStatus.isPending &&
                setUserStatus.variables?.user.id === user.id
              }
              aria-label={`${user.status === "active" ? "ปิด" : "เปิด"}ใช้งาน ${user.fullName}`}
              onChange={(active) =>
                setUserStatus.mutate({
                  user,
                  status: active ? "active" : "inactive",
                })
              }
            />
            <Tag color={STATUS_COLORS[user.status]}>
              {STATUS_LABELS[user.status]}
            </Tag>
          </Space>
        ),
      },
      {
        title: "จัดการ",
        key: "actions",
        width: 145,
        fixed: "right",
        render: (_, user) => (
          <Space>
            <Tooltip title="Product & Feature Access">
              <Button
                type="text"
                aria-label={`จัดการสิทธิ์ของ ${user.fullName}`}
                icon={<KeyOutlined />}
                onClick={() => setAccessUser(user)}
              />
            </Tooltip>
            <PermissionGuard permission={PERMISSIONS.USER_EDIT}>
              <Tooltip title="แก้ไข">
                <Button
                  type="text"
                  aria-label={`แก้ไข ${user.fullName}`}
                  icon={<EditOutlined />}
                  onClick={() => openEditor(user)}
                />
              </Tooltip>
            </PermissionGuard>
            <PermissionGuard permission={PERMISSIONS.USER_DELETE}>
              <Tooltip title="ลบ">
                <Button
                  type="text"
                  danger
                  aria-label={`ลบ ${user.fullName}`}
                  icon={<DeleteOutlined />}
                  onClick={() => confirmDelete(user)}
                />
              </Tooltip>
            </PermissionGuard>
          </Space>
        ),
      },
  ];

  const changePage = (pagination: TablePaginationConfig) =>
    setFilters((current) => ({
      ...current,
      page: pagination.current || 1,
      pageSize: pagination.pageSize || PAGINATION.DEFAULT_PAGE_SIZE,
    }));

  return (
    <PageTransition>
      <div className="users-page">
        <div className="users-page-header">
          <div>
            <h1 className="text-xl font-bold">
              User Management - จัดการผู้ใช้งาน
            </h1>
          </div>
          <Space wrap>
            <PermissionGuard permission={PERMISSIONS.USER_CREATE}>
              <Button
                size="large"
                icon={<UserAddOutlined />}
                onClick={() => openEditor()}
              >
                เพิ่มผู้ใช้งานใหม่
              </Button>
            </PermissionGuard>
            <PermissionGuard permission={PERMISSIONS.USER_EDIT}>
              <Button
                type="primary"
                size="large"
                icon={<PlusOutlined />}
                onClick={() => setBulkAccessOpen(true)}
              >
                เพิ่ม User เข้า Product
              </Button>
            </PermissionGuard>
          </Space>
        </div>

        <Card className="users-filter-card">
          <Input
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            prefix={<SearchOutlined />}
            placeholder="ค้นหาชื่อ อีเมล หรือ Username"
            allowClear
          />
          <Select
            value={filters.status}
            onChange={(status) =>
              setFilters((current) => ({ ...current, status, page: 1 }))
            }
            allowClear
            placeholder="ทุกสถานะ"
            options={[
              { label: "Active", value: "active" },
              { label: "Deactive", value: "inactive" },
              { label: "Pending", value: "pending" },
            ]}
          />
          <Button
            icon={<ReloadOutlined />}
            onClick={() => {
              setSearch("");
              setFilters({ page: 1, pageSize: 10 });
            }}
          >
            ล้างตัวกรอง
          </Button>
        </Card>

        {users.isError ? (
          <Alert
            type="error"
            showIcon
            message="โหลดรายชื่อผู้ใช้งานไม่สำเร็จ"
            description={errorMessage(users.error, "กรุณาลองอีกครั้ง")}
            action={
              <Button onClick={() => users.refetch()}>ลองอีกครั้ง</Button>
            }
          />
        ) : users.isLoading ? (
          <TableSkeleton />
        ) : (
          <Card className="users-table-card">
            <Table<User>
              rowKey="id"
              columns={columns}
              dataSource={users.data?.data ?? []}
              loading={users.isFetching}
              scroll={{ x: 980 }}
              onChange={changePage}
              pagination={{
                current: filters.page,
                pageSize: filters.pageSize,
                total: users.data?.total ?? 0,
                showSizeChanger: true,
                pageSizeOptions: [...PAGINATION.PAGE_SIZE_OPTIONS],
                showTotal: (total) => `ทั้งหมด ${total} คน`,
              }}
              locale={{ emptyText: "ไม่พบผู้ใช้งานตามเงื่อนไขที่เลือก" }}
            />
          </Card>
        )}
      </div>

      <UserFormModal
        open={formOpen}
        user={editingUser}
        loading={createUser.isPending || updateUser.isPending}
        onCancel={() => {
          setFormOpen(false);
          setEditingUser(undefined);
        }}
        onSubmit={submitUser}
      />
      <UserAccessDrawer
        open={Boolean(accessUser)}
        user={accessUser}
        products={productAccess}
        onClose={() => setAccessUser(undefined)}
        onChanged={refreshAccess}
      />
      <BulkProductAccessModal
        open={bulkAccessOpen}
        products={productAccess}
        onClose={() => setBulkAccessOpen(false)}
        onSaved={() => {
          refreshAccess();
          void client.invalidateQueries({ queryKey: QUERY_KEYS.USERS });
        }}
      />
    </PageTransition>
  );
}
