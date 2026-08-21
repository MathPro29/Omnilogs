import { Button, Input, Select } from "antd";
import { PlusOutlined, SearchOutlined } from "@ant-design/icons";
import type { ProductOption } from "@/services/product.service";
import type { ProductEnvironment } from "@/types/apiKey.types";
interface Props {
  products: ProductOption[];
  environments: ProductEnvironment[];
  productId?: number;
  environmentId?: number;
  search: string;
  canCreate: boolean;
  onProduct: (value: number) => void;
  onEnvironment: (value?: number) => void;
  onSearch: (value: string) => void;
  onGenerate: () => void;
}
export function ConnectionHeader(props: Props) {
  return (
    <>
      <div className="connections-title">
        <div>
          <h1>API Keys Connections</h1>
          <p>
            Manage API keys and connection settings used by external services to
            send logs into OmniLogs.
          </p>
        </div>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          disabled={!props.canCreate || !props.productId}
          onClick={props.onGenerate}
        >
          Generate API Key
        </Button>
      </div>
      <div className="connections-filters">
        <Select
          showSearch
          optionFilterProp="label"
          value={props.productId}
          placeholder="Select product"
          onChange={props.onProduct}
          options={props.products.map((product) => ({
            value: product.id,
            label: product.name,
          }))}
        />
        <Select
          allowClear
          value={props.environmentId}
          placeholder="All environments"
          onChange={props.onEnvironment}
          options={props.environments.map((environment) => ({
            value: environment.environmentId,
            label: environment.name,
          }))}
        />
        <Input
          prefix={<SearchOutlined />}
          value={props.search}
          onChange={(event) => props.onSearch(event.target.value)}
          placeholder="Search key name or prefix"
          allowClear
        />
      </div>
    </>
  );
}
