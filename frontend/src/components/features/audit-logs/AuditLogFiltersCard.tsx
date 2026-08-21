import { Button, Card as AntCard, Col, Flex, Input, Row, Select } from "antd";
import { SearchOutlined } from "@ant-design/icons";
import { DatePicker, type DateRange } from "@/components";
import type { ProductOption } from "@/services/product.service";

export type { DateRange };

export interface FilterCardProps {
  keywordInput: string;
  products: ProductOption[];
  productId?: number;
  activityOptions: string[];
  action?: string;
  resourceType?: string;
  result?: string;
  dateRange: DateRange;
  setKeywordInput: (value: string) => void;
  setProductId: (value?: number) => void;
  setKeyword: (value: string) => void;
  setAction: (value?: string) => void;
  setResourceType: (value?: string) => void;
  setResult: (value?: string) => void;
  setDateRange: (value: DateRange) => void;
  setPage: (value: number) => void;
  resetFilters: () => void;
}

function readableAction(action: string) {
  return action
    .replaceAll("_", " ")
    .toLowerCase()
    .replace(/\b\w/g, (letter) => letter.toUpperCase());
}

const AuditLogFiltersCard = ({
  keywordInput,
  activityOptions,
  action,
  result,
  dateRange,
  setKeywordInput,
  setKeyword,
  setAction,
  setResult,
  setDateRange,
  setPage,
  resetFilters,
}: FilterCardProps) => {
  return (
    <AntCard className="audit-filter-card" bordered={false}>
      <Row gutter={[12, 12]}>
        <Col xs={24} md={12} lg={6}>
          <Input
            allowClear
            prefix={<SearchOutlined />}
            placeholder="ค้นหา Activity, Path, Resource ID..."
            value={keywordInput}
            onChange={(e) => setKeywordInput(e.target.value)}
            onPressEnter={() => {
              setKeyword(keywordInput.trim());
              setPage(1);
            }}
          />
        </Col>
        <Col xs={12} sm={8} md={6} lg={4}>
          <Select
            allowClear
            showSearch
            placeholder="Activity"
            value={action}
            onChange={(value) => {
              setAction(value);
              setPage(1);
            }}
            options={activityOptions.map((value) => ({
              value,
              label: readableAction(value),
            }))}
          />
        </Col>
        <Col xs={12} sm={8} md={6} lg={3}>
          <Select
            allowClear
            placeholder="Result"
            value={result}
            onChange={(value) => {
              setResult(value);
              setPage(1);
            }}
            options={["SUCCESS", "FAILED", "DENIED"].map((value) => ({
              value,
              label: value,
            }))}
          />
        </Col>

        <Col xs={24} md={12} lg={8}>
          <DatePicker.RangePicker
            value={dateRange}
            onChange={(value) => {
              setDateRange(value);
              setPage(1);
            }}
          />
        </Col>
      </Row>
      <Flex justify="flex-end" gap={8} className="audit-filter-actions">
        <Button type="text" onClick={resetFilters}>
          ล้างตัวกรอง
        </Button>
        <Button
          type="primary"
          icon={<SearchOutlined />}
          onClick={() => {
            setKeyword(keywordInput.trim());
            setPage(1);
          }}
        >
          ค้นหา
        </Button>
      </Flex>
    </AntCard>
  );
};

export default AuditLogFiltersCard;
