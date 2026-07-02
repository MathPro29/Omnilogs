import { useEffect, useState, useCallback } from 'react';
import {
  Card,
  Form,
  Select,
  Button,
  DatePicker,
  Typography,
  Table,
  Tag,
  Modal,
  message,
  Divider,
  Row,
  Col,
  Empty,
  Tooltip,
} from 'antd';
import {
  ShieldCheckIcon,
  UserPlusIcon,
  TrashIcon,
  LockClosedIcon,
} from '@heroicons/react/24/outline';
import { Navigate } from 'react-router-dom';
import { useAuthStore, useAppStore } from '@/store';
import { ROUTES } from '@/constants';
import { productAdminService, userService } from '@/services';
import { PageTransition } from '@/components';
import type {
  Product,
  Project,
  ProjectFeature,
  ProductRoleDefinition,
  User,
} from '@/types';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { confirm } = Modal;

interface GrantFormValues {
  productId: number;
  projectId?: number;
  categoryIds?: number[];
  userIds: number[];
  roleId: number;
  expiresAt?: dayjs.Dayjs;
}

interface ScopeTableRow {
  key: string;
  scopeId: number;
  membershipId: number;
  userId: number;
  userFullName: string;
  userEmail: string;
  roleName: string;
  scopeLevel: 'PRODUCT' | 'PROJECT' | 'CATEGORY';
  projectName?: string;
  featureName?: string;
  isActive: boolean;
  expiresAt?: string | null;
}

interface ScopePayloadPreview {
  scopeLevel: 'PRODUCT' | 'PROJECT' | 'CATEGORY';
  projectId?: number | null;
  categoryId?: number | null;
}

function scopePayloadKey(value: ScopePayloadPreview): string {
  return `${value.scopeLevel}:${value.projectId ?? 'null'}:${value.categoryId ?? 'null'}`;
}

function scopeTableRowKey(value: {
  membershipId: number;
  scopeLevel: 'PRODUCT' | 'PROJECT' | 'CATEGORY';
  projectId?: number | null;
  categoryId?: number | null;
}): string {
  return `${value.membershipId}:${value.scopeLevel}:${value.projectId ?? 'null'}:${value.categoryId ?? 'null'}`;
}

export function AccessControlPage() {
  const setBreadcrumbs = useAppStore((state) => state.setBreadcrumbs);
  const roles = useAuthStore((state) => state.roles);
  
  // ตรวจสอบบทบาทของ platform user ว่าเป็น god, owner, หรือ superadmin หรือไม่
  const isGodOrOwner = roles.includes('god') || roles.includes('owner') || roles.includes('superadmin') || roles.includes('super_admin');

  const [form] = Form.useForm<GrantFormValues>();
  const watchedFormValues = Form.useWatch([], form) as GrantFormValues | undefined;
  
  // Data lists
  const [products, setProducts] = useState<Product[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [features, setFeatures] = useState<ProjectFeature[]>([]);
  const [productRoles, setProductRoles] = useState<ProductRoleDefinition[]>([]);
  const [systemUsers, setSystemUsers] = useState<User[]>([]);
  
  // Selection states
  const [selectedProductId, setSelectedProductId] = useState<number | null>(null);
  const [selectedProjectId, setSelectedProjectId] = useState<number | null>(null);
  
  // Table state
  const [tableData, setTableData] = useState<ScopeTableRow[]>([]);
  
  // Loading states
  const [loadingInitial, setLoadingInitial] = useState(true);
  const [loadingData, setLoadingData] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const buildScopePayloads = (values: GrantFormValues): ScopePayloadPreview[] => {
    if (values.projectId && values.categoryIds && values.categoryIds.length > 0) {
      return values.categoryIds.map((categoryId) => ({
        scopeLevel: 'CATEGORY',
        projectId: values.projectId,
        categoryId,
      }));
    }

    if (values.projectId) {
      return [
        {
          scopeLevel: 'PROJECT',
          projectId: values.projectId,
          categoryId: null,
        },
      ];
    }

    return [
      {
        scopeLevel: 'PRODUCT',
        projectId: null,
        categoryId: null,
      },
    ];
  };

  // Set breadcrumbs
  useEffect(() => {
    setBreadcrumbs([
      { title: 'จัดการผู้ใช้' },
      { title: 'บทบาทและสิทธิ์ (จัดการสิทธิ์เข้าถึง)', path: ROUTES.ROLES },
    ]);
  }, [setBreadcrumbs]);

  // Load initial products and users
  const loadInitialData = useCallback(async () => {
    try {
      setLoadingInitial(true);
      const [productsData, usersData] = await Promise.all([
        productAdminService.listProducts(),
        userService.getUsers({ page: 1, pageSize: 1000 }),
      ]);
      setProducts(productsData);
      setSystemUsers(usersData.data);
    } catch {
      message.error('เกิดข้อผิดพลาดในการโหลดข้อมูลเริ่มต้น');
    } finally {
      setLoadingInitial(false);
    }
  }, []);

  useEffect(() => {
    if (isGodOrOwner) {
      loadInitialData();
    }
  }, [isGodOrOwner, loadInitialData]);

  // Load projects, features, roles, and current scopes when product changes
  const loadProductDependentData = useCallback(async (productId: number) => {
    try {
      setLoadingData(true);
      const [projectsData, rolesData, membershipsData] = await Promise.all([
        productAdminService.listProjects(productId),
        productAdminService.listRoles(productId),
        productAdminService.listMemberships(productId),
      ]);

      setProjects(projectsData);
      setProductRoles(rolesData);
      
      // Load scopes for each membership
      const rows: ScopeTableRow[] = [];
      await Promise.all(
        membershipsData.map(async (membership) => {
          try {
            const scopes = await productAdminService.listScopes(productId, membership.membershipId);
            
            // Map member user details
            const matchedUser = systemUsers.find((u) => String(u.id) === String(membership.userId));
            const userFullName = matchedUser?.fullName || `User ID: ${membership.userId}`;
            const userEmail = matchedUser?.email || 'N/A';
            
            // Map role details
            const matchedRole = rolesData.find((r) => r.roleId === membership.roleId);
            const roleName = matchedRole?.roleName || `Role ID: ${membership.roleId}`;

            scopes.forEach((scope) => {
              // Get project name if scoped to project or category
              let projectName = '';
              if (scope.projectId) {
                const proj = projectsData.find((p) => p.projectId === scope.projectId);
                projectName = proj?.projectName || `Project ID: ${scope.projectId}`;
              }

              // Load features if category is scoped
              rows.push({
                key: `${membership.membershipId}-${scope.scopeId}`,
                scopeId: scope.scopeId,
                membershipId: membership.membershipId,
                userId: membership.userId,
                userFullName,
                userEmail,
                roleName,
                scopeLevel: scope.scopeLevel,
                projectName: scope.projectId ? projectName : undefined,
                featureName: scope.categoryId ? `Feature ID: ${scope.categoryId}` : undefined,
                isActive: scope.isActive,
                expiresAt: membership.expiresAt,
              });
            });
          } catch (err) {
            console.error('Failed to load scopes for membership', membership.membershipId, err);
          }
        })
      );

      // Now query category details for category level scopes to resolve category/feature names
      // (Do this sequentially or parallel after rows are gathered)
      for (const row of rows) {
        if (row.scopeLevel === 'CATEGORY' && row.projectName) {
          const membership = membershipsData.find((m) => m.membershipId === row.membershipId);
          if (membership && row.projectName) {
            const proj = projectsData.find((p) => p.projectName === row.projectName);
            if (proj) {
              try {
                const projFeatures = await productAdminService.listFeatures(productId, proj.projectId);
                // Try finding matching category
                const scopeItem = rows.find((r) => r.key === row.key);
                if (scopeItem) {
                  // Find the target scope detail from our mocked/raw DB
                  const scopesForMem = await productAdminService.listScopes(productId, row.membershipId);
                  const thisScope = scopesForMem.find((s) => s.scopeId === row.scopeId);
                  if (thisScope && thisScope.categoryId) {
                    const feat = projFeatures.find((f) => f.categoryId === thisScope.categoryId);
                    scopeItem.featureName = feat?.categoryName || `Feature ID: ${thisScope.categoryId}`;
                  }
                }
              } catch (e) {
                console.error(e);
              }
            }
          }
        }
      }

      setTableData(rows);
    } catch {
      message.error('เกิดข้อผิดพลาดในการโหลดข้อมูลโครงสร้างระบบ');
    } finally {
      setLoadingData(false);
    }
  }, [systemUsers]);

  // Load features when project changes
  const loadProjectFeatures = useCallback(async (productId: number, projectId: number) => {
    try {
      const featuresData = await productAdminService.listFeatures(productId, projectId);
      setFeatures(featuresData);
    } catch {
      message.error('เกิดข้อผิดพลาดในการโหลดข้อมูล Feature');
    }
  }, []);

  // Handle Product selection change
  const handleProductChange = (productId: number) => {
    setSelectedProductId(productId);
    setSelectedProjectId(null);
    setFeatures([]);
    setProjects([]);
    setProductRoles([]);
    setTableData([]);
    form.setFieldsValue({ projectId: undefined, categoryIds: undefined });
    loadProductDependentData(productId);
  };

  // Handle Project selection change
  const handleProjectChange = (projectId: number) => {
    setSelectedProjectId(projectId);
    setFeatures([]);
    form.setFieldsValue({ categoryIds: undefined });
    if (selectedProductId) {
      loadProjectFeatures(selectedProductId, projectId);
    }
  };

  // Handle Revoke Scope Access
  const handleRevokeScope = (record: ScopeTableRow) => {
    confirm({
      title: 'ยืนยันการเพิกถอนสิทธิ์',
      content: `ต้องการเพิกถอนสิทธิ์เข้าถึงของ "${record.userFullName}" สำหรับ ${
        record.scopeLevel === 'PRODUCT'
          ? 'ทั้ง Product'
          : record.scopeLevel === 'PROJECT'
          ? `โปรเจกต์: ${record.projectName}`
          : `ฟีเจอร์: ${record.featureName} (${record.projectName})`
      } หรือไม่?`,
      okText: 'เพิกถอนสิทธิ์',
      okType: 'danger',
      cancelText: 'ยกเลิก',
      onOk: async () => {
        try {
          if (selectedProductId) {
            await productAdminService.deleteScope(selectedProductId, record.membershipId, record.scopeId);
            message.success('เพิกถอนสิทธิ์เข้าถึงสำเร็จ');
            loadProductDependentData(selectedProductId);
          }
        } catch {
          message.error('เกิดข้อผิดพลาดในการเพิกถอนสิทธิ์');
        }
      },
    });
  };

  // Handle Submit Form
  const handleFinish = async (values: GrantFormValues) => {
    setSubmitting(true);
    try {
      const { productId, userIds, roleId, expiresAt } = values;
      const formattedExpiresAt = expiresAt ? expiresAt.toISOString() : null;
      const scopePayloads = buildScopePayloads(values);
      const rolesById = new Map(productRoles.map((role) => [role.roleId, role]));
      const projectsById = new Map(projects.map((project) => [project.projectId, project]));
      const featuresById = new Map(features.map((feature) => [feature.categoryId, feature]));
      const affectedRows: ScopeTableRow[] = [];
      const memberships = await productAdminService.listMemberships(productId);

      // Loop through all selected users and grant access
      for (const userId of userIds) {
        // 1. Check if user already has membership
        let membership = memberships.find((m) => String(m.userId) === String(userId));

        if (!membership) {
          // Create new membership
          try {
            membership = await productAdminService.createMembership(productId, {
              userId,
              roleId,
              expiresAt: formattedExpiresAt,
            });
            memberships.push(membership);
          } catch (err: any) {
            throw new Error(err?.response?.data?.error?.message || `Create membership failed for user ${userId}`);
          }
        } else if (membership.roleId !== roleId) {
          // If membership role has changed, update it
          try {
            membership = await productAdminService.updateMembership(productId, membership.membershipId, {
              roleId,
              expiresAt: formattedExpiresAt,
            });
          } catch (err: any) {
            throw new Error(err?.response?.data?.error?.message || `Update membership failed for user ${userId}`);
          }
        }

        const matchedUser = systemUsers.find((u) => String(u.id) === String(userId));
        const roleName = rolesById.get(membership.roleId)?.roleName || `Role ID: ${membership.roleId}`;

        // 2. Create target scope(s) for this membership
        const existingScopes = await productAdminService.listScopes(productId, membership.membershipId);
        const existingScopeKeys = new Set(
          existingScopes.map((scope) =>
            scopePayloadKey({
              scopeLevel: scope.scopeLevel,
              projectId: scope.projectId ?? null,
              categoryId: scope.categoryId ?? null,
            })
          )
        );

        for (const scopePayload of scopePayloads) {
          const existingScope = existingScopes.find(
            (scope) =>
              scope.scopeLevel === scopePayload.scopeLevel &&
              (scope.projectId ?? null) === (scopePayload.projectId ?? null) &&
              (scope.categoryId ?? null) === (scopePayload.categoryId ?? null)
          );

          if (existingScopeKeys.has(scopePayloadKey(scopePayload)) && existingScope) {
            affectedRows.push({
              key: scopeTableRowKey({
                membershipId: membership.membershipId,
                scopeLevel: existingScope.scopeLevel,
                projectId: existingScope.projectId ?? null,
                categoryId: existingScope.categoryId ?? null,
              }),
              scopeId: existingScope.scopeId,
              membershipId: membership.membershipId,
              userId,
              userFullName: matchedUser?.fullName || `User ID: ${userId}`,
              userEmail: matchedUser?.email || 'N/A',
              roleName,
              scopeLevel: existingScope.scopeLevel,
              projectName: existingScope.projectId
                ? projectsById.get(existingScope.projectId)?.projectName || `Project ID: ${existingScope.projectId}`
                : undefined,
              featureName: existingScope.categoryId
                ? featuresById.get(existingScope.categoryId)?.categoryName || `Feature ID: ${existingScope.categoryId}`
                : undefined,
              isActive: existingScope.isActive,
              expiresAt: membership.expiresAt,
            });
            continue;
          }
          try {
            const createdScope = await productAdminService.createScope(productId, membership.membershipId, scopePayload);
            affectedRows.push({
              key: scopeTableRowKey({
                membershipId: membership.membershipId,
                scopeLevel: createdScope.scopeLevel,
                projectId: createdScope.projectId ?? null,
                categoryId: createdScope.categoryId ?? null,
              }),
              scopeId: createdScope.scopeId,
              membershipId: membership.membershipId,
              userId,
              userFullName: matchedUser?.fullName || `User ID: ${userId}`,
              userEmail: matchedUser?.email || 'N/A',
              roleName,
              scopeLevel: createdScope.scopeLevel,
              projectName: createdScope.projectId
                ? projectsById.get(createdScope.projectId)?.projectName || `Project ID: ${createdScope.projectId}`
                : undefined,
              featureName: createdScope.categoryId
                ? featuresById.get(createdScope.categoryId)?.categoryName || `Feature ID: ${createdScope.categoryId}`
                : undefined,
              isActive: createdScope.isActive,
              expiresAt: membership.expiresAt,
            });
          } catch (err: any) {
            throw new Error(
              err?.response?.data?.error?.message ||
                `Create scope failed for user ${userId}: ${JSON.stringify(scopePayload)}`
            );
          }
        }
      }

      message.success(`มอบสิทธิ์การเข้าถึงให้ ${userIds.length} ผู้ใช้งานสำเร็จ`);
      form.setFieldsValue({ userIds: [], expiresAt: undefined, categoryIds: [] });
      setTableData((current) => {
        const deduped = new Map<string, ScopeTableRow>();
        const nextRoleName = rolesById.get(roleId)?.roleName;
        current
          .map((row) =>
            userIds.includes(row.userId)
              ? {
                  ...row,
                  roleName: nextRoleName || row.roleName,
                  expiresAt: formattedExpiresAt,
                }
              : row
          )
          .forEach((row) => deduped.set(row.key, row));
        affectedRows.forEach((row) => deduped.set(row.key, row));
        return Array.from(deduped.values());
      });
    } catch (err: any) {
      console.error(err);
      message.error(err.message || 'เกิดข้อผิดพลาดในการมอบสิทธิ์');
    } finally {
      setSubmitting(false);
    }
  };

  // Route protection redirect
  if (!isGodOrOwner) {
    return <Navigate to={ROUTES.FORBIDDEN} replace />;
  }

  // Table Columns
  const columns = [
    {
      title: 'ผู้ใช้งาน',
      key: 'user',
      render: (_: any, record: ScopeTableRow) => (
        <div>
          <div className="font-semibold text-gray-800 dark:text-gray-200">{record.userFullName}</div>
          <div className="text-xs text-gray-500">{record.userEmail}</div>
        </div>
      ),
    },
    {
      title: 'บทบาท (Product Role)',
      dataIndex: 'roleName',
      key: 'roleName',
      render: (roleName: string) => <Tag color="blue">{roleName}</Tag>,
    },
    {
      title: 'ระดับการเข้าถึง (Scope Level)',
      dataIndex: 'scopeLevel',
      key: 'scopeLevel',
      render: (level: string) => {
        const colors = {
          PRODUCT: 'purple',
          PROJECT: 'cyan',
          CATEGORY: 'orange',
        };
        const labels = {
          PRODUCT: 'Product (ทั้งหมด)',
          PROJECT: 'Project (โปรเจกต์)',
          CATEGORY: 'Feature (ฟีเจอร์)',
        };
        return <Tag color={colors[level as keyof typeof colors]}>{labels[level as keyof typeof labels] || level}</Tag>;
      },
    },
    {
      title: 'เป้าหมายสิทธิ์ (Target Scope)',
      key: 'target',
      render: (_: any, record: ScopeTableRow) => {
        if (record.scopeLevel === 'PRODUCT') {
          return <span className="text-gray-400">—</span>;
        }
        if (record.scopeLevel === 'PROJECT') {
          return (
            <div>
              <span className="text-xs text-gray-400">โปรเจกต์:</span>{' '}
              <span className="font-medium text-gray-700 dark:text-gray-300">{record.projectName}</span>
            </div>
          );
        }
        return (
          <div>
            <div>
              <span className="text-xs text-gray-400">โปรเจกต์:</span>{' '}
              <span className="text-gray-600 dark:text-gray-400 text-xs">{record.projectName}</span>
            </div>
            <div>
              <span className="text-xs text-gray-400">ฟีเจอร์:</span>{' '}
              <span className="font-medium text-gray-800 dark:text-gray-200">{record.featureName}</span>
            </div>
          </div>
        );
      },
    },
    {
      title: 'วันหมดอายุสิทธิ์',
      dataIndex: 'expiresAt',
      key: 'expiresAt',
      render: (date: string | null) =>
        date ? <span className="text-xs">{dayjs(date).format('DD/MM/YYYY')}</span> : <Tag color="default">ถาวร</Tag>,
    },
    {
      title: 'การจัดการ',
      key: 'action',
      render: (_: any, record: ScopeTableRow) => (
        <Tooltip title="เพิกถอนสิทธิ์การเข้าถึง">
          <Button
            type="text"
            danger
            icon={<TrashIcon className="w-4 h-4" />}
            onClick={() => handleRevokeScope(record)}
          />
        </Tooltip>
      ),
    },
  ];
  const scopePreview = buildScopePayloads(
    watchedFormValues ?? ({ productId: 0, roleId: 0, userIds: [] } as GrantFormValues)
  );

  return (
    <PageTransition>
      <div className="space-y-6 max-w-6xl mx-auto">
        
        {/* Page Header */}
        <div className="flex items-center justify-between">
          <div>
            <Title level={3} className="flex items-center gap-2 mb-1">
              <ShieldCheckIcon className="w-7 h-7 text-indigo-600" />
              การจัดการสิทธิ์การเข้าถึงสมาชิก (Grant Access Panel)
            </Title>
            <Text type="secondary">
              มอบสิทธิ์การเข้าถึงในระดับ Product, Project หรือ Feature (Category) ให้แก่ผู้ใช้งานครั้งละหลายคน
            </Text>
          </div>
          <Tag color="gold" icon={<LockClosedIcon className="w-3.5 h-3.5 inline mr-1" />} className="py-1 px-3">
            เฉพาะระดับ Owner / Superadmin / GOD
          </Tag>
        </div>

        <Row gutter={[24, 24]}>
          {/* Form Card */}
          <Col xs={24} lg={10}>
            <Card
              title={
                <span className="flex items-center gap-2">
                  <UserPlusIcon className="w-5 h-5 text-indigo-600" />
                  ฟอร์มกำหนดสิทธิ์การเข้าใช้งาน
                </span>
              }
              bordered={false}
              className="shadow-sm rounded-lg"
            >
              <Form
                form={form}
                layout="vertical"
                onFinish={handleFinish}
                initialValues={{ userIds: [] }}
              >
                
                {/* 1. Select Product */}
                <Form.Item
                  label="1. เลือก Product"
                  name="productId"
                  rules={[{ required: true, message: 'กรุณาเลือก Product' }]}
                >
                  <Select
                    placeholder="กรุณาเลือก Product"
                    options={products.map((p) => ({ value: p.productId, label: `${p.productName} (${p.productCode})` }))}
                    onChange={handleProductChange}
                    loading={loadingInitial}
                  />
                </Form.Item>

                {/* 2. Select Project */}
                <Form.Item
                  label="2. เลือก Project (ไม่จำเป็น — เว้นไว้ถ้าต้องการให้สิทธิ์ทั้ง Product)"
                  name="projectId"
                >
                  <Select
                    placeholder={selectedProductId ? "เลือกโปรเจกต์เป้าหมาย" : "กรุณาเลือก Product ก่อน"}
                    disabled={!selectedProductId}
                    options={projects.map((p) => ({ value: p.projectId, label: p.projectName }))}
                    onChange={handleProjectChange}
                    allowClear
                  />
                </Form.Item>

                {/* 3. Select Feature */}
                <Form.Item
                  label="3. เลือก Feature/Category (ไม่จำเป็น — เลือกได้มากกว่า 1 ฟีเจอร์ หรือเว้นไว้หากต้องการสิทธิ์ทั้ง Project)"
                  name="categoryIds"
                >
                  <Select
                    mode="multiple"
                    placeholder={selectedProjectId ? "เลือกฟีเจอร์เป้าหมาย" : "กรุณาเลือก Project ก่อน"}
                    disabled={!selectedProjectId}
                    options={features.map((f) => ({ value: f.categoryId, label: f.categoryName }))}
                    allowClear
                  />
                </Form.Item>

                <Divider className="my-4" />

                {/* 4. Select Users */}
                <Form.Item
                  label="4. เลือกผู้ใช้งานที่ต้องการมอบสิทธิ์ (เลือกได้มากกว่า 1 คน)"
                  name="userIds"
                  rules={[{ required: true, type: 'array', min: 1, message: 'กรุณาเลือกผู้ใช้อย่างน้อย 1 คน' }]}
                >
                  <Select
                    mode="multiple"
                    placeholder="ค้นหาและเลือกผู้ใช้"
                    style={{ width: '100%' }}
                    optionFilterProp="label"
                    options={systemUsers.map((u) => ({
                      value: Number(u.id),
                      label: `${u.fullName} (${u.email})`,
                    }))}
                    allowClear
                  />
                </Form.Item>

                {/* 5. Select Role */}
                <Form.Item
                  label="5. กำหนดบทบาทสิทธิ์ (Product Role)"
                  name="roleId"
                  rules={[{ required: true, message: 'กรุณาเลือกบทบาท' }]}
                >
                  <Select
                    placeholder={selectedProductId ? "เลือกบทบาทสิทธิ์" : "กรุณาเลือก Product ก่อน"}
                    disabled={!selectedProductId}
                    options={productRoles.map((r) => ({ value: r.roleId, label: r.roleName }))}
                  />
                </Form.Item>

                {/* 6. Expiry Date */}
                <Form.Item
                  label="6. วันสิ้นสุดการได้รับสิทธิ์ (หากกำหนดสิทธิ์ถาวร ไม่ต้องระบุ)"
                  name="expiresAt"
                >
                  <DatePicker style={{ width: '100%' }} placeholder="เลือกวันสิ้นสุดสิทธิ์" format="DD/MM/YYYY" />
                </Form.Item>

                <Form.Item className="mb-0">
                  <Button
                    type="primary"
                    htmlType="submit"
                    block
                    loading={submitting}
                    disabled={!selectedProductId}
                    className="bg-indigo-600 hover:bg-indigo-500"
                  >
                    มอบสิทธิ์เข้าใช้งาน
                  </Button>
                </Form.Item>
                <div className="mt-3 rounded border border-gray-200 bg-gray-50 p-3">
                  <div className="mb-2 text-xs font-semibold text-gray-700">Scope payload preview</div>
                  <pre className="overflow-x-auto text-xs text-gray-700">{JSON.stringify(scopePreview, null, 2)}</pre>
                </div>
              </Form>
            </Card>
          </Col>

          {/* Current Grants list Table */}
          <Col xs={24} lg={14}>
            <Card
              title={
                <div className="flex items-center justify-between">
                  <span>ตารางสิทธิ์การใช้งานปัจจุบัน</span>
                  {selectedProductId && (
                    <Text type="secondary" className="text-xs">
                      (Product ID: {selectedProductId})
                    </Text>
                  )}
                </div>
              }
              bordered={false}
              className="shadow-sm rounded-lg"
            >
              {!selectedProductId ? (
                <Empty
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                  description={
                    <span className="text-gray-400">
                      กรุณาเลือก Product ในฟอร์มเพื่อดูตารางสิทธิ์การใช้งาน
                    </span>
                  }
                />
              ) : (
                <Table
                  dataSource={tableData}
                  columns={columns}
                  loading={loadingData}
                  pagination={{ pageSize: 6 }}
                  scroll={{ x: true }}
                  size="middle"
                />
              )}
            </Card>
          </Col>
        </Row>
      </div>
    </PageTransition>
  );
}
