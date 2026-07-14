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

  const load = async () => {
    if (!productId) return;
    setLoading(true);
    try {
      const [requestItems, secretItems] = await Promise.all([
        sensitiveLogService.listRequests(productId),
        sensitiveLogService.listSecrets(productId),
      ]);
      setRequests(requestItems);
      setSecrets(secretItems);
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

  const submit = async (values: { secret_id: string; reason: string }) => {
    if (!productId) return;
    try {
      await sensitiveLogService.createRequest({
        product_id: productId,
        secret_id: values.secret_id,
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
          : r.field_path || r.secret_id || r.log_id || "Sensitive log";
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
            <Form.Item
              name="secret_id"
              label="Protected field"
              rules={[
                { required: true, message: "Please choose a protected field" },
              ]}
            >
              <Select
                showSearch
                placeholder={
                  productId
                    ? "Choose a protected field"
                    : "Select product first"
                }
                disabled={!productId}
                options={secrets.map((secret) => ({
                  value: secret.secret_id,
                  label: `${secret.field_key} (${secret.source_section})`,
                }))}
                filterOption={(input, option) =>
                  String(option?.label || "")
                    .toLowerCase()
                    .includes(input.toLowerCase())
                }
              />
            </Form.Item>
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
              disabled={!productId || secrets.length === 0}
            >
              Submit request
            </Button>
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
        <AuditSecretRequestsPanel canReview={canReview} />
      </div>
    </PageTransition>
  );
}
