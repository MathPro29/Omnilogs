import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Space, Typography, Modal, Button, Select, Input, Table, Tag, Form, Alert, Row, Col, Badge, message, Spin, Switch } from 'antd';
import { useQuery, useMutation } from '@tanstack/react-query';
import { SearchOutlined, EyeOutlined, SendOutlined, ReloadOutlined, LineChartOutlined } from '@ant-design/icons';
import { productAdminService } from '@/services';
import { useAppStore, useAutoLogStore } from '@/store';
import { ROUTES } from '@/constants';
import type { MainLog, ProjectFeature } from '@/types';

const { Text, Title, Paragraph } = Typography;

function getScopedAutoSendFeatures(features: ProjectFeature[], selectedCategoryId?: number | null) {
  const activeFeatures = features.filter((feature) => feature.isActive);
  if (!selectedCategoryId) {
    return activeFeatures;
  }

  return activeFeatures.filter((feature) => {
    const pathIds = feature.pathIds?.split(',').map((value) => value.trim()) || [];
    return feature.categoryId === selectedCategoryId || pathIds.includes(String(selectedCategoryId));
  });
}

export function DashboardPage() {
  const setBreadcrumbs = useAppStore((state) => state.setBreadcrumbs);
  const navigate = useNavigate();

  // States for Logs Monitor Filters
  const [selectedProductId, setSelectedProductId] = useState<number | null>(null);
  const [selectedEnvironmentId, setSelectedEnvironmentId] = useState<number | null>(null);
  const [selectedProjectId, setSelectedProjectId] = useState<number | null>(null);
  const [selectedCategoryId, setSelectedCategoryId] = useState<number | null>(null);
  const [logLevel, setLogLevel] = useState<string | undefined>(undefined);
  const [logKeyword, setLogKeyword] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [selectedLog, setSelectedLog] = useState<MainLog | null>(null);
  const [isAutoRefresh, setIsAutoRefresh] = useState(true);

  // States for Log Generator Form
  const [generatorForm] = Form.useForm();
  const generatorProductId = Form.useWatch('productId', generatorForm);
  const generatorProjectId = Form.useWatch('projectId', generatorForm);

  // Elastic Online Status
  // refresh every 5 seconds
  const { data: elasticStatus = null, isLoading: isLoadingElasticStatus } = useQuery({
    queryKey: ['elastic-online'],
    queryFn: () => productAdminService.checkElasticOnline(),
    refetchInterval: 15000,
    refetchIntervalInBackground: true,
  });

  // สถานะ Auto Send อยู่ใน Global Store จึงไม่หายเมื่อ Dashboard ถูก Unmount
  const isAutoSending = useAutoLogStore((state) => state.isRunning);
  const startAutoSending = useAutoLogStore((state) => state.start);
  const stopAutoSending = useAutoLogStore((state) => state.stop);

  const handleToggleAutoSending = async () => {
    if (isAutoSending) {
      stopAutoSending();
      message.info('หยุดส่ง Log อัตโนมัติแล้ว');
      return;
    }

    try {
      const values = await generatorForm.validateFields([
        'productId',
        'environmentId',
        'eventType',
      ]);
      const allValues = generatorForm.getFieldsValue();
      const scopedFeatures = getScopedAutoSendFeatures(generatorFeatures, allValues.categoryId);
      
      let featureFullPath = null;
      let featurePathIds = null;
      if (allValues.categoryId) {
        const feature = generatorFeatures.find(f => f.categoryId === allValues.categoryId);
        if (feature) {
          featureFullPath = feature.fullPath || null;
          featurePathIds = feature.pathIds || null;
        }
      }

      startAutoSending({
        productId: values.productId,
        environmentId: values.environmentId,
        projectId: allValues.projectId || undefined,
        categoryId: allValues.categoryId || undefined,
        featureFullPath,
        featurePathIds,
        eventType: values.eventType,
        features: scopedFeatures,
      });
      message.success('เริ่มส่ง Log อัตโนมัติแล้ว สามารถเปลี่ยนหน้าได้');
    } catch {
      message.warning('กรุณากรอกข้อมูลที่จำเป็นก่อนเริ่ม Auto Send');
    }
  };

  useEffect(() => {
    setBreadcrumbs([{ title: 'แดชบอร์ด', path: ROUTES.DASHBOARD }]);
  }, [setBreadcrumbs]);

  // ----------------------------------------------------
  // 1. Queries for Products, Environments & Projects
  // ----------------------------------------------------
  
  // Products
  const { data: products = [], isLoading: isLoadingProducts } = useQuery({
    queryKey: ['products-admin'],
    queryFn: () => productAdminService.listProducts(),
  });

  // Selected Product for Monitor Filter
  const monitorProduct = products.find((p) => p.productId === selectedProductId);
  const monitorEnvironments = monitorProduct?.productEnvironments || [];

  // Projects for Monitor Filter
  const { data: monitorProjects = [], isLoading: isLoadingMonitorProjects } = useQuery({
    queryKey: ['projects-admin-monitor', selectedProductId],
    queryFn: () => productAdminService.listProjects(selectedProductId!),
    enabled: !!selectedProductId,
  });

  // Selected Product for Generator Form
  const generatorProduct = products.find((p) => p.productId === generatorProductId);
  const generatorEnvironments = generatorProduct?.productEnvironments || [];

  // Projects for Generator Form
  const { data: generatorProjects = [], isLoading: isLoadingGeneratorProjects } = useQuery({
    queryKey: ['projects-admin-generator', generatorProductId],
    queryFn: () => productAdminService.listProjects(generatorProductId!),
    enabled: !!generatorProductId,
  });

  // Features for Generator Form
  const { data: generatorFeatures = [], isLoading: isLoadingGeneratorFeatures } = useQuery({
    queryKey: ['features-admin-generator', generatorProductId, generatorProjectId],
    queryFn: () => productAdminService.listFeatures(generatorProductId!, generatorProjectId!),
    enabled: !!generatorProductId && !!generatorProjectId,
  });

  // Features for Monitor Filter
  const { data: monitorFeatures = [], isLoading: isLoadingMonitorFeatures } = useQuery({
    queryKey: ['features-admin-monitor', selectedProductId, selectedProjectId],
    queryFn: () => productAdminService.listFeatures(selectedProductId!, selectedProjectId!),
    enabled: !!selectedProductId && !!selectedProjectId,
  });

  // Sync monitor environment options when product changes
  useEffect(() => {
    if (monitorEnvironments.length > 0) {
      setSelectedEnvironmentId(monitorEnvironments[0].environmentId);
    } else {
      setSelectedEnvironmentId(null);
    }
    setSelectedProjectId(null);
    setSelectedCategoryId(null);
    setCurrentPage(1);
  }, [selectedProductId, monitorEnvironments]);

  // Sync monitor category when project changes
  useEffect(() => {
    setSelectedCategoryId(null);
  }, [selectedProjectId]);

  // Automatically select the first product when products list is loaded
  useEffect(() => {
    if (products.length > 0 && selectedProductId === null) {
      setSelectedProductId(products[0].productId);
    }
  }, [products, selectedProductId]);

  // Synchronize generator form fields when monitor selections change
  useEffect(() => {
    if (selectedProductId) {
      generatorForm.setFieldValue('productId', selectedProductId);
    }
  }, [selectedProductId, generatorForm]);

  useEffect(() => {
    if (selectedEnvironmentId) {
      generatorForm.setFieldValue('environmentId', selectedEnvironmentId);
    } else {
      generatorForm.setFieldValue('environmentId', null);
    }
  }, [selectedEnvironmentId, generatorForm]);

  useEffect(() => {
    if (selectedProjectId) {
      generatorForm.setFieldValue('projectId', selectedProjectId);
    } else {
      generatorForm.setFieldValue('projectId', null);
    }
  }, [selectedProjectId, generatorForm]);

  useEffect(() => {
    if (selectedCategoryId) {
      generatorForm.setFieldValue('categoryId', selectedCategoryId);
    } else {
      generatorForm.setFieldValue('categoryId', null);
    }
  }, [selectedCategoryId, generatorForm]);

  // ----------------------------------------------------
  // 2. Fetching & Generating Logs
  // ----------------------------------------------------

  // Fetch Logs Query
  const { data: logsData, isLoading: isLoadingLogs, refetch: refetchLogs } = useQuery({
    queryKey: ['dashboard-logs', selectedProductId, selectedEnvironmentId, selectedProjectId, selectedCategoryId, logLevel, logKeyword, currentPage, pageSize],
    queryFn: () =>
      productAdminService.searchLogs({
        productId: selectedProductId!,
        environmentId: selectedEnvironmentId || undefined,
        projectId: selectedProjectId || undefined,
        categoryId: selectedCategoryId || undefined,
        level: logLevel || undefined,
        keyword: logKeyword || undefined,
        page: currentPage,
        perPage: pageSize,
      }),
    enabled: !!selectedProductId,
    refetchInterval: isAutoRefresh ? 5000 : false,
    refetchIntervalInBackground: true,
  });

  const logs = logsData?.data || [];
  const totalLogs = logsData?.total || 0;

  // Log Generator Mutation
  const generateLogMutation = useMutation({
    mutationFn: (values: {
      productId: number;
      environmentId: number;
      projectId?: number;
      categoryId?: number;
      logLevel: string;
      message: string;
      eventType: string;
    }) => {
      let featureFullPath = null;
      let featurePathIds = null;
      if (values.categoryId) {
        const feature = generatorFeatures.find(f => f.categoryId === values.categoryId);
        if (feature) {
          featureFullPath = feature.fullPath || null;
          featurePathIds = feature.pathIds || null;
        }
      }
      return productAdminService.importLogs({
        productId: values.productId,
        environmentId: values.environmentId,
        projectId: values.projectId || undefined,
        categoryId: values.categoryId || undefined,
        featureFullPath,
        featurePathIds,
        logLevel: values.logLevel,
        message: values.message,
        eventType: values.eventType,
      });
    },
    onSuccess: (_, variables) => {
      message.success('สร้าง Logs และดึงเข้า Elasticsearch สำเร็จ!');
      // If the generated product matches the currently viewed product, reload logs list
      if (variables.productId === selectedProductId) {
        refetchLogs();
      } else {
        // Auto switch monitor to the generated product to show it
        setSelectedProductId(variables.productId);
        setSelectedEnvironmentId(variables.environmentId);
        if (variables.projectId) setSelectedProjectId(variables.projectId);
      }
    },
    onError: (err: any) => {
      message.error(`เกิดข้อผิดพลาดในการส่งข้อมูล: ${err?.message || err}`);
    },
  });

  const handleSearch = () => {
    setCurrentPage(1);
    refetchLogs();
  };

  const getLevelColor = (level: string | null | undefined) => {
    if (!level) return 'default';
    switch (level.toUpperCase()) {
      case 'ERROR':
      case 'FATAL':
        return 'error';
      case 'WARN':
      case 'WARNING':
        return 'warning';
      case 'INFO':
        return 'success';
      case 'DEBUG':
        return 'processing';
      default:
        return 'default';
    }
  };

  return (
    <div style={{ padding: '8px' }}>
      <Space direction="vertical" style={{ width: '100%' }} size="large">
        
        {/* Welcome & Dashboard Status Header */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '12px' }}>
          <div>
            <Title level={3} style={{ margin: 0, fontWeight: 700 }}>
              Elasticsearch Logs Monitoring Hub & Generator
            </Title>
            <Text type="secondary">
              จำลองส่งข้อมูล Logs และติดตามผลการจัดเก็บของระบบผ่าน Elasticsearch แบบเรียลไทม์
            </Text>
          </div>
          {/* [STATUS : Elastic Online] */}
          <Space>
            <Button
              type="primary"
              icon={<LineChartOutlined />}
              onClick={() => navigate(ROUTES.LOGS_EXPLORER)}
              style={{ borderRadius: '8px' }}
            >
              เปิด Logs Explorer
            </Button>
            {isLoadingElasticStatus ? (
              <Spin size="small" />
            ) : elasticStatus?.data.status === 'online' ? (
              <Badge status="success" text="ระบบทำงานปกติ (Elastic Online)" />
            ) : (
              <Badge status="error" text="ระบบทำงานผิดปกติ (Elastic Offline)" />
            )}
          </Space>
        </div>

        <Row gutter={[16, 16]}>
          {/* Column 1: Log Generator Card */}
          <Col xs={24} lg={8}>
            <Card
              title={
                <Space>
                  <SendOutlined style={{ color: '#1890ff' }} />
                  <span>เครื่องมือจำลองสร้าง Logs (Log Generator)</span>
                </Space>
              }
              bordered={false}
              className="shadow-sm"
              style={{ borderRadius: '12px', height: '100%' }}
            >
              <Paragraph type="secondary" style={{ fontSize: '13px' }}>
                กำหนดสเปกและเนื้อหาของ Log จากนั้นระบบจะยิงเข้าสู่ Postgres Queue และเรียก Worker ดึงลง Elasticsearch ทันที
              </Paragraph>
              
              <Form
                form={generatorForm}
                layout="vertical"
                initialValues={{
                  logLevel: 'INFO',
                  eventType: 'manual-trigger',
                  message: 'User completed action successfully',
                }}
                onFinish={(values) => generateLogMutation.mutate(values)}
                onValuesChange={(changedValues) => {
                  if ('productId' in changedValues) {
                    setSelectedProductId(changedValues.productId);
                  }
                  if ('environmentId' in changedValues) {
                    setSelectedEnvironmentId(changedValues.environmentId);
                  }
                  if ('projectId' in changedValues) {
                    setSelectedProjectId(changedValues.projectId);
                  }
                  if ('categoryId' in changedValues) {
                    setSelectedCategoryId(changedValues.categoryId);
                  }
                }}
              >
                <Form.Item
                  label="1. ผลิตภัณฑ์ปลายทาง (Product)"
                  name="productId"
                  rules={[{ required: true, message: 'กรุณาเลือกผลิตภัณฑ์' }]}
                >
                  <Select
                    placeholder="เลือกผลิตภัณฑ์..."
                    loading={isLoadingProducts}
                    onChange={() => {
                      generatorForm.setFieldValue('environmentId', null);
                      generatorForm.setFieldValue('projectId', null);
                    }}
                    options={products.map((p) => ({
                      value: p.productId,
                      label: `${p.productName} (${p.productCode})`,
                    }))}
                  />
                </Form.Item>

                <Form.Item
                  label="2. สภาพแวดล้อม (Environment)"
                  name="environmentId"
                  rules={[{ required: true, message: 'กรุณาเลือกสภาพแวดล้อม' }]}
                >
                  <Select
                    placeholder={generatorProductId ? "เลือกสภาพแวดล้อม..." : "เลือกผลิตภัณฑ์ก่อน..."}
                    disabled={!generatorProductId || generatorEnvironments.length === 0}
                    options={generatorEnvironments.map((e) => ({
                      value: e.environmentId,
                      label: `${e.environmentName} (${e.environmentCode})`,
                    }))}
                  />
                </Form.Item>

                <Form.Item label="3. โครงการย่อย (Project - Optional)" name="projectId">
                  <Select
                    allowClear
                    placeholder={generatorProductId ? "เลือกโครงการ..." : "เลือกผลิตภัณฑ์ก่อน..."}
                    disabled={!generatorProductId || generatorProjects.length === 0}
                    loading={isLoadingGeneratorProjects}
                    options={generatorProjects.map((p) => ({
                      value: p.projectId,
                      label: `${p.projectName} (${p.projectCode})`,
                    }))}
                  />
                </Form.Item>

                <Form.Item label="4. หมวดหมู่ (Category / Feature - Optional)" name="categoryId">
                  <Select
                    allowClear
                    placeholder={generatorProjectId ? "เลือกหมวดหมู่..." : "เลือกโครงการก่อน..."}
                    disabled={!generatorProjectId || generatorFeatures.length === 0}
                    loading={isLoadingGeneratorFeatures}
                    options={generatorFeatures.map((f) => ({
                      value: f.categoryId,
                      label: `${f.categoryName} (${f.fullPath})`,
                    }))}
                  />
                </Form.Item>

                <Form.Item
                  label="5. ระดับความสำคัญ (Log Level)"
                  name="logLevel"
                  rules={[{ required: true }]}
                >
                  <Select
                    options={['DEBUG', 'INFO', 'WARN', 'ERROR'].map((val) => ({
                      value: val,
                      label: val,
                    }))}
                  />
                </Form.Item>

                <Form.Item
                  label="6. ประเภทงาน (Event Type / Log Type)"
                  name="eventType"
                  rules={[{ required: true, message: 'กรุณากรอกประเภทงาน' }]}
                >
                  <Input placeholder="เช่น user-login, checkout, api-call" />
                </Form.Item>

                <Form.Item
                  label="7. ข้อความ Log (Message)"
                  name="message"
                  rules={[{ required: true, message: 'กรุณากรอกข้อความของ Log' }]}
                >
                  <Input.TextArea rows={3} placeholder="กรอกข้อความแจ้งเตือนหรือข้อมูลที่ต้องการเก็บบันทึก..." />
                </Form.Item>

                <Form.Item style={{ marginBottom: '12px' }}>
                  <Button
                    danger={isAutoSending}
                    type={isAutoSending ? 'primary' : 'default'}
                    onClick={handleToggleAutoSending}
                    block
                  >
                    {isAutoSending ? 'หยุดส่ง Log อัตโนมัติ (Stop Auto Send)' : 'ส่ง Log อัตโนมัติทุก 3 วินาที (Start Auto Send)'}
                  </Button>
                </Form.Item>

                <Form.Item style={{ marginBottom: 0 }}>
                  <Button
                    type="primary"
                    htmlType="submit"
                    block
                    icon={<SendOutlined />}
                    loading={generateLogMutation.isPending}
                  >
                    ส่งข้อความ Log เข้าสู่ระบบ
                  </Button>
                </Form.Item>
              </Form>

              <div style={{ marginTop: '24px', borderTop: '1px solid #f0f0f0', paddingTop: '16px' }}>
                <Title level={5} style={{ marginBottom: '16px' }}>
                  Hierarchy Validation Test Cases (จำลองเงื่อนไข)
                </Title>
                <Alert 
                  type="info" 
                  showIcon 
                  message="กดปุ่มด้านล่างเพื่อทดสอบส่งค่า ID ที่ผิดพลาดเข้าสู่ระบบ" 
                  style={{ marginBottom: '16px' }} 
                />
                
                <Space direction="vertical" style={{ width: '100%' }}>
                  <Button 
                    block 
                    onClick={() => {
                      const vals = generatorForm.getFieldsValue();
                      if(!vals.productId) { message.warning('เลือก Product หลักก่อน'); return; }
                      generateLogMutation.mutate({ ...vals, environmentId: 999999, message: 'Test: Invalid Environment' });
                    }}
                  >
                    Case 2: Invalid Environment (ผิด Product)
                  </Button>
                  <Button 
                    block 
                    onClick={() => {
                      const vals = generatorForm.getFieldsValue();
                      if(!vals.productId || !vals.environmentId) { message.warning('เลือก Product & Environment ก่อน'); return; }
                      generateLogMutation.mutate({ ...vals, projectId: undefined, categoryId: 666666, message: 'Test: Orphan Category' });
                    }}
                  >
                    Case 3: Orphan Category (มี Category แต่ไม่มี Project)
                  </Button>
                  <Button 
                    block 
                    onClick={() => {
                      const vals = generatorForm.getFieldsValue();
                      if(!vals.productId || !vals.environmentId) { message.warning('เลือก Product & Environment ก่อน'); return; }
                      generateLogMutation.mutate({ ...vals, projectId: 555555, message: 'Test: Project Mismatch' });
                    }}
                  >
                    Case 4: Project Mismatch (Project คนละ Product หรือถูกปิด)
                  </Button>
                  <Button 
                    block 
                    onClick={() => {
                      const vals = generatorForm.getFieldsValue();
                      if(!vals.productId || !vals.environmentId || !vals.projectId) { message.warning('เลือก Product, Environment, Project ก่อน'); return; }
                      generateLogMutation.mutate({ ...vals, categoryId: 111111, message: 'Test: Category Mismatch' });
                    }}
                  >
                    Case 5: Category Mismatch (Category คนละ Project)
                  </Button>
                </Space>
              </div>
            </Card>
          </Col>

          {/* Column 2: Logs Monitor / Viewer */}
          <Col xs={24} lg={16}>
            <Space direction="vertical" style={{ width: '100%' }} size="middle">
              
              {/* Filters Card */}
              <Card 
                bordered={false} 
                className="shadow-sm" 
                style={{ borderRadius: '12px' }}
                title={
                  <span style={{ fontWeight: 600 }}>ตัวกรองการสืบค้น (Logs Search Filter)</span>
                }
                extra={
                  selectedProductId && (
                    <Space>
                      <Switch
                        checked={isAutoRefresh}
                        onChange={(checked) => setIsAutoRefresh(checked)}
                        checkedChildren="Auto"
                        unCheckedChildren="Manual"
                      />
                      <Button
                        type="text"
                        icon={<ReloadOutlined spin={isLoadingLogs} />}
                        onClick={() => refetchLogs()}
                        style={{ display: 'flex', alignItems: 'center' }}
                      >
                        โหลดข้อมูลใหม่
                      </Button>
                    </Space>
                  )
                }
              >
                <Form layout="vertical">
                  <Row gutter={[12, 12]} align="bottom">
                    <Col xs={24} sm={12} md={8}>
                      <Form.Item label="ผลิตภัณฑ์ (Product)" style={{ marginBottom: 0 }}>
                        <Select
                          showSearch
                          placeholder="กรุณาเลือกผลิตภัณฑ์..."
                          value={selectedProductId}
                          onChange={(val) => setSelectedProductId(val)}
                          loading={isLoadingProducts}
                          filterOption={(input, option) =>
                            (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
                          }
                          options={products.map((p) => ({
                            value: p.productId,
                            label: `${p.productName} (${p.productCode})`,
                          }))}
                          style={{ width: '100%' }}
                        />
                      </Form.Item>
                    </Col>
                    
                    <Col xs={24} sm={12} md={8}>
                      <Form.Item label="สภาพแวดล้อม (Environment)" style={{ marginBottom: 0 }}>
                        <Select
                          placeholder={selectedProductId ? "เลือกสภาพแวดล้อม..." : "เลือกผลิตภัณฑ์ก่อน..."}
                          value={selectedEnvironmentId}
                          onChange={(val) => setSelectedEnvironmentId(val)}
                          disabled={!selectedProductId || monitorEnvironments.length === 0}
                          options={monitorEnvironments.map((e) => ({
                            value: e.environmentId,
                            label: `${e.environmentName} (${e.environmentCode})`,
                          }))}
                          style={{ width: '100%' }}
                        />
                      </Form.Item>
                    </Col>

                    <Col xs={24} sm={12} md={8}>
                      <Form.Item label="โครงการ (Project)" style={{ marginBottom: 0 }}>
                        <Select
                          allowClear
                          placeholder={selectedProductId ? "ทั้งหมด..." : "เลือกผลิตภัณฑ์ก่อน..."}
                          value={selectedProjectId}
                          onChange={(val) => setSelectedProjectId(val)}
                          disabled={!selectedProductId || monitorProjects.length === 0}
                          loading={isLoadingMonitorProjects}
                          options={monitorProjects.map((p) => ({
                            value: p.projectId,
                            label: `${p.projectName} (${p.projectCode})`,
                          }))}
                          style={{ width: '100%' }}
                        />
                      </Form.Item>
                    </Col>

                    <Col xs={24} sm={12} md={8}>
                      <Form.Item label="หมวดหมู่ (Category)" style={{ marginBottom: 0 }}>
                        <Select
                          allowClear
                          placeholder={selectedProjectId ? "ทั้งหมด..." : "เลือกโครงการก่อน..."}
                          value={selectedCategoryId}
                          onChange={(val) => setSelectedCategoryId(val)}
                          disabled={!selectedProjectId || monitorFeatures.length === 0}
                          loading={isLoadingMonitorFeatures}
                          options={monitorFeatures.map((f) => ({
                            value: f.categoryId,
                            label: `${f.categoryName} (${f.fullPath})`,
                          }))}
                          style={{ width: '100%' }}
                        />
                      </Form.Item>
                    </Col>

                    <Col xs={12} sm={12} md={8}>
                      <Form.Item label="ระดับความรุนแรง (Log Level)" style={{ marginBottom: 0 }}>
                        <Select
                          allowClear
                          placeholder="ทั้งหมด"
                          value={logLevel}
                          onChange={(val) => setLogLevel(val)}
                          options={['DEBUG', 'INFO', 'WARN', 'ERROR'].map((val) => ({
                            value: val,
                            label: val,
                          }))}
                          style={{ width: '100%' }}
                        />
                      </Form.Item>
                    </Col>

                    <Col xs={24} sm={24} md={16}>
                      <Form.Item label="คำค้นหา (Keyword)" style={{ marginBottom: 0 }}>
                        <Input.Search
                          placeholder="ค้นหาข้อความ, พาธ, Trace ID..."
                          value={logKeyword}
                          onChange={(e) => setLogKeyword(e.target.value)}
                          onSearch={handleSearch}
                          enterButton={
                            <Button type="primary" icon={<SearchOutlined />}>
                              ค้นหา
                            </Button>
                          }
                          allowClear
                        />
                      </Form.Item>
                    </Col>
                  </Row>
                </Form>
              </Card>

              {/* Logs Table Card */}
              {selectedProductId ? (
                <Card bordered={false} className="shadow-sm" style={{ borderRadius: '12px' }}>
                  <Table
                    rowKey="logId"
                    loading={isLoadingLogs}
                    dataSource={logs}
                    locale={{ emptyText: 'ไม่พบข้อมูล Logs สำหรับเงื่อนไขที่เลือก' }}
                    pagination={{
                      current: currentPage,
                      pageSize: pageSize,
                      total: totalLogs,
                      showSizeChanger: true,
                      pageSizeOptions: ['10', '20', '50', '100'],
                      onChange: (page, size) => {
                        setCurrentPage(page);
                        setPageSize(size);
                      },
                      showTotal: (total, range) => `แสดงผล ${range[0]}-${range[1]} จากทั้งหมด ${total} รายการ`,
                    }}
                    columns={[
                      {
                        title: 'เวลาบันทึก (Timestamp)',
                        dataIndex: 'timestamp',
                        key: 'timestamp',
                        width: 200,
                        render: (value: string | null | undefined) => (
                          <Text code style={{ fontFamily: 'monospace' }}>
                            {value ? new Date(value).toLocaleString('th-TH') : '-'}
                          </Text>
                        ),
                      },
                      {
                        title: 'ระดับ (Level)',
                        dataIndex: 'level',
                        key: 'level',
                        width: 100,
                        render: (value: string | null | undefined) => (
                          <Tag color={getLevelColor(value)} style={{ fontWeight: 600 }}>
                            {value?.toUpperCase() || 'INFO'}
                          </Tag>
                        ),
                      },
                      {
                        title: 'ประเภท (Log Type)',
                        dataIndex: 'logType',
                        key: 'logType',
                        width: 120,
                        render: (value: string | null | undefined) => (
                          <Tag bordered={false}>{value || 'N/A'}</Tag>
                        ),
                      },
                      {
                        title: 'ข้อความบันทึก (Message)',
                        dataIndex: 'message',
                        key: 'message',
                        render: (value: string | null | undefined, record: any) => (
                          <div style={{ maxWidth: '450px', wordBreak: 'break-all' }}>
                            <Text strong style={{ display: 'block' }}>{value || '-'}</Text>
                            {record.path && (
                              <Text type="secondary" style={{ fontSize: '11px', fontFamily: 'monospace' }}>
                                [{record.method || 'GET'}] {record.path}
                              </Text>
                            )}
                          </div>
                        ),
                      },
                      {
                        title: 'การจัดการ',
                        key: 'action',
                        width: 110,
                        render: (_: any, record: MainLog) => (
                          <Button
                            type="link"
                            icon={<EyeOutlined />}
                            onClick={() => setSelectedLog(record)}
                          >
                            เปิดดู
                          </Button>
                        ),
                      },
                    ]}
                  />
                </Card>
              ) : (
                <Alert
                  message="กรุณาเลือกผลิตภัณฑ์ทางขวามือ"
                  description="เพื่อสืบค้นข้อมูล Logs จาก Elasticsearch กรุณาเลือกผลิตภัณฑ์ที่ต้องการติดตามข้อมูล และระบุสภาพแวดล้อมเพื่อเริ่มต้นตรวจสอบการทำงาน"
                  type="info"
                  showIcon
                  style={{ borderRadius: '12px', padding: '24px' }}
                />
              )}
            </Space>
          </Col>
        </Row>
      </Space>

      {/* Log Details Modal */}
      <Modal
        title={
          <div style={{ borderBottom: '1px solid #f0f0f0', paddingBottom: '10px' }}>
            <Title level={4} style={{ margin: 0 }}>
              รายละเอียด Log Document
            </Title>
            <Text type="secondary">ID: {selectedLog?.logId}</Text>
          </div>
        }
        open={!!selectedLog}
        onCancel={() => setSelectedLog(null)}
        footer={[
          <Button key="close" type="primary" onClick={() => setSelectedLog(null)}>
            ปิดหน้าต่าง
          </Button>,
        ]}
        width={800}
      >
        {selectedLog && (
          <div style={{ marginTop: '15px' }}>
            <Row gutter={[16, 16]} style={{ marginBottom: '15px' }}>
              <Col span={8}>
                <Card size="small" title="Timestamp" bordered={false} style={{ background: '#f5f5f5' }}>
                  <Text>{selectedLog.timestamp ? new Date(selectedLog.timestamp).toLocaleString('th-TH') : '-'}</Text>
                </Card>
              </Col>
              <Col span={8}>
                <Card size="small" title="Log Level" bordered={false} style={{ background: '#f5f5f5' }}>
                  <Tag color={getLevelColor(selectedLog.level)}>{selectedLog.level || 'INFO'}</Tag>
                </Card>
              </Col>
              <Col span={8}>
                <Card size="small" title="Method & Status" bordered={false} style={{ background: '#f5f5f5' }}>
                  <Text>{selectedLog.method || 'N/A'} - {selectedLog.statusCode || '-'}</Text>
                </Card>
              </Col>
            </Row>

            <Title level={5}>Raw JSON Document (Elasticsearch payload)</Title>
            <pre
              style={{
                background: '#1e1e1e',
                color: '#d4d4d4',
                padding: '15px',
                borderRadius: '8px',
                overflowX: 'auto',
                whiteSpace: 'pre-wrap',
                fontFamily: 'monospace',
                fontSize: '12px',
                maxHeight: '400px',
              }}
            >
              {JSON.stringify(selectedLog.raw || selectedLog, null, 2)}
            </pre>
          </div>
        )}
      </Modal>
    </div>
  );
}
