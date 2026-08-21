import {
  Card,
  Typography,
  Tag,
  Button,
  Tabs,
  Space,
  Tooltip,
  message,
  Descriptions,
  Empty,
} from "antd";
import {
  CopyOutlined,
  CodeOutlined,
  InfoCircleOutlined,
  FileTextOutlined,
  CloseOutlined,
} from "@ant-design/icons";
import type { SearchRecord } from "@/features/log-explorer/services/log-search.service";

const { Text } = Typography;

interface RawViewerProps {
  record: SearchRecord | null;
  onClose?: () => void;
}

export function RawViewer({ record, onClose }: RawViewerProps) {
  if (!record) {
    return (
      <Card className="h-full flex items-center justify-center min-h-[400px]">
        <Empty description="เลือกรายการ Log ในตารางเพื่อดู Raw Log" />
      </Card>
    );
  }

  const timestamp = String(record["@timestamp"] ?? record.timestamp ?? "N/A");
  const level = String(
    record.level ?? record.payload?.log_level ?? "INFO",
  ).toUpperCase();
  const service = String(record.service ?? record.payload?.service ?? "N/A");
  const messageText = String(record.message ?? record.payload?.message ?? "");

  const handleCopyJson = () => {
    navigator.clipboard.writeText(JSON.stringify(record, null, 2));
    message.success("คัดลอก JSON เรียบร้อยแล้ว");
  };

  const handleCopyMessage = () => {
    navigator.clipboard.writeText(messageText);
    message.success("คัดลอกข้อความ Log เรียบร้อยแล้ว");
  };

  const getLevelTag = (logLevel: string) => {
    switch (logLevel) {
      case "ERROR":
      case "FATAL":
      case "CRITICAL":
        return <Tag color="#FF3D89">{logLevel}</Tag>;
      case "WARN":
      case "WARNING":
        return <Tag color="#D4B106">{logLevel}</Tag>;
      case "DEBUG":
      case "TRACE":
        return <Tag color="default">{logLevel}</Tag>;
      default:
        return <Tag color="processing">{logLevel}</Tag>;
    }
  };

  return (
    <Card
      title={
        <Space size="middle">
          <FileTextOutlined
            style={{ color: "var(--color-primary, #1F8457)" }}
          />
          <span>Raw Log Viewer</span>
          {getLevelTag(level)}
        </Space>
      }
      extra={
        <Space>
          <Tooltip title="คัดลอก JSON ทั้งหมด">
            <Button
              icon={<CopyOutlined />}
              onClick={handleCopyJson}
              size="small"
            >
              Copy JSON
            </Button>
          </Tooltip>
          {onClose && (
            <Button
              icon={<CloseOutlined />}
              type="text"
              onClick={onClose}
              size="small"
            />
          )}
        </Space>
      }
      className="h-full border border-gray-200 dark:border-gray-800 shadow-sm"
    >
      <Tabs
        defaultActiveKey="raw"
        items={[
          {
            key: "raw",
            label: (
              <span>
                <CodeOutlined /> Raw JSON
              </span>
            ),
            children: (
              <div className="space-y-4">
                <div className="flex justify-between items-center bg-gray-50 p-3 rounded-lg border border-gray-200 dark:border-gray-700">
                  <div>
                    <Text type="secondary" className="text-xs block text-black">
                      Timestamp
                    </Text>
                    <Text strong className="font-mono text-xs">
                      {timestamp}
                    </Text>
                  </div>
                  <div>
                    <Text type="secondary" className="text-xs block">
                      Service
                    </Text>
                    <Tag
                      color="default"
                      className="bg-[#F6F6F6] text-[#1F8457] border-[#D9CAB3] font-medium"
                    >
                      {service}
                    </Tag>
                  </div>
                </div>

                <div>
                  <div className="flex justify-between items-center mb-1">
                    <Text strong className="text-xs text-gray-500">
                      MESSAGE
                    </Text>
                    <Button
                      type="link"
                      size="small"
                      icon={<CopyOutlined />}
                      onClick={handleCopyMessage}
                    >
                      Copy
                    </Button>
                  </div>
                  <div className="p-3 bg-gray-900 text-gray-100 rounded-md font-mono text-xs overflow-x-auto whitespace-pre-wrap break-words leading-relaxed max-h-[160px]">
                    {messageText || "(empty message)"}
                  </div>
                </div>

                <div>
                  <Text strong className="text-xs text-gray-500 mb-1 block">
                    FULL PAYLOAD (JSON)
                  </Text>
                  <pre className="p-3 bg-black text-white rounded-md font-mono text-xs overflow-x-auto leading-relaxed border border-gray-800 max-h-[320px]">
                    {JSON.stringify(record, null, 2)}
                  </pre>
                </div>
              </div>
            ),
          },
          {
            key: "overview",
            label: (
              <span>
                <InfoCircleOutlined /> Overview
              </span>
            ),
            children: (
              <Descriptions
                bordered
                column={1}
                size="small"
                items={Object.entries(record).map(([key, val]) => ({
                  key,
                  label: (
                    <span className="font-mono font-semibold text-xs">
                      {key}
                    </span>
                  ),
                  children: (
                    <span className="font-mono text-xs break-all">
                      {typeof val === "object"
                        ? JSON.stringify(val)
                        : String(val ?? "null")}
                    </span>
                  ),
                }))}
              />
            ),
          },
        ]}
      />
    </Card>
  );
}
