import { CopyOutlined } from "@ant-design/icons";
import { Button, Tabs, message } from "antd";
import { getOmniLogsIngestUrl } from "@/utils/omnilogs-url";

const copy = async (value: string) => { await navigator.clipboard.writeText(value); message.success("Copied"); };
const exampleData = { timestamp: new Date().toISOString(), level: "INFO", service: "YOUR_SERVICE", request_path: "/health", request_method: "GET", message: "Omnilogs connection test" };

export function ConnectionGuide({ productId, secret = "YOUR_API_KEY" }: { productId?: number; environmentCode?: string; secret?: string }) {
  const slug = productId ? `product-${productId}` : "product";
  const body = { queue_key: `${slug}-web-logs`, source_type: "SERVICE", source_platform: "HTTP", logs: [{ sequence_no: 1, source_type: "SERVICE", source_platform: "HTTP", data: { ...exampleData, service: `${slug}-web` } }] };
  const curl = `curl --request POST \\
  --url ${getOmniLogsIngestUrl()} \\
  --header "X-API-Key: ${secret}" \\
  --header "Content-Type: application/json" \\
  --data '${JSON.stringify(body, null, 2)}'`;
  return <section className="connection-guide"><div><h3>Connection guide</h3><span>POST /api/v1/ingest/logs</span></div><pre>{curl}</pre><Button icon={<CopyOutlined />} onClick={() => void copy(curl)}>Copy cURL example</Button><Tabs items={[{ key: "payload", label: "Payload", children: <div className="connection-example"><pre>{JSON.stringify(body, null, 2)}</pre><Button size="small" icon={<CopyOutlined />} onClick={() => void copy(JSON.stringify(body, null, 2))}>Copy JSON</Button></div> }]} /></section>;
}
