import { Button, Divider, Input, InputNumber, Select, Tag } from "antd";
import { ClearOutlined, SearchOutlined } from "@ant-design/icons";
import type {
  ProductEnvironmentOption,
  ProductFeatureOption,
  ProductOption,
  ProductProjectOption,
} from "@/services/product.service";
import type { SearchRule } from "@/features/log-explorer/services/log-search.service";
import SelectProductField from "./SelectProductField";

interface SearchPanelProps {
  query: string;
  productId?: number;
  projectId?: number;
  environmentId?: number;
  categoryId?: number;
  subFeatureId?: number;
  products: ProductOption[];
  projects?: ProductProjectOption[];
  environments?: ProductEnvironmentOption[];
  categories?: ProductFeatureOption[];
  subFeatures?: ProductFeatureOption[];
  timeRange: string;
  rules: SearchRule[];
  loading: boolean;
  dirty: boolean;
  onQueryChange: (value: string) => void;
  onProductChange: (value: number) => void;
  onProjectChange: (value?: number) => void;
  onEnvironmentChange: (value?: number) => void;
  onCategoryChange: (value?: number) => void;
  onSubFeatureChange: (value?: number) => void;
  onTimeRangeChange: (value: string) => void;
  onRulesChange: (rules: SearchRule[]) => void;
  onSearch: () => void;
  onClear: () => void;
}

const levelOptions = [
  "TRACE",
  "DEBUG",
  "INFO",
  "NOTICE",
  "WARN",
  "WARNING",
  "ERROR",
  "CRITICAL",
  "PANIC",
  "FATAL",
];
const getRule = (rules: SearchRule[], field: string) =>
  rules.find((rule) => rule.field === field);

export function SearchPanel(props: SearchPanelProps) {
  const levelRule = props.rules.find((rule) =>
    rule.field.includes("log_level"),
  );
  const statusRule = getRule(props.rules, "payload.status_code");
  const project = props.projects?.find(
    (item) => item.project_id === props.projectId,
  );
  const environment = props.environments?.find(
    (item) => item.environment_id === props.environmentId,
  );
  const category = props.categories?.find(
    (item) => item.category_id === props.categoryId,
  );
  const subFeature = props.subFeatures?.find(
    (item) => item.category_id === props.subFeatureId,
  );

  const setRule = (
    field: string,
    label: string,
    type: SearchRule["type"],
    value?: string | number | null,
  ) => {
    const rest = props.rules.filter((rule) => rule.field !== field);
    const normalized = typeof value === "string" ? value.trim() : value;
    if (normalized === undefined || normalized === null || normalized === "") {
      props.onRulesChange(rest);
      return;
    }
    props.onRulesChange([
      ...rest,
      {
        id: `${field}-${String(normalized)}`,
        field,
        label,
        type,
        operator: "eq",
        value: normalized,
        enabled: true,
      },
    ]);
  };

  const activeFilters = [
    props.query.trim()
      ? {
          key: "query",
          label: `Search: ${props.query.trim()}`,
          remove: () => props.onQueryChange(""),
        }
      : undefined,
    props.environmentId
      ? {
          key: "environment",
          label: `Environment: ${environment?.environment_name ?? props.environmentId}`,
          remove: () => props.onEnvironmentChange(undefined),
        }
      : undefined,
    props.projectId
      ? {
          key: "project",
          label: `Project: ${project?.project_name ?? props.projectId}`,
          remove: () => props.onProjectChange(undefined),
        }
      : undefined,
    props.categoryId
      ? {
          key: "category",
          label: `Feature: ${category?.category_name ?? props.categoryId}`,
          remove: () => props.onCategoryChange(undefined),
        }
      : undefined,
    props.subFeatureId
      ? {
          key: "sub-feature",
          label: `Sub-feature: ${subFeature?.category_name ?? props.subFeatureId}`,
          remove: () => props.onSubFeatureChange(undefined),
        }
      : undefined,
    props.timeRange !== "all"
      ? {
          key: "time",
          label: `Time: ${props.timeRange}`,
          remove: () => props.onTimeRangeChange("all"),
        }
      : undefined,
    ...props.rules
      .filter((rule) => rule.enabled)
      .map((rule) => ({
        key: rule.field,
        label: `${rule.label}: ${String(rule.value)}`,
        remove: () =>
          props.onRulesChange(
            props.rules.filter((item) => item.field !== rule.field),
          ),
      })),
  ].filter((item): item is { key: string; label: string; remove: () => void } =>
    Boolean(item),
  );

  return (
    <aside className="log-search-panel" aria-label="Search and filters">
      <div className="log-search-panel__heading">
        {props.dirty && <Tag color="gold">ยังไม่ถูกค้นหา</Tag>}
      </div>
      <label className="log-search-field">
        <span>ค้นหา</span>
        <Input
          size="large"
          prefix={<SearchOutlined />}
          value={props.query}
          onChange={(event) => props.onQueryChange(event.target.value)}
          onPressEnter={props.onSearch}
          placeholder="ค้นหา Log เช่น timeout"
          aria-label="Search logs"
          allowClear
        />
      </label>

      <Divider titlePlacement="start" plain>
        Scope
      </Divider>
      <SelectProductField
        productId={props.productId}
        environmentId={props.environmentId}
        projectId={props.projectId}
        categoryId={props.categoryId}
        subFeatureId={props.subFeatureId}
        products={props.products}
        projects={props.projects}
        environments={props.environments}
        categories={props.categories}
        subFeatures={props.subFeatures}
        onProductChange={props.onProductChange}
        onEnvironmentChange={props.onEnvironmentChange}
        onProjectChange={props.onProjectChange}
        onCategoryChange={props.onCategoryChange}
        onSubFeatureChange={props.onSubFeatureChange}
      />

      <Divider titlePlacement="start" plain>
        Time & details
      </Divider>
      <div className="log-filter-stack">
        <label>
          ช่วงเวลา
          <Select
            value={props.timeRange}
            onChange={props.onTimeRangeChange}
            options={[
              { value: "15m", label: "15 นาที" },
              { value: "1h", label: "1 ชั่วโมง" },
              { value: "24h", label: "24 ชั่วโมง" },
              { value: "7d", label: "7 วัน" },
              { value: "all", label: "ทั้งหมด" },
            ]}
          />
        </label>
        <label>
          Log level
          <Select
            allowClear
            placeholder="All levels"
            value={levelRule?.value as string | undefined}
            onChange={(value) =>
              setRule("payload.log_level", "Level", "keyword", value)
            }
            options={levelOptions.map((level) => ({
              value: level,
              label: level,
            }))}
          />
        </label>
        <label>
          Status code
          <InputNumber
            min={100}
            max={599}
            precision={0}
            controls={false}
            value={
              typeof statusRule?.value === "number"
                ? statusRule.value
                : undefined
            }
            onChange={(value) =>
              setRule("payload.status_code", "Status", "number", value)
            }
            placeholder="e.g. 404, 500"
          />
        </label>
      </div>

      {activeFilters.length > 0 && (
        <div className="log-active-filters">
          <div>
            <span>ตัวกรองที่เลือก</span>
            <small>{activeFilters.length}</small>
          </div>
          <div>
            {activeFilters.map((filter) => (
              <Tag
                key={filter.key}
                closable
                onClose={(event) => {
                  event.preventDefault();
                  filter.remove();
                }}
              >
                {filter.label}
              </Tag>
            ))}
          </div>
        </div>
      )}
      <div className="log-search-panel__actions">
        <Button
          icon={<ClearOutlined />}
          disabled={!activeFilters.length}
          onClick={props.onClear}
        >
          รีเซ็ต
        </Button>
        <Button
          type="primary"
          icon={<SearchOutlined />}
          loading={props.loading}
          disabled={!props.productId}
          onClick={props.onSearch}
        >
          {props.dirty ? "ใช้ตัวกรอง" : "ค้นหา"}
        </Button>
      </div>
    </aside>
  );
}
