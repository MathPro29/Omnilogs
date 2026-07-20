import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import type { TableColumnsType } from "antd";
import {
  App,
  Button,
  Card,
  Col,
  DatePicker,
  Descriptions,
  Drawer,
  Empty,
  Flex,
  Input,
  Row,
  Select,
  Skeleton,
  Space,
  Table,
  Tag,
  Typography,
} from "antd";
import {
  AuditOutlined,
  EyeOutlined,
  ReloadOutlined,
  SearchOutlined,
} from "@ant-design/icons";
import dayjs, { type Dayjs } from "dayjs";
import { auditService, type AuditLog, type AuditLogFilters } from "@/services";
import "@/styles/pages/audit-logs.css";

const { Title, Text } = Typography;
const { RangePicker } = DatePicker;

type DateRange = [Dayjs | null, Dayjs | null] | null;

const resultColor: Record<string, string> = {
  SUCCESS: "success",
  FAILED: "error",
  DENIED: "warning",
};

const methodColor: Record<string, string> = {
  GET: "blue",
  POST: "green",
  PUT: "gold",
  PATCH: "purple",
  DELETE: "red",
};

function readableAction(action: string) {
  return action
    .replaceAll("_", " ")
    .toLowerCase()
    .replace(/\b\w/g, (letter) => letter.toUpperCase());
}

function JsonBlock({ value }: { value: unknown }) {
  return (
    <pre className="audit-json">{JSON.stringify(value ?? {}, null, 2)}</pre>
  );
}

export function AuditLogsPage() {
  const { message } = App.useApp();
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [keywordInput, setKeywordInput] = useState("");
  const [keyword, setKeyword] = useState("");
  const [action, setAction] = useState<string>();
  const [resourceType, setResourceType] = useState<string>();
  const [result, setResult] = useState<string>();
  const [dateRange, setDateRange] = useState<DateRange>(null);
  const [selectedAuditId, setSelectedAuditId] = useState<string>();

  const filters = useMemo<AuditLogFilters>(
    () => ({
      page,
      per_page: pageSize,
      keyword: keyword || undefined,
      action,
      resource_type: resourceType,
      result,
      date_from: dateRange?.[0]?.toISOString(),
      date_to: dateRange?.[1]?.toISOString(),
    }),
    [page, pageSize, keyword, action, resourceType, result, dateRange],
  );

  const audits = useQuery({
    queryKey: ["audit-logs", filters],
    queryFn: () => auditService.list(filters),
  });

  const detail = useQuery({
    queryKey: ["audit-log", selectedAuditId],
    queryFn: () => auditService.getById(selectedAuditId!),
    enabled: Boolean(selectedAuditId),
  });

  const resetFilters = () => {
    setKeywordInput("");
    setKeyword("");
    setAction(undefined);
    setResourceType(undefined);
    setResult(undefined);
    setDateRange(null);
    setPage(1);
  };

  const columns: TableColumnsType<AuditLog> = [
    {
      title: "เวลา",
      dataIndex: "created_at",
      width: 178,
      render: (value?: string) => (
        <div>
          <Text strong>{value ? dayjs(value).format("DD MMM YYYY") : "—"}</Text>
          <br />
          <Text type="secondary" className="audit-time">
            {value ? dayjs(value).format("HH:mm:ss") : ""}
          </Text>
        </div>
      ),
    },
    {
      title: "Activity",
      dataIndex: "action",
      width: 230,
      render: (value: string, record) => (
        <Space size={10}>
          <span className="audit-activity-icon">
            <AuditOutlined />
          </span>
          <div>
            <Text strong>{readableAction(value)}</Text>
            <br />
            <Text type="secondary" className="audit-resource">
              {record.resource_type}
              {record.resource_id ? ` · ${record.resource_id}` : ""}
            </Text>
          </div>
        </Space>
      ),
    },
    {
      title: "Actor",
      dataIndex: "actor_user_id",
      width: 120,
      render: (value?: number) =>
        value ? (
          <Text>User #{value}</Text>
        ) : (
          <Text type="secondary">System</Text>
        ),
    },
    {
      title: "Request",
      key: "request",
      ellipsis: true,
      render: (_, record) => (
        <Space size={8}>
          {record.method && (
            <Tag
              color={methodColor[record.method] || "default"}
              bordered={false}
            >
              {record.method}
            </Tag>
          )}
          <Text code className="audit-path">
            {record.path || "—"}
          </Text>
        </Space>
      ),
    },
    {
      title: "Result",
      dataIndex: "result",
      width: 112,
      render: (value: string) => (
        <Tag color={resultColor[value] || "default"} bordered={false}>
          {value}
        </Tag>
      ),
    },
    {
      title: "",
      key: "detail",
      width: 54,
      fixed: "right",
      render: (_, record) => (
        <Button
          type="text"
          shape="circle"
          aria-label="ดูรายละเอียด"
          icon={<EyeOutlined />}
          onClick={() => setSelectedAuditId(record.audit_id)}
        />
      ),
    },
  ];

  return (
    <div className="audit-page">
      <Flex justify="space-between" align="flex-start" gap={16} wrap>
        <div>
          <Space size={10}>
            <Title level={2} className="audit-title">
              Audit Logs
            </Title>
          </Space>
          <br />
        </div>
        <Button
          icon={<ReloadOutlined />}
          onClick={() => audits.refetch()}
          loading={audits.isFetching}
        >
          รีเฟรช
        </Button>
      </Flex>

      <Card className="audit-filter-card" bordered={false}>
        <Row gutter={[12, 12]}>
          <Col xs={24} md={12} lg={6}>
            <Input
              allowClear
              prefix={<SearchOutlined />}
              placeholder="ค้นหา Activity, Path, Resource ID..."
              value={keywordInput}
              onChange={(e) => setKeywordInput(e.target.value)}
              onPressEnter={() => {
                setKeyword(keywordInput.trim());
                setPage(1);
              }}
            />
          </Col>
          <Col xs={12} sm={8} md={6} lg={3}>
            <Select
              allowClear
              showSearch
              placeholder="Activity"
              value={action}
              onChange={(value) => {
                setAction(value);
                setPage(1);
              }}
              options={[
                "LOGIN",
                "LOGOUT",
                "CREATE",
                "UPDATE",
                "DELETE",
                "SEARCH_LOGS",
                "VIEW_LOG_DETAIL",
                "EXPORT_LOGS",
              ].map((value) => ({ value, label: readableAction(value) }))}
            />
          </Col>
          <Col xs={12} sm={8} md={6} lg={3}>
            <Select
              allowClear
              placeholder="Result"
              value={result}
              onChange={(value) => {
                setResult(value);
                setPage(1);
              }}
              options={["SUCCESS", "FAILED", "DENIED"].map((value) => ({
                value,
                label: value,
              }))}
            />
          </Col>
          <Col xs={24} sm={8} md={6} lg={4}>
            <Select
              allowClear
              showSearch
              placeholder="Resource"
              value={resourceType}
              onChange={(value) => {
                setResourceType(value);
                setPage(1);
              }}
              options={[
                "AUTH",
                "USER",
                "PRODUCT",
                "PROJECT",
                "MAIN_LOG",
                "POLICY",
                "ROLE",
                "MEMBERSHIP",
              ].map((value) => ({ value, label: readableAction(value) }))}
            />
          </Col>
          <Col xs={24} md={12} lg={8}>
            <RangePicker
              showTime={{
                defaultValue: [dayjs("00:00:00", "HH:mm:ss"), dayjs("23:59:59", "HH:mm:ss")],
              }}
              format="YYYY-MM-DD HH:mm:ss"
              value={dateRange}
              onChange={(value) => {
                setDateRange(value);
                setPage(1);
              }}
              style={{ width: "100%" }}
            />
          </Col>
        </Row>
        <Flex justify="flex-end" gap={8} className="audit-filter-actions">
          <Button type="text" onClick={resetFilters}>
            ล้างตัวกรอง
          </Button>
          <Button
            type="primary"
            icon={<SearchOutlined />}
            onClick={() => {
              setKeyword(keywordInput.trim());
              setPage(1);
            }}
          >
            ค้นหา
          </Button>
        </Flex>
      </Card>

      <Card className="audit-table-card" bordered={false}>
        <Table<AuditLog>
          className="audit-table"
          rowKey="audit_id"
          columns={columns}
          dataSource={audits.data?.data || []}
          loading={audits.isLoading}
          scroll={{ x: 980 }}
          locale={{
            emptyText: audits.isError ? (
              <Empty description="ไม่สามารถโหลด Audit Logs ได้">
                <Button onClick={() => audits.refetch()}>ลองอีกครั้ง</Button>
              </Empty>
            ) : (
              <Empty description="ไม่พบ Activity" />
            ),
          }}
          pagination={{
            current: page,
            pageSize,
            total: audits.data?.total || 0,
            showSizeChanger: true,
            pageSizeOptions: [10, 20, 50, 100],
            showTotal: (total, range) =>
              `${range[0]}–${range[1]} จาก ${total} รายการ`,
            onChange: (nextPage, nextSize) => {
              setPage(nextSize !== pageSize ? 1 : nextPage);
              setPageSize(nextSize);
            },
          }}
          onRow={(record) => ({
            onDoubleClick: () => setSelectedAuditId(record.audit_id),
          })}
        />
      </Card>

      <Drawer
        title={
          <Space>
            <AuditOutlined />
            รายละเอียด Activity
          </Space>
        }
        open={Boolean(selectedAuditId)}
        onClose={() => setSelectedAuditId(undefined)}
        width={640}
        destroyOnHidden
      >
        {detail.isLoading ? (
          <Skeleton active paragraph={{ rows: 10 }} />
        ) : detail.isError ? (
          <Empty description="ไม่สามารถโหลดรายละเอียดได้">
            <Button
              onClick={() => {
                detail.refetch();
                message.info("กำลังโหลดข้อมูลอีกครั้ง");
              }}
            >
              ลองอีกครั้ง
            </Button>
          </Empty>
        ) : (
          detail.data && (
            <Space direction="vertical" size={24} style={{ width: "100%" }}>
              <div>
                <Text type="secondary">Activity</Text>
                <Title level={4} style={{ margin: "4px 0 8px" }}>
                  {readableAction(detail.data.action)}
                </Title>
                <Tag
                  color={resultColor[detail.data.result] || "default"}
                  bordered={false}
                >
                  {detail.data.result}
                </Tag>
              </div>
              <Descriptions
                column={1}
                size="small"
                bordered
                items={[
                  {
                    key: "audit",
                    label: "Audit ID",
                    children: (
                      <Text copyable code>
                        {detail.data.audit_id}
                      </Text>
                    ),
                  },
                  {
                    key: "time",
                    label: "เวลา",
                    children: detail.data.created_at
                      ? dayjs(detail.data.created_at).format(
                          "DD MMM YYYY, HH:mm:ss",
                        )
                      : "—",
                  },
                  {
                    key: "actor",
                    label: "Actor",
                    children: detail.data.actor_user_id
                      ? `User #${detail.data.actor_user_id}`
                      : "System",
                  },
                  {
                    key: "resource",
                    label: "Resource",
                    children: `${detail.data.resource_type}${detail.data.resource_id ? ` · ${detail.data.resource_id}` : ""}`,
                  },
                  {
                    key: "request",
                    label: "Request",
                    children: (
                      <Space>
                        {detail.data.method && (
                          <Tag color={methodColor[detail.data.method]}>
                            {detail.data.method}
                          </Tag>
                        )}
                        <Text code>{detail.data.path || "—"}</Text>
                      </Space>
                    ),
                  },
                  {
                    key: "ip",
                    label: "IP Address",
                    children: detail.data.ip_address || "—",
                  },
                  {
                    key: "requestId",
                    label: "Request ID",
                    children: detail.data.request_id ? (
                      <Text copyable>{detail.data.request_id}</Text>
                    ) : (
                      "—"
                    ),
                  },
                  {
                    key: "traceId",
                    label: "Trace ID",
                    children: detail.data.trace_id ? (
                      <Text copyable>{detail.data.trace_id}</Text>
                    ) : (
                      "—"
                    ),
                  },
                ]}
              />
              <div>
                <Title level={5}>Metadata</Title>
                <JsonBlock value={detail.data.metadata} />
              </div>
              {detail.data.user_agent && (
                <div>
                  <Title level={5}>User Agent</Title>
                  <Text type="secondary">{detail.data.user_agent}</Text>
                </div>
              )}
            </Space>
          )
        )}
      </Drawer>
    </div>
  );
}
