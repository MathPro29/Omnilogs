import { useState } from "react";
import { InputNumber, Radio, Select, Space, Button } from "antd";
import type { ProductOption } from "@/services/product.service";
import { DatePicker, type DateRange } from "@/components";

interface SearchScopeControlsProps {
  productId?: number;
  products: ProductOption[];
  isLoadingProducts: boolean;
  hasProductError: boolean;
  timeRange: string;
  logic: "AND" | "OR";
  onProductChange: (value: number) => void;
  onTimeRangeChange: (value: string) => void;
  onLogicChange: (value: "AND" | "OR") => void;
  onSearch?: () => void;
  isSearching?: boolean;
  searchDisabled?: boolean;
}

export function SearchScopeControls({
  productId,
  products,
  isLoadingProducts,
  hasProductError,
  timeRange,
  logic,
  onProductChange,
  onTimeRangeChange,
  onLogicChange,
  onSearch,
  isSearching,
  searchDisabled,
}: SearchScopeControlsProps) {
  const isCustomMin =
    timeRange.endsWith("m") &&
    !["1m", "5m", "10m", "15m", "30m"].includes(timeRange);
  const [customMin, setCustomMin] = useState<number>(
    parseInt(timeRange, 10) || 10,
  );

  const [dateRange, setDateRange] = useState<DateRange>(null);

  const handleDateRangeChange = (value: DateRange) => {
    setDateRange(value);
    if (value?.[0]) {
      onTimeRangeChange(value[0].toISOString());
    }
  };

  const handleCustomChange = (val: number | null) => {
    const min = val && val > 0 ? val : 1;
    setCustomMin(min);
    onTimeRangeChange(`${min}m`);
  };

  return (
    <div className="search-builder-scope">
      <div className="scope-item">
        <label>Product</label>
        <Select
          value={productId}
          onChange={onProductChange}
          loading={isLoadingProducts}
          status={hasProductError ? "error" : undefined}
          placeholder="Select Product"
          notFoundContent={
            hasProductError ? "Failed to load products" : undefined
          }
          optionFilterProp="label"
          showSearch
          style={{ minWidth: 170 }}
          options={products.map((product) => ({
            label: product.name,
            value: product.id,
          }))}
        />
      </div>
      <div className="scope-item">
        <label>Time Range</label>
        <Space>
          <Select
            value={isCustomMin ? "custom" : timeRange}
            onChange={(val) => {
              if (val === "custom") {
                onTimeRangeChange(`${customMin}m`);
              } else {
                onTimeRangeChange(val);
              }
            }}
            style={{ minWidth: 180 }}
            options={[
              { label: "Every 1 minute (1m)", value: "1m" },
              { label: "Every 5 minutes (5m)", value: "5m" },
              { label: "Every 10 minutes (10m)", value: "10m" },
              { label: "Every 15 minutes (15m)", value: "15m" },
              { label: "Every 30 minutes (30m)", value: "30m" },
              { label: "Every 1 hour (1h)", value: "1h" },
              { label: "Every 24 hours (24h)", value: "24h" },
              { label: "Every 7 days (7d)", value: "7d" },
              { label: "Custom (Every n mins)", value: "custom" },
              { label: "All Logs", value: "all" },
            ]}
          />
          {isCustomMin && (
            <InputNumber
              min={1}
              max={10080}
              value={customMin}
              addonAfter="mins"
              onChange={handleCustomChange}
              style={{ width: 120 }}
            />
          )}
        </Space>
      </div>
      <div className="scope-item">
        <label>Date Range</label>
        <DatePicker.RangePicker
          value={dateRange}
          onChange={handleDateRangeChange}
          style={{ minWidth: 320 }}
        />
      </div>
      <div className="scope-item">
        <label>Match Logic</label>
        <Radio.Group
          value={logic}
          onChange={(event) =>
            onLogicChange(event.target.value as "AND" | "OR")
          }
          options={[
            { label: "Match All (AND)", value: "AND" },
            { label: "Match Any (OR)", value: "OR" },
          ]}
          optionType="button"
        />
      </div>
      {onSearch && (
        <div className="scope-item ml-auto align-self-end">
          <label className="invisible">Search</label>
          <Button
            type="primary"
            disabled={searchDisabled}
            loading={isSearching}
            onClick={onSearch}
            className="h-[32px] px-5 font-medium rounded-lg"
          >
            Search Logs
          </Button>
        </div>
      )}
    </div>
  );
}
