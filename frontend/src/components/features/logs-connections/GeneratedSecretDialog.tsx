import { useState } from "react";
import { Button, Modal, message } from "antd";
import {
  CheckOutlined,
  CopyOutlined,
  DownloadOutlined,
} from "@ant-design/icons";
import type { GeneratedApiKey, ProductEnvironment } from "@/types/apiKey.types";
import type { ProductOption } from "@/services/product.service";
import { getOmniLogsBaseUrl } from "@/utils/omnilogs-url";
import { ConnectionGuide } from "./ConnectionGuide";
const copy = async (value: string) => {
  await navigator.clipboard.writeText(value);
  message.success("Copied securely");
};
export function GeneratedSecretDialog({
  value,
  product,
  environment,
  onClose,
}: {
  value: GeneratedApiKey | null;
  product?: ProductOption;
  environment?: ProductEnvironment;
  onClose: () => void;
}) {
  const [closing, setClosing] = useState(false);

  if (!value) return null;
  const connectionSlug = `product-${value.productId}`;
  const envText = `OMNILOGS_BASE_URL=${getOmniLogsBaseUrl()}
OMNILOGS_API_KEY=${value.secret}
OMNILOGS_PRODUCT_ID=${value.productId}
OMNILOGS_ENVIRONMENT_CODE=${environment?.code ?? "production"}
OMNILOGS_QUEUE_KEY=${connectionSlug}-web-logs
OMNILOGS_SERVICE=${connectionSlug}-web
# Omnilogs จะ map Project / Feature จาก service, request_path และ request_method
# ไม่ต้องกำหนด OMNILOGS_ROUTE_KEY หรือ OMNILOGS_EVENT_NAME แบบตายตัว
`;
  const download = () => {
    const url = URL.createObjectURL(
      new Blob([envText], { type: "text/plain" }),
    );
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = ".env.omnilogs.example";
    anchor.click();
    URL.revokeObjectURL(url);
  };
  return (
    <Modal
      open={!closing}
      width={720}
      title="API key generated"
      closable={false}
      maskClosable={false}
      keyboard={false}
      destroyOnHidden
      afterClose={() => {
        if (closing) {
          setClosing(false);
          onClose();
        }
      }}
      footer={
        <Button
          type="primary"
          icon={<CheckOutlined />}
          onClick={() => setClosing(true)}
        >
          I understand
        </Button>
      }
    >
      <div className="generated-warning">
        <strong>This API key will only be shown once.</strong>
        <span>Store it securely before closing this window.</span>
      </div>
      <div className="generated-secret">
        <code>{value.secret}</code>
        <Button icon={<CopyOutlined />} onClick={() => void copy(value.secret)}>
          Copy API Key
        </Button>
      </div>
      <dl className="generated-meta">
        <div>
          <dt>Key prefix</dt>
          <dd>{value.prefix}</dd>
        </div>
        <div>
          <dt>Product</dt>
          <dd>{product?.name ?? value.productId}</dd>
        </div>
        <div>
          <dt>Environment</dt>
          <dd>{environment?.name ?? "All environments"}</dd>
        </div>
        <div>
          <dt>Permissions</dt>
          <dd>{value.permissions.join(", ")}</dd>
        </div>
      </dl>
      <div className="generated-actions">
        <Button icon={<DownloadOutlined />} onClick={download}>
          Download environment example
        </Button>
      </div>
      <ConnectionGuide
        productId={value.productId}
        environmentCode={environment?.code}
        secret={value.secret}
      />
    </Modal>
  );
}
