import { useMemo, useState } from "react";
import { Modal } from "antd";
import { PageTransition } from "@/components";
import type { ApiKeyRecord, GeneratedApiKey } from "../types/apiKey.types";
import { useConnectionStore } from "@/features/logs-connections/store/connection.store";
import { useProductScope } from "@/features/logs-connections/hooks/useProductScope";
import { useApiKeys } from "@/features/logs-connections/hooks/useApiKeys";
import { useApiKeyPermissions } from "@/features/logs-connections/hooks/useApiKeyPermissions";
import { ConnectionHeader } from "@/components/features/logs-connections/ConnectionHeader";
import { ConnectionSummary } from "@/components/features/logs-connections/ConnectionSummary";
import { ApiKeyTable } from "@/components/features/logs-connections/ApiKeyTable";
import { GenerateApiKeyDrawer } from "@/components/features/logs-connections/GenerateApiKeyDrawer";
import { GeneratedSecretDialog } from "@/components/features/logs-connections/GeneratedSecretDialog";
import { ApiKeyDetailDrawer } from "@/components/features/logs-connections/ApiKeyDetailDrawer";
import "@/styles/pages/logs-connections.css";

export function LogsConnectionsPage() {
  const store = useConnectionStore();
  const scope = useProductScope();
  const access = useApiKeyPermissions(scope.productId);
  const apiKeys = useApiKeys(scope.productId);
  const [secret, setSecret] = useState<GeneratedApiKey | null>(null);
  const keys = useMemo(
    () =>
      (apiKeys.query.data ?? []).filter(
        (key) =>
          (!store.environmentId || key.environmentId === store.environmentId) &&
          (!store.search ||
            `${key.name} ${key.prefix}`
              .toLowerCase()
              .includes(store.search.toLowerCase())),
      ),
    [apiKeys.query.data, store.environmentId, store.search],
  );
  const selected = (apiKeys.query.data ?? []).find(
    (key) => key.keyId === store.selectedKeyId,
  );
  const product = scope.products.find(
    (item: { id: number; name: string }) => item.id === scope.productId,
  );
  const environment = scope.environments.find(
    (item: { environmentId: number }) =>
      item.environmentId === (secret?.environmentId ?? selected?.environmentId),
  );
  const confirmActive = (key: ApiKeyRecord, active: boolean) =>
    Modal.confirm({
      title: `${active ? "Enable" : "Disable"} ${key.name}?`,
      content: active
        ? "The key can send logs again immediately."
        : "Requests using this key will be rejected until it is enabled.",
      okText: active ? "Enable" : "Disable",
      onOk: () => apiKeys.setActive.mutate({ keyId: key.keyId, active }),
    });
  const confirmRevoke = (key: ApiKeyRecord) =>
    Modal.confirm({
      title: `Revoke ${key.name}?`,
      content: "Revocation is permanent. This key cannot be enabled again.",
      okText: "Revoke key",
      okButtonProps: { danger: true },
      onOk: () => apiKeys.revoke.mutate(key.keyId),
    });
  return (
    <PageTransition>
      <main className="logs-connections-page">
        <ConnectionHeader
          products={scope.products}
          environments={scope.environments}
          productId={scope.productId}
          environmentId={store.environmentId}
          search={store.search}
          canCreate={access.canCreate}
          onProduct={store.setProductId}
          onEnvironment={store.setEnvironmentId}
          onSearch={store.setSearch}
          onGenerate={() => store.setGenerateOpen(true)}
        />
        <ConnectionSummary keys={keys} />
        <div className="connections-table-shell">
          <ApiKeyTable
            keys={keys}
            products={scope.products}
            environments={scope.environments}
            loading={apiKeys.query.isLoading}
            canUpdate={access.canUpdate}
            canDelete={access.canDelete}
            onOpen={(key) => store.selectKey(key.keyId)}
            onActive={confirmActive}
            onRevoke={confirmRevoke}
          />
        </div>
        <GenerateApiKeyDrawer
          open={store.generateOpen}
          productName={product?.name}
          environments={scope.environments}
          loading={apiKeys.create.isPending}
          onClose={() => store.setGenerateOpen(false)}
          onSubmit={(input) =>
            apiKeys.create.mutate(input, {
              onSuccess: (created) => {
                store.setGenerateOpen(false);
                setSecret(created);
              },
            })
          }
        />
        <GeneratedSecretDialog
          value={secret}
          product={product}
          environment={environment}
          onClose={() => setSecret(null)}
        />
        <ApiKeyDetailDrawer
          value={selected}
          product={product}
          environment={environment}
          onClose={() => store.selectKey(undefined)}
        />
      </main>
    </PageTransition>
  );
}
