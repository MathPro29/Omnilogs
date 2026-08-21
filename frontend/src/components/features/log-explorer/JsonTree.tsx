import { memo, useMemo, useState } from "react";
import { Button, Input, message } from "antd";
import {
  CopyOutlined,
  SearchOutlined,
  FilterOutlined,
  StarFilled,
  StarOutlined,
} from "@ant-design/icons";
import { maskSensitive } from "@/features/log-explorer/utils/log-utils";

const normalPath = (path: string) => {
  let p = path.trim();
  p = p.replace(/^\$\.?/, "");
  p = p.replace(/^raw\./, "");
  p = p.replace(/^(data|fields|payload)\./, "");
  return p || "$";
};

interface JsonTreeProps {
  value: unknown;
  rootPath?: string;
  initialExpandedAll?: boolean | null;
  favoritePaths?: string[];
  favoriteBusy?: boolean;
  busyPath?: string | null;
  onFilter?: (path: string, value: unknown, exclude: boolean) => void;
  onFavorite?: (path: string, name: string, value: unknown) => void;
}
const copy = async (value: string, label: string) => {
  await navigator.clipboard.writeText(value);
  message.success(`${label} copied`);
};
const JsonNode = memo(function JsonNode({
  name,
  value,
  path,
  depth,
  search,
  expandedAll,
  favoritePaths,
  favoriteBusy,
  busyPath,
  onFilter,
  onFavorite,
}: {
  name: string;
  value: unknown;
  path: string;
  depth: number;
  search: string;
  expandedAll: boolean | null;
  favoritePaths: string[];
  favoriteBusy?: boolean;
  busyPath?: string | null;
  onFilter?: JsonTreeProps["onFilter"];
  onFavorite?: JsonTreeProps["onFavorite"];
}) {
  const [open, setOpen] = useState(depth < 2);
  const isObject = value !== null && typeof value === "object";
  const visible = expandedAll ?? open;
  const display = typeof value === "string" ? `"${value}"` : String(value);
  const normalizedNodePath = normalPath(path);
  const favorite = favoritePaths.includes(normalizedNodePath);
  const nodeBusy = busyPath ? busyPath === normalizedNodePath : Boolean(favoriteBusy);
  const match = Boolean(
    search && `${name} ${display}`.toLowerCase().includes(search.toLowerCase()),
  );
  return (
    <div
      className={`json-node ${match ? "is-match" : ""}`}
      style={{ paddingLeft: depth * 14 }}
    >
      <div className="json-node__line">
        {isObject ? (
          <button
            type="button"
            className="json-toggle"
            onClick={() => setOpen((current) => !current)}
            aria-label={`${visible ? "Collapse" : "Expand"} ${name}`}
          >
            {visible ? "-" : "+"}
          </button>
        ) : (
          <span className="json-toggle-placeholder" />
        )}
        <button
          type="button"
          className="json-key"
          onClick={() => isObject && setOpen((current) => !current)}
        >
          {favorite && <StarFilled className="json-favorite" />} {name}
        </button>
        {!isObject && (
          <>
            <span className="json-colon">:</span>
            <span className={`json-value json-value--${typeof value}`}>
              {display}
            </span>
            <Button
              type="text"
              size="small"
              icon={<FilterOutlined />}
              title="Add to filter"
              onClick={() => onFilter?.(path, value, false)}
              style={{ marginLeft: 8, color: '#1677ff' }}
            />
            {onFavorite && (
              <Button type="text" size="small" icon={favorite ? <StarFilled /> : <StarOutlined />} title={favorite ? "Remove from favorites" : "Add to favorite"} aria-label={favorite ? ("Remove " + name + " from favorites") : ("Add " + name + " to favorites")} loading={nodeBusy} onClick={() => onFavorite(path, name, value)} style={{ marginLeft: 4, color: favorite ? "#faad14" : undefined }} />
            )}
          </>
        )}
        {isObject && (
          <span className="json-count">
            {Array.isArray(value)
              ? `[${value.length}]`
              : `{${Object.keys(value as object).length}}`}
          </span>
        )}
      </div>
      {isObject &&
        visible &&
        Object.entries(value as Record<string, unknown>).map(([key, child]) => (
          <JsonNode
            key={key}
            name={key}
            value={child}
            path={Array.isArray(value) ? `${path}[${key}]` : `${path}.${key}`}
            depth={depth + 1}
            search={search}
            expandedAll={expandedAll}
            favoritePaths={favoritePaths}
            favoriteBusy={favoriteBusy}
            busyPath={busyPath}
            onFilter={onFilter}
            onFavorite={onFavorite}
          />
        ))}
    </div>
  );
});
export function JsonTree({
  value,
  rootPath = "$",
  initialExpandedAll = null,
  favoritePaths = [],
  favoriteBusy = false,
  busyPath = null,
  onFilter,
  onFavorite,
}: JsonTreeProps) {
  const [search, setSearch] = useState("");
  const [expandedAll, setExpandedAll] = useState<boolean | null>(
    initialExpandedAll,
  );
  const masked = useMemo(() => maskSensitive(value), [value]);
  return (
    <div className="json-tree">
      <div className="json-tree__toolbar">
        <Input
          size="small"
          prefix={<SearchOutlined />}
          placeholder="Find in JSON"
          value={search}
          onChange={(event) => setSearch(event.target.value)}
          allowClear
        />
        <Button size="small" onClick={() => setExpandedAll(true)}>
          Expand all
        </Button>
        <Button size="small" onClick={() => setExpandedAll(false)}>
          Collapse all
        </Button>
        <Button
          size="small"
          icon={<CopyOutlined />}
          onClick={() => void copy(JSON.stringify(masked, null, 2), "JSON")}
        >
          Copy
        </Button>
      </div>
      <div className="json-tree__content" role="tree">
        <JsonNode
          name="root"
          value={masked}
          path={rootPath}
          depth={0}
          search={search}
          expandedAll={expandedAll}
          favoritePaths={favoritePaths}
          favoriteBusy={favoriteBusy}
          busyPath={busyPath}
          onFilter={onFilter}
          onFavorite={onFavorite}
        />
      </div>
    </div>
  );
}
