import { useEffect, useMemo, useState } from 'react';
import {
  Alert,
  Button,
  Card,
  Checkbox,
  Col,
  Descriptions,
  Empty,
  Form,
  Input,
  List,
  Modal,
  Row,
  Select,
  Space,
  Statistic,
  Switch,
  Tag,
  Tree,
  Typography,
  message,
} from 'antd';
import type { DataNode } from 'antd/es/tree';
import {
  FolderIcon,
  PlusIcon,
  Squares2X2Icon,
  TrashIcon,
} from '@heroicons/react/24/outline';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { PageTransition } from '@/components';
import { ROUTES } from '@/constants';
import { productAdminService } from '@/services';
import { useAppStore, useAuthStore } from '@/store';
import type { Project, ProjectFeature } from '@/types';

const { Title, Text } = Typography;
const ROOT_PARENT_VALUE = '__root__';

type FeatureTreeNode = DataNode & {
  key: string;
  children?: FeatureTreeNode[];
};

function includesIgnoreCase(value: string | null | undefined, keyword: string) {
  return (value || '').toLowerCase().includes(keyword.toLowerCase());
}

function buildFeatureChildrenMap(features: ProjectFeature[]) {
  const byParent = new Map<number | null, ProjectFeature[]>();
  features.forEach((feature) => {
    const items = byParent.get(feature.parentId ?? null) || [];
    items.push(feature);
    byParent.set(feature.parentId ?? null, items);
  });
  return byParent;
}

function buildFeatureTree(features: ProjectFeature[], selectedFeatureId: number | null, searchTerm: string): FeatureTreeNode[] {
  const byParent = buildFeatureChildrenMap(features);
  const keyword = searchTerm.trim().toLowerCase();

  const toNode = (feature: ProjectFeature): FeatureTreeNode | null => {
    const children = ((byParent.get(feature.categoryId) || []).map(toNode).filter(Boolean) as FeatureTreeNode[]);
    const matched =
      !keyword ||
      includesIgnoreCase(feature.categoryName, keyword) ||
      includesIgnoreCase(feature.categoryCode, keyword) ||
      includesIgnoreCase(feature.fullPath, keyword) ||
      includesIgnoreCase(feature.categoryType, keyword);

    if (!matched && children.length === 0) {
      return null;
    }

    return {
      key: String(feature.categoryId),
      title: (
        <div className="flex items-center gap-2 flex-wrap">
          <span
            style={{
              fontWeight: selectedFeatureId === feature.categoryId ? 700 : 400,
              textDecoration: feature.isActive ? 'none' : 'line-through',
              opacity: feature.isActive ? 1 : 0.6,
            }}
          >
            {feature.categoryName}
          </span>
          <Tag color="blue">{feature.categoryCode}</Tag>
          <Tag>{`L${feature.level}`}</Tag>
          {feature.categoryType && <Tag color="gold">{feature.categoryType}</Tag>}
          {!feature.isActive && <Tag color="error">Inactive</Tag>}
        </div>
      ),
      children,
    };
  };

  return ((byParent.get(null) || []).map(toNode).filter(Boolean) as FeatureTreeNode[]);
}

function getExpandedKeysForSearch(features: ProjectFeature[], searchTerm: string) {
  const keyword = searchTerm.trim().toLowerCase();
  if (!keyword) {
    return features.map((feature) => String(feature.categoryId));
  }

  const keys = new Set<string>();
  features.forEach((feature) => {
    const matched =
      includesIgnoreCase(feature.categoryName, keyword) ||
      includesIgnoreCase(feature.categoryCode, keyword) ||
      includesIgnoreCase(feature.fullPath, keyword) ||
      includesIgnoreCase(feature.categoryType, keyword);

    if (matched && feature.pathIds) {
      feature.pathIds.split(',').forEach((id) => keys.add(id));
    }
  });
  return Array.from(keys);
}

function getSelectableParentOptions(features: ProjectFeature[], editingFeature: ProjectFeature | null) {
  if (!editingFeature) {
    return features;
  }

  const blockedIds = new Set(
    features
      .filter((feature) => feature.pathIds?.split(',').includes(String(editingFeature.categoryId)))
      .map((feature) => feature.categoryId)
  );

  return features.filter((feature) => !blockedIds.has(feature.categoryId));
}

function getScopeLabel(selectedProductId: number | null, selectedProject: Project | null, selectedFeature: ProjectFeature | null) {
  if (selectedFeature) {
    return `Feature: ${selectedFeature.categoryName}`;
  }
  if (selectedProject) {
    return `Project: ${selectedProject.projectName}`;
  }
  if (selectedProductId) {
    return `Product #${selectedProductId}`;
  }
  return 'No scope selected';
}

export function ProductCatalogPage() {
  const queryClient = useQueryClient();
  const setBreadcrumbs = useAppStore((state) => state.setBreadcrumbs);
  const roles = useAuthStore((state) => state.roles);
  const isPlatformAdmin =
    roles.includes('god') || roles.includes('owner') || roles.includes('superadmin') || roles.includes('super_admin');

  const [selectedProductId, setSelectedProductId] = useState<number | null>(null);
  const [selectedProjectId, setSelectedProjectId] = useState<number | null>(null);
  const [selectedFeatureId, setSelectedFeatureId] = useState<number | null>(null);
  const [featureSearch, setFeatureSearch] = useState('');
  const [expandedKeys, setExpandedKeys] = useState<string[]>([]);
  const [productModalOpen, setProductModalOpen] = useState(false);
  const [projectModalOpen, setProjectModalOpen] = useState(false);
  const [featureModalOpen, setFeatureModalOpen] = useState(false);
  const [isBulkDeleteMode, setIsBulkDeleteMode] = useState(false);
  const [selectedProductIds, setSelectedProductIds] = useState<number[]>([]);

  const [editingProject, setEditingProject] = useState<Project | null>(null);
  const [editingFeature, setEditingFeature] = useState<ProjectFeature | null>(null);

  const [productForm] = Form.useForm();
  const [projectForm] = Form.useForm();
  const [featureForm] = Form.useForm();

  useEffect(() => {
    setBreadcrumbs([{ title: 'Product Catalog', path: ROUTES.PRODUCTS }]);
  }, [setBreadcrumbs]);

  const productsQuery = useQuery({
    queryKey: ['products-admin'],
    queryFn: () => productAdminService.listProducts(),
  });

  const projectsQuery = useQuery({
    queryKey: ['projects-admin', selectedProductId],
    queryFn: () => productAdminService.listProjects(selectedProductId!),
    enabled: !!selectedProductId,
  });

  const featuresQuery = useQuery({
    queryKey: ['features-admin', selectedProductId, selectedProjectId],
    queryFn: () => productAdminService.listFeatures(selectedProductId!, selectedProjectId!),
    enabled: !!selectedProductId && !!selectedProjectId,
  });

  const apiKeysQuery = useQuery({
    queryKey: ['product-api-keys', selectedProductId],
    queryFn: () => productAdminService.listApiKeys(selectedProductId!),
    enabled: !!selectedProductId,
  });

  const products = useMemo(() => productsQuery.data || [], [productsQuery.data]);
  const projects = useMemo(() => projectsQuery.data || [], [projectsQuery.data]);
  const features = useMemo(() => featuresQuery.data || [], [featuresQuery.data]);
  const selectedProduct = useMemo(
    () => products.find((product) => product.productId === selectedProductId) || null,
    [products, selectedProductId]
  );
  const selectedProject = useMemo(
    () => projects.find((project) => project.projectId === selectedProjectId) || null,
    [projects, selectedProjectId]
  );
  const selectedFeature = useMemo(
    () => features.find((feature) => feature.categoryId === selectedFeatureId) || null,
    [features, selectedFeatureId]
  );
  const rootFeatureCount = useMemo(() => features.filter((feature) => !feature.parentId).length, [features]);
  const inactiveFeatureCount = useMemo(() => features.filter((feature) => !feature.isActive).length, [features]);
  const activeApiKeys = useMemo(() => (apiKeysQuery.data || []).filter((item) => item.isActive), [apiKeysQuery.data]);
  const canImportLogs = activeApiKeys.length > 0 && !!selectedProduct?.productEnvironments?.length;

  useEffect(() => {
    if (!selectedProductId && products.length) {
      setSelectedProductId(products[0].productId);
    }
  }, [products, selectedProductId]);

  useEffect(() => {
    if (!selectedProjectId && projects.length) {
      setSelectedProjectId(projects[0].projectId);
    }
  }, [projects, selectedProjectId]);

  useEffect(() => {
    setExpandedKeys(getExpandedKeysForSearch(features, featureSearch));
  }, [features, featureSearch]);

  const featureTree = useMemo(
    () => buildFeatureTree(features, selectedFeatureId, featureSearch),
    [features, selectedFeatureId, featureSearch]
  );
  const selectableParentOptions = useMemo(
    () => getSelectableParentOptions(features, editingFeature),
    [features, editingFeature]
  );

  const createProductMutation = useMutation({
    mutationFn: (values: { productName: string; productCode?: string }) => productAdminService.createProduct(values),
    onSuccess: () => {
      message.success('Product created');
      setProductModalOpen(false);
      productForm.resetFields();
      queryClient.invalidateQueries({ queryKey: ['products-admin'] });
    },
  });

  const deleteProductMutation = useMutation({
    mutationFn: (productId: number) => productAdminService.deleteProduct(productId),
    onSuccess: () => {
      message.success('Product deleted');
      setSelectedProductId(null);
      setSelectedProjectId(null);
      setSelectedFeatureId(null);
      queryClient.invalidateQueries({ queryKey: ['products-admin'] });
    },
    onError: (err: any) => {
      message.error(err?.response?.data?.message || 'Delete failed');
    },
  });

  const bulkDeleteProductsMutation = useMutation({
    mutationFn: (productIds: number[]) => productAdminService.bulkDeleteProducts(productIds),
    onSuccess: () => {
      message.success('Selected products deleted');
      setSelectedProductIds([]);
      setIsBulkDeleteMode(false);
      setSelectedProductId(null);
      setSelectedProjectId(null);
      setSelectedFeatureId(null);
      queryClient.invalidateQueries({ queryKey: ['products-admin'] });
    },
    onError: (err: any) => {
      message.error(err?.response?.data?.message || 'Bulk delete failed');
    },
  });


  const createProjectMutation = useMutation({
    mutationFn: (values: { projectName: string; projectCode?: string }) =>
      productAdminService.createProject(selectedProductId!, values),
    onSuccess: () => {
      message.success('Project created');
      setProjectModalOpen(false);
      projectForm.resetFields();
      queryClient.invalidateQueries({ queryKey: ['projects-admin', selectedProductId] });
    },
  });

  const updateProjectMutation = useMutation({
    mutationFn: (values: { projectName: string; projectCode?: string; isActive?: boolean }) =>
      productAdminService.updateProject(selectedProductId!, editingProject!.projectId, values),
    onSuccess: () => {
      message.success('Project updated');
      setProjectModalOpen(false);
      setEditingProject(null);
      projectForm.resetFields();
      queryClient.invalidateQueries({ queryKey: ['projects-admin', selectedProductId] });
    },
  });

  const deleteProjectMutation = useMutation({
    mutationFn: (projectId: number) => productAdminService.deleteProject(selectedProductId!, projectId),
    onSuccess: () => {
      message.success('Project deleted');
      setSelectedProjectId(null);
      setSelectedFeatureId(null);
      queryClient.invalidateQueries({ queryKey: ['projects-admin', selectedProductId] });
    },
    onError: (err: any) => {
      message.error(err?.response?.data?.message || 'Delete project failed');
    },
  });

  const createFeatureMutation = useMutation({
    mutationFn: (values: { categoryName: string; categoryCode?: string; categoryType?: string; parentId?: string | number }) =>
      productAdminService.createFeature(selectedProductId!, selectedProjectId!, {
        categoryName: values.categoryName,
        categoryCode: values.categoryCode,
        categoryType: values.categoryType || null,
        parentId: values.parentId === ROOT_PARENT_VALUE ? null : Number(values.parentId ?? selectedFeatureId ?? null),
      }),
    onSuccess: () => {
      message.success('Feature created');
      setFeatureModalOpen(false);
      featureForm.resetFields();
      queryClient.invalidateQueries({ queryKey: ['features-admin', selectedProductId, selectedProjectId] });
    },
  });

  const updateFeatureMutation = useMutation({
    mutationFn: (values: { categoryName: string; categoryType?: string; parentId?: string | number; isActive?: boolean }) =>
      productAdminService.updateFeature(selectedProductId!, selectedProjectId!, editingFeature!.categoryId, {
        categoryName: values.categoryName,
        categoryType: values.categoryType || null,
        parentId: values.parentId === ROOT_PARENT_VALUE ? null : values.parentId ? Number(values.parentId) : undefined,
        isActive: values.isActive,
      }),
    onSuccess: () => {
      message.success('Feature updated');
      setFeatureModalOpen(false);
      setEditingFeature(null);
      featureForm.resetFields();
      queryClient.invalidateQueries({ queryKey: ['features-admin', selectedProductId, selectedProjectId] });
    },
    onError: () => {
      message.error('Unable to update feature');
    },
  });

  const deleteFeatureMutation = useMutation({
    mutationFn: (featureId: number) => productAdminService.deleteFeature(selectedProductId!, selectedProjectId!, featureId),
    onSuccess: () => {
      message.success('Feature deleted');
      setSelectedFeatureId(null);
      queryClient.invalidateQueries({ queryKey: ['features-admin', selectedProductId, selectedProjectId] });
    },
    onError: () => {
      message.error('Delete failed. Remove child features first.');
    },
  });

  const handleOpenProjectCreate = () => {
    setEditingProject(null);
    projectForm.resetFields();
    setProjectModalOpen(true);
  };

  const handleOpenProjectEdit = (project: Project) => {
    setEditingProject(project);
    projectForm.setFieldsValue({
      projectName: project.projectName,
      projectCode: project.projectCode,
      isActive: project.isActive,
    });
    setProjectModalOpen(true);
  };

  const handleOpenFeatureCreate = (parentId: number | null) => {
    setEditingFeature(null);
    setSelectedFeatureId(parentId);
    featureForm.resetFields();
    featureForm.setFieldsValue({ parentId: parentId ?? ROOT_PARENT_VALUE });
    setFeatureModalOpen(true);
  };

  const handleOpenFeatureEdit = (feature: ProjectFeature) => {
    setEditingFeature(feature);
    featureForm.setFieldsValue({
      categoryName: feature.categoryName,
      categoryCode: feature.categoryCode,
      categoryType: feature.categoryType,
      parentId: feature.parentId ?? ROOT_PARENT_VALUE,
      isActive: feature.isActive,
    });
    setFeatureModalOpen(true);
  };

  const handleToggleFeatureStatus = (feature: ProjectFeature) => {
    productAdminService
      .updateFeature(selectedProductId!, selectedProjectId!, feature.categoryId, { isActive: !feature.isActive })
      .then(() => {
        message.success(feature.isActive ? 'Feature deactivated' : 'Feature activated');
        queryClient.invalidateQueries({ queryKey: ['features-admin', selectedProductId, selectedProjectId] });
      })
      .catch(() => {
        message.error('Unable to change feature status');
      });
  };

  return (
    <PageTransition>
      <div className="space-y-6">
        <div className="flex items-center justify-between flex-wrap gap-4">
          <div>
            <Title level={4} className="mb-1">
              Product Hierarchy
            </Title>
            <Text type="secondary">
              Use hierarchy as the entry point for logs by product, project, and feature, with import support after API key setup.
            </Text>
          </div>
          {isPlatformAdmin && (
            <Button type="primary" icon={<PlusIcon className="w-4 h-4" />} onClick={() => setProductModalOpen(true)}>
              Create Product
            </Button>
          )}
        </div>

        <Row gutter={[16, 16]}>
          <Col xs={12} md={6}>
            <Card><Statistic title="Products" value={products.length} /></Card>
          </Col>
          <Col xs={12} md={6}>
            <Card><Statistic title="Projects" value={projects.length} /></Card>
          </Col>
          <Col xs={12} md={6}>
            <Card><Statistic title="Root Features" value={rootFeatureCount} /></Card>
          </Col>
          <Col xs={12} md={6}>
            <Card><Statistic title="Active API Keys" value={activeApiKeys.length} /></Card>
          </Col>
        </Row>

        <Row gutter={[16, 16]}>
          <Col xs={24} lg={5}>
            <Card
              title="Products"
              extra={
                isPlatformAdmin ? (
                  isBulkDeleteMode ? (
                    <Space>
                      <Button
                        size="small"
                        danger
                        disabled={selectedProductIds.length === 0}
                        loading={bulkDeleteProductsMutation.isPending}
                        onClick={() =>
                          Modal.confirm({
                            title: `Delete ${selectedProductIds.length} products?`,
                            content: 'Are you sure you want to delete the selected products and all their contents? This action cannot be undone.',
                            okButtonProps: { danger: true },
                            onOk: () => bulkDeleteProductsMutation.mutate(selectedProductIds),
                          })
                        }
                      >
                        Delete ({selectedProductIds.length})
                      </Button>
                      <Button
                        size="small"
                        onClick={() => {
                          setIsBulkDeleteMode(false);
                          setSelectedProductIds([]);
                        }}
                      >
                        Cancel
                      </Button>
                    </Space>
                  ) : (
                    <Space>
                      <Tag>{products.length}</Tag>
                      <Button
                        size="small"
                        onClick={() => {
                          setIsBulkDeleteMode(true);
                          setSelectedProductIds([]);
                        }}
                      >
                        Select
                      </Button>
                    </Space>
                  )
                ) : (
                  <Tag>{products.length}</Tag>
                )
              }
            >
              <List
                loading={productsQuery.isLoading}
                dataSource={products}
                locale={{ emptyText: <Empty description="No products yet" /> }}
                renderItem={(item) => (
                  <List.Item
                    className={`cursor-pointer rounded-lg px-3 ${selectedProductId === item.productId && !isBulkDeleteMode ? 'bg-slate-50' : ''}`}
                    onClick={() => {
                      if (isBulkDeleteMode) {
                        if (selectedProductIds.includes(item.productId)) {
                          setSelectedProductIds(selectedProductIds.filter((id) => id !== item.productId));
                        } else {
                          setSelectedProductIds([...selectedProductIds, item.productId]);
                        }
                      } else {
                        setSelectedProductId(item.productId);
                        setSelectedProjectId(null);
                        setSelectedFeatureId(null);
                        setFeatureSearch('');
                      }
                    }}
                    actions={
                      isPlatformAdmin && !isBulkDeleteMode
                        ? [
                            <Button
                              key={`delete-product-${item.productId}`}
                              size="small"
                              type="link"
                              danger
                              icon={<TrashIcon className="w-3.5 h-3.5" />}
                              onClick={(event) => {
                                event.stopPropagation();
                                Modal.confirm({
                                  title: `Delete ${item.productName}?`,
                                  content: 'Are you sure you want to delete this product and all its contents? This action cannot be undone.',
                                  okButtonProps: { danger: true },
                                  onOk: () => deleteProductMutation.mutate(item.productId),
                                });
                              }}
                            />,
                          ]
                        : undefined
                    }
                  >
                    <div className="w-full flex items-center justify-between gap-3">
                      <Space>
                        {isBulkDeleteMode ? (
                          <Checkbox
                            checked={selectedProductIds.includes(item.productId)}
                            onChange={() => {}} // toggled by List.Item onClick
                          />
                        ) : (
                          <FolderIcon className="w-4 h-4 text-slate-500" />
                        )}
                        <Text strong>{item.productName}</Text>
                      </Space>
                      <Tag color={item.isActive ? 'green' : 'default'}>{item.productCode}</Tag>
                    </div>
                  </List.Item>
                )}
              />
            </Card>
          </Col>

          <Col xs={24} lg={5}>
            <Card
              title="Projects"
              extra={
                <Button size="small" type="link" disabled={!selectedProductId} onClick={handleOpenProjectCreate}>
                  Add Project
                </Button>
              }
            >
              <List
                loading={projectsQuery.isLoading}
                dataSource={projects}
                locale={{ emptyText: <Empty description="Select a product first" /> }}
                renderItem={(item) => (
                  <List.Item
                    className={`cursor-pointer rounded-lg px-3 ${selectedProjectId === item.projectId ? 'bg-slate-50' : ''}`}
                    onClick={() => {
                      setSelectedProjectId(item.projectId);
                      setSelectedFeatureId(null);
                    }}
                    actions={[
                      <Button
                        key={`edit-project-${item.projectId}`}
                        size="small"
                        type="link"
                        onClick={(event) => {
                          event.stopPropagation();
                          handleOpenProjectEdit(item);
                        }}
                      >
                        Edit
                      </Button>,
                      isPlatformAdmin && (
                        <Button
                          key={`delete-project-${item.projectId}`}
                          size="small"
                          type="link"
                          danger
                          icon={<TrashIcon className="w-3.5 h-3.5" />}
                          onClick={(event) => {
                            event.stopPropagation();
                            Modal.confirm({
                              title: `Delete Project ${item.projectName}?`,
                              content: 'Are you sure you want to delete this project and all its contents? This action cannot be undone.',
                              okButtonProps: { danger: true },
                              onOk: () => deleteProjectMutation.mutate(item.projectId),
                            });
                          }}
                        />
                      )
                    ]}
                  >
                    <div className="w-full flex items-center justify-between gap-3">
                      <Space>
                        <Text strong>{item.projectName}</Text>
                        {!item.isActive && <Tag color="error">Inactive</Tag>}
                      </Space>
                      <Tag color="purple">{item.projectCode}</Tag>
                    </div>
                  </List.Item>
                )}
              />
            </Card>
          </Col>

          <Col xs={24} lg={6}>
            <Card
              title={selectedProject ? `Features: ${selectedProject.projectName}` : 'Features'}
              extra={
                <Space wrap>
                  <Button size="small" icon={<PlusIcon className="w-4 h-4" />} disabled={!selectedProjectId} onClick={() => handleOpenFeatureCreate(null)}>
                    Root
                  </Button>
                  <Button size="small" icon={<Squares2X2Icon className="w-4 h-4" />} disabled={!selectedFeatureId} onClick={() => handleOpenFeatureCreate(selectedFeatureId)}>
                    Child
                  </Button>
                </Space>
              }
            >
              <div className="space-y-3">
                <Input.Search
                  allowClear
                  placeholder="Search feature by name, code, path"
                  value={featureSearch}
                  onChange={(event) => setFeatureSearch(event.target.value)}
                />
                {selectedProjectId ? (
                  featureTree.length ? (
                    <Tree
                      treeData={featureTree}
                      selectedKeys={selectedFeatureId ? [String(selectedFeatureId)] : []}
                      expandedKeys={expandedKeys}
                      onExpand={(keys) => setExpandedKeys(keys.map(String))}
                      onSelect={(keys) => setSelectedFeatureId(keys.length ? Number(keys[0]) : null)}
                    />
                  ) : (
                    <Empty description="No matching features" />
                  )
                ) : (
                  <Empty description="Select a project first" />
                )}
              </div>
            </Card>
          </Col>

          <Col xs={24} lg={8}>
            <div className="space-y-4">
              <Card title="Selected Scope" extra={<Tag color="geekblue">{getScopeLabel(selectedProductId, selectedProject, selectedFeature)}</Tag>}>
                {selectedProductId ? (
                  <div className="space-y-4">
                    <Descriptions size="small" column={1}>
                      <Descriptions.Item label="Product">{selectedProduct?.productName || '-'}</Descriptions.Item>
                      <Descriptions.Item label="Project">{selectedProject?.projectName || 'All in product'}</Descriptions.Item>
                      <Descriptions.Item label="Feature">{selectedFeature?.categoryName || 'All in scope'}</Descriptions.Item>
                      <Descriptions.Item label="Feature Path">{selectedFeature?.fullPath || '-'}</Descriptions.Item>
                      <Descriptions.Item label="Inactive Features">{inactiveFeatureCount}</Descriptions.Item>
                    </Descriptions>

                    {selectedFeature && (
                      <Space direction="vertical" className="w-full">
                        <Button block onClick={() => handleOpenFeatureEdit(selectedFeature)}>Edit Feature</Button>
                        <Button block onClick={() => handleToggleFeatureStatus(selectedFeature)}>
                          {selectedFeature.isActive ? 'Deactivate Feature' : 'Activate Feature'}
                        </Button>
                        <Button
                          block
                          danger
                          icon={<TrashIcon className="w-4 h-4" />}
                          onClick={() =>
                            Modal.confirm({
                              title: `Delete ${selectedFeature.categoryName}?`,
                              content: 'Delete works only when there are no child nodes.',
                              okButtonProps: { danger: true },
                              onOk: () => deleteFeatureMutation.mutate(selectedFeature.categoryId),
                            })
                          }
                        >
                          Delete Feature
                        </Button>
                      </Space>
                    )}
                  </div>
                ) : (
                  <Empty description="Select a product or node" />
                )}
              </Card>

              <Card title="API Key Connection">
                {selectedProductId ? (
                  canImportLogs ? (
                    <Alert
                      type="success"
                      showIcon
                      message="API key connected"
                      description={`Found ${activeApiKeys.length} active key(s). You can import logs into the selected scope.`}
                    />
                  ) : (
                    <Alert
                      type="warning"
                      showIcon
                      message="API key not ready"
                      description="Create or activate a product API key first, then assign an environment to import logs."
                    />
                  )
                ) : (
                  <Empty description="Select a product first" />
                )}
              </Card>
            </div>
          </Col>
        </Row>

        <Modal
          title="Create Product"
          open={productModalOpen}
          onCancel={() => setProductModalOpen(false)}
          onOk={() => productForm.submit()}
          confirmLoading={createProductMutation.isPending}
          width={550}
        >
          <Form
            form={productForm}
            layout="vertical"
            onFinish={(values) => createProductMutation.mutate(values)}
            initialValues={{
              environments: [
                { environmentCode: 'DEV', environmentName: 'Development' },
                { environmentCode: 'PROD', environmentName: 'Production' },
              ],
            }}
          >
            <Form.Item name="productName" label="Product Name" rules={[{ required: true }]}>
              <Input />
            </Form.Item>
            <Form.Item name="productCode" label="Product Code">
              <Input placeholder="Optional" />
            </Form.Item>
            <Form.List name="environments">
              {(fields, { add, remove }) => (
                <>
                  <div style={{ marginBottom: 8, fontWeight: 'bold' }}>Product Environments</div>
                  {fields.map(({ key, name, ...restField }) => (
                    <Space key={key} style={{ display: 'flex', marginBottom: 8 }} align="baseline">
                      <Form.Item {...restField} name={[name, 'environmentCode']} rules={[{ required: true }]} noStyle>
                        <Input placeholder="DEV, PROD" style={{ width: 180 }} />
                      </Form.Item>
                      <Form.Item {...restField} name={[name, 'environmentName']} rules={[{ required: true }]} noStyle>
                        <Input placeholder="Development" style={{ width: 220 }} />
                      </Form.Item>
                      {fields.length > 1 && (
                        <Button type="link" danger onClick={() => remove(name)} style={{ padding: 0 }}>
                          Remove
                        </Button>
                      )}
                    </Space>
                  ))}
                  <Form.Item>
                    <Button type="dashed" onClick={() => add()} block>
                      + Add Environment
                    </Button>
                  </Form.Item>
                </>
              )}
            </Form.List>
          </Form>
        </Modal>

        <Modal
          title={editingProject ? 'Edit Project' : 'Create Project'}
          open={projectModalOpen}
          onCancel={() => setProjectModalOpen(false)}
          onOk={() => projectForm.submit()}
          confirmLoading={editingProject ? updateProjectMutation.isPending : createProjectMutation.isPending}
        >
          <Form
            form={projectForm}
            layout="vertical"
            onFinish={(values) => (editingProject ? updateProjectMutation.mutate(values) : createProjectMutation.mutate(values))}
          >
            <Form.Item name="projectName" label="Project Name" rules={[{ required: true }]}>
              <Input />
            </Form.Item>
            {!editingProject && (
              <Form.Item name="projectCode" label="Project Code">
                <Input placeholder="Optional" />
              </Form.Item>
            )}
            {editingProject && (
              <Form.Item name="isActive" label="Status" valuePropName="checked">
                <Switch checkedChildren="Active" unCheckedChildren="Inactive" />
              </Form.Item>
            )}
          </Form>
        </Modal>

        <Modal
          title={editingFeature ? 'Edit Feature' : selectedFeatureId ? 'Create Child Feature' : 'Create Root Feature'}
          open={featureModalOpen}
          onCancel={() => setFeatureModalOpen(false)}
          onOk={() => featureForm.submit()}
          confirmLoading={editingFeature ? updateFeatureMutation.isPending : createFeatureMutation.isPending}
        >
          <Form
            form={featureForm}
            layout="vertical"
            onFinish={(values) => (editingFeature ? updateFeatureMutation.mutate(values) : createFeatureMutation.mutate(values))}
          >
            <Form.Item name="categoryName" label="Feature Name" rules={[{ required: true }]}>
              <Input />
            </Form.Item>
            {!editingFeature && (
              <Form.Item name="categoryCode" label="Feature Code">
                <Input placeholder="Optional" />
              </Form.Item>
            )}
            <Form.Item name="categoryType" label="Feature Type">
              <Input placeholder="Optional" />
            </Form.Item>
            <Form.Item name="parentId" label="Parent Feature" initialValue={ROOT_PARENT_VALUE}>
              <Select
                options={[
                  { value: ROOT_PARENT_VALUE, label: 'Root Level' },
                  ...selectableParentOptions.map((feature) => ({
                    value: feature.categoryId,
                    label: `${feature.fullPath || feature.categoryName} (${feature.categoryCode})`,
                  })),
                ]}
              />
            </Form.Item>
            {editingFeature && (
              <Form.Item name="isActive" label="Status" valuePropName="checked">
                <Switch checkedChildren="Active" unCheckedChildren="Inactive" />
              </Form.Item>
            )}
          </Form>
        </Modal>
      </div>
    </PageTransition>
  );
}
