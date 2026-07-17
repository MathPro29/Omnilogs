import { useEffect, useState, useCallback, useMemo } from "react";
import {
  Card,
  Table,
  Button,
  Input,
  Select,
  Tag,
  Space,
  Typography,
  Modal,
  message,
  Tooltip,
  Avatar,
  Row,
  Col,
  Form,
} from "antd";
import type { ColumnsType, TablePaginationConfig } from "antd/es/table";
import {
  MagnifyingGlassIcon,
  PlusIcon,
  PencilSquareIcon,
  TrashIcon,
  ArrowDownTrayIcon,
  FunnelIcon,
} from "@heroicons/react/24/outline";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { motion } from "motion/react";
import { userService } from "@/services";
import { useAppStore } from "@/store";
import {
  PERMISSIONS,
  QUERY_KEYS,
  ROUTES,
  PAGINATION,
  STATUS_LABELS,
  STATUS_COLORS,
  DEFAULT_ADMIN_FEATURES,
} from "@/constants";
import { formatDate, getInitials, debounce } from "@/utils";
import { PageTransition, PermissionGuard, TableSkeleton } from "@/components";
import type { User, UserFilterParams } from "@/types";

const { Title, Text } = Typography;
const { confirm } = Modal;

const FEATURE_OPTIONS = [
  { label: "แดชบอร์ด (Dashboard)", value: PERMISSIONS.FEATURE_DASHBOARD },
  { label: "Logs Explorer", value: PERMISSIONS.FEATURE_LOGS_EXPLORER },
  { label: "จัดการผลิตภัณฑ์ (Products)", value: PERMISSIONS.FEATURE_PRODUCTS },
  { label: "Retention", value: PERMISSIONS.FEATURE_RETENTION_TEST },
  { label: "จัดการผู้ใช้งาน (Users)", value: PERMISSIONS.FEATURE_USERS },
  { label: "จัดการบทบาทและสิทธิ์ (Roles)", value: PERMISSIONS.FEATURE_ROLES },
  { label: "บันทึกกิจกรรม (Audit Logs)", value: PERMISSIONS.FEATURE_AUDITS },
  { label: "Sensitive Log Access", value: PERMISSIONS.FEATURE_SENSITIVE_ACCESS },
  { label: "Custom Fields", value: PERMISSIONS.FEATURE_CUSTOM_FIELDS },
  { label: "Product Roles", value: PERMISSIONS.FEATURE_PRODUCT_ROLES },
  { label: "API Test Page", value: PERMISSIONS.FEATURE_API_TEST },
];

export function UsersPage() {
  const queryClient = useQueryClient();
  const setBreadcrumbs = useAppStore((state) => state.setBreadcrumbs);

  const [filters, setFilters] = useState<UserFilterParams>({
    page: PAGINATION.DEFAULT_PAGE,
    pageSize: PAGINATION.DEFAULT_PAGE_SIZE,
    search: "",
    status: undefined,
    department: "",
  });

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [form] = Form.useForm();

  useEffect(() => {
    setBreadcrumbs([
      { title: "จัดการผู้ใช้" },
      { title: "ผู้ใช้งาน", path: ROUTES.USERS },
    ]);
  }, [setBreadcrumbs]);

  // Fetch users
  const { data, isLoading } = useQuery({
    queryKey: [...QUERY_KEYS.USERS, filters],
    queryFn: () => userService.getUsers(filters),
  });

  // Create mutation
  const createMutation = useMutation({
    mutationFn: userService.createUser,
    onSuccess: () => {
      message.success("สร้างผู้ใช้งานสำเร็จ");
      setIsModalOpen(false);
      form.resetFields();
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.USERS });
    },
    onError: (err: any) => {
      const errMsg = err?.response?.data?.message || err?.message || "เกิดข้อผิดพลาดในการสร้างผู้ใช้งาน";
      message.error(errMsg);
    },
  });

  // Update mutation
  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<User> & { role?: string } }) =>
      userService.updateUser(id, data),
    onSuccess: () => {
      message.success("แก้ไขข้อมูลผู้ใช้งานสำเร็จ");
      setIsModalOpen(false);
      setEditingUser(null);
      form.resetFields();
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.USERS });
    },
    onError: (err: any) => {
      const errMsg = err?.response?.data?.message || err?.message || "เกิดข้อผิดพลาดในการแก้ไขข้อมูลผู้ใช้งาน";
      message.error(errMsg);
    },
  });

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: userService.deleteUser,
    onSuccess: () => {
      message.success("ลบผู้ใช้สำเร็จ");
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.USERS });
    },
    onError: () => {
      message.error("เกิดข้อผิดพลาดในการลบผู้ใช้");
    },
  });

  // eslint-disable-next-line react-hooks/exhaustive-deps
  const debouncedSearch = useCallback(
    debounce((value: string) => {
      setFilters((prev) => ({ ...prev, search: value, page: 1 }));
    }, 400),
    [],
  );

  // useCallback: ส่งเป็น props ให้ Table column render
  const handleDelete = useCallback((user: User) => {
    confirm({
      title: "ยืนยันการลบ",
      content: `ต้องการลบผู้ใช้ "${user.fullName}" หรือไม่?`,
      okText: "ลบ",
      okType: "danger",
      cancelText: "ยกเลิก",
      onOk: () => deleteMutation.mutate(user.id),
    });
  }, [deleteMutation]);

  // useCallback: ส่งเป็น onChange ให้ Table
  const handleTableChange = useCallback((pagination: TablePaginationConfig) => {
    setFilters((prev) => ({
      ...prev,
      page: pagination.current || 1,
      pageSize: pagination.pageSize || PAGINATION.DEFAULT_PAGE_SIZE,
    }));
  }, []);

  // useCallback: ส่งเป็น onClick ให้ Button
  const handleResetFilters = useCallback(() => {
    setFilters({
      page: PAGINATION.DEFAULT_PAGE,
      pageSize: PAGINATION.DEFAULT_PAGE_SIZE,
      search: "",
      status: undefined,
      department: "",
    });
  }, []);

  const handleFormSubmit = (values: any) => {
    const role = values.role;
    const isAdminRole = ["god", "owner", "superadmin"].includes(role);
    const permissions = isAdminRole ? [] : (values.permissions || []);

    if (editingUser) {
      updateMutation.mutate({
        id: editingUser.id,
        data: {
          fullName: values.fullName,
          email: values.email,
          phone: values.phone,
          status: values.status,
          role: values.role,
          permissions,
        },
      });
    } else {
      createMutation.mutate({
        username: values.username,
        fullName: values.fullName,
        email: values.email,
        phone: values.phone,
        password: values.password,
        role: values.role,
        permissions,
      });
    }
  };

  // useMemo: columns config ไม่เปลี่ยน ไม่ต้องสร้างใหม่ทุก render
  const columns = useMemo<ColumnsType<User>>(() => [
    {
      title: "ผู้ใช้งาน",
      key: "user",
      render: (_, record) => (
        <div className="flex items-center gap-3">
          <Avatar
            size={40}
            style={{
              backgroundColor: 'var(--color-primary)',
              fontWeight: 600,
              fontSize: "0.75rem",
            }}
          >
            {getInitials(record.fullName)}
          </Avatar>
          <div>
            <div className="font-medium" style={{ color: "#000000D9" }}>
              {record.fullName}
            </div>
            <div className="text-xs" style={{ color: "#00000073" }}>
              {record.email}
            </div>
          </div>
        </div>
      ),
    },
    {
      title: "แผนก",
      dataIndex: "department",
      key: "department",
      render: (dept: string) => dept || "-",
    },
    {
      title: "ตำแหน่ง",
      dataIndex: "position",
      key: "position",
      render: (pos: string) => pos || "-",
    },
    {
      title: "บทบาท",
      dataIndex: "roles",
      key: "roles",
      render: (roles: User["roles"]) => (
        <div className="flex gap-1 flex-wrap">
          {roles.map((r) => (
            <Tag key={r.id} color="purple">
              {r.name}
            </Tag>
          ))}
        </div>
      ),
    },
    {
      title: "สิทธิ์การเข้าถึงฟีเจอร์",
      key: "permissions",
      render: (_, record) => {
        const role = record.roles?.[0]?.name || "user";
        const isAdminRole = ["god", "owner", "superadmin"].includes(role);
        if (isAdminRole) {
          return <Tag color="gold">ทั้งหมด (All)</Tag>;
        }

        let displayPermissions = record.permissions || [];
        if (role.toLowerCase() === 'admin') {
          displayPermissions = Array.from(new Set([...DEFAULT_ADMIN_FEATURES, ...displayPermissions]));
        }

        if (displayPermissions.length === 0) {
          return <span className="text-gray-400 font-normal">—</span>;
        }
        return (
          <div className="flex gap-1 flex-wrap max-w-xs">
            {displayPermissions.map((perm) => {
              const opt = FEATURE_OPTIONS.find((o) => o.value === perm);
              return (
                <Tag key={perm} color="blue" bordered={false}>
                  {opt ? opt.label.split(" (")[0] : perm}
                </Tag>
              );
            })}
          </div>
        );
      },
    },
    {
      title: "สถานะ",
      dataIndex: "status",
      key: "status",
      render: (status: string) => (
        <Tag color={STATUS_COLORS[status]}>{STATUS_LABELS[status]}</Tag>
      ),
    },
    {
      title: "วันที่สร้าง",
      dataIndex: "createdAt",
      key: "createdAt",
      render: (date: string) => formatDate(date),
    },
    {
      title: "การดำเนินการ",
      key: "actions",
      width: 120,
      render: (_, record) => (
        <Space size="small">
          <PermissionGuard permission={PERMISSIONS.USER_EDIT}>
            <Tooltip title="แก้ไข">
              <Button
                type="text"
                size="small"
                icon={<PencilSquareIcon className="w-4 h-4" />}
                onClick={() => {
                  setEditingUser(record);
                  const role = record.roles?.[0]?.name || "user";
                  let recordPerms = record.permissions || [];
                  if (role.toLowerCase() === 'admin' && recordPerms.length === 0) {
                    recordPerms = DEFAULT_ADMIN_FEATURES;
                  }
                  form.setFieldsValue({
                    fullName: record.fullName,
                    email: record.email,
                    phone: record.phone,
                    role: role,
                    status: record.status,
                    permissions: recordPerms,
                  });
                  setIsModalOpen(true);
                }}
              />
            </Tooltip>
          </PermissionGuard>
          <PermissionGuard permission={PERMISSIONS.USER_DELETE}>
            <Tooltip title="ลบ">
              <Button
                type="text"
                danger
                size="small"
                icon={<TrashIcon className="w-4 h-4" />}
                onClick={() => handleDelete(record)}
              />
            </Tooltip>
          </PermissionGuard>
        </Space>
      ),
    },
  ], [handleDelete]);

  if (isLoading) {
    return (
      <PageTransition>
        <div className="space-y-6">
          <div>
            <Title level={4} className="mb-1">
              ผู้ใช้งาน
            </Title>
            <Text type="secondary">จัดการข้อมูลผู้ใช้งานในระบบ</Text>
          </div>
          <TableSkeleton />
        </div>
      </PageTransition>
    );
  }

  return (
    <PageTransition>
      <div className="space-y-6">
        {/* Page Header */}
        <div className="flex items-center justify-between flex-wrap gap-4">
          <div>
            <Title level={4} className="mb-1">
              ผู้ใช้งาน
            </Title>
            <Text type="secondary">
              จัดการข้อมูลผู้ใช้งานในระบบ ({data?.total || 0} รายการ)
            </Text>
          </div>
          <div className="flex gap-2">
            <PermissionGuard permission={PERMISSIONS.USER_EXPORT}>
              <Button icon={<ArrowDownTrayIcon className="w-4 h-4" />}>
                ส่งออก
              </Button>
            </PermissionGuard>
            <PermissionGuard permission={PERMISSIONS.USER_CREATE}>
              <Button
                type="primary"
                icon={<PlusIcon className="w-4 h-4" />}
                onClick={() => {
                  setEditingUser(null);
                  form.resetFields();
                  setIsModalOpen(true);
                }}
              >
                เพิ่มผู้ใช้
              </Button>
            </PermissionGuard>
          </div>
        </div>

        {/* Filter Bar */}
        <motion.div
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1, duration: 0.25 }}
        >
          <Card bordered={false} className="stat-card">
            <Row gutter={[12, 12]} align="middle">
              <Col xs={24} sm={12} md={8}>
                <Input
                  placeholder="ค้นหาชื่อ, อีเมล, ชื่อผู้ใช้..."
                  prefix={
                    <MagnifyingGlassIcon className="w-4 h-4 text-gray-400" />
                  }
                  onChange={(e) => debouncedSearch(e.target.value)}
                  allowClear
                />
              </Col>
              <Col xs={12} sm={6} md={4}>
                <Select
                  placeholder="สถานะ"
                  className="w-full"
                  allowClear
                  value={filters.status || undefined}
                  onChange={(value) =>
                    setFilters((prev) => ({ ...prev, status: value, page: 1 }))
                  }
                  options={[
                    { label: "ใช้งาน", value: "active" },
                    { label: "ไม่ใช้งาน", value: "inactive" },
                    { label: "รอดำเนินการ", value: "pending" },
                  ]}
                />
              </Col>
              <Col xs={12} sm={6} md={4}>
                <Select
                  placeholder="แผนก"
                  className="w-full"
                  allowClear
                  value={filters.department || undefined}
                  onChange={(value) =>
                    setFilters((prev) => ({
                      ...prev,
                      department: value || "",
                      page: 1,
                    }))
                  }
                  options={[
                    { label: "IT", value: "IT" },
                    { label: "HR", value: "HR" },
                    { label: "Finance", value: "Finance" },
                    { label: "Sales", value: "Sales" },
                    { label: "Marketing", value: "Marketing" },
                    { label: "Operations", value: "Operations" },
                  ]}
                />
              </Col>
              <Col>
                <Button
                  icon={<FunnelIcon className="w-4 h-4" />}
                  onClick={handleResetFilters}
                >
                  ล้างตัวกรอง
                </Button>
              </Col>
            </Row>
          </Card>
        </motion.div>

        {/* Table */}
        <motion.div
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2, duration: 0.25 }}
        >
          <Card bordered={false} className="stat-card">
            <Table
              className="admin-table"
              columns={columns}
              dataSource={data?.data || []}
              rowKey="id"
              loading={isLoading}
              pagination={{
                current: filters.page,
                pageSize: filters.pageSize,
                total: data?.total || 0,
                showSizeChanger: true,
                pageSizeOptions: [...PAGINATION.PAGE_SIZE_OPTIONS],
                showTotal: (total, range) =>
                  `${range[0]}-${range[1]} จาก ${total} รายการ`,
              }}
              onChange={handleTableChange}
              scroll={{ x: 900 }}
            />
          </Card>
        </motion.div>
      </div>

      {/* Create / Edit User Modal */}
      <Modal
        title={editingUser ? "แก้ไขข้อมูลผู้ใช้งาน" : "เพิ่มผู้ใช้งานใหม่"}
        open={isModalOpen}
        onCancel={() => {
          setIsModalOpen(false);
          setEditingUser(null);
          form.resetFields();
        }}
        onOk={() => form.submit()}
        okText="บันทึก"
        cancelText="ยกเลิก"
        confirmLoading={createMutation.isPending || updateMutation.isPending}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleFormSubmit}
          initialValues={{ role: "user", status: "active" }}
        >
          {!editingUser && (
            <Form.Item
              name="username"
              label="ชื่อผู้ใช้ (Username)"
              rules={[
                { required: true, message: "กรุณากรอกชื่อผู้ใช้" },
                { min: 3, message: "ชื่อผู้ใช้ต้องมีอย่างน้อย 3 ตัวอักษร" },
              ]}
            >
              <Input placeholder="เช่น somsom" />
            </Form.Item>
          )}
          <Form.Item
            name="fullName"
            label="ชื่อ-นามสกุล (Full Name)"
            rules={[{ required: true, message: "กรุณากรอกชื่อ-นามสกุล" }]}
          >
            <Input placeholder="เช่น สมชาย ใจดี" />
          </Form.Item>
          <Form.Item
            name="email"
            label="อีเมล (Email)"
            rules={[
              { required: true, message: "กรุณากรอกอีเมล" },
              { type: "email", message: "รูปแบบอีเมลไม่ถูกต้อง" },
            ]}
          >
            <Input placeholder="เช่น somchai@company.com" />
          </Form.Item>
          <Form.Item
            name="phone"
            label="เบอร์โทรศัพท์ (Phone Number)"
            rules={[
              {
                pattern: /^[0-9]{8,20}$/,
                message: "รูปแบบเบอร์โทรศัพท์ไม่ถูกต้อง",
              },
            ]}
          >
            <Input placeholder="เช่น 0891234567" />
          </Form.Item>
          {!editingUser && (
            <Form.Item
              name="password"
              label="รหัสผ่าน (Password)"
              rules={[
                { required: true, message: "กรุณากรอกรหัสผ่าน" },
                { min: 8, message: "รหัสผ่านต้องมีอย่างน้อย 8 ตัวอักษร" },
              ]}
            >
              <Input.Password placeholder="รหัสผ่านสำหรับเข้าสู่ระบบ" />
            </Form.Item>
          )}
          <Form.Item
            name="role"
            label="บทบาทในระบบ (Platform Role)"
            rules={[{ required: true, message: "กรุณาเลือกบทบาท" }]}
          >
            <Select
              onChange={(value) => {
                if (value === "admin") {
                  form.setFieldsValue({ permissions: DEFAULT_ADMIN_FEATURES });
                } else if (["god", "owner", "superadmin"].includes(value)) {
                  form.setFieldsValue({ permissions: [] });
                }
              }}
              options={[
                { label: "GOD", value: "god" },
                { label: "Owner", value: "owner" },
                { label: "Superadmin", value: "superadmin" },
                { label: "Admin", value: "admin" },
                { label: "User", value: "user" },
              ]}
            />
          </Form.Item>
          <Form.Item
            noStyle
            shouldUpdate={(prevValues, currentValues) => prevValues.role !== currentValues.role}
          >
            {({ getFieldValue }) => {
              const role = getFieldValue("role");
              const isAdminRole = ["god", "owner", "superadmin"].includes(role);
              return (
                <Form.Item
                  name="permissions"
                  label="ฟังก์ชันที่สามารถเข้าถึงได้ (Feature Access)"
                  extra="สิทธิ์ของกลุ่มผู้ดูแลระบบ (Admin/GOD/Superadmin) จะเข้าถึงได้ทุกหน้าโดยอัตโนมัติ"
                >
                  <Select
                    mode="multiple"
                    placeholder={isAdminRole ? "มีสิทธิ์เข้าถึงทุกฟังก์ชันโดยอัตโนมัติ" : "เลือกฟังก์ชันที่อนุญาตให้เข้าถึง"}
                    disabled={isAdminRole}
                    options={FEATURE_OPTIONS}
                    style={{ width: "100%" }}
                  />
                </Form.Item>
              );
            }}
          </Form.Item>
          {editingUser && (
            <Form.Item
              name="status"
              label="สถานะ (Status)"
              rules={[{ required: true, message: "กรุณาเลือกสถานะ" }]}
            >
              <Select
                options={[
                  { label: "ใช้งาน (Active)", value: "active" },
                  { label: "ไม่ใช้งาน (Inactive)", value: "inactive" },
                ]}
              />
            </Form.Item>
          )}
        </Form>
      </Modal>
    </PageTransition>
  );
}
