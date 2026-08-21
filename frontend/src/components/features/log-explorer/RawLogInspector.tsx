import { useEffect, useState } from "react";
import {
  Alert,
  Button,
  Empty,
  message,
  Segmented,
  Tabs,
  Tag,
  Tooltip,
} from "antd";
import {
  AlertOutlined,
  CloseOutlined,
  DeleteOutlined,
  FullscreenOutlined,
  PlusOutlined,
  StarFilled,
} from "@ant-design/icons";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type {
  SearchRecord,
  SearchRule,
} from "@/features/log-explorer/services/log-search.service";
import { customFieldFavoriteService } from "@/features/log-explorer/services/custom-field-favorite.service";
import { fieldGroups } from "@/features/log-explorer/utils/log-utils";
import { useAuthStore } from "@/store";
import {
  addUserFavoriteId,
  favoriteLogColumnKey,
  getUserFavoriteIds,
  removeUserFavoriteId,
  useLogExplorerStore,
} from "@/features/log-explorer/store/useLogExplorerStore";
import { PrettyLogView, SummaryCard } from "./SummaryCard";
import { JsonTree } from "./JsonTree";
import { AddFeatureModal } from "./AddFeatureModal";
import { LogRouteMappingModal } from "./LogRouteMappingModal";

interface Props {
  record: SearchRecord | null;
  productId?: number;
  fullscreen: boolean;
  onFullscreen: (value: boolean) => void;
  onClose: () => void;
  onAddRule: (rule: SearchRule) => void;
}
const typeOf = (value: unknown) =>
  Array.isArray(value) ? "array" : value === null ? "null" : typeof value;
const normalPath = (path: string) => {
  let p = path.trim();
  p = p.replace(/^\$\.?/, "");
  p = p.replace(/^raw\./, "");
  p = p.replace(/^(data|fields|payload)\./, "");
  return p || "$";
};
const changeValue = (value: unknown) => {
  if (value === undefined) return "Not set";
  if (value === null) return "null";
  if (typeof value === "string") return value || "Empty string";
  if (typeof value === "object") return JSON.stringify(value, null, 2);
  return String(value);
};

export function RawLogInspector({
  record,
  productId,
  fullscreen,
  onFullscreen,
  onClose,
  onAddRule,
}: Props) {
  const [view, setView] = useState<"sections" | "raw">("sections");
  const [optimisticFavorites, setOptimisticFavorites] = useState<string[]>([]);
  const [isAddFeatureOpen, setIsAddFeatureOpen] = useState(false);
  const [isRouteMappingOpen, setIsRouteMappingOpen] = useState(false);
  const currentUser = useAuthStore((state) => state.currentUser);
  const userId = currentUser?.id;
  const explorer = useLogExplorerStore();
  const client = useQueryClient();
  const favoritesQuery = useQuery({
    queryKey: ["custom-fields", "favorites", productId],
    queryFn: () => customFieldFavoriteService.list(productId!),
    enabled: Boolean(productId),
  });
  const allFavorites = favoritesQuery.data ?? [];
  const userFavIds = getUserFavoriteIds(userId, productId);
  const favorites =
    userFavIds === null
      ? []
      : allFavorites.filter((field) =>
          userFavIds.includes(field.field_definition_id),
        );
  const favoritePaths = [...new Set([
    ...favorites.map((field) => normalPath(field.field_path ?? field.field_key)),
    ...optimisticFavorites,
  ])];
  const refreshFavorites = () =>
    void client.invalidateQueries({
      queryKey: ["custom-fields", "favorites", productId],
    });
  const addFavorite = useMutation({
    mutationFn: ({
      path,
      name,
      value,
    }: {
      path: string;
      name: string;
      value: unknown;
    }) =>
      customFieldFavoriteService.add({
        product_id: productId!,
        field_path: normalPath(path),
        display_name: name,
        sample_value: value,
        detected_type: typeOf(value),
      }),
    onMutate: ({ path }) => {
      const normalizedPath = normalPath(path);
      setOptimisticFavorites((current) => current.includes(normalizedPath) ? current : [...current, normalizedPath]);
    },
    onSuccess: (field, variables) => {
      setOptimisticFavorites((current) => current.filter((path) => path !== normalPath(variables.path)));
      message.success("Favorite field added");
      if (userId && productId && field?.field_definition_id) {
        addUserFavoriteId(userId, productId, field.field_definition_id);
        const colKey = favoriteLogColumnKey(field.field_definition_id);
        if (!explorer.columns.includes(colKey)) {
          explorer.setColumns([...explorer.columns, colKey]);
        }
        client.setQueryData<typeof allFavorites>(
          ["custom-fields", "favorites", productId],
          (old) => {
            const list = old ?? [];
            if (list.some((item) => item.field_definition_id === field.field_definition_id)) return list;
            return [...list, field];
          },
        );
      }
      refreshFavorites();
    },
    onError: (_, variables) => {
      setOptimisticFavorites((current) => current.filter((path) => path !== normalPath(variables.path)));
      message.error("Could not add favorite field");
    },
  });
  const removeFavorite = useMutation({
    mutationFn: customFieldFavoriteService.remove,
    onSuccess: (_, fieldDefinitionId) => {
      message.success("Favorite field removed");
      if (userId && productId) {
        removeUserFavoriteId(userId, productId, fieldDefinitionId);
        const colKey = favoriteLogColumnKey(fieldDefinitionId);
        if (explorer.columns.includes(colKey)) {
          explorer.setColumns(explorer.columns.filter((c) => c !== colKey));
        }
        client.setQueryData<typeof allFavorites>(
          ["custom-fields", "favorites", productId],
          (old) => (old ?? []).filter((item) => item.field_definition_id !== fieldDefinitionId),
        );
      }
      refreshFavorites();
    },
    onError: () => message.error("Could not remove favorite field"),
  });
  useEffect(() => {
    const handler = (event: KeyboardEvent) =>
      event.key === "Escape" && (fullscreen ? onFullscreen(false) : onClose());
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [fullscreen, onClose, onFullscreen]);
  if (!record)
    return (
      <aside className="log-inspector log-inspector--empty">
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description="Select a log to inspect its structured data"
        />
      </aside>
    );
  const groups = fieldGroups(record);
  const requestPath = [
    record.path,
    record.request_path,
    (record.payload as Record<string, unknown> | undefined)?.request_path,
    (groups.request as Record<string, unknown>).path,
  ].find(
    (value): value is string =>
      typeof value === "string" && value.trim().length > 0,
  );
  const busyPath = addFavorite.variables ? normalPath(addFavorite.variables.path) : null;
  const addFilter = (path: string, value: unknown, exclude: boolean) =>
    onAddRule({
      id: crypto.randomUUID(),
      field: path.replace(/^\$\./, ""),
      label: path.split(".").at(-1) ?? path,
      type: typeof value === "number" ? "number" : "keyword",
      operator: exclude ? "neq" : "eq",
      value: typeof value === "number" ? value : String(value),
      enabled: true,
    });
  const toggleFavorite = (path: string, name: string, value: unknown) => {
    if (!productId) {
      message.warning("Select a product before adding a favorite");
      return;
    }
    const targetPath = normalPath(path);
    const field = favorites.find(
      (item) => normalPath(item.field_path ?? item.field_key) === targetPath,
    );
    if (field) removeFavorite.mutate(field.field_definition_id);
    else addFavorite.mutate({ path, name, value });
  };
  const tree = (value: unknown, rootPath: string) => (
    <JsonTree
      value={value}
      rootPath={rootPath}
      onFilter={addFilter}
      favoritePaths={favoritePaths}
      busyPath={busyPath}
      onFavorite={toggleFavorite}
    />
  );
  const isEmpty = (value: unknown) =>
    value === undefined ||
    value === null ||
    (Array.isArray(value) && value.length === 0) ||
    (!Array.isArray(value) &&
      typeof value === "object" &&
      Object.keys(value as object).length === 0);
  const present = (
    items: Array<{
      key: string;
      label: string;
      value: unknown;
      rootPath: string;
    }>,
  ) =>
    items
      .filter((item) => !isEmpty(item.value))
      .map((item) => ({
        key: item.key,
        label: item.label,
        children: tree(item.value, item.rootPath),
      }));
  const section = (value: unknown, rootPath: string, description: string) =>
    isEmpty(value) ? (
      <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={description} />
    ) : (
      tree(value, rootPath)
    );
  const customContent = (
    <div className="log-custom-fields">
      <div className="log-favorite-fields">
        <h3>
          <StarFilled /> Favorite fields
        </h3>
        {favorites.length ? (
          favorites.map((field) => (
            <div className="log-favorite-field" key={field.field_definition_id}>
              <div>
                <strong>{field.display_name ?? field.field_key}</strong>
                <span>{field.field_path ?? field.field_key}</span>
              </div>
              <Button
                type="text"
                danger
                icon={<DeleteOutlined />}
                loading={removeFavorite.isPending}
                onClick={() => removeFavorite.mutate(field.field_definition_id)}
                aria-label={`Remove ${field.display_name ?? field.field_key} from favorites`}
              />
            </div>
          ))
        ) : (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description="No favorite fields yet"
          />
        )}
      </div>
      {section(
        groups.custom,
        "$.custom_fields",
        "No custom fields in this log",
      )}
    </div>
  );
  const changeContent = groups.changes.length ? (
    <section className="log-change-list" aria-label="Changed fields">
      <header>
        <strong>
          {groups.changes.length} changed field
          {groups.changes.length === 1 ? "" : "s"}
        </strong>
        <span>Values recorded before and after the update</span>
      </header>
      <div className="log-change-table" role="table">
        <div className="log-change-table__head" role="row">
          <span role="columnheader">Field</span>
          <span role="columnheader">Before</span>
          <span role="columnheader">After</span>
        </div>
        {groups.changes.map((change, index) => (
          <div
            className="log-change-table__row"
            role="row"
            key={`${change.path}-${index}`}
          >
            <code role="cell">{change.path}</code>
            <pre role="cell">{changeValue(change.before)}</pre>
            <pre role="cell">{changeValue(change.after)}</pre>
          </div>
        ))}
      </div>
    </section>
  ) : null;
  const requestItems = present([
    {
      key: "request-overview",
      label: "Overview",
      value: groups.request,
      rootPath: "$.request",
    },
    {
      key: "request-headers",
      label: "Headers",
      value: groups.requestHeaders,
      rootPath: "$.request.headers",
    },
    {
      key: "request-query",
      label: "Query",
      value: groups.requestQuery,
      rootPath: "$.request.query",
    },
    {
      key: "request-body",
      label: "Body",
      value: groups.requestBody,
      rootPath: "$.request.body",
    },
  ]);
  const responseItems = present([
    {
      key: "response-overview",
      label: "Overview",
      value: groups.response,
      rootPath: "$.response",
    },
    {
      key: "response-headers",
      label: "Headers",
      value: groups.responseHeaders,
      rootPath: "$.response.headers",
    },
    {
      key: "response-body",
      label: "Body",
      value: groups.responseBody,
      rootPath: "$.response.body",
    },
    {
      key: "response-error",
      label: "Error",
      value: groups.responseError,
      rootPath: "$.response.error",
    },
  ]);
  const responseOverview = groups.response as Record<string, unknown>;
  const responseBody = groups.responseBody as Record<string, unknown>;
  const responseErrorObject = (responseBody?.err ?? responseOverview?.err) as Record<string, unknown> | undefined;
  const responseFailureReason =
    responseErrorObject?.Message ??
    responseErrorObject?.message ??
    responseBody?.message ??
    responseOverview?.message ??
    responseBody?.error_message;
  const responseFailures = responseBody?.fail ?? responseOverview?.fail;
  const responseFailureItems = [
    responseFailureReason ? `Reason: ${String(responseFailureReason)}` : "",
    Array.isArray(responseFailures) && responseFailures.length ? `Failed operation: ${responseFailures.join("; ")}` : "",
  ].filter(Boolean);
  const responseFailureText = responseFailureItems.join("\n");
  const requestMethod = String((groups.request as Record<string, unknown>).method ?? record.method ?? "");

  const items = [
    {
      key: "summary",
      label: "Summary",
      children: <SummaryCard record={record} />,
    },
    {
      key: "pretty",
      label: "Pretty view",
      children: <PrettyLogView record={record} />,
    },
    ...(changeContent
      ? [
          {
            key: "changes",
            label: `Changes (${groups.changes.length})`,
            children: changeContent,
          },
        ]
      : []),
    {
      key: "payload",
      label: "Event payload",
      children: (
        <section className="log-inspector-section">
          {section(groups.payload, "$.payload", "No event payload captured")}
        </section>
      ),
    },
    {
      key: "actor",
      label: "Actor",
      children: (
        <section className="log-inspector-section">
          {section(groups.author, "$.author", "ไม่พบ ACTOR ใน Log นี้")}
        </section>
      ),
    },
    {
      key: "request",
      label: "Request",
      children: (
        <section className="log-inspector-section">
          <header className="log-inspector-section__header">
            <div>
              <div>
                <strong>Incoming request</strong>
                {requestMethod ? <Tag>{requestMethod}</Tag> : null}
              </div>
            </div>
            <span>
              {requestPath ?? "Request data captured by the source service"}
            </span>
          </header>
          {requestItems.length ? (
            <Tabs size="small" items={requestItems} />
          ) : (
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              description="No request data captured"
            />
          )}
        </section>
      ),
    },
    {
      key: "response",
      label: "Response",
      children: (
        <section className="log-inspector-section p-4">
          {responseFailureItems.length > 0 && (
            <Alert type="error" showIcon className="log-response-failure mb-4" message="Failure reason" description={<pre style={{ whiteSpace: "pre-wrap", margin: 0, fontFamily: "inherit" }}>{responseFailureText}</pre>} />
          )}
          {responseItems.length ? (
            <Tabs size="small" items={responseItems} />
          ) : (
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              description="No response data captured"
            />
          )}
        </section>
      ),
    },
    {
      key: "context",
      label: "Context",
      children: (
        <Tabs
          size="small"
          items={[
            {
              key: "metadata",
              label: "Metadata",
              children: section(
                groups.metadata,
                "$.metadata",
                "No metadata captured",
              ),
            },
            {
              key: "trace",
              label: "Trace",
              children: section(
                groups.trace,
                "$.trace",
                "No trace information captured",
              ),
            },
            { key: "custom", label: "Custom fields", children: customContent },
          ]}
        />
      ),
    },
  ];
  const visibleItems = items.filter(
    (item) =>
      (item.key !== "request" || requestItems.length > 0) &&
      (item.key !== "response" || responseItems.length > 0),
  );

  const payloadRecord = record.payload as Record<string, unknown> | undefined;
  const isMapped =
    record.routing_status === "CLASSIFIED" ||
    payloadRecord?.routing_status === "CLASSIFIED" ||
    payloadRecord?.mapping_status === "mapped";
  const isUnclassified =
    !isMapped &&
    (record.routing_status === "UNCLASSIFIED" ||
      payloadRecord?.routing_status === "UNCLASSIFIED" ||
      payloadRecord?.mapping_status === "unmapped" ||
      !record.category_id);

  const customFeatureCode =
    (groups.custom as Record<string, unknown> | undefined)?.sub_feature_code ??
    (groups.custom as Record<string, unknown> | undefined)?.feature_code ??
    (groups.custom as Record<string, unknown> | undefined)?.category_code ??
    "";

  return (
    <aside
      className={`log-inspector ${fullscreen ? "is-fullscreen" : ""}`}
      aria-label="Raw log inspector"
    >
      <header className="log-inspector__toolbar">
        <div>
          <strong>Log inspector</strong>
        </div>
        <Segmented
          size="small"
          value={view}
          onChange={(value) => setView(value as "sections" | "raw")}
          options={[
            { value: "sections", label: "Sections" },
            { value: "raw", label: "Raw JSON" },
          ]}
        />
        <Tooltip title="ขยายเต็มจอ">
          <Button
            type="text"
            icon={<FullscreenOutlined />}
            onClick={() => onFullscreen(!fullscreen)}
          />
        </Tooltip>
        <Tooltip title="ปิด">
          <Button
            type="text"
            icon={<CloseOutlined />}
            onClick={onClose}
            aria-label="Close inspector"
          />
        </Tooltip>
      </header>
      <div className="log-inspector__body">
        {(isUnclassified || (isMapped && !record.category_id)) && (
          <Alert
            type="warning"
            showIcon
            icon={<AlertOutlined />}
            className="mb-4"
            message={
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                <span className="text-xs">
                  Log นี้ยังไม่ได้ถูกจับคู่ Feature/Category (หรือยังไม่มี
                  Feature ในระบบ)
                </span>
                <div className="flex items-center gap-2">
                  <Button
                    size="small"
                    type="primary"
                    ghost
                    icon={<PlusOutlined />}
                    onClick={() => setIsAddFeatureOpen(true)}
                  >
                    เพิ่ม Feature / Category
                  </Button>
                  {!isMapped && (
                    <Button
                      size="small"
                      onClick={() => setIsRouteMappingOpen(true)}
                    >
                      Map Route
                    </Button>
                  )}
                </div>
              </div>
            }
          />
        )}
        {view === "raw" ? (
          tree(record, "$")
        ) : (
          <Tabs tabPosition="top" items={visibleItems} />
        )}
      </div>

      <AddFeatureModal
        open={isAddFeatureOpen}
        productId={
          productId ||
          (record.product_id ? Number(record.product_id) : undefined)
        }
        projectId={record.project_id ? Number(record.project_id) : undefined}
        initialCode={
          typeof customFeatureCode === "string" ? customFeatureCode : undefined
        }
        onClose={() => setIsAddFeatureOpen(false)}
        onSuccess={() => {
          void client.invalidateQueries({ queryKey: ["logs"] });
        }}
      />
      <LogRouteMappingModal
        open={isRouteMappingOpen}
        record={record}
        productId={
          productId ||
          (record.product_id ? Number(record.product_id) : undefined)
        }
        onClose={() => setIsRouteMappingOpen(false)}
      />
    </aside>
  );
}
