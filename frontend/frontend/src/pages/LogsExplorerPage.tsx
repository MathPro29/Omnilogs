import { useEffect, useState, useMemo, useCallback, type ReactNode } from 'react';
import { Card, Space, Typography, Modal, Button, Select, Input, Table, Tag, Form, Row, Col, Spin, Switch, Statistic, Tooltip as AntTooltip, message } from 'antd';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  SearchOutlined,
  EyeOutlined,
  ReloadOutlined,
  LineChartOutlined,
  ThunderboltOutlined,
  DatabaseOutlined,
  StarOutlined,
  StarFilled,
} from '@ant-design/icons';
import {
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
import { customFieldService, type CustomField } from '@/services/custom-field.service';
import { useAppStore, useAuthStore } from '@/store';
import { ROUTES } from '@/constants';
import type { MainLog } from '@/types';

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

const normalizeFavoriteFieldPath = (path: string) => {
  const normalized = path.replace(/^raw\./, '');
  if (normalized === 'payload') return '$';
  return normalized.replace(/^payload\./, '');
};

const detectRawJSONType = (value: unknown): string => {
  if (Array.isArray(value)) return 'array';
  if (value !== null && typeof value === 'object') return 'object';
  if (typeof value === 'boolean') return 'boolean';
  if (typeof value === 'number') return 'number';
  if (typeof value === 'string' && value.includes('-') && !Number.isNaN(Date.parse(value))) return 'datetime';
  return 'string';
};

type JSONFavoriteTreeProps = {
  value: unknown;
  path?: string;
  level?: number;
  favoritePaths: Set<string>;
  onToggleFavorite: (path: string, value: unknown, dataType: string, isFavorited: boolean) => void;
};

const JSONFavoriteRow = ({ label, path, canFavorite, isFavorited, onToggle, children }: {
  label?: ReactNode;
  path: string;
  canFavorite: boolean;
  isFavorited: boolean;
  onToggle: () => void;
  children: ReactNode;
}) => (
  <div
    className="json-fav-row"
    style={{ display: 'flex', alignItems: 'flex-start', gap: 0, position: 'relative', paddingRight: canFavorite ? 28 : 0 }}
  >
    <div style={{ flex: 1, minWidth: 0 }}>
      {label}{children}
    </div>
    {canFavorite && (
      <AntTooltip title={isFavorited ? 'นำออกจาก Custom Field' : 'เพิ่มไปยัง Custom Field'}>
        <Button
          type="text"
          size="small"
          icon={isFavorited ? <StarFilled /> : <StarOutlined />}
          aria-label={isFavorited ? `Remove ${path} from favorite` : `Add ${path} to favorite`}
          onClick={onToggle}
          className={isFavorited ? 'json-fav-star json-fav-star' : 'json-fav-star'}
          style={{
            color: '#faad14',
            padding: '0 4px',
            height: 20,
            width: 20,
            position: 'absolute',
            right: 0,
            top: 2,
            opacity: isFavorited ? 1 : 0.4,
          }}
        />
      </AntTooltip>
    )}
  </div>
);

const JSONFavoriteTree = ({ value, path = '', level = 0, favoritePaths, onToggleFavorite }: JSONFavoriteTreeProps) => {
  const indent = level * 16;

  // Leaf / primitive value
  if (value === null || typeof value !== 'object') {
    return (
      <Text style={{ color: typeof value === 'string' ? '#ce9178' : typeof value === 'number' ? '#b5cea8' : typeof value === 'boolean' ? '#569cd6' : '#d4d4d4' }}>
        {JSON.stringify(value)}
      </Text>
    );
  }

  // Array
  if (Array.isArray(value)) {
    if (value.length === 0) return <Text type="secondary">[]</Text>;
    return (
      <>
        <Text type="secondary">[</Text>
        {value.map((item, index) => (
          <div key={`${path}[${index}]`} style={{ marginLeft: indent + 16, lineHeight: 1.7 }}>
            <JSONFavoriteTree value={item} path={`${path}[]`} level={level + 1} favoritePaths={favoritePaths} onToggleFavorite={onToggleFavorite} />
            {index < value.length - 1 && <Text type="secondary">,</Text>}
          </div>
        ))}
        <Text type="secondary">]</Text>
      </>
    );
  }

  // Object
  const entries = Object.entries(value as Record<string, unknown>);
  if (entries.length === 0) return <Text type="secondary">{'{}'}</Text>;

  return (
    <>
      <Text type="secondary">{'{'}</Text>
      {entries.map(([key, child], idx) => {
        const childPath = path ? `${path}.${key}` : key;
        const childCanFav = !!childPath;
        const childIsFav = favoritePaths.has(normalizeFavoriteFieldPath(childPath));
        const childDataType = detectRawJSONType(child);
        return (
          <div key={childPath} style={{ marginLeft: indent + 16, lineHeight: 1.7 }}>
            <JSONFavoriteRow
              label={<Text style={{ color: '#9cdcfe' }}>{JSON.stringify(key)}: </Text>}
              path={childPath}
              canFavorite={childCanFav}
              isFavorited={childIsFav}
              onToggle={() => onToggleFavorite(childPath, child, childDataType, childIsFav)}
            >
              <JSONFavoriteTree value={child} path={childPath} level={level + 1} favoritePaths={favoritePaths} onToggleFavorite={onToggleFavorite} />
              {idx < entries.length - 1 && <Text type="secondary">,</Text>}
            </JSONFavoriteRow>
          </div>
        );
      })}
      <Text type="secondary">{'}'}</Text>
    </>
  );
};

const getCustomFieldValue = (record: any, path: string): unknown => {
  if (!record) return undefined;
  if (normalizeFavoriteFieldPath(path) === '$') {
    return record.raw?.payload || record.payload || record.raw || record;
  }
  const normalizedPath = path.replace(/^payload\./, '').replace(/^custom_fields\./, '');
  const segments = normalizedPath.split('.').filter(Boolean);
  const candidates = [
    record.customFields,
    record.raw?.payload?.custom_fields,
    record.raw?.custom_fields,
    record.raw?.payload,
    record.raw,
    record.payload,
  ];

  const readPath = (value: any, index: number): unknown => {
    if (index >= segments.length) return value;
    if (Array.isArray(value)) {
      const values = value.map((item) => readPath(item, index)).filter((item) => item !== undefined && item !== null);
      return values.length > 0 ? values : undefined;
    }
    if (!value || typeof value !== 'object') return undefined;
    const segment = segments[index];
    if (segment.endsWith('[]')) {
      const arrayValue = value[segment.slice(0, -2)];
      return Array.isArray(arrayValue) ? readPath(arrayValue, index + 1) : undefined;
    }
    return segment in value ? readPath(value[segment], index + 1) : undefined;
  };

  for (const candidate of candidates) {
    const current = readPath(candidate, 0);
    if (current !== undefined && current !== null) return current;
  }
  return undefined;
};

const formatCustomFieldValue = (field: CustomField | undefined, value: unknown): string => {
  if (value === undefined || value === null || value === '') return '—';
  if (field?.data_type === 'enum') {
    return field.enum_options?.find((option) => option.option_value === String(value))?.option_label || String(value);
  }
  if (typeof value === 'boolean') return value ? 'ใช่' : 'ไม่ใช่';
  if (typeof value === 'object') return JSON.stringify(value);
  return String(value);
};

const filterJsonObject = (obj: any, term: string): any => {
  if (!term) return obj;
  if (typeof obj !== 'object' || obj === null) return obj;

  const cleanTerm = term.toLowerCase().trim();
  if (cleanTerm === '') return obj;
  
  if (Array.isArray(obj)) {
    return obj
      .map(item => filterJsonObject(item, term))
      .filter(item => {
        if (typeof item === 'object' && item !== null) {
          return Object.keys(item).length > 0;
        }
        return String(item).toLowerCase().includes(cleanTerm);
      });
  }

  const result: any = {};
  for (const [key, value] of Object.entries(obj)) {
    if (key.toLowerCase().includes(cleanTerm)) {
      result[key] = value;
      continue;
    }
    if (typeof value === 'object' && value !== null) {
      const filteredValue = filterJsonObject(value, term);
      if (filteredValue && (Array.isArray(filteredValue) ? filteredValue.length > 0 : Object.keys(filteredValue).length > 0)) {
        result[key] = filteredValue;
      }
    } else if (String(value).toLowerCase().includes(cleanTerm)) {
      result[key] = value;
    }
  }
  return result;
};

const getVisibleCustomFieldsForLog = (log: MainLog, allCustomFields: CustomField[]) => {
  const raw = (log.raw || log) as any;
  const projId = raw?.payload?.project_id || raw?.project_id || log.productId;
  // Category / Feature might not be set, so check categoryId
  const catId = raw?.payload?.category_id || raw?.category_id || (log as any).categoryId;
  
  return allCustomFields.filter((field) => 
    field.is_active &&
    field.is_favorite &&
    field.is_visible &&
    (field.product_id == null || field.product_id === log.productId) &&
    (field.project_id == null || field.project_id === projId) &&
    (field.category_id == null || field.category_id === catId)
  );
};

export function LogsExplorerPage() {
  const setBreadcrumbs = useAppStore((state) => state.setBreadcrumbs);
  const accessToken = useAuthStore((state) => state.accessToken);

  // Filter States
  const [selectedProductId, setSelectedProductId] = useState<number | null>(null);
  const [selectedEnvironmentId, setSelectedEnvironmentId] = useState<number | null>(null);
  const [selectedProjectIds, setSelectedProjectIds] = useState<number[]>([]);
  const [selectedCategoryIds, setSelectedCategoryIds] = useState<number[]>([]);
  const [selectedLevels, setSelectedLevels] = useState<string[]>([]);
  const [selectedCustomFieldPaths, setSelectedCustomFieldPaths] = useState<string[]>([]);
  const [customFieldFilterPath, setCustomFieldFilterPath] = useState<string>();
  const [customFieldFilterValue, setCustomFieldFilterValue] = useState<string>();
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [selectedLog, setSelectedLog] = useState<MainLog | null>(null);

  // States for JSON filtering
  const [jsonFilterTerm, setJsonFilterTerm] = useState('');

  useEffect(() => {
    if (!selectedLog) {
      setJsonFilterTerm('');
    }
  }, [selectedLog]);

  useEffect(() => {
    setBreadcrumbs([
      { title: 'แดชบอร์ด', path: ROUTES.DASHBOARD },
      { title: 'Logs Explorer' },
    ]);
  }, [setBreadcrumbs]);

  // -- Keyword search
  const [rawKeyword, setRawKeyword] = useState('');
  const [logKeyword, setLogKeyword] = useState('');

  useEffect(() => {
    const handler = setTimeout(() => {
      setLogKeyword((rawKeyword || '').trim());
    }, 500); // 500ms debounce delay

    return () => {
      clearTimeout(handler);
    };
  }, [rawKeyword]);

  // ─── Data Queries ─────────────────────────

  const { data: products = [], isLoading: isLoadingProducts } = useQuery({
    queryKey: ['products-admin'],
    queryFn: () => productAdminService.listProducts(),
  });

  const monitorProduct = products.find((p) => p.productId === selectedProductId);
  const monitorEnvironments = monitorProduct?.productEnvironments || [];

  const { data: customFields = [], refetch: refetchCustomFields } = useQuery({
    queryKey: ['explorer-custom-fields', selectedProductId],
    queryFn: () => customFieldService.list(selectedProductId || undefined),
    enabled: !!selectedProductId,
  });
  const scopedCustomFields = customFields.filter((field: CustomField) =>
    field.is_active &&
    field.is_favorite &&
    (field.product_id == null || field.product_id === selectedProductId) &&
    (field.project_id == null || selectedProjectIds.length === 0 || selectedProjectIds.includes(field.project_id)) &&
    (field.category_id == null || selectedCategoryIds.length === 0 || selectedCategoryIds.includes(field.category_id))
  );
  const visibleCustomFields = scopedCustomFields.filter((field) => field.is_visible);
  const filterableCustomFields = scopedCustomFields.filter((field) =>
    field.is_filterable &&
    ![''].includes(field.field_key)
  );

  const addFavoriteFromLog = async (path: string, sampleValue: unknown, detectedType: string) => {
    if (!selectedProductId) return;
    const canonicalPath = normalizeFavoriteFieldPath(path);
    await customFieldService.favorite({
      product_id: selectedProductId,
      field_path: canonicalPath,
      sample_value: sampleValue,
      detected_type: detectedType,
    });
    await refetchCustomFields();
    setSelectedCustomFieldPaths((current) => (
      current.length > 0 && !current.includes(canonicalPath) ? [...current, canonicalPath] : current
    ));
  };

  const favoritePaths = useMemo(() => {
    const paths = new Set<string>();
    for (const field of customFields) {
      if (field.is_favorite && field.is_active) {
        const p = field.field_path || `custom_fields.${field.field_key}`;
        paths.add(normalizeFavoriteFieldPath(p));
      }
    }
    return paths;
  }, [customFields]);

  const toggleFavoriteFromRawJSON = async (path: string, sampleValue: unknown, detectedType: string, isFavorited: boolean) => {
    try {
      if (isFavorited) {
        const canonicalPath = normalizeFavoriteFieldPath(path);
        const field = customFields.find((f) => normalizeFavoriteFieldPath(f.field_path || `custom_fields.${f.field_key}`) === canonicalPath && f.is_favorite);
        if (field) {
          await customFieldService.removeFavorite(field.field_definition_id);
          await refetchCustomFields();
          message.success(`นำ ${canonicalPath} ออกจาก Favorite แล้ว`);
        }
      } else {
        await addFavoriteFromLog(path, sampleValue, detectedType);
        message.success(`เพิ่ม ${path} เป็น Favorite แล้ว`);
      }
    } catch (error: unknown) {
      const apiError = error as { response?: { data?: { error?: string } }; message?: string };
      message.error(apiError.response?.data?.error || apiError.message || 'ดำเนินการไม่สำเร็จ');
    }
  };

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
      selectedCustomFieldPaths,
      customFieldFilterPath,
      customFieldFilterValue,
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
        customFieldPath: customFieldFilterPath,
        customFieldValue: customFieldFilterValue,
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
    setSelectedCustomFieldPaths([]);
    setCustomFieldFilterPath(undefined);
    setCustomFieldFilterValue(undefined);
    setSelectedEnvironmentId(null);
    setCurrentPage(1);
  }, [selectedProductId]);

  useEffect(() => {
    setSelectedCategoryIds([]);
  }, [selectedProjectIds]);

  useEffect(() => {
    setSelectedCustomFieldPaths([]);
    setCustomFieldFilterPath(undefined);
    setCustomFieldFilterValue(undefined);
  }, [selectedProjectIds, selectedCategoryIds]);

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

  const tableColumns = useMemo(() => {
    const cols: any[] = [
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
    ];

    // Add selected custom fields; an empty selection means all visible fields.
    const selectedPaths = new Set(selectedCustomFieldPaths);
    const pathsToRender = visibleCustomFields
      .map((field) => field.field_path || `custom_fields.${field.field_key}`)
      .filter((path) => selectedCustomFieldPaths.length > 0 && selectedPaths.has(path));

    pathsToRender.forEach((path) => {
      const field = visibleCustomFields.find(
        (f) => f.field_path === path || `custom_fields.${f.field_key}` === path
      );
      const displayName = field?.display_name || field?.field_key || path;
      cols.push({
        title: `${displayName} (Custom)`,
        key: path,
        width: 150,
        render: (_: any, record: any) => {
          const val = getCustomFieldValue(record, path);
          return <Text style={{ fontSize: 12 }}>{formatCustomFieldValue(field, val)}</Text>;
        },
      });
    });

    cols.push({
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
    });

    cols.push({
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
    });

    return cols;
  }, [
    selectedCustomFieldPaths,
    visibleCustomFields,
    getProjectName,
    getCategoryName,
  ]);

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



              {filterableCustomFields.map((field) => {
                const path = field.field_path || `custom_fields.${field.field_key}`;
                const isCurrent = customFieldFilterPath === path;
                const currentVal = isCurrent ? customFieldFilterValue : undefined;
                
                return (
                  <Col xs={24} sm={12} md={6} key={field.field_definition_id}>
                    <Form.Item 
                      label={
                        <Space>
                          <span>{field.display_name || field.field_key}</span>
                          {isCurrent && <Tag color="green" style={{ fontSize: '10px', height: 'auto', lineHeight: '14px', padding: '0 4px', margin: 0 }}>กรองอยู่</Tag>}
                        </Space>
                      } 
                      style={{ marginBottom: 0 }}
                    >
                      {field.data_type === 'enum' ? (
                        <Select
                          allowClear
                          placeholder="เลือกค่า..."
                          value={currentVal}
                          onChange={(val) => {
                            if (val === undefined) {
                              setCustomFieldFilterPath(undefined);
                              setCustomFieldFilterValue(undefined);
                            } else {
                              setCustomFieldFilterPath(path);
                              setCustomFieldFilterValue(val);
                            }
                            setCurrentPage(1);
                          }}
                          options={(field.enum_options || []).filter((opt) => opt.is_active).map((opt) => ({
                            value: opt.option_value,
                            label: (
                              <Space>
                                {opt.color_code && (
                                  <span style={{ display: 'inline-block', width: 8, height: 8, borderRadius: '50%', backgroundColor: opt.color_code }} />
                                )}
                                <span>{opt.option_label}</span>
                              </Space>
                            ),
                          }))}
                          style={{ width: '100%' }}
                        />
                      ) : field.data_type === 'boolean' ? (
                        <Select
                          allowClear
                          placeholder="เลือก..."
                          value={currentVal}
                          onChange={(val) => {
                            if (val === undefined) {
                              setCustomFieldFilterPath(undefined);
                              setCustomFieldFilterValue(undefined);
                            } else {
                              setCustomFieldFilterPath(path);
                              setCustomFieldFilterValue(val);
                            }
                            setCurrentPage(1);
                          }}
                          options={[{ value: 'true', label: 'ใช่' }, { value: 'false', label: 'ไม่ใช่' }]}
                          style={{ width: '100%' }}
                        />
                      ) : (
                        <Input
                          allowClear
                          placeholder="กรอกค่า..."
                          value={currentVal}
                          onChange={(e) => {
                            const val = e.target.value;
                            if (!val) {
                              setCustomFieldFilterPath(undefined);
                              setCustomFieldFilterValue(undefined);
                            } else {
                              setCustomFieldFilterPath(path);
                              setCustomFieldFilterValue(val);
                            }
                            setCurrentPage(1);
                          }}
                        />
                      )}
                    </Form.Item>
                  </Col>
                );
              })}

              <Col xs={24} sm={12} md={12}>
                <Form.Item
                  label="Custom Fields ที่แสดงในตาราง"
                  style={{ marginBottom: 0 }}
                >
                  <Select
                    mode="multiple"
                    allowClear
                    maxTagCount="responsive"
                    placeholder={visibleCustomFields.length > 0 ? 'เลือก Custom Fields ที่ต้องการแสดง' : 'ไม่มี Custom Fields ที่เปิดเผย'}
                    value={selectedCustomFieldPaths}
                    onChange={(paths) => {
                      setSelectedCustomFieldPaths(paths);
                      setCurrentPage(1);
                    }}
                    disabled={visibleCustomFields.length === 0}
                    options={visibleCustomFields.map((field) => ({
                      value: field.field_path || `custom_fields.${field.field_key}`,
                      label: field.display_name || field.field_key,
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
                    placeholder="ค้นหา message, path, trace id, request id, error, stack trace..."
                    value={rawKeyword}
                    onChange={(e) => {
                      setRawKeyword(e.target.value);
                      setCurrentPage(1);
                    }}
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
                    disabled={!selectedProductId}
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
                        formatter={(value, name) => [value ?? 0, String(name)]}
                        labelFormatter={(label, payload) => {
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
              columns={tableColumns}
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
        width={800}
        footer={null}
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

            {(() => {
              const visibleFieldsForLog = getVisibleCustomFieldsForLog(selectedLog, customFields)
                .filter((field) => selectedCustomFieldPaths.length === 0 || selectedCustomFieldPaths.includes(field.field_path || `custom_fields.${field.field_key}`));
              if (visibleFieldsForLog.length === 0) return null;

              return (
                <Card size="small" title="Custom Fields" style={{ marginBottom: 16 }}>
                  <Row gutter={[16, 12]}>
                    {visibleFieldsForLog.map((field) => {
                      const value = getCustomFieldValue(selectedLog, field.field_path || `custom_fields.${field.field_key}`);
                      const isSelected = selectedCustomFieldPaths.includes(field.field_path || `custom_fields.${field.field_key}`);
                      if (!isSelected && (value === undefined || value === null)) return null;
                      return (
                        <Col xs={24} sm={12} key={field.field_definition_id}>
                          <Text type="secondary">{field.display_name || field.field_key}</Text>
                          <div><Text strong>{formatCustomFieldValue(field, value)}</Text></div>
                        </Col>
                      );
                    })}
                  </Row>
                </Card>
              );
            })()}

            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '8px', flexWrap: 'wrap', gap: '8px' }}>
              <Space size="small">
                <Title level={5} style={{ margin: 0 }}>Raw JSON Document</Title>
                </Space>
              <Input
                placeholder="กรองคีย์หรือค่าใน JSON..."
                style={{ width: '220px' }}
                value={jsonFilterTerm}
                onChange={(e) => setJsonFilterTerm(e.target.value)}
                allowClear
                size="small"
              />
            </div>
            <div
              style={{
                background: '#1e1e1e',
                padding: '15px',
                borderRadius: '8px',
                overflowX: 'auto',
                whiteSpace: 'pre-wrap',
                fontFamily: 'monospace',
                fontSize: '12px',
                maxHeight: '400px',
              }}
            >
              {(() => {
                const jsonDoc = jsonFilterTerm
                  ? filterJsonObject(selectedLog.raw || selectedLog, jsonFilterTerm)
                  : selectedLog.raw || selectedLog;
                return <JSONFavoriteTree value={jsonDoc} favoritePaths={favoritePaths} onToggleFavorite={(path, value, dataType, isFav) => { void toggleFavoriteFromRawJSON(path, value, dataType, isFav); }} />;
              })()}
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
}
