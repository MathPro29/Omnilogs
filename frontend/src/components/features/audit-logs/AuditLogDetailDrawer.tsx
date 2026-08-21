import { useMemo } from "react";
import {
  App,
  Button,
  Descriptions,
  Drawer as AntDrawer,
  Empty,
  Space,
  Skeleton,
  Tag,
  Typography,
} from "antd";
import { AuditOutlined } from "@ant-design/icons";
import dayjs from "dayjs";
import { useQuery } from "@tanstack/react-query";
import { auditService } from "@/features/audit-logs/services/audit.service";
import { userService } from "@/services";
import { useAuthStore } from "@/store";
import { buildUserDisplayMap } from "@/utils/userDisplay";

const { Title, Text } = Typography;

function JsonBlock({ value }: { value: unknown }) {
  return (
    <pre className="audit-json">{JSON.stringify(value ?? {}, null, 2)}</pre>
  );
}

interface AuditLogDetailDrawerProps {
  selectedAuditId?: string;
  onClose: () => void;
}

export function AuditLogDetailDrawer({
  selectedAuditId,
  onClose,
}: AuditLogDetailDrawerProps) {
  const resultColor: Record<string, string> = {
    SUCCESS: "success",
    FAILED: "error",
    DENIED: "warning",
  };

  const methodColor: Record<string, string> = {
    GET: "blue",
    POST: "green",
    PUT: "gold",
    PATCH: "cyan",
    DELETE: "red",
  };

  function readableAction(action: string) {
    return action
      .replaceAll("_", " ")
      .toLowerCase()
      .replace(/\b\w/g, (letter) => letter.toUpperCase());
  }

  const { message } = App.useApp();
  const currentUser = useAuthStore((state) => state.currentUser);

  const usersQuery = useQuery({
    queryKey: ["users", "list-all"],
    queryFn: () => userService.getUsers({ page: 1, pageSize: 100 }),
  });

  const userMap = useMemo(
    () => buildUserDisplayMap(currentUser, usersQuery.data),
    [currentUser, usersQuery.data],
  );

  const detail = useQuery({
    queryKey: ["audit-log", selectedAuditId],
    queryFn: () => auditService.getById(selectedAuditId!),
    enabled: Boolean(selectedAuditId),
  });

  return (
    <AntDrawer
      title={
        <Space>
          <AuditOutlined />
          รายละเอียด Activity
        </Space>
      }
      open={Boolean(selectedAuditId)}
      onClose={onClose}
      size={640}
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
                  children: (() => {
                    const actorId = detail.data.actor_user_id !== undefined ? Number(detail.data.actor_user_id) : undefined;
                    const mappedName = actorId !== undefined ? userMap[actorId] : undefined;
                    return (
                      detail.data.actor_name ||
                      mappedName ||
                      detail.data.actor_email ||
                      (actorId ? `User #${actorId}` : "System")
                    );
                  })(),
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
    </AntDrawer>
  );
}

export default AuditLogDetailDrawer;
