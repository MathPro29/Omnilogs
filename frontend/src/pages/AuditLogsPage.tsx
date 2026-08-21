import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import type { TableColumnsType } from "antd";
import { Button, Flex, Space, Tag, Typography } from "antd";
import { AuditOutlined, EyeOutlined } from "@ant-design/icons";
import dayjs, { type Dayjs } from "dayjs";
import {
  userService,
  productService,
} from "@/services";
import { auditService, type AuditLog, type AuditLogFilters } from "@/features/audit-logs/services/audit.service";
import { useAuthStore } from "@/store";
import { buildUserDisplayMap } from "@/utils/userDisplay";
import "@/styles/pages/audit-logs.css";
import RefreshButton from "@/components/features/audit-logs/RefreshButton";
import AuditLogFiltersCard from "@/components/features/audit-logs/AuditLogFiltersCard";
import AuditLogTableCard from "@/components/features/audit-logs/AuditLogTableCard";
import { AuditLogDetailDrawer } from "@/components/features";

const { Title, Text } = Typography;

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

export function AuditLogsPage() {
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [keywordInput, setKeywordInput] = useState("");
  const [keyword, setKeyword] = useState("");
  const [productId, setProductId] = useState<number>();
  const [action, setAction] = useState<string>();
  const [resourceType, setResourceType] = useState<string>();
  const [result, setResult] = useState<string>();
  const [dateRange, setDateRange] = useState<DateRange>(null);
  const [selectedAuditId, setSelectedAuditId] = useState<string>();

  const currentUser = useAuthStore((state) => state.currentUser);

  const filters = useMemo<AuditLogFilters>(
    () => ({
      page,
      per_page: pageSize,
      keyword: keyword || undefined,
      product_id: productId,
      action,
      resource_type: resourceType,
      result,
      date_from: dateRange?.[0]?.toISOString(),
      date_to: dateRange?.[1]?.toISOString(),
    }),
    [
      page,
      pageSize,
      keyword,
      productId,
      action,
      resourceType,
      result,
      dateRange,
    ],
  );

  const audits = useQuery({
    queryKey: ["audit-logs", filters],
    queryFn: () => auditService.list(filters),
  });

  const productsQuery = useQuery({
    queryKey: ["products", "audit-filter"],
    queryFn: productService.listOptions,
  });

  const activitiesQuery = useQuery({
    queryKey: ["audit-activities", productId],
    queryFn: () => auditService.listActivities(productId),
  });

  const usersQuery = useQuery({
    queryKey: ["users", "list-all"],
    queryFn: () => userService.getUsers({ page: 1, pageSize: 100 }),
  });

  const userMap = useMemo(
    () => buildUserDisplayMap(currentUser, usersQuery.data),
    [currentUser, usersQuery.data],
  );

  const resetFilters = () => {
    setKeywordInput("");
    setKeyword("");
    setProductId(undefined);
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
            <Text
              strong
              onClick={() => setSelectedAuditId(record.audit_id)}
              style={{ cursor: "pointer" }}
            >
              {readableAction(value)}
            </Text>
            <br />
            <Text type="secondary" className="audit-resource">
              {record.resource_type}
              {record.resource_id ? ` · ${record.resource_id}` : ""}
              {record.product_id ? ` · Product #${record.product_id}` : ""}
            </Text>
          </div>
        </Space>
      ),
    },
    {
      title: "Actor",
      key: "actor",
      width: 160,
      render: (_, record) => {
        const actorId =
          record.actor_user_id !== undefined
            ? Number(record.actor_user_id)
            : undefined;
        const mappedName = actorId !== undefined ? userMap[actorId] : undefined;
        const actorName =
          record.actor_name ||
          mappedName ||
          record.actor_email ||
          (actorId ? `User #${actorId}` : undefined);
        return actorName ? (
          <Text>{actorName}</Text>
        ) : (
          <Text type="secondary">System</Text>
        );
      },
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
              Audit Logs - การกระทำของผู้ใช้
            </Title>
          </Space>
          <br />
        </div>
        <RefreshButton filters={filters} />
      </Flex>

      <AuditLogFiltersCard
        keywordInput={keywordInput}
        setKeywordInput={setKeywordInput}
        setKeyword={setKeyword}
        products={productsQuery.data ?? []}
        productId={productId}
        setProductId={setProductId}
        activityOptions={activitiesQuery.data ?? []}
        action={action}
        setAction={setAction}
        resourceType={resourceType}
        setResourceType={setResourceType}
        result={result}
        setResult={setResult}
        dateRange={dateRange}
        setDateRange={setDateRange}
        setPage={setPage}
        resetFilters={resetFilters}
      />

      <AuditLogTableCard
        columns={columns}
        audits={audits}
        page={page}
        pageSize={pageSize}
        setPage={setPage}
        setPageSize={setPageSize}
        setSelectedAuditId={setSelectedAuditId}
      />

      <AuditLogDetailDrawer
        selectedAuditId={selectedAuditId}
        onClose={() => setSelectedAuditId(undefined)}
      />
    </div>
  );
}
