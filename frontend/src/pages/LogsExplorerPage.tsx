import { useEffect, useState, useMemo, useCallback } from 'react';
import { Card, Space, Typography, Modal, Button, Select, Input, Table, Tag, Form, Row, Col, Badge, message, Spin, Switch, Statistic, Tooltip as AntTooltip } from 'antd';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  SearchOutlined,
  EyeOutlined,
  ReloadOutlined,
  LineChartOutlined,
  ThunderboltOutlined,
  DatabaseOutlined,
} from '@ant-design/icons';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  Area,
  AreaChart,
} from 'recharts';
import { productAdminService, normalizeMainLog } from '@/services';
import { useAppStore, useAuthStore } from '@/store';
import { ROUTES } from '@/constants';
import type { MainLog, Project, ProjectFeature } from '@/types';

const { Text, Title, Paragraph } = Typography;

// Color scheme for log levels
const LEVEL_COLORS: Record<string, string> = {
  DEBUG: '#8b5cf6',
  INFO: '#10b981',
  WARN: '#f59e0b',
  ERROR: '#ef4444',
  FATAL: '#dc2626',
};

const LEVEL_OPTIONS = ['DEBUG', 'INFO', 'WARN', 'ERROR'];

export function LogsExplorerPage() {
  const setBreadcrumbs = useAppStore((state) => state.setBreadcrumbs);
  const accessToken = useAuthStore((state) => state.accessToken);

  // Filter States
  const [selectedProductId, setSelectedProductId] = useState<number | null>(null);
  const [selectedEnvironmentId, setSelectedEnvironmentId] = useState<number | null>(null);
  const [selectedProjectIds, setSelectedProjectIds] = useState<number[]>([]);
  const [selectedCategoryIds, setSelectedCategoryIds] = useState<number[]>([]);
  const [selectedLevels, setSelectedLevels] = useState<string[]>([]);
  const [logKeyword, setLogKeyword] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [selectedLog, setSelectedLog] = useState<MainLog | null>(null);

  useEffect(() => {
    setBreadcrumbs([
      { title: 'แดชบอร์ด', path: ROUTES.DASHBOARD },
      { title: 'Logs Explorer' },
    ]);
  }, [setBreadcrumbs]);

  // ─── Data Queries ─────────────────────────

  const { data: products = [], isLoading: isLoadingProducts } = useQuery({
    queryKey: ['products-admin'],
    queryFn: () => productAdminService.listProducts(),
  });

  const monitorProduct = products.find((p) => p.productId === selectedProductId);
  const monitorEnvironments = monitorProduct?.productEnvironments || [];

  const { data: monitorProjects = [], isLoading: isLoadingProjects } = useQuery({
    queryKey: ['projects-admin-explorer', selectedProductId],
    queryFn: () => productAdminService.listProjects(selectedProductId!),
    enabled: !!selectedProductId,
  });

  // Fetch features for all selected projects
  const { data: allFeatures = [] } = useQuery({
    queryKey: ['features-admin-explorer', selectedProductId, selectedProjectIds],
    queryFn: async () => {
      if (!selectedProductId || selectedProjectIds.length === 0) return [];
      const results = await Promise.all(
        selectedProjectIds.map((projId) =>
          productAdminService.listFeatures(selectedProductId!, projId)
        )
      );
      return results.flat();
    },
    enabled: !!selectedProductId && selectedProjectIds.length > 0,
  });

  // Search Logs (multi-select)
  const {
    data: logsData,
    isLoading: isLoadingLogs,
    refetch: refetchLogs,
  } = useQuery({
    queryKey: [
      'explorer-logs',
      selectedProductId,
      selectedEnvironmentId,
      selectedProjectIds,
      selectedCategoryIds,
      selectedLevels,
      logKeyword,
      currentPage,
      pageSize,
    ],
    queryFn: () =>
      productAdminService.searchLogsMulti({
        productId: selectedProductId!,
        environmentId: selectedEnvironmentId || undefined,
        projectIds: selectedProjectIds.length > 0 ? selectedProjectIds : undefined,
        categoryIds: selectedCategoryIds.length > 0 ? selectedCategoryIds : undefined,
        levels: selectedLevels.length > 0 ? selectedLevels : undefined,
        keyword: logKeyword || undefined,
        page: currentPage,
        perPage: pageSize,
      }),
    enabled: !!selectedProductId,
    refetchInterval: false, // Replaced by SSE Live Tail
  });

  const queryClient = useQueryClient();

  // Handle Live Tail SSE connection
  useEffect(() => {
    if (!autoRefresh || !selectedProductId || !accessToken) return;

    const apiBase = import.meta.env.VITE_API_BASE_URL || '/api/v1';
    let url = `${apiBase}/logs/live?product_id=${selectedProductId}&token=${encodeURIComponent(accessToken)}`;
    if (selectedEnvironmentId) {
      url += `&environment_id=${selectedEnvironmentId}`;
    }

    const eventSource = new EventSource(url);

    eventSource.onmessage = (event) => {
      try {
        const parsedData = JSON.parse(event.data);
        if (Array.isArray(parsedData) && parsedData.length > 0) {
          const newLogs = parsedData.map(normalizeMainLog);
          
          queryClient.setQueriesData(
            { queryKey: ['explorer-logs', selectedProductId] },
            (oldData: any) => {
              if (!oldData || !oldData.data) return oldData;
              const combined = [...newLogs, ...oldData.data];
              const sliced = combined.slice(0, 1000);
              return {
                ...oldData,
                data: sliced,
                total: oldData.total + newLogs.length
              };
            }
          );
        }
      } catch (error) {
        console.error('Failed to parse SSE data', error);
      }
    };

    eventSource.onerror = (error) => {
      console.error('EventSource error:', error);
    };

    return () => {
      eventSource.close();
    };
  }, [
    autoRefresh,
    selectedProductId,
    selectedEnvironmentId,
    queryClient,
    selectedProjectIds,
    selectedCategoryIds,
    selectedLevels,
    logKeyword,
    currentPage,
    pageSize,
    accessToken,
  ]);

  // Stats for charts
  const { data: statsData, isLoading: isLoadingStats } = useQuery({
    queryKey: [
      'explorer-stats',
      selectedProductId,
      selectedEnvironmentId,
      selectedProjectIds,
      selectedCategoryIds,
    ],
    queryFn: () =>
      productAdminService.getDashboardLogStats({
        productId: selectedProductId!,
        environmentId: selectedEnvironmentId || undefined,
        projectIds: selectedProjectIds.length > 0 ? selectedProjectIds : undefined,
        categoryIds: selectedCategoryIds.length > 0 ? selectedCategoryIds : undefined,
      }),
    enabled: !!selectedProductId,
    refetchInterval: autoRefresh ? 15000 : false,
  });

  const logs = logsData?.data || [];
  const totalLogs = logsData?.total || 0;

  // ─── Derived Chart Data ─────────────────────────

  const chartData = useMemo(() => {
    if (!statsData?.logs_over_time?.buckets) return [];
    return statsData.logs_over_time.buckets.map((bucket: any) => {
      const time = new Date(bucket.key_as_string);
      const entry: Record<string, any> = {
        time: time.toLocaleTimeString('th-TH', { hour: '2-digit', minute: '2-digit' }),
        fullTime: time.toLocaleString('th-TH'),
        total: bucket.doc_count,
      };
      // Add per-level counts
      if (bucket.by_level?.buckets) {
        for (const lvl of bucket.by_level.buckets) {
          entry[lvl.key] = lvl.doc_count;
        }
      }
      return entry;
    });
  }, [statsData]);

  const levelSummary = useMemo(() => {
    if (!statsData?.log_levels?.buckets) return [];
    return statsData.log_levels.buckets.map((b: any) => ({
      level: b.key,
      count: b.doc_count,
    }));
  }, [statsData]);

  // Get unique levels from chart data for dynamic line rendering
  const activeChartLevels = useMemo(() => {
    const levels = new Set<string>();
    chartData.forEach((d: any) => {
      LEVEL_OPTIONS.forEach((lvl) => {
        if (d[lvl] !== undefined && d[lvl] > 0) levels.add(lvl);
      });
    });
    return Array.from(levels);
  }, [chartData]);

  // ─── Effects ─────────────────────────

  useEffect(() => {
    if (monitorEnvironments.length > 0 && !selectedEnvironmentId) {
      setSelectedEnvironmentId(monitorEnvironments[0].environmentId);
    }
  }, [monitorEnvironments, selectedEnvironmentId]);

  useEffect(() => {
    setSelectedProjectIds([]);
    setSelectedCategoryIds([]);
    setSelectedEnvironmentId(null);
    setCurrentPage(1);
  }, [selectedProductId]);

  useEffect(() => {
    setSelectedCategoryIds([]);
  }, [selectedProjectIds]);

  // ─── Helpers ─────────────────────────

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

  const getProjectName = useCallback(
    (projectId: number | null | undefined) => {
      if (!projectId) return null;
      const proj = monitorProjects.find((p) => p.projectId === projectId);
      return proj?.projectName || null;
    },
    [monitorProjects]
  );

  const getCategoryName = useCallback(
    (categoryId: number | null | undefined) => {
      if (!categoryId) return null;
      const cat = allFeatures.find((f) => f.categoryId === categoryId);
      return cat?.categoryName || null;
    },
    [allFeatures]
  );

  return (
    <div style={{ padding: '8px' }}>
      <Space direction="vertical" style={{ width: '100%' }} size="large">
        {/* Header */}
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            flexWrap: 'wrap',
            gap: '12px',
          }}
        >
          <div>
            <Title level={3} style={{ margin: 0, fontWeight: 700 }}>
              <LineChartOutlined style={{ marginRight: 8, color: 'var(--color-primary)' }} />
              Logs Explorer
            </Title>
            <Text type="secondary">
              สำรวจ Logs หลายประเภท หลาย Sub-Category พร้อมกัน พร้อมกราฟเส้นแสดงแนวโน้ม
            </Text>
          </div>
          <Space>
            <AntTooltip title={autoRefresh ? 'ปิด Live Tail' : 'เปิด Live Tail (สตรีม Real-time)'}>
              <Switch
                checked={autoRefresh}
                onChange={setAutoRefresh}
                checkedChildren={<ThunderboltOutlined />}
                unCheckedChildren="Live"
              />
            </AntTooltip>
            {selectedProductId && (
              <Button
                type="text"
                icon={<ReloadOutlined spin={isLoadingLogs} />}
                onClick={() => refetchLogs()}
              >
                Refresh
              </Button>
            )}
          </Space>
        </div>

        {/* Filters */}
        <Card
          bordered={false}
          className="shadow-sm"
          style={{ borderRadius: '12px' }}
          title={
            <span style={{ fontWeight: 600 }}>
              <SearchOutlined style={{ marginRight: 8 }} />
              ตัวกรองการสืบค้น (Multi-Select Filters)
            </span>
          }
        >
          <Form layout="vertical">
            <Row gutter={[12, 12]} align="bottom">
              <Col xs={24} sm={12} md={6}>
                <Form.Item label="ผลิตภัณฑ์ (Product)" style={{ marginBottom: 0 }}>
                  <Select
                    showSearch
                    placeholder="เลือกผลิตภัณฑ์..."
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

              <Col xs={24} sm={12} md={6}>
                <Form.Item label="สภาพแวดล้อม (Environment)" style={{ marginBottom: 0 }}>
                  <Select
                    placeholder={selectedProductId ? 'เลือกสภาพแวดล้อม...' : 'เลือกผลิตภัณฑ์ก่อน...'}
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

              <Col xs={24} sm={12} md={6}>
                <Form.Item label="โครงการ (Projects — เลือกได้หลายรายการ)" style={{ marginBottom: 0 }}>
                  <Select
                    mode="multiple"
                    allowClear
                    maxTagCount="responsive"
                    placeholder={selectedProductId ? 'เลือกโครงการ...' : 'เลือกผลิตภัณฑ์ก่อน...'}
                    value={selectedProjectIds}
                    onChange={(val) => setSelectedProjectIds(val)}
                    disabled={!selectedProductId || monitorProjects.length === 0}
                    loading={isLoadingProjects}
                    options={monitorProjects.map((p) => ({
                      value: p.projectId,
                      label: `${p.projectName} (${p.projectCode})`,
                    }))}
                    style={{ width: '100%' }}
                  />
                </Form.Item>
              </Col>

              <Col xs={24} sm={12} md={6}>
                <Form.Item
                  label="หมวดหมู่ (Categories — เลือกได้หลายรายการ)"
                  style={{ marginBottom: 0 }}
                >
                  <Select
                    mode="multiple"
                    allowClear
                    maxTagCount="responsive"
                    placeholder={
                      selectedProjectIds.length > 0 ? 'เลือกหมวดหมู่...' : 'เลือกโครงการก่อน...'
                    }
                    value={selectedCategoryIds}
                    onChange={(val) => setSelectedCategoryIds(val)}
                    disabled={selectedProjectIds.length === 0 || allFeatures.length === 0}
                    options={allFeatures.map((f) => ({
                      value: f.categoryId,
                      label: `${f.categoryName} (${f.fullPath || f.categoryCode})`,
                    }))}
                    style={{ width: '100%' }}
                  />
                </Form.Item>
              </Col>

              <Col xs={12} sm={12} md={6}>
                <Form.Item
                  label="ระดับ Log Level (เลือกได้หลายรายการ)"
                  style={{ marginBottom: 0 }}
                >
                  <Select
                    mode="multiple"
                    allowClear
                    maxTagCount="responsive"
                    placeholder="ทั้งหมด"
                    value={selectedLevels}
                    onChange={(val) => setSelectedLevels(val)}
                    options={LEVEL_OPTIONS.map((val) => ({
                      value: val,
                      label: (
                        <span>
                          <span
                            style={{
                              display: 'inline-block',
                              width: 8,
                              height: 8,
                              borderRadius: '50%',
                              backgroundColor: LEVEL_COLORS[val],
                              marginRight: 6,
                            }}
                          />
                          {val}
                        </span>
                      ),
                    }))}
                    style={{ width: '100%' }}
                  />
                </Form.Item>
              </Col>

              <Col xs={24} sm={24} md={18}>
                <Form.Item label="คำค้นหา (Keyword)" style={{ marginBottom: 0 }}>
                  <Input.Search
                    placeholder="ค้นหาข้อความ, พาธ, Trace ID..."
                    value={logKeyword}
                    onChange={(e) => setLogKeyword(e.target.value)}
                    onSearch={() => {
                      setCurrentPage(1);
                      refetchLogs();
                    }}
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

        {/* Stats Summary + Chart */}
        {selectedProductId && (
          <Row gutter={[16, 16]}>
            {/* Level Summary Cards */}
            <Col xs={24}>
              <Row gutter={[12, 12]}>
                {levelSummary.map((item: { level: string; count: number }) => (
                  <Col xs={12} sm={6} md={6} lg={3} key={item.level}>
                    <Card
                      bordered={false}
                      className="shadow-sm"
                      style={{
                        borderRadius: '12px',
                        borderLeft: `4px solid ${LEVEL_COLORS[item.level] || '#999'}`,
                      }}
                      size="small"
                    >
                      <Statistic
                        title={
                          <span style={{ fontSize: '12px', color: '#666' }}>
                            {item.level}
                          </span>
                        }
                        value={item.count}
                        valueStyle={{
                          color: LEVEL_COLORS[item.level] || '#999',
                          fontWeight: 700,
                          fontSize: '20px',
                        }}
                      />
                    </Card>
                  </Col>
                ))}
                {levelSummary.length === 0 && !isLoadingStats && (
                  <Col span={24}>
                    <Text type="secondary">ยังไม่มีข้อมูลสถิติ</Text>
                  </Col>
                )}
              </Row>
            </Col>

            {/* Line Chart */}
            <Col xs={24}>
              <Card
                bordered={false}
                className="shadow-sm"
                style={{ borderRadius: '12px' }}
                title={
                  <Space>
                    <LineChartOutlined style={{ color: 'var(--color-primary)' }} />
                    <span style={{ fontWeight: 600 }}>Logs Over Time (Per Hour)</span>
                  </Space>
                }
              >
                {isLoadingStats ? (
                  <div style={{ textAlign: 'center', padding: '60px 0' }}>
                    <Spin size="large" />
                  </div>
                ) : chartData.length > 0 ? (
                  <ResponsiveContainer width="100%" height={320}>
                    <AreaChart data={chartData} margin={{ top: 10, right: 30, left: 0, bottom: 0 }}>
                      <defs>
                        {activeChartLevels.map((level) => (
                          <linearGradient key={level} id={`color${level}`} x1="0" y1="0" x2="0" y2="1">
                            <stop offset="5%" stopColor={LEVEL_COLORS[level]} stopOpacity={0.3} />
                            <stop offset="95%" stopColor={LEVEL_COLORS[level]} stopOpacity={0} />
                          </linearGradient>
                        ))}
                      </defs>
                      <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                      <XAxis
                        dataKey="time"
                        tick={{ fontSize: 11, fill: '#999' }}
                        axisLine={{ stroke: '#e0e0e0' }}
                      />
                      <YAxis
                        tick={{ fontSize: 11, fill: '#999' }}
                        axisLine={{ stroke: '#e0e0e0' }}
                        allowDecimals={false}
                      />
                      <Tooltip
                        contentStyle={{
                          background: 'rgba(0,0,0,0.85)',
                          border: 'none',
                          borderRadius: '8px',
                          color: '#fff',
                          fontSize: 12,
                        }}
                        labelStyle={{ color: '#ccc', marginBottom: 4 }}
                        formatter={(value: number, name: string) => [value, name]}
                        labelFormatter={(label: string, payload: any[]) => {
                          const first = payload?.[0]?.payload;
                          return first?.fullTime || label;
                        }}
                      />
                      <Legend
                        iconType="circle"
                        wrapperStyle={{ fontSize: 12, paddingTop: '8px' }}
                      />
                      {activeChartLevels.map((level) => (
                        <Area
                          key={level}
                          type="monotone"
                          dataKey={level}
                          stroke={LEVEL_COLORS[level]}
                          strokeWidth={2}
                          fill={`url(#color${level})`}
                          dot={false}
                          activeDot={{ r: 4, strokeWidth: 2 }}
                        />
                      ))}
                    </AreaChart>
                  </ResponsiveContainer>
                ) : (
                  <div style={{ textAlign: 'center', padding: '60px 0', color: '#999' }}>
                    <DatabaseOutlined style={{ fontSize: 48, marginBottom: 12, display: 'block' }} />
                    <Text type="secondary">ยังไม่มีข้อมูลกราฟ</Text>
                  </div>
                )}
              </Card>
            </Col>
          </Row>
        )}

        {/* Logs Table */}
        {selectedProductId ? (
          <Card
            bordered={false}
            className="shadow-sm"
            style={{ borderRadius: '12px' }}
            title={
              <Space>
                <DatabaseOutlined style={{ color: 'var(--color-primary)' }} />
                <span style={{ fontWeight: 600 }}>
                  รายการ Logs ({totalLogs.toLocaleString()} รายการ)
                </span>
                {autoRefresh && (
                  <Space size={6} style={{ marginLeft: 8 }}>
                    <span className="live-dot" />
                    <span style={{ fontSize: '11px', color: 'var(--color-success)', fontWeight: 'bold', letterSpacing: '0.05em' }}>
                      LIVE TAIL
                    </span>
                  </Space>
                )}
              </Space>
            }
          >
            <Table
              rowKey="logId"
              loading={isLoadingLogs}
              dataSource={logs}
              locale={{ emptyText: 'ไม่พบข้อมูล Logs สำหรับเงื่อนไขที่เลือก' }}
              scroll={{ x: 900 }}
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
                showTotal: (total, range) =>
                  `แสดงผล ${range[0]}-${range[1]} จากทั้งหมด ${total} รายการ`,
              }}
              columns={[
                {
                  title: 'เวลา (Timestamp)',
                  dataIndex: 'timestamp',
                  key: 'timestamp',
                  width: 170,
                  render: (value: string | null | undefined) => (
                    <Text code style={{ fontFamily: 'monospace', fontSize: '11px' }}>
                      {value ? new Date(value).toLocaleString('th-TH') : '-'}
                    </Text>
                  ),
                },
                {
                  title: 'Level',
                  dataIndex: 'level',
                  key: 'level',
                  width: 90,
                  render: (value: string | null | undefined) => (
                    <Tag color={getLevelColor(value)} style={{ fontWeight: 600 }}>
                      {value?.toUpperCase() || 'INFO'}
                    </Tag>
                  ),
                },
                {
                  title: 'Project',
                  key: 'project',
                  width: 130,
                  render: (_: any, record: any) => {
                    const projId = record.raw?.payload?.project_id || record.raw?.project_id;
                    const name = getProjectName(projId);
                    return name ? (
                      <Tag bordered={false} color="blue">
                        {name}
                      </Tag>
                    ) : (
                      <Text type="secondary" style={{ fontSize: 11 }}>—</Text>
                    );
                  },
                },
                {
                  title: 'Category',
                  key: 'category',
                  width: 140,
                  render: (_: any, record: any) => {
                    const catId = record.raw?.payload?.category_id || record.raw?.category_id;
                    const name = getCategoryName(catId);
                    const path = record.raw?.payload?.feature_full_path;
                    return name ? (
                      <AntTooltip title={path || name}>
                        <Tag bordered={false} color="purple">
                          {name}
                        </Tag>
                      </AntTooltip>
                    ) : (
                      <Text type="secondary" style={{ fontSize: 11 }}>—</Text>
                    );
                  },
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
                  title: 'ข้อความ (Message)',
                  dataIndex: 'message',
                  key: 'message',
                  render: (value: string | null | undefined, record: any) => (
                    <div style={{ maxWidth: '400px', wordBreak: 'break-all' }}>
                      <Text strong style={{ display: 'block' }}>
                        {value || '-'}
                      </Text>
                      {record.path && (
                        <Text
                          type="secondary"
                          style={{ fontSize: '11px', fontFamily: 'monospace' }}
                        >
                          [{record.method || 'GET'}] {record.path}
                        </Text>
                      )}
                    </div>
                  ),
                },
                {
                  title: '',
                  key: 'action',
                  width: 80,
                  render: (_: any, record: MainLog) => (
                    <Button
                      type="link"
                      icon={<EyeOutlined />}
                      onClick={() => setSelectedLog(record)}
                      size="small"
                    >
                      ดู
                    </Button>
                  ),
                },
              ]}
            />
          </Card>
        ) : (
          <Card
            bordered={false}
            className="shadow-sm"
            style={{
              borderRadius: '12px',
              textAlign: 'center',
              padding: '60px 24px',
              background: 'linear-gradient(135deg, #f5f3ff 0%, #ede9fe 100%)',
            }}
          >
            <LineChartOutlined
              style={{ fontSize: 64, color: 'var(--color-primary)', marginBottom: 16 }}
            />
            <Title level={4} style={{ color: '#333' }}>
              เลือกผลิตภัณฑ์เพื่อเริ่มต้นสำรวจ Logs
            </Title>
            <Paragraph type="secondary">
              เลือกผลิตภัณฑ์ด้านบน จากนั้นสามารถเลือกหลาย Project, หลาย Category, หลาย Level
              พร้อมกันได้ ระบบจะแสดงกราฟเส้นข้อมูลและตาราง Logs ทันที
            </Paragraph>
          </Card>
        )}
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
                <Card
                  size="small"
                  title="Timestamp"
                  bordered={false}
                  style={{ background: '#f5f5f5' }}
                >
                  <Text>
                    {selectedLog.timestamp
                      ? new Date(selectedLog.timestamp).toLocaleString('th-TH')
                      : '-'}
                  </Text>
                </Card>
              </Col>
              <Col span={8}>
                <Card
                  size="small"
                  title="Log Level"
                  bordered={false}
                  style={{ background: '#f5f5f5' }}
                >
                  <Tag color={getLevelColor(selectedLog.level)}>
                    {selectedLog.level || 'INFO'}
                  </Tag>
                </Card>
              </Col>
              <Col span={8}>
                <Card
                  size="small"
                  title="Method & Status"
                  bordered={false}
                  style={{ background: '#f5f5f5' }}
                >
                  <Text>
                    {selectedLog.method || 'N/A'} - {selectedLog.statusCode || '-'}
                  </Text>
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
