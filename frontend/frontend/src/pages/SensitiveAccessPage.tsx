import { useEffect, useState } from "react";
import {
  Alert,
  Button,
  Card,
  Form,
  Input,
  Select,
  Space,
  Table,
  Tag,
  Typography,
  Modal,
  message,
} from "antd";
import { LockClosedIcon } from "@heroicons/react/24/outline";
import { PageTransition } from "@/components";
import { ROUTES } from "@/constants";
import {
  productAdminService,
  sensitiveLogService,
  type SensitiveAccessRequest,
  type SensitiveSecretOption,
} from "@/services";
import { useAppStore, useAuthStore } from "@/store";
import type { MainLog } from "@/types";
import { AuditSecretRequestsPanel } from "@/components/AuditSecretRequestsPanel";

const { Title, Text } = Typography;

export function SensitiveAccessPage() {
  const setBreadcrumbs = useAppStore((state) => state.setBreadcrumbs);
  const roles = useAuthStore((state) => state.roles);
  const canReview = roles.some((role) =>
    ["god", "owner", "superadmin", "super_admin"].includes(role),
  );
  const [products, setProducts] = useState<
    Array<{ productId: number; productName: string }>
  >([]);
  const [productId, setProductId] = useState<number>();
  const [requests, setRequests] = useState<SensitiveAccessRequest[]>([]);
  const [secrets, setSecrets] = useState<SensitiveSecretOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [requestTarget, setRequestTarget] = useState<"field" | "main_log">("field");
  const [canViewRawMainLogs, setCanViewRawMainLogs] = useState(false);
  const [mainLogs, setMainLogs] = useState<MainLog[]>([]);
  const [selectedMainLogId, setSelectedMainLogId] = useState<string>();
  const [rawMainLog, setRawMainLog] = useState<unknown>(null);
  const [loadingRawMainLog, setLoadingRawMainLog] = useState(false);

  const load = async () => {
    if (!productId) return;
    setLoading(true);
    try {
      const [requestItems, secretItems, mainLogPermission, rawPermission] = await Promise.all([
        sensitiveLogService.listRequests(productId),
        sensitiveLogService.listSecrets(productId),
        productAdminService.checkPermission(productId, "LOG", "READ"),
        productAdminService.checkPermission(productId, "LOG", "VIEW_SENSITIVE"),
      ]);
      setRequests(requestItems);
      setSecrets(secretItems);
      const approvedMainLogsRequest = requestItems.some((request) =>
        request.field_path === "$main_logs" &&
        request.approval_status === "APPROVED" &&
        (!request.expires_at || new Date(request.expires_at).getTime() > Date.now()),
      );
      const canRead = mainLogPermission.allowed || rawPermission.allowed || approvedMainLogsRequest;
      const [mainLogsResponse] = await Promise.all([
        canRead
          ? productAdminService.searchLogs({ productId, page: 1, perPage: 100 })
          : Promise.resolve({ data: [], total: 0, page: 1, perPage: 100 }),
      ]);
      setCanViewRawMainLogs(rawPermission.allowed || approvedMainLogsRequest);
      setMainLogs(mainLogsResponse.data);
      setSelectedMainLogId(mainLogsResponse.data[0]?.logId);


    } catch (e: any) {
      message.error(e.message || "Unable to load requests");
    } finally {
      setLoading(false);
    }
  };
  useEffect(() => {
    setBreadcrumbs([
      { title: "Sensitive logs" },
      { title: "Access requests", path: ROUTES.SENSITIVE_ACCESS },
    ]);
    productAdminService
      .listProducts()
      .then((items) => {
        setProducts(items);
        if (items[0]) setProductId(items[0].productId);
      })
      .catch(() => message.error("Unable to load products"));
  }, [setBreadcrumbs]);
  useEffect(() => {
    load();
  }, [productId]);

  const submit = async (values: { secret_id?: string; log_id?: string; reason: string }) => {
    if (!productId) return;
    try {
      await sensitiveLogService.createRequest({
        product_id: productId,
        ...(requestTarget === "main_log" ? { field_path: "$main_logs" } : { secret_id: values.secret_id }),
        reason: values.reason,
      });
      message.success("Access request submitted");
      await load();
    } catch (e: any) {
      message.error(e.message || "Unable to submit request");
    }
  };
  const review = async (
    request: SensitiveAccessRequest,
    status: "APPROVED" | "REJECTED",
  ) => {
    if (!productId) return;
    try {
      await sensitiveLogService.reviewRequest(
        productId,
        request.request_id,
        status,
      );
      message.success(`Request ${status.toLowerCase()}`);
      await load();
    } catch (e: any) {
      message.error(e.message || "Unable to review request");
    }
  };

  const handleViewRawMainLog = async () => {
    if (!productId || !selectedMainLogId) return;
    setLoadingRawMainLog(true);
    try {
      const value = await sensitiveLogService.revealRawMainLog(productId, selectedMainLogId);
      setRawMainLog(value);
    } catch (e: any) {
      message.error(e?.response?.data?.message || e?.message || "Unable to reveal raw Main Log");
    } finally {
      setLoadingRawMainLog(false);
    }
  };

  const columns = [
    {
      title: "Request",
      dataIndex: "request_id",
      render: (v: string) => (
        <Text copyable={{ text: v }}>{v.slice(0, 8)}...</Text>
      ),
    },
    { title: "User", dataIndex: "user_id" },
    {
      title: "Target",
      render: (_: unknown, r: SensitiveAccessRequest) => {
        const match = secrets.find((item) => item.secret_id === r.secret_id);
        return match
          ? `${match.field_key} (${match.source_section})`
          : r.field_path === "$main_logs" ? "All Main Logs" : r.field_path || r.secret_id || r.log_id || "Sensitive log";
      },
    },
    { title: "Reason", dataIndex: "reason", ellipsis: true },
    {
      title: "Status",
      render: (_: unknown, r: SensitiveAccessRequest) => {
        const expired =
          r.expires_at && new Date(r.expires_at).getTime() < Date.now();
        if (expired) return <Tag color="red">EXPIRED</Tag>;
        const v = r.approval_status;
        return (
          <Tag
            color={
              v === "APPROVED" ? "green" : v === "REJECTED" ? "red" : "gold"
            }
          >
            {v}
          </Tag>
        );
      },
    },
    ...(canReview
      ? [
          {
            title: "Action",
            render: (_: unknown, r: SensitiveAccessRequest) =>
              r.approval_status === "PENDING" ? (
                <Space>
                  <Button
                    size="small"
                    type="primary"
                    onClick={() => review(r, "APPROVED")}
                  >
                    Approve
                  </Button>
                  <Button
                    size="small"
                    danger
                    onClick={() => review(r, "REJECTED")}
                  >
                    Reject
                  </Button>
                </Space>
              ) : null,
          },
        ]
      : []),
  ];
  return (
    <PageTransition>
      <div className="space-y-6">
        <div>
          <Title level={2}>Sensitive log access</Title>
          <Text type="secondary">
            Request temporary access to masked log values. Every reveal is
            audited.
          </Text>
        </div>
        <Card>
          <Select
            className="w-full max-w-md"
            value={productId}
            onChange={setProductId}
            options={products.map((p) => ({
              value: p.productId,
              label: p.productName,
            }))}
            placeholder="Select product"
          />
        </Card>
        <Card
          title={
            <Space>
              <LockClosedIcon className="h-5 w-5" />
              Request access
            </Space>
          }
        >
          <Alert
            className="mb-4"
            type="info"
            showIcon
            message="เลือก protected field จากรายการด้านล่างได้เลย ไม่ต้องกรอก ID เอง"
          />
          <Form layout="vertical" onFinish={submit}>

            <Form.Item label="Access target">
              <Select
                value={requestTarget}
                onChange={setRequestTarget}
                options={[
                  { value: "field", label: "Protected field" },
                  { value: "main_log", label: "Main Log" },
                ]}
              />
            </Form.Item>
            {requestTarget === "main_log" ? (
              <>
                <Select
                  className="w-full mb-3"
                  value={selectedMainLogId}
                  onChange={setSelectedMainLogId}
                  placeholder={mainLogs.length ? "Select a Main Log" : "No Main Logs available"}
                  options={mainLogs.map((log) => ({
                    value: log.logId,
                    label: log.timestamp + " | " + (log.message || log.logId),
                  }))}
                  disabled={!mainLogs.length}
                />
                <Button
                  type="primary"
                  onClick={handleViewRawMainLog}
                  loading={loadingRawMainLog}
                  disabled={!canViewRawMainLogs || !selectedMainLogId}
                  className="mb-3"
                >
                  View Raw Main Log
                </Button>
                <Alert
                  className="mb-4"
                  type={canViewRawMainLogs ? "success" : "warning"}
                  showIcon
                  message={
                    canViewRawMainLogs
                      ? "You can view raw Main Logs."
                      : "You do not have raw Main Log access. Submit a request for temporary product-level access."
                  }
                />
              </>            ) : (
              <Form.Item
                name="secret_id"
                label="Protected field"
                rules={[{ required: true, message: "Please choose a protected field" }]}
              >
                <Select
                  showSearch
                  placeholder={productId ? "Choose a protected field" : "Select product first"}
                  disabled={!productId}
                  options={secrets.map((secret) => ({
                    value: secret.secret_id,
                    label: secret.field_key + " (" + secret.source_section + ")",
                  }))}
                  filterOption={(input, option) =>
                    String(option?.label || "").toLowerCase().includes(input.toLowerCase())
                  }
                />
              </Form.Item>
            )}

                        {!(requestTarget === "main_log" && canViewRawMainLogs) && (
              <>
<Form.Item
              name="reason"
              label="Why is access needed?"
              rules={[{ required: true, message: "Please provide a reason" }]}
            >
              <Input.TextArea
                rows={3}
                placeholder="Explain what you need to investigate"
              />
            </Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              disabled={!productId || (requestTarget === "main_log" && canViewRawMainLogs) || (requestTarget === "field" && secrets.length === 0)}
            >
              Submit request
            </Button>
              </>
            )}
          </Form>
        </Card>
        <Card
          title={canReview ? "Requests awaiting review" : "My access requests"}
          extra={
            <Button onClick={load} loading={loading}>
              Refresh
            </Button>
          }
        >
          <Table
            rowKey="request_id"
            loading={loading}
            dataSource={requests}
            scroll={{ x: 700 }}
            columns={columns}
          />
        </Card>
        <Modal
          title="Raw Main Log Data"
          open={rawMainLog !== null}
          onCancel={() => setRawMainLog(null)}
          footer={null}
          width={800}
        >
          <pre style={{ maxHeight: 600, overflow: "auto", whiteSpace: "pre-wrap" }}>
            {JSON.stringify(rawMainLog, null, 2)}
          </pre>
        </Modal>
        <AuditSecretRequestsPanel canReview={canReview} />
      </div>
    </PageTransition>
  );
}
