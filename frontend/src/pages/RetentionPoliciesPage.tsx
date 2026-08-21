import React, { useEffect, useState } from "react";
import {
  Input,
  Select,
  Table,
  Tag,
  Switch,
  Button,
  Popconfirm,
  Tooltip,
  Space,
  Drawer,
  Form,
  Radio,
  InputNumber,
  Modal,
  Empty,
  Card,
  Typography,
  Alert,
  DatePicker,
  TimePicker,
  Descriptions,
  message,
} from "antd";
import {
  PlusOutlined,
  SearchOutlined,
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
  CheckCircleOutlined,
  FolderOutlined,
  DatabaseOutlined,
  InfoCircleOutlined,
  CalendarOutlined,
  DeleteFilled,
  CloudOutlined,
  CompressOutlined,
  InboxOutlined,
  RollbackOutlined,
  EyeOutlined,
  DownloadOutlined,
} from "@ant-design/icons";
import { AnimatePresence, motion } from "motion/react";
import { PageTransition } from "@/components";
import { useRetentionPolicies } from "@/features/retention-policies/hooks/useRetentionPolicies";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { productService } from "@/services/product.service";
import { retentionPolicyService } from "@/features/retention-policies/services/retentionPolicy.service";
import type {
  RetentionPolicy,
  RetentionMode,
  CreateRetentionPolicyPayload,
  RetentionUnit,
  SimulationResult,
} from "@/features/retention-policies/types/retentionPolicy.types";
import type { ColumnsType } from "antd/es/table";
import dayjs, { type Dayjs } from "dayjs";
import "@/styles/pages/retention-policies.css";

const { Title } = Typography;

/* ─── Thai Labels ─── */
const MODE_LABEL: Record<string, string> = {
  WEEKLY: "รายสัปดาห์",
  MONTHLY: "รายเดือน",
  DAILY: "รายวัน",
  CUSTOM: "กำหนดเอง",
};
const MODE_COLOR: Record<string, string> = {
  WEEKLY: "blue",
  MONTHLY: "purple",
  DAILY: "cyan",
  CUSTOM: "orange",
};
const ACTION_LABEL: Record<string, string> = {
  DELETE: "ลบถาวร",
  ARCHIVE_COLD: "จัดเก็บ",
  COMPRESS_GZIP: "บีบอัด",
};
const ACTION_STYLE: Record<string, string> = {
  DELETE: "delete",
  ARCHIVE_COLD: "archive",
  COMPRESS_GZIP: "compress",
};
const ACTION_ICON: Record<string, React.ReactNode> = {
  DELETE: <DeleteFilled />,
  ARCHIVE_COLD: <CloudOutlined />,
  COMPRESS_GZIP: <CompressOutlined />,
};

function folderSummary(mode: string, val: number, totalDays: number): string {
  if (totalDays <= 0) return `${val} วัน`;
  if (mode === "WEEKLY") return `${totalDays} วัน / ${val} สัปดาห์`;
  if (mode === "MONTHLY") return `${totalDays} วัน / ${val} เดือน`;
  if (mode === "DAILY") return `${totalDays} วัน`;
  return `${totalDays} วัน`;
}

/* ─── Main Page ─── */
export const RetentionPoliciesPage: React.FC = () => {
  const productsQuery = useQuery({
    queryKey: ["products", "retention-policy-scope"],
    queryFn: productService.listOptions,
  });
  const [productId, setProductId] = useState<number>();
  const [environmentId, setEnvironmentId] = useState<number>();

  useEffect(() => {
    if (!productsQuery.data?.length) return;
    if (
      !productId ||
      !productsQuery.data.some((product) => product.id === productId)
    ) {
      setProductId(productsQuery.data[0].id);
      setEnvironmentId(undefined);
    }
  }, [productId, productsQuery.data]);

  const environmentsQuery = useQuery({
    queryKey: ["products", productId, "environments"],
    queryFn: () => productService.listEnvironments(productId!),
    enabled: Boolean(productId),
  });

  useEffect(() => {
    const environments = environmentsQuery.data ?? [];
    if (!environments.length) {
      setEnvironmentId(undefined);
      return;
    }
    if (
      !environmentId ||
      !environments.some(
        (environment) => environment.environment_id === environmentId,
      )
    ) {
      setEnvironmentId(environments[0].environment_id);
    }
  }, [environmentId, environmentsQuery.data]);

  const {
    policies,
    isLoading,
    isError,
    error,
    refetch,
    stats,
    isStatsLoading,
    createPolicy,
    isCreating,
    updatePolicy,
    isUpdating,
    deletePolicy,
    togglePolicy,
    triggerPolicyNow,
    isTriggering,
    simulationData,
  } = useRetentionPolicies(productId, environmentId);

  const [drawerOpen, setDrawerOpen] = useState(false);
  const [editing, setEditing] = useState<RetentionPolicy | null>(null);
  const [backfillPolicy, setBackfillPolicy] = useState<RetentionPolicy | null>(
    null,
  );
  const [inspectPolicy, setInspectPolicy] = useState<RetentionPolicy | null>(
    null,
  );
  const [rollbackOpen, setRollbackOpen] = useState(false);
  const [search, setSearch] = useState("");
  const [modeFilter, setModeFilter] = useState("ALL");

  const filtered = policies.filter((p) => {
    const q = search.toLowerCase();
    const matchSearch =
      !q ||
      p.name.toLowerCase().includes(q) ||
      p.description.toLowerCase().includes(q);
    const matchMode = modeFilter === "ALL" || p.retention_mode === modeFilter;
    return matchSearch && matchMode;
  });

  const openCreate = () => {
    setEditing(null);
    setDrawerOpen(true);
  };
  const openEdit = (p: RetentionPolicy) => {
    setEditing(p);
    setDrawerOpen(true);
  };

  const handleSave = async (payload: CreateRetentionPolicyPayload) => {
    if (editing) await updatePolicy({ policyId: editing.policy_id, payload });
    else await createPolicy(payload);
  };

  const activeCount = policies.filter((p) => p.is_active).length;
  const scopeLoading = productsQuery.isLoading || environmentsQuery.isLoading;

  /* ─── Columns ─── */
  const columns: ColumnsType<RetentionPolicy> = [
    {
      title: "นโยบาย",
      dataIndex: "name",
      key: "name",
      render: (name, r) => (
        <div className="retention-policy-name">
          <strong>{name}</strong>
          {r.description && (
            <span className="retention-policy-desc">{r.description}</span>
          )}
        </div>
      ),
    },
    {
      title: "โหมด",
      dataIndex: "retention_mode",
      key: "mode",
      width: 130,
      render: (m: RetentionMode) => (
        <Tag color={MODE_COLOR[m]} className="retention-mode-tag">
          {MODE_LABEL[m]}
        </Tag>
      ),
    },
    {
      title: "ระยะจัดเก็บรักษา",
      key: "retention",
      width: 200,
      render: (_, r) => (
        <div className="retention-summary">
          <CalendarOutlined />
          <span>
            {folderSummary(r.retention_mode, r.retention_value, r.total_days)}
          </span>
        </div>
      ),
    },
    {
      title: "รอบตรวจ/หมดอายุถัดไป",
      key: "next_purge_at",
      render: (_, r) => (
        <Tooltip
          title={
            r.next_purge_at
              ? `${dayjs(r.next_purge_at).format("YYYY-MM-DD HH:mm:ss")} · ${scheduleLabel(r.cron_schedule, r.schedule_timezone)}`
              : "ยังไม่กำหนด"
          }
        >
          <div>
            <span>{expiresInText(r.next_purge_at)}</span>
            <div className="text-xs text-slate-500">
              {scheduleLabel(r.cron_schedule, r.schedule_timezone)}
            </div>
          </div>
        </Tooltip>
      ),
    },
    {
      title: "เมื่อหมดอายุ",
      dataIndex: "auto_purge_action",
      key: "action",
      width: 130,
      render: (a: string) => (
        <span className={`retention-action-badge ${ACTION_STYLE[a] || ""}`}>
          {ACTION_ICON[a]} {ACTION_LABEL[a] || a}
        </span>
      ),
    },
    {
      title: "สถานะ",
      dataIndex: "is_active",
      key: "is_active",
      width: 100,
      align: "center",
      render: (v: boolean, r: RetentionPolicy) => (
        <Switch
          size="small"
          checked={v}
          onChange={(c) => togglePolicy({ policyId: r.policy_id, isActive: c })}
        />
      ),
    },
    {
      title: "จัดการ",
      key: "actions",
      width: 370,
      align: "right",
      render: (_, r) => (
        <Space size={4}>
          <Tooltip title="ดูจำนวน Logs และพื้นที่ของ Policy">
            <Button
              type="text"
              size="small"
              icon={<EyeOutlined />}
              disabled={!r.is_active || !environmentId}
              onClick={() => setInspectPolicy(r)}
            >
              ดู Logs
            </Button>
          </Tooltip>
          <Tooltip
            title={
              r.archive_enabled && r.is_active
                ? "เลือกช่วงเวลาและ TAG เพื่อสำรองข้อมูลแบบ Manual ไปยัง Storage เดียวกับ Policy"
                : "ต้องเปิดใช้งาน Policy และ Archive ก่อน"
            }
          >
            <Button
              type="text"
              size="small"
              icon={<InboxOutlined />}
              disabled={!r.is_active || !r.archive_enabled}
              onClick={() => setBackfillPolicy(r)}
            >
              Manual Backup
            </Button>
          </Tooltip>
          <Tooltip
            title={
              r.archive_enabled
                ? "หยุดการสร้าง Archive ต่อเนื่อง"
                : "เริ่มการสร้าง Archive ต่อเนื่อง"
            }
          >
            <Button
              type="text"
              size="small"
              icon={r.archive_enabled ? <CloudOutlined /> : <InboxOutlined />}
              disabled={!r.is_active}
              onClick={() =>
                updatePolicy({
                  policyId: r.policy_id,
                  payload: { archive_enabled: !r.archive_enabled },
                })
              }
              loading={isUpdating}
            >
              {r.archive_enabled ? "หยุด Archive" : "เริ่ม Archive"}
            </Button>
          </Tooltip>
          <Tooltip title="แก้ไข">
            <Button
              type="text"
              size="small"
              icon={<EditOutlined />}
              onClick={() => openEdit(r)}
            />
          </Tooltip>
          <Popconfirm
            title="ทดสอบ Retention Worker ตอนนี้?"
            description="ระบบจะส่ง Policy นี้เข้าคิว Retention จริง และอาจสร้าง Backup/Verify/Purge ตามค่าที่ตั้งไว้"
            okText="เริ่มทดสอบ"
            cancelText="ยกเลิก"
            onConfirm={() => triggerPolicyNow(r.policy_id)}
          >
            <Tooltip title="ทดสอบ Retention Worker">
              <Button
                type="text"
                size="small"
                icon={<ReloadOutlined />}
                disabled={!r.is_active || !r.archive_enabled}
                loading={isTriggering}
                aria-label="ทดสอบ Retention Worker"
              >
                ทดสอบ
              </Button>
            </Tooltip>
          </Popconfirm>
          <Popconfirm
            title="ลบนโยบายนี้?"
            description="การดำเนินการนี้ไม่สามารถย้อนกลับได้"
            onConfirm={() => deletePolicy(r.policy_id)}
            okText="ลบ"
            cancelText="ยกเลิก"
            okButtonProps={{ danger: true }}
          >
            <Tooltip title="ลบ">
              <Button
                type="text"
                size="small"
                danger
                icon={<DeleteOutlined />}
              />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <PageTransition>
      <div className="retention-page">
        {/* Header */}
        <div className="retention-page-header">
          <div>
            <Title level={2} className="retention-title">
              Retention Policies — การเก็บรักษาข้อมูล
            </Title>
          </div>
          <Space>
            <Select
              loading={productsQuery.isLoading}
              value={productId}
              placeholder="เลือก Product"
              style={{ minWidth: 180 }}
              options={(productsQuery.data ?? []).map((product) => ({
                label: product.name,
                value: product.id,
              }))}
              onChange={(value: number) => {
                setProductId(value);
                setEnvironmentId(undefined);
              }}
            />
            <Select
              loading={environmentsQuery.isLoading}
              disabled={!productId || !environmentsQuery.data?.length}
              value={environmentId}
              placeholder="เลือก Environment"
              style={{ minWidth: 180 }}
              options={(environmentsQuery.data ?? []).map((environment) => ({
                label: `${environment.environment_name} (${environment.environment_code})`,
                value: environment.environment_id,
              }))}
              onChange={setEnvironmentId}
            />
            <Button
              icon={<ReloadOutlined spin={isLoading} />}
              onClick={() => refetch()}
              disabled={isLoading || !environmentId}
            >
              รีเฟรช
            </Button>
            <Button
              icon={<RollbackOutlined />}
              onClick={() => setRollbackOpen(true)}
              disabled={!productId || !environmentId}
            >
              Rollback
            </Button>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={openCreate}
              disabled={!productId || !environmentId}
              className="bg-[#1F8457] hover:!bg-[#ffffff]"
            >
              สร้าง Policy ใหม่
            </Button>
          </Space>
        </div>

        {!scopeLoading && !productsQuery.data?.length && (
          <Alert
            type="info"
            showIcon
            message="ยังไม่มี Product ที่เข้าถึงได้"
            description="Retention Policy จะแสดงเมื่อบัญชีนี้มีสิทธิ์เข้าถึง Product อย่างน้อยหนึ่งรายการ"
          />
        )}
        {!scopeLoading && productId && !environmentsQuery.data?.length && (
          <Alert
            type="info"
            showIcon
            message="Product นี้ยังไม่มี Environment"
            description="กรุณาสร้าง Environment ก่อนตั้งค่า Retention Policy"
          />
        )}
        {isError && (
          <Alert
            type="error"
            showIcon
            message="โหลด Retention Policy ไม่สำเร็จ"
            description={
              error instanceof Error ? error.message : "กรุณาลองใหม่อีกครั้ง"
            }
            action={<Button onClick={() => void refetch()}>ลองใหม่</Button>}
          />
        )}

        {/* Stats Cards */}
        <div className="retention-stats">
          <Card className="retention-stat-card" loading={isStatsLoading}>
            <div className="retention-stat-icon green">
              <FolderOutlined />
            </div>
            <div className="retention-stat-value">
              {stats?.total_policies ?? policies.length}
            </div>
            <div className="retention-stat-label">นโยบายทั้งหมด</div>
          </Card>
          <Card className="retention-stat-card" loading={isStatsLoading}>
            <div className="retention-stat-icon blue">
              <CheckCircleOutlined />
            </div>
            <div className="retention-stat-value">
              {stats?.active_policies ?? activeCount}
            </div>
            <div className="retention-stat-label">เปิดใช้งานอยู่</div>
          </Card>
          <Card className="retention-stat-card" loading={isStatsLoading}>
            <div className="retention-stat-icon orange">
              <DatabaseOutlined />
            </div>
            <div className="retention-stat-value">
              {formatBytes(
                stats?.total_storage_bytes ??
                  (stats?.total_storage_gb ?? 0) * 1024 * 1024 * 1024,
              )}
            </div>
            <Tooltip
              title={
                stats
                  ? `Active Elasticsearch ${formatBytes(stats.active_storage_bytes)} • Archive GZIP ${formatBytes(stats.archive_gzip_storage_bytes)} • Ready ZIP ${formatBytes(stats.ready_zip_storage_bytes)}`
                  : undefined
              }
            >
              <div className="retention-stat-label">พื้นที่ใช้งานจริง</div>
            </Tooltip>
          </Card>
          <Card className="retention-stat-card" loading={isStatsLoading}>
            <div className="retention-stat-icon purple">
              <CalendarOutlined />
            </div>
            <div className="retention-stat-value">
              {stats?.scheduled_purges_24h ?? 0}
            </div>
            <div className="retention-stat-label">รอดำเนินการ (24 ชม.)</div>
          </Card>
        </div>

        {/* Filters */}
        <Card className="retention-filter-card">
          <Input
            placeholder="ค้นหานโยบาย..."
            prefix={<SearchOutlined style={{ color: "#bfbfbf" }} />}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            allowClear
          />
          <Select
            value={modeFilter}
            onChange={setModeFilter}
            style={{ width: 160 }}
            options={[
              { label: "ทุกโหมด", value: "ALL" },
              { label: "รายสัปดาห์", value: "WEEKLY" },
              { label: "รายเดือน", value: "MONTHLY" },
              { label: "รายวัน", value: "DAILY" },
              { label: "กำหนดเอง", value: "CUSTOM" },
            ]}
          />
          {(search || modeFilter !== "ALL") && (
            <Button
              type="link"
              onClick={() => {
                setSearch("");
                setModeFilter("ALL");
              }}
            >
              ล้างตัวกรอง
            </Button>
          )}
        </Card>

        {/* Table */}
        <Card className="retention-table-card">
          <Table
            className="retention-table"
            columns={columns}
            dataSource={filtered}
            rowKey="policy_id"
            loading={isLoading}
            pagination={
              filtered.length > 10
                ? {
                    pageSize: 10,
                    showSizeChanger: true,
                    showTotal: (total) => `ทั้งหมด ${total} รายการ`,
                  }
                : false
            }
            size="middle"
            locale={{
              emptyText: (
                <Empty
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                  description="ยังไม่มีนโยบายการเก็บรักษาข้อมูล"
                />
              ),
            }}
          />
        </Card>

        {/* Drawer */}
        <PolicyDrawer
          open={drawerOpen}
          onClose={() => setDrawerOpen(false)}
          onSave={handleSave}
          initial={editing}
          productId={productId!}
          environmentId={environmentId!}
          loading={isCreating || isUpdating}
        />

        {/* Simulate Modal */}
        <SimulateModal result={simulationData ?? null} />
        <BackfillLogsModal
          open={Boolean(backfillPolicy)}
          policy={backfillPolicy}
          productId={productId}
          environmentId={environmentId}
          onClose={() => setBackfillPolicy(null)}
        />
        <PolicyLogsModal
          open={Boolean(inspectPolicy)}
          policy={inspectPolicy}
          productId={productId}
          environmentId={environmentId}
          onClose={() => setInspectPolicy(null)}
        />
        <RollbackArchiveModal
          open={rollbackOpen}
          productId={productId}
          environmentId={environmentId}
          onClose={() => setRollbackOpen(false)}
        />
      </div>
    </PageTransition>
  );
};
/* ──────────────────────────────────────────
/* โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€โ”€
   Policy Drawer (Create / Edit) — Thai
   ────────────────────────────────────────── */
type ArchiveDateRange = [Dayjs | null, Dayjs | null] | null;

function archiveDate(value: Dayjs): string {
  return value.startOf("day").format("YYYY-MM-DDTHH:mm:ssZ");
}

function instantAtStartOfDay(value: Dayjs): string {
  return value.startOf("day").format("YYYY-MM-DDTHH:mm:ssZ");
}

function scheduleTime(cron?: string): Dayjs {
  const [minute = "0", hour = "1"] = (cron ?? "0 1 * * *").trim().split(/\s+/);
  return dayjs()
    .hour(Number(hour) || 0)
    .minute(Number(minute) || 0)
    .second(0);
}

function scheduleLabel(cron?: string, timezone?: string): string {
  const value = scheduleTime(cron);
  return `${value.format("HH:mm")} (${timezone || "Asia/Bangkok"})`;
}

function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const unitIndex = Math.min(
    Math.floor(Math.log(bytes) / Math.log(1024)),
    units.length - 1,
  );
  const value = bytes / 1024 ** unitIndex;
  return `${value.toFixed(value >= 100 || unitIndex === 0 ? 0 : value >= 10 ? 1 : 2)} ${units[unitIndex]}`;
}

function expiresInText(value?: string): string {
  if (!value) return "ยังไม่กำหนด";
  const target = dayjs(value);
  const minutes = target.diff(dayjs(), "minute");
  if (minutes <= 0) return "ถึงกำหนดแล้ว";
  if (minutes < 60) return `อีก ${minutes} นาที`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `อีก ${hours} ชม.`;
  return `อีก ${Math.floor(hours / 24)} วัน`;
}

function errorText(error: unknown): string {
  if (error && typeof error === "object") {
    const normalized = error as {
      message?: unknown;
      data?: { error?: unknown; message?: unknown };
    };
    if (typeof normalized.data?.error === "string") {
      return normalized.data.error;
    }
    if (typeof normalized.data?.message === "string") {
      return normalized.data.message;
    }
    if (typeof normalized.message === "string") {
      return normalized.message;
    }
  }
  return error instanceof Error
    ? error.message
    : "เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง";
}

function policyDefaultRange(policy: RetentionPolicy): ArchiveDateRange {
  const value = policy.active_retention_value ?? policy.retention_value ?? 30;
  const rawUnit = policy.active_retention_unit ?? policy.retention_unit;
  const unit: "day" | "week" | "month" | "year" =
    rawUnit === "WEEK" || rawUnit === "WEEKS"
      ? "week"
      : rawUnit === "MONTH" || rawUnit === "MONTHS"
        ? "month"
        : rawUnit === "YEAR" || rawUnit === "YEARS"
          ? "year"
          : "day";
  return [dayjs().subtract(value, unit).startOf("day"), dayjs().startOf("day")];
}

function PolicyLogsModal({
  open,
  policy,
  productId,
  environmentId,
  onClose,
}: {
  open: boolean;
  policy: RetentionPolicy | null;
  productId?: number;
  environmentId?: number;
  onClose: () => void;
}) {
  const queryClient = useQueryClient();
  const [range, setRange] = useState<ArchiveDateRange>(null);
  const [importFile, setImportFile] = useState<File>();
  const [snapshotArchiveId, setSnapshotArchiveId] = useState<string>();
  const readyDownloadsQuery = useQuery({
    queryKey: ["ready-archive-downloads", productId, environmentId],
    queryFn: () =>
      retentionPolicyService.listReadyDownloads(productId!, environmentId!),
    enabled: open && Boolean(productId && environmentId),
    refetchInterval: open ? 30_000 : false,
  });
  const readyDownloads = (readyDownloadsQuery.data ?? []).filter(
    (download) =>
      download.policy_id == null || download.policy_id === policy?.policy_id,
  );
  const previewMutation = useMutation({
    mutationFn: (payload: { date_from: string; date_to: string }) =>
      retentionPolicyService.previewPolicy(productId!, {
        environment_id: environmentId!,
        policy_id: policy!.policy_id,
        ...payload,
        group_by: "day",
      }),
    onError: (error: unknown) =>
      message.error(`ดูข้อมูล Logs ไม่สำเร็จ: ${errorText(error)}`),
  });
  const downloadMutation = useMutation({
    mutationFn: (payload: { date_from: string; date_to: string }) =>
      retentionPolicyService.downloadArchives(productId!, {
        environment_id: environmentId!,
        policy_id: policy!.policy_id,
        ...payload,
      }),
    onSuccess: () => message.success("เริ่มดาวน์โหลด ZIP แล้ว"),
    onError: (error: unknown) =>
      message.error(`ดาวน์โหลด ZIP ไม่สำเร็จ: ${errorText(error)}`),
  });
  const readyDownloadMutation = useMutation({
    mutationFn: (downloadId: string) =>
      retentionPolicyService.downloadReadyArchive(
        productId!,
        environmentId!,
        downloadId,
      ),
    onSuccess: () => {
      message.success("ดาวน์โหลดไฟล์ Ready to load แล้ว");
      queryClient.invalidateQueries({
        queryKey: ["ready-archive-downloads", productId, environmentId],
      });
    },
    onError: (error: unknown) =>
      message.error(
        `ดาวน์โหลดไฟล์ Ready to load ไม่สำเร็จ: ${errorText(error)}`,
      ),
  });
  const snapshotRestoreMutation = useMutation({
    mutationFn: () =>
      retentionPolicyService.restoreSnapshotForSearch(
        productId!,
        environmentId!,
        snapshotArchiveId!,
      ),
    onSuccess: (result) => {
      message.success("Restore Snapshot สำเร็จ กำลังเปิด Logs Explorer");
      const query = new URLSearchParams({
        product: String(productId),
        environment_id: String(environmentId),
        archive_id: result.archive_id,
      });
      window.open(
        `/logs-explorer?${query.toString()}`,
        "_blank",
        "noopener,noreferrer",
      );
    },
    onError: (error: unknown) =>
      message.error(`Restore Snapshot ไม่สำเร็จ: ${errorText(error)}`),
  });
  const importMutation = useMutation({
    mutationFn: () =>
      retentionPolicyService.importArchive(
        productId!,
        environmentId!,
        importFile!,
      ),
    onSuccess: (result) => {
      message.success(`นำเข้า Archive สำเร็จ ${result.imported_count} รายการ`);
      queryClient.invalidateQueries({
        queryKey: ["retention-archives", productId, environmentId],
      });
      queryClient.invalidateQueries({
        queryKey: ["retention-policy-stats", productId, environmentId],
      });
      setImportFile(undefined);
    },
    onError: (error: unknown) =>
      message.error(`นำเข้า Archive ไม่สำเร็จ: ${errorText(error)}`),
  });

  useEffect(() => {
    if (!open || !policy) return;
    setRange(policyDefaultRange(policy));
    setSnapshotArchiveId(undefined);
    setImportFile(undefined);
    previewMutation.reset();
    downloadMutation.reset();
    readyDownloadMutation.reset();
  }, [open, policy?.policy_id]);

  const rangePayload =
    range?.[0] && range[1]
      ? {
          date_from: archiveDate(range[0]),
          date_to: archiveDate(range[1].add(1, "day")),
        }
      : null;
  const inspect = () => {
    if (!rangePayload) return message.warning("กรุณาเลือกช่วงวันที่ก่อน");
    previewMutation.mutate(rangePayload);
  };
  const download = () => {
    if (!rangePayload) return message.warning("กรุณาเลือกช่วงวันที่ก่อน");
    downloadMutation.mutate(rangePayload);
  };
  const restoreSnapshot = () => {
    if (!snapshotArchiveId)
      return message.warning("กรุณาเลือก Snapshot ที่ Verify แล้ว");
    snapshotRestoreMutation.mutate();
  };

  return (
    <Modal
      title={
        <Space>
          <EyeOutlined /> Logs และพื้นที่ของ Policy
        </Space>
      }
      open={open}
      onCancel={onClose}
      width={680}
      destroyOnClose
      footer={[
        <Button key="close" onClick={onClose}>
          ปิด
        </Button>,
        <Button
          key="inspect"
          type="primary"
          icon={<EyeOutlined />}
          onClick={inspect}
          loading={previewMutation.isPending}
        >
          ดูข้อมูล
        </Button>,
        <Button
          key="download"
          icon={<DownloadOutlined />}
          onClick={download}
          loading={downloadMutation.isPending}
        >
          Download ZIP
        </Button>,
        <Button
          key="snapshot"
          icon={<CloudOutlined />}
          onClick={restoreSnapshot}
          disabled={!snapshotArchiveId}
          loading={snapshotRestoreMutation.isPending}
        >
          เปิด Snapshot ใน Explorer
        </Button>,
        <Button
          key="import"
          icon={<InboxOutlined />}
          onClick={() => importMutation.mutate()}
          disabled={!importFile}
          loading={importMutation.isPending}
        >
          Import ZIP
        </Button>,
      ]}
    >
      <Alert
        type="info"
        showIcon
        message={`Policy: ${policy?.name ?? "-"}`}
        description="จำนวน Logs และขนาดพื้นที่คำนวณจาก Elasticsearch ตามช่วงวันที่เลือก ส่วน ZIP จะรวมเฉพาะ Archive ที่สร้างไว้แล้วในช่วงนั้น"
        className="mb-4"
      />
      <Card
        size="small"
        title="Ready to load"
        loading={readyDownloadsQuery.isLoading}
        className="mb-4"
      >
        {readyDownloads.length === 0 ? (
          <Typography.Text type="secondary">
            ยังไม่มี ZIP ที่รอโหลด ระบบจะแจ้งรายการเมื่อ Archive
            ถึงเวลาจัดเก็บเป็น ZIP
          </Typography.Text>
        ) : (
          <Space direction="vertical" className="w-full" size={8}>
            {readyDownloads.map((download) => (
              <div
                key={download.download_id}
                className="flex items-center justify-between gap-3"
              >
                <div>
                  <Typography.Text strong>{download.file_name}</Typography.Text>
                  <div>
                    <Typography.Text type="secondary">
                      {download.archive_count.toLocaleString()} Archive ·{" "}
                      {(download.size_bytes / 1024 / 1024).toFixed(2)} MB
                    </Typography.Text>
                  </div>
                </div>
                <Button
                  size="small"
                  icon={<DownloadOutlined />}
                  loading={readyDownloadMutation.isPending}
                  onClick={() =>
                    readyDownloadMutation.mutate(download.download_id)
                  }
                >
                  Download ZIP
                </Button>
              </div>
            ))}
          </Space>
        )}
      </Card>
      <DatePicker.RangePicker
        className="w-full"
        value={range}
        format="YYYY-MM-DD"
        placeholder={["วันเริ่มต้น", "วันสิ้นสุด"]}
        onChange={(value) => {
          setRange(value as ArchiveDateRange);
          previewMutation.reset();
        }}
      />
      {previewMutation.data && (
        <Descriptions
          className="mt-5"
          size="small"
          bordered
          column={1}
          items={[
            {
              key: "logs",
              label: "มี Logs หรือไม่",
              children:
                previewMutation.data.historical.document_count > 0
                  ? "มี"
                  : "ไม่มี",
            },
            {
              key: "count",
              label: "จำนวน Logs",
              children:
                previewMutation.data.historical.document_count.toLocaleString(),
            },
            {
              key: "size",
              label: "พื้นที่โดยประมาณ",
              children: `${(previewMutation.data.historical.estimated_size_bytes / 1024 / 1024).toFixed(2)} MB`,
            },
            {
              key: "oldest",
              label: "Log เก่าที่สุด",
              children: previewMutation.data.historical.oldest_log ?? "-",
            },
            {
              key: "newest",
              label: "Log ใหม่ที่สุด",
              children: previewMutation.data.historical.newest_log ?? "-",
            },
            {
              key: "archive",
              label: "สถานะ Archive",
              children:
                previewMutation.data.historical.archive_status ||
                "NOT_ARCHIVED",
            },
          ]}
        />
      )}
    </Modal>
  );
}

function BackfillLogsModal({
  open,
  policy,
  productId,
  environmentId,
  onClose,
}: {
  open: boolean;
  policy: RetentionPolicy | null;
  productId?: number;
  environmentId?: number;
  onClose: () => void;
}) {
  const queryClient = useQueryClient();
  const [range, setRange] = useState<ArchiveDateRange>(null);
  const [backupTag, setBackupTag] = useState("");
  const [previewKey, setPreviewKey] = useState("");
  const rangePayload =
    range?.[0] && range[1]
      ? {
          date_from: archiveDate(range[0]),
          date_to: archiveDate(range[1].add(1, "day")),
        }
      : null;
  const currentKey = rangePayload
    ? `${rangePayload.date_from}:${rangePayload.date_to}`
    : "";

  const previewMutation = useMutation({
    mutationFn: () =>
      retentionPolicyService.previewPolicy(productId!, {
        environment_id: environmentId!,
        policy_id: policy!.policy_id,
        ...rangePayload!,
        group_by: "day",
      }),
    onSuccess: () => setPreviewKey(currentKey),
    onError: (error: unknown) =>
      message.error(`ดูตัวอย่างไม่สำเร็จ: ${errorText(error)}`),
  });
  const importMutation = useMutation({
    mutationFn: async () => {
      const result = await retentionPolicyService.createArchive(productId!, {
        environment_id: environmentId!,
        policy_id: policy!.policy_id,
        ...rangePayload!,
        backup_type: "MANUAL",
        backup_tag: backupTag.trim() || undefined,
      });
      const verification = await Promise.all(
        result.archives.map((archive) =>
          retentionPolicyService.verifyArchive(
            productId!,
            environmentId!,
            archive.archive_id,
            policy!.policy_id,
          ),
        ),
      );
      return { archives: result.archives, verification };
    },
    onSuccess: (result) => {
      queryClient.invalidateQueries({
        queryKey: ["retention-archives", productId, environmentId],
      });
      if (result.archives.length === 0)
        return message.info("ไม่พบ Log ในช่วงวันที่ที่เลือก");
      message.success(
        `นำเข้าและยืนยัน ${result.archives.length} Archive แล้ว และนำ Logs ออกจาก Log Explorer`,
      );
    },
    onError: (error: unknown) =>
      message.error(`นำเข้า Log เก่าไม่สำเร็จ: ${errorText(error)}`),
  });

  useEffect(() => {
    if (!open) return;
    setRange(null);
    setPreviewKey("");
    setBackupTag("");
    previewMutation.reset();
    importMutation.reset();
  }, [open, policy?.policy_id]);

  const preview = () => {
    if (!rangePayload || !policy || !productId || !environmentId)
      return message.warning("กรุณาเลือกช่วงวันที่ก่อน");
    previewMutation.mutate();
  };
  const startImport = () => {
    if (!rangePayload || !previewMutation.data || previewKey !== currentKey)
      return message.warning("กรุณาดูตัวอย่างของช่วงวันที่นี้ก่อนนำเข้า");
    importMutation.mutate();
  };
  const canImport =
    Boolean(rangePayload) &&
    previewKey === currentKey &&
    (previewMutation.data?.archive_candidate_count ?? 0) > 0;
  const importButton = (
    <Button
      key="import"
      type="primary"
      disabled={!canImport}
      loading={importMutation.isPending}
      onClick={startImport}
    >
      นำเข้าและยืนยัน Archive
    </Button>
  );

  return (
    <Modal
      title={
        <Space>
          <InboxOutlined /> Manual Backup
        </Space>
      }
      open={open}
      onCancel={onClose}
      width={640}
      destroyOnClose
      footer={[
        <Button key="close" onClick={onClose}>
          ปิด
        </Button>,
        <Button
          key="preview"
          onClick={preview}
          loading={previewMutation.isPending}
        >
          ดูตัวอย่าง
        </Button>,
        importButton,
      ]}
    >
      <Alert
        type="info"
        showIcon
        message="สำรองแบบ Manual ด้วย Flow เดียวกับ Policy Backup"
        description="ระบบจะติด TAG, กันช่วงข้อมูลซ้ำ, สร้าง GZIP และอัปโหลดไปยัง R2 เมื่อ Policy เลือก R2 จากนั้นจึง Verify ก่อนดำเนินการกับ Active Logs"
        className="mb-4"
      />
      <Typography.Text strong>Policy: {policy?.name ?? "-"}</Typography.Text>
      <Input
        className="mt-3"
        value={backupTag}
        maxLength={180}
        onChange={(event) => setBackupTag(event.target.value)}
        placeholder="TAG (ไม่บังคับ) เช่น manual-release-2026-08"
        addonBefore="TAG"
      />
      <DatePicker.RangePicker
        className="mt-3 w-full"
        value={range}
        format="YYYY-MM-DD"
        placeholder={["วันเริ่มต้น", "วันสิ้นสุด"]}
        onChange={(value) => {
          setRange(value as ArchiveDateRange);
          setPreviewKey("");
          previewMutation.reset();
        }}
      />
      {previewMutation.data && previewKey === currentKey && (
        <Descriptions
          className="mt-5"
          size="small"
          bordered
          column={1}
          items={[
            {
              key: "logs",
              label: "Log ที่จะ Archive",
              children:
                previewMutation.data.archive_candidate_count.toLocaleString(),
            },
            {
              key: "size",
              label: "ขนาดโดยประมาณ",
              children: `${(previewMutation.data.historical.estimated_size_bytes / 1024 / 1024).toFixed(2)} MB`,
            },
            {
              key: "oldest",
              label: "Log เก่าที่สุด",
              children: previewMutation.data.historical.oldest_log ?? "-",
            },
            {
              key: "newest",
              label: "Log ใหม่ที่สุด",
              children: previewMutation.data.historical.newest_log ?? "-",
            },
          ]}
        />
      )}
      {importMutation.data && (
        <Alert
          className="mt-4"
          type="success"
          showIcon
          message={`สร้าง Archive สำเร็จ ${importMutation.data.archives.length} รายการ`}
          description="หากต้องการคืน Log ให้ใช้ปุ่ม Rollback Archive ด้านบน"
        />
      )}
    </Modal>
  );
}

function RollbackArchiveModal({
  open,
  productId,
  environmentId,
  onClose,
}: {
  open: boolean;
  productId?: number;
  environmentId?: number;
  onClose: () => void;
}) {
  const queryClient = useQueryClient();
  const [period, setPeriod] = useState<"DAY" | "WEEK" | "MONTH" | "YEAR">(
    "DAY",
  );
  const [periodValue, setPeriodValue] = useState<Dayjs | null>(dayjs());
  const [featureId, setFeatureId] = useState<number>();
  const [subFeatureId, setSubFeatureId] = useState<number>();
  const rangeStart = periodValue
    ? periodValue.startOf(
        period === "DAY"
          ? "day"
          : period === "WEEK"
            ? "week"
            : period === "MONTH"
              ? "month"
              : "year",
      )
    : null;
  const rangeEnd = rangeStart
    ? rangeStart.add(
        1,
        period === "DAY"
          ? "day"
          : period === "WEEK"
            ? "week"
            : period === "MONTH"
              ? "month"
              : "year",
      )
    : null;
  const featuresQuery = useQuery({
    queryKey: ["retention-rollback-features", productId],
    queryFn: () => productService.listFeatures(productId!),
    enabled: open && Boolean(productId),
  });
  const features = featuresQuery.data ?? [];
  const featureOptions = features.filter(
    (feature) => !feature.parent_id || feature.level === 1,
  );
  const subFeatureOptions = featureId
    ? features.filter(
        (feature) =>
          feature.parent_id === featureId ||
          feature.path_ids?.split(",").includes(String(featureId)),
      )
    : [];
  const archivesQuery = useQuery({
    queryKey: [
      "retention-archives",
      productId,
      environmentId,
      featureId,
      subFeatureId,
      rangeStart?.valueOf(),
      rangeEnd?.valueOf(),
    ],
    queryFn: () =>
      retentionPolicyService.listArchives(productId!, environmentId!, {
        feature_id: featureId,
        sub_feature_id: subFeatureId,
        date_from: rangeStart ? instantAtStartOfDay(rangeStart) : undefined,
        date_to: rangeEnd ? instantAtStartOfDay(rangeEnd) : undefined,
      }),
    enabled:
      open && Boolean(productId && environmentId && rangeStart && rangeEnd),
  });
  const archives = (archivesQuery.data ?? []).filter(
    (archive) => archive.status === "VERIFIED",
  );
  const selectedArchives =
    rangeStart && rangeEnd
      ? archives.filter((archive) => {
          if (!archive.date_from || !archive.date_to) return false;
          return (
            dayjs(archive.date_from).isBefore(rangeEnd) &&
            dayjs(archive.date_to).isAfter(rangeStart)
          );
        })
      : [];
  const selectedLogCount = selectedArchives.reduce(
    (total, archive) => total + archive.document_count,
    0,
  );
  useEffect(() => {
    if (open) {
      setPeriod("DAY");
      setPeriodValue(dayjs());
      setFeatureId(undefined);
      setSubFeatureId(undefined);
    }
  }, [open, productId, environmentId]);
  const restoreMutation = useMutation({
    mutationFn: async () => {
      let restoredCount = 0;
      let skippedCount = 0;
      const failures: Array<{ archiveId: string; error: string }> = [];
      for (const archive of selectedArchives) {
        try {
          const result = await retentionPolicyService.restoreArchive(
            productId!,
            environmentId!,
            archive.archive_id,
            "SKIP_EXISTING",
          );
          restoredCount += result.restored_count;
          skippedCount += result.skipped_count;
        } catch (error) {
          failures.push({
            archiveId: archive.archive_id,
            error: errorText(error),
          });
        }
      }
      return { restoredCount, skippedCount, failures };
    },
    onSuccess: (result) => {
      queryClient.invalidateQueries({
        queryKey: ["retention-archives", productId, environmentId],
      });
      queryClient.invalidateQueries({
        queryKey: ["retention-policy-stats", productId, environmentId],
      });
      if (result.failures.length === 0) {
        message.success(
          `คืน Log แล้ว ${result.restoredCount.toLocaleString()} รายการ${result.skippedCount ? `, ข้าม ${result.skippedCount.toLocaleString()} รายการที่มีอยู่แล้ว` : ""}`,
        );
      } else {
        const failedIds = result.failures
          .map((failure) => failure.archiveId.slice(0, 8))
          .join(", ");
        message.error(
          `Rollback ไม่ครบ ${result.failures.length} Archive (${failedIds}) — Archive ต้นฉบับยังอยู่และสามารถลองใหม่ได้`,
          10,
        );
      }
      const query = new URLSearchParams({
        product: String(productId),
        environment_id: String(environmentId),
        time_range: "all",
      });
      window.open(
        `/logs-explorer?${query.toString()}`,
        "_blank",
        "noopener,noreferrer",
      );
      if (result.failures.length === 0) {
        onClose();
      }
    },
    onError: (error: unknown) =>
      message.error(`Rollback ไม่สำเร็จ: ${errorText(error)}`),
  });
  return (
    <Modal
      title={
        <Space>
          <RollbackOutlined /> Rollback Archive
        </Space>
      }
      open={open}
      onCancel={onClose}
      width={640}
      footer={[
        <Button key="close" onClick={onClose}>
          ปิด
        </Button>,
        <Popconfirm
          key="restore"
          title={`ยืนยัน Rollback ${selectedArchives.length} Archive?`}
          description="ระบบจะคืน Logs ไปยัง Log Explorer โดยไม่เขียนทับ Logs ที่คืนไว้แล้ว"
          okText="คืน Log"
          cancelText="ยกเลิก"
          onConfirm={() => restoreMutation.mutate()}
        >
          <Button
            type="primary"
            disabled={!periodValue || selectedArchives.length === 0}
            loading={restoreMutation.isPending}
          >
            คืน Log จาก Archive
          </Button>
        </Popconfirm>,
      ]}
    >
      <Alert
        type="info"
        showIcon
        message="เลือกช่วงเวลาที่ต้องการ Rollback"
        description="ระบบจะคืน Logs จาก GZIP เข้า Log Explorer จากนั้นสามารถค้นหาและใช้ Filter ได้เหมือน Logs ปกติ"
        className="mb-4"
      />
      <Radio.Group
        className="mb-3"
        optionType="button"
        buttonStyle="solid"
        value={period}
        onChange={(event) => {
          setPeriod(event.target.value);
          setPeriodValue(null);
        }}
      >
        <Radio.Button value="DAY">ตามวัน</Radio.Button>
        <Radio.Button value="WEEK">ตามสัปดาห์</Radio.Button>
        <Radio.Button value="MONTH">ตามเดือน</Radio.Button>
        <Radio.Button value="YEAR">ตามปี</Radio.Button>
      </Radio.Group>
      <DatePicker
        className="mb-3 w-full"
        picker={
          period === "DAY"
            ? "date"
            : period === "WEEK"
              ? "week"
              : period === "MONTH"
                ? "month"
                : "year"
        }
        value={periodValue}
        onChange={setPeriodValue}
        placeholder={
          period === "DAY"
            ? "เลือกวันที่"
            : period === "WEEK"
              ? "เลือกสัปดาห์"
              : period === "MONTH"
                ? "เลือกเดือน"
                : "เลือกปี"
        }
      />
      <Space direction="vertical" className="w-full" size="small">
        <Select
          className="w-full"
          loading={featuresQuery.isLoading}
          value={featureId}
          placeholder="เลือก Feature ก่อน"
          onChange={(value) => {
            setFeatureId(value);
            setSubFeatureId(undefined);
          }}
          allowClear
          options={featureOptions.map((feature) => ({
            value: feature.category_id,
            label: feature.full_path ?? feature.category_name,
          }))}
        />
        <Select
          className="w-full"
          loading={featuresQuery.isLoading}
          value={subFeatureId}
          disabled={!featureId}
          placeholder={
            featureId ? "เลือก Sub-feature (ถ้ามี)" : "ต้องเลือก Feature ก่อน"
          }
          onChange={(value) => {
            setSubFeatureId(value);
          }}
          allowClear
          options={subFeatureOptions.map((feature) => ({
            value: feature.category_id,
            label: feature.full_path ?? feature.category_name,
          }))}
        />
      </Space>
      <Alert
        className="mt-4"
        type={selectedArchives.length ? "success" : "warning"}
        showIcon
        message={
          selectedArchives.length
            ? `พบ ${selectedArchives.length} Archive รวม ${selectedLogCount.toLocaleString()} Logs`
            : "ไม่พบ Archive ในช่วงเวลาที่เลือก"
        }
        description={
          rangeStart && rangeEnd
            ? `${rangeStart.format("DD/MM/YYYY")} ถึง ${rangeEnd.subtract(1, "millisecond").format("DD/MM/YYYY")}`
            : undefined
        }
      />
    </Modal>
  );
}

function PolicyDrawer({
  open,
  onClose,
  onSave,
  initial,
  productId,
  environmentId,
  loading,
}: {
  open: boolean;
  onClose: () => void;
  onSave: (p: CreateRetentionPolicyPayload) => Promise<void>;
  initial: RetentionPolicy | null;
  productId: number;
  environmentId: number;
  loading: boolean;
}) {
  const [form] = Form.useForm();
  const mode: RetentionMode = Form.useWatch("retention_mode", form) || "WEEKLY";
  const archiveEnabled = Boolean(Form.useWatch("archive_enabled", form));

  React.useEffect(() => {
    if (!open) return;
    if (initial) {
      form.setFieldsValue({
        name: initial.name,
        description: initial.description,
        retention_mode: initial.retention_mode,
        retention_unit: initial.retention_unit,
        retention_value: initial.retention_value,
        environment_id: initial.environment_id,
        auto_purge_action: "COMPRESS_GZIP",
        storage_provider: initial.storage_provider,
        bucket_name: initial.bucket_name,
        cron_schedule: initial.cron_schedule,
        schedule_time: scheduleTime(initial.cron_schedule),
        schedule_timezone: initial.schedule_timezone || "Asia/Bangkok",
        is_active: initial.is_active,
        archive_enabled: initial.archive_enabled,
        apply_to_existing_logs: initial.apply_to_existing_logs,
      });
    } else {
      form.resetFields();
    }
  }, [open, initial, form]);

  const handleFinish = async (
    v: Omit<CreateRetentionPolicyPayload, "product_id">,
  ) => {
    const selectedTime = form.getFieldValue("schedule_time") as
      | Dayjs
      | undefined;
    const cronSchedule = selectedTime
      ? `${selectedTime.minute()} ${selectedTime.hour()} * * *`
      : "0 1 * * *";
    const unitMap: Record<string, RetentionUnit> = {
      WEEKLY: "WEEKS",
      MONTHLY: "MONTHS",
      DAILY: "DAYS",
      CUSTOM: "DAYS",
    };
    await onSave({
      product_id: productId,
      environment_id: environmentId,
      name: v.name,
      description: v.description || "",
      retention_mode: v.retention_mode,
      retention_unit: unitMap[v.retention_mode] || "DAYS",
      retention_value: v.retention_value || 1,
      folder_structure: v.folder_structure,
      auto_purge_action: "COMPRESS_GZIP",
      storage_provider: v.storage_provider,
      bucket_name: v.bucket_name,
      cron_schedule: cronSchedule,
      schedule_timezone: v.schedule_timezone || "Asia/Bangkok",
      is_active: v.is_active ?? true,
      active_retention_value: v.retention_value || 1,
      active_retention_unit: unitMap[v.retention_mode] || "DAYS",
      archive_enabled: v.archive_enabled,
      // Archive follows the retention window above and is kept until the
      // user stops it. Active logs are never deleted by archive verification.
      delete_active_after_archive: Boolean(v.archive_enabled),
      archive_never_delete: Boolean(v.archive_enabled),
      apply_to_existing_logs: v.apply_to_existing_logs,
    });
    onClose();
  };

  const unitSuffix =
    mode === "WEEKLY" ? "สัปดาห์" : mode === "MONTHLY" ? "เดือน" : "วัน";

  const modeHints: Record<string, string> = {
    WEEKLY: "ล็อกหมุนเวียนใน 7 โฟลเดอร์ต่อสัปดาห์",
    MONTHLY: "4 โฟลเดอร์ย่อยรายสัปดาห์ แต่ละโฟลเดอร์มี 7 วัน",
    DAILY: "1 โฟลเดอร์ต่อวัน",
    CUSTOM: "โครงสร้างโฟลเดอร์กำหนดเอง",
  };

  return (
    <Drawer
      className="retention-drawer"
      title={initial ? "แก้ไข Policy" : "สร้าง Policy ใหม่"}
      width={500}
      open={open}
      onClose={onClose}
      destroyOnClose
      extra={
        <Space>
          <Button onClick={onClose}>ยกเลิก</Button>
          <Button
            type="primary"
            onClick={() => form.submit()}
            loading={loading}
            className="bg-[#1F8457]"
          >
            {initial ? "บันทึก" : "สร้าง"}
          </Button>
        </Space>
      }
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={handleFinish}
        initialValues={{
          retention_mode: "WEEKLY",
          retention_value: 1,
          auto_purge_action: "COMPRESS_GZIP",
          storage_provider: "R2",
          schedule_time: scheduleTime("0 1 * * *"),
          schedule_timezone: "Asia/Bangkok",
          archive_enabled: false,
          apply_to_existing_logs: false,
          is_active: true,
        }}
      >
        <Form.Item
          name="name"
          label="ชื่อ Policy"
          rules={[
            {
              required: true,
              message: "กรุณากรอกชื่อ Policy",
            },
          ]}
        >
          <Input placeholder="เช่น Weekly Hot Logs" />
        </Form.Item>

        <Form.Item name="description" label="คำอธิบาย (ไม่บังคับ)">
          <Input.TextArea rows={2} placeholder="คำอธิบาย (ไม่บังคับ)" />
        </Form.Item>

        <Form.Item
          name="retention_mode"
          label="ช่วงเวลาที่ต้องการจัดเก็บ"
          rules={[{ required: true }]}
        >
          <Radio.Group optionType="button" buttonStyle="solid">
            <Radio.Button value="DAILY">รายวัน</Radio.Button>
            <Radio.Button value="WEEKLY">รายสัปดาห์</Radio.Button>
            <Radio.Button value="MONTHLY">รายเดือน</Radio.Button>
            <Radio.Button value="CUSTOM">กำหนดเอง</Radio.Button>
          </Radio.Group>
        </Form.Item>

        {/* Mode hint */}
        <div className="retention-mode-hint">
          <InfoCircleOutlined />
          {modeHints[mode]}
        </div>

        <Form.Item
          name="retention_value"
          label={`เก็บรักษาเป็นเวลา (${unitSuffix})`}
          rules={[{ required: true }]}
        >
          <InputNumber min={1} max={365} className="w-full" />
        </Form.Item>

        <Form.Item name="auto_purge_action" label="เมื่อ Log หมดช่วง Active">
          <Select
            options={[
              {
                label: "สร้าง ZIP และแจ้งเตือน Ready to Download",
                value: "COMPRESS_GZIP",
              },
            ]}
          />
        </Form.Item>

        <Form.Item name="storage_provider" label="ผู้ให้บริการจัดเก็บ">
          <Select
            options={[
              {
                label: "Cloudflare R2",
                value: "R2",
              },
              {
                label: "ภายในเครื่อง (Local)",
                value: "LOCAL",
              },
              { label: "AWS S3", value: "S3" },
              { label: "Google Cloud Storage", value: "GCS" },
            ]}
          />
        </Form.Item>

        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
          <Form.Item
            name="schedule_time"
            label="เวลาตรวจวันหมดอายุ/สำรองอัตโนมัติ"
            rules={[{ required: true, message: "กรุณาเลือกเวลา" }]}
          >
            <TimePicker className="w-full" format="HH:mm" minuteStep={5} />
          </Form.Item>
          <Form.Item
            name="schedule_timezone"
            label="Timezone"
            rules={[{ required: true }]}
          >
            <Select
              showSearch
              options={[
                { label: "Asia/Bangkok (UTC+7)", value: "Asia/Bangkok" },
                { label: "UTC", value: "UTC" },
                { label: "Asia/Singapore (UTC+8)", value: "Asia/Singapore" },
                { label: "Asia/Tokyo (UTC+9)", value: "Asia/Tokyo" },
              ]}
            />
          </Form.Item>
        </div>

        <Form.Item
          name="archive_enabled"
          valuePropName="checked"
          label="เก็บ Archive ต่อเนื่อง"
          extra="ใช้ช่วงเวลาที่ต้องการจัดเก็บด้านบนเป็นหน้าต่าง Archive และทำซ้ำไปเรื่อย ๆ จนกว่าจะกดหยุด"
        >
          <Switch checkedChildren="เปิด" unCheckedChildren="ปิด" />
        </Form.Item>

        {archiveEnabled && (
          <>
            <Alert
              type="info"
              showIcon
              message="ระบบจะสร้าง Archive ตามช่วงเวลาที่ตั้งไว้ด้านบน"
              description="ตัวอย่าง: เลือกเก็บ 7 วัน ระบบจะหมุนเก็บ Archive ของช่วง 7 วันล่าสุดต่อเนื่องทุกวัน จนกว่าจะกด “หยุด Archive” หรือปิด Policy"
              className="mb-4"
            />
            <Alert
              type="warning"
              showIcon
              message="Logs ที่ Archive แล้วจะไม่แสดงใน Log Explorer"
              description="ระบบจะเก็บเป็น GZIP และลบออกจาก Active หลัง Verify สำเร็จ หากต้องการตรวจสอบย้อนหลังให้ใช้ Rollback ตามวัน สัปดาห์ หรือเดือน"
            />
          </>
        )}

        <Form.Item
          name="apply_to_existing_logs"
          valuePropName="checked"
          label="อนุญาตให้นำ Log เก่าเข้า Policy"
          extra="เปิดแล้วจะใช้ปุ่ม “นำเข้า Log เก่า” เพื่อเลือกช่วงวันและยืนยัน Archive ได้"
        >
          <Switch
            checkedChildren="นำไปใช้กับ log เดิม"
            unCheckedChildren="ใช้กับ log ใหม่เท่านั้น"
          />
        </Form.Item>

        <Form.Item
          name="is_active"
          valuePropName="checked"
          label="เปิดใช้งานทันที"
        >
          <Switch checkedChildren="เปิด" unCheckedChildren="ปิด" />
        </Form.Item>
      </Form>
    </Drawer>
  );
}

/* ──────────────────────────────────────────
   Simulate Modal — Thai
   ────────────────────────────────────────── */
function SimulateModal({ result }: { result: SimulationResult | null }) {
  return (
    <div className="space-y-4 py-2">
      <AnimatePresence mode="wait">
        {result && (
          <motion.div
            key="result"
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.25 }}
          >
            <div className="simulate-result-grid">
              <div className="simulate-stat">
                <div className="simulate-stat-value total">
                  {result.total_retention_days}
                </div>
                <div className="simulate-stat-label">จำนวนวันทั้งหมด</div>
              </div>
              <div className="simulate-stat">
                <div className="simulate-stat-value active">
                  {result.active_folders_count}
                </div>
                <div className="simulate-stat-label">โฟลเดอร์ที่ใช้งาน</div>
              </div>
              <div className="simulate-stat">
                <div className="simulate-stat-value purge">
                  {result.purge_folders_count}
                </div>
                <div className="simulate-stat-label">รอลบ</div>
              </div>
            </div>
            <div className="simulate-footer-info">
              วันที่ตัดรอบ: <code>{result.calculated_cutoff_date}</code>
              <span>·</span>
              ประหยัดได้ ~{result.estimated_savings_mb.toFixed(0)} MB
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
