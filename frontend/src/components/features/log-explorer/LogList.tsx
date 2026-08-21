/* eslint-disable react-hooks/refs -- dnd-kit sortable exposes ref callback bindings during render. */
import { memo, useEffect, useMemo, useRef, useState } from "react";
import type { CSSProperties } from "react";
import {
  Button,
  Checkbox,
  Empty,
  Popover,
  Select,
  Skeleton,
  Tag,
  Tooltip,
} from "antd";
import {
  AppstoreOutlined,
  HolderOutlined,
  StarFilled,
  MenuFoldOutlined,
  FilterOutlined,
} from "@ant-design/icons";
import {
  DndContext,
  PointerSensor,
  closestCenter,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import type { DragEndEvent } from "@dnd-kit/core";
import {
  SortableContext,
  arrayMove,
  verticalListSortingStrategy,
  useSortable,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import type { FavoriteField } from "@/features/log-explorer/services/custom-field-favorite.service";
import type { SearchRecord } from "@/features/log-explorer/services/log-search.service";
import {
  getCustomFieldValue,
  getLogErrorInfo,
  getRecordId,
  getValue,
} from "@/features/log-explorer/utils/log-utils";
import {
  favoriteFieldDefinitionId,
  favoriteLogColumnKey,
  isFavoriteLogColumn,
  useLogExplorerStore,
} from "@/features/log-explorer/store/useLogExplorerStore";
import type {
  BuiltInLogColumn,
  LogColumn,
} from "@/features/log-explorer/store/useLogExplorerStore";

const builtInLabels: Record<BuiltInLogColumn, string> = {
  timestamp: "Timestamp",
  level: "Level",
  product: "Product",
  hierarchy: "Hierarchy",
  message: "Message",
};
const builtInColumns = Object.keys(builtInLabels) as BuiltInLogColumn[];
const getFavoriteLabel = (field: FavoriteField) =>
  field.display_name ?? field.field_key;
const getColumnLabel = (column: LogColumn, favorites: FavoriteField[]) => {
  if (!isFavoriteLogColumn(column)) return builtInLabels[column];
  const field = favorites.find(
    (item) => item.field_definition_id === favoriteFieldDefinitionId(column),
  );
  return field ? getFavoriteLabel(field) : "Custom field";
};

//  LogList
const getColumnWidth = (column: LogColumn) => {
  if (column === "timestamp") return "156px";
  if (column === "level") return "76px";
  if (column === "product") return "112px";
  if (column === "hierarchy") return "minmax(210px,1.1fr)";
  if (column === "message") return "minmax(240px,1.4fr)";
  return "minmax(160px,.8fr)";
};

const formatTimestamp = (raw: string, timeZone: string = "Asia/Bangkok") => {
  if (!raw || raw === "—" || raw === "-") return "—";
  try {
    const date = new Date(raw);
    if (isNaN(date.getTime())) return raw;

    const targetTz =
      timeZone === "local"
        ? Intl.DateTimeFormat().resolvedOptions().timeZone
        : timeZone;

    return date
      .toLocaleString("sv-SE", { timeZone: targetTz })
      .replace("T", " ");
  } catch {
    return raw;
  }
};

const renderTimestampCell = (
  value: string,
  query: string,
  timeZone: string,
) => {
  if (!value || value === "—") return "—";
  const formatted = formatTimestamp(value, timeZone);
  const bkkFormatted = formatTimestamp(value, "Asia/Bangkok");
  const utcFormatted = formatTimestamp(value, "UTC");

  return (
    <Tooltip
      title={
        <div style={{ fontSize: "12px", lineHeight: "1.6" }}>
          <div>
            <strong>Bangkok (GMT+7):</strong> {bkkFormatted}
          </div>
          <div>
            <strong>UTC:</strong> {utcFormatted}
          </div>
          <div style={{ color: "#9ca3af", marginTop: "2px" }}>
            <strong>Raw:</strong> {value}
          </div>
        </div>
      }
      placement="top"
    >
      <span>
        <Highlight text={formatted} query={query} />
      </span>
    </Tooltip>
  );
};

const renderHierarchyCell = (value: string, query: string) => {
  if (!value || value === "—") return "—";
  const segments = value.split(/\s*[/>]\s*/).filter(Boolean);
  if (segments.length <= 1) {
    return <Highlight text={value} query={query} />;
  }
  return (
    <Tooltip title={`Hierarchy: ${value}`} placement="top">
      <span className="log-hierarchy-breadcrumbs">
        {segments.map((seg, idx) => (
          <span key={idx} className="log-hierarchy-crumb-wrap">
            {idx > 0 && <span className="log-hierarchy-sep">›</span>}
            <span
              className={`log-hierarchy-crumb ${
                idx === segments.length - 1 ? "is-leaf" : ""
              }`}
            >
              <Highlight text={seg} query={query} />
            </span>
          </span>
        ))}
      </span>
    </Tooltip>
  );
};

interface LogListProps {
  records: SearchRecord[];
  total: number;
  loading: boolean;
  hasMore: boolean;
  query: string;
  activeId: string | null;
  selectedIds: string[];
  columns: LogColumn[];
  favorites: FavoriteField[];
  onSelect: (id: string) => void;
  onToggle: (id: string) => void;
  onColumnsChange: (columns: LogColumn[]) => void;
  onFavoriteReorder: (fieldDefinitionIds: number[]) => void;
  onLoadMore: () => void;
  onOpen: () => void;
}

const Highlight = ({ text, query }: { text: string; query: string }) => {
  if (!query.trim()) return text;
  const index = text.toLowerCase().indexOf(query.toLowerCase());
  return index < 0 ? (
    text
  ) : (
    <>
      {text.slice(0, index)}
      <mark>{text.slice(index, index + query.length)}</mark>
      {text.slice(index + query.length)}
    </>
  );
};

function SortableColumnItem({
  column,
  label,
  selected,
  onToggle,
}: {
  column: LogColumn;
  label: string;
  selected: boolean;
  onToggle: (selected: boolean) => void;
}) {
  const sortable = useSortable({ id: column });
  const style: CSSProperties = {
    transform: CSS.Transform.toString(sortable.transform),
    transition: sortable.transition,
    opacity: sortable.isDragging ? 0.55 : 1,
  };
  return (
    <div
      ref={sortable.setNodeRef}
      style={style}
      className="log-column-chooser__item"
    >
      <button
        type="button"
        className="log-column-chooser__drag"
        ref={sortable.setActivatorNodeRef}
        {...sortable.attributes}
        {...sortable.listeners}
        aria-label={`Drag ${label}`}
      >
        <HolderOutlined />
      </button>
      <Checkbox
        checked={selected}
        onChange={(event) => onToggle(event.target.checked)}
      >
        {label}
      </Checkbox>
    </div>
  );
}

function ColumnChooser({
  columns,
  favorites,
  onColumnsChange,
  onFavoriteReorder,
}: {
  columns: LogColumn[];
  favorites: FavoriteField[];
  onColumnsChange: (columns: LogColumn[]) => void;
  onFavoriteReorder: (fieldDefinitionIds: number[]) => void;
}) {
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
  );
  const availableColumns = useMemo(() => {
    const favoriteColumns = favorites.map((field) =>
      favoriteLogColumnKey(field.field_definition_id),
    );
    return [
      ...columns,
      ...builtInColumns.filter((column) => !columns.includes(column)),
      ...favoriteColumns.filter((column) => !columns.includes(column)),
    ];
  }, [columns, favorites]);
  const onDragEnd = ({ active, over }: DragEndEvent) => {
    if (!over || active.id === over.id) return;
    const oldIndex = availableColumns.indexOf(active.id as LogColumn);
    const newIndex = availableColumns.indexOf(over.id as LogColumn);
    if (oldIndex < 0 || newIndex < 0) return;
    const next = arrayMove(availableColumns, oldIndex, newIndex);
    onColumnsChange(next.filter((column) => columns.includes(column)));
    onFavoriteReorder(
      next.filter(isFavoriteLogColumn).map(favoriteFieldDefinitionId),
    );
  };
  const toggleColumn = (column: LogColumn, selected: boolean) =>
    onColumnsChange(
      selected
        ? [...columns, column]
        : columns.filter((item) => item !== column),
    );
  return (
    <div className="log-column-chooser" aria-label="Choose and reorder columns">
      <div className="log-column-chooser__title">
        <strong>Displayed columns</strong>
        <span>Drag to reorder</span>
      </div>
      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragEnd={onDragEnd}
      >
        <SortableContext
          items={availableColumns}
          strategy={verticalListSortingStrategy}
        >
          <div className="log-column-chooser__list">
            {availableColumns.map((column) => (
              <SortableColumnItem
                key={column}
                column={column}
                label={getColumnLabel(column, favorites)}
                selected={columns.includes(column)}
                onToggle={(selected) => toggleColumn(column, selected)}
              />
            ))}
          </div>
        </SortableContext>
      </DndContext>
    </div>
  );
}

const LogRow = memo(function LogRow({
  record,
  index,
  active,
  columns,
  favorites,
  query,
  timeZone,
  onSelect,
  onOpen,
}: {
  record: SearchRecord;
  index: number;
  active: boolean;
  columns: LogColumn[];
  favorites: FavoriteField[];
  query: string;
  timeZone: string;
  onSelect: () => void;
  onOpen: () => void;
}) {
  const errorInfo = getLogErrorInfo(record);
  const messageVal = getValue(
    record,
    "message",
    "msg",
    "log",
    "text",
    "description",
    "error",
    "error_message",
    "payload.message",
    "data.message",
  );
  let levelVal = getValue(
    record,
    "level",
    "log_level",
    "severity",
    "logType",
    "payload.log_level",
    "payload.level",
    "data.level",
    "data.log_level",
  );
  if (levelVal === "—" && messageVal !== "—") {
    const match = messageVal.match(
      /^(CRITICAL|FATAL|ERROR|WARN|WARNING|INFO|DEBUG|TRACE):?/i,
    );
    if (match) {
      levelVal = match[1].toUpperCase();
      if (levelVal === "WARNING") levelVal = "WARN";
    }
  }

  const values: Record<string, string> = {
    timestamp: getValue(
      record,
      "@timestamp",
      "timestamp",
      "created_at",
      "received_at",
    ),
    level: levelVal,
    product: getValue(
      record,
      "product_name",
      "product_code",
      "detected_product_code",
      "product",
      "service",
      "app_name",
      "app",
      "payload.product_code",
      "data.product_code",
      "product_id",
    ),
    hierarchy: getValue(
      record,
      "feature_full_path",
      "payload.feature_full_path",
      "data.feature_full_path",
      "project_name",
      "payload.project_name",
      "project_id",
      "payload.project_id",
      "metadata.actor.project_id",
      "payload.metadata.actor.project_id",
      "actor.project_id",
      "payload.actor.project_id",
      "custom_fields.project_id",
      "payload.custom_fields.project_id",
    ),
    message: messageVal,
  };
  for (const field of favorites)
    values[favoriteLogColumnKey(field.field_definition_id)] =
      getCustomFieldValue(record, field.field_path ?? field.field_key);
  const gridStyle = {
    "--log-grid-template": ["40px", ...columns.map(getColumnWidth)].join(" "),
  } as CSSProperties;

  return (
    <div
      className={`log-row ${active ? "is-active" : ""}`}
      style={gridStyle}
      role="row"
      tabIndex={active || index === 0 ? 0 : -1}
      onClick={onSelect}
      onDoubleClick={onOpen}
      aria-selected={active}
    >
      <div role="cell" className="log-select-cell"></div>
      {columns.map((column) => {
        const value = values[column] ?? "—";
        return (
          <div
            key={column}
            role="cell"
            className={`log-cell ${isFavoriteLogColumn(column) ? "log-cell--custom" : `log-cell--${column}`}`}
            title={
              column === "timestamp" || column === "hierarchy"
                ? undefined
                : value
            }
          >
            {column === "timestamp" ? (
              renderTimestampCell(value, query, timeZone)
            ) : column === "level" ? (
              errorInfo.severity === "warn" ? (
                <Tag color="#D4B106" style={{ margin: 0, fontWeight: 700 }}>
                  {value !== "—" ? value : "WARN"}
                </Tag>
              ) : errorInfo.severity === "error" ? (
                <Tag color="#FF3D89" style={{ margin: 0, fontWeight: 700 }}>
                  {value !== "—" ? value : "ERROR"}
                </Tag>
              ) : value !== "—" ? (
                <Tag color="#1890FF" style={{ margin: 0, fontWeight: 600 }}>
                  {value}
                </Tag>
              ) : (
                "—"
              )
            ) : column === "message" ? (
              <Highlight text={value} query={query} />
            ) : column === "hierarchy" ? (
              renderHierarchyCell(value, query)
            ) : (
              value
            )}
          </div>
        );
      })}
    </div>
  );
});

export function LogList(props: LogListProps) {
  const sentinel = useRef<HTMLDivElement>(null);
  const [focused, setFocused] = useState(0);
  const [timeZone, setTimeZone] = useState<string>(() => {
    try {
      return localStorage.getItem("omnilogs_timezone") || "Asia/Bangkok";
    } catch {
      return "Asia/Bangkok";
    }
  });
  const explorer = useLogExplorerStore();
  const { hasMore, onLoadMore } = props;
  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => entry.isIntersecting && hasMore && onLoadMore(),
      { rootMargin: "240px" },
    );
    if (sentinel.current) observer.observe(sentinel.current);
    return () => observer.disconnect();
  }, [hasMore, onLoadMore]);
  const gridStyle = {
    "--log-grid-template": ["40px", ...props.columns.map(getColumnWidth)].join(
      " ",
    ),
  } as CSSProperties;
  const onColumnsOrderChange = (nextColumns: LogColumn[]) => {
    props.onColumnsChange(nextColumns);
  };
  return (
    <section
      className="log-list"
      aria-label="Log results"
      onKeyDown={(event) => {
        if (!props.records.length) return;
        if (event.key === "ArrowDown" || event.key === "ArrowUp") {
          event.preventDefault();
          const next = Math.max(
            0,
            Math.min(
              props.records.length - 1,
              focused + (event.key === "ArrowDown" ? 1 : -1),
            ),
          );
          setFocused(next);
          props.onSelect(getRecordId(props.records[next], next));
        }
        if (event.key === "Enter") props.onOpen();
      }}
    >
      <header className="log-list__toolbar">
        <div style={{ display: "flex", alignItems: "center", gap: "12px" }}>
          <Button
            icon={
              explorer.filtersOpen ? <MenuFoldOutlined /> : <FilterOutlined />
            }
            onClick={() => explorer.setFiltersOpen(!explorer.filtersOpen)}
            aria-label="Toggle filters"
          >
            {explorer.filtersOpen ? "ซ่อนตัวกรอง" : "แสดงตัวกรอง"}
          </Button>

          <div>
            <strong>{props.total.toLocaleString()}</strong>
            <span> logs</span>
          </div>
        </div>
        <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
          <Select
            value={timeZone}
            onChange={(val) => {
              setTimeZone(val);
              try {
                localStorage.setItem("omnilogs_timezone", val);
              } catch {
                // Ignore storage error
              }
            }}
            options={[
              { value: "Asia/Bangkok", label: "GMT+7 (Bangkok)" },
              { value: "UTC", label: "UTC (+00:00)" },
            ]}
            aria-label="Select timezone"
            style={{ width: 155 }}
          />
          <Popover
            trigger="click"
            placement="bottomRight"
            content={
              <ColumnChooser
                columns={props.columns}
                favorites={props.favorites}
                onColumnsChange={onColumnsOrderChange}
                onFavoriteReorder={props.onFavoriteReorder}
              />
            }
          >
            <Button icon={<AppstoreOutlined />}>Columns</Button>
          </Popover>
        </div>
      </header>
      <div className="log-table" role="table">
        <div className="log-table__header" style={gridStyle} role="row">
          <div className="log-select-cell" />
          {props.columns.map((column) => (
            <div
              key={column}
              role="columnheader"
              className={`log-cell ${isFavoriteLogColumn(column) ? "log-cell--custom" : `log-cell--${column}`}`}
            >
              {isFavoriteLogColumn(column) && (
                <StarFilled className="log-column-star" />
              )}{" "}
              {getColumnLabel(column, props.favorites)}
            </div>
          ))}
        </div>
        <div className="log-table__body">
          {props.records.map((record, index) => {
            const id = getRecordId(record, index);
            return (
              <LogRow
                key={id}
                record={record}
                index={index}
                active={props.activeId === id}
                columns={props.columns}
                favorites={props.favorites}
                query={props.query}
                timeZone={timeZone}
                onSelect={() => props.onSelect(id)}
                onOpen={props.onOpen}
              />
            );
          })}
          {!props.loading && !props.records.length && (
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              description="No logs found for this search"
            />
          )}
          {props.loading && (
            <div className="log-list__loading">
              <Skeleton active paragraph={{ rows: 3 }} />
            </div>
          )}
          <div ref={sentinel} className="log-list__sentinel" />
        </div>
      </div>
    </section>
  );
}
