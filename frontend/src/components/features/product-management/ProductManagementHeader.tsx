import { Button} from "antd";
import { PlusOutlined } from "@ant-design/icons";

interface ProductManagementHeaderProps {
  roleLabel: string;
  canCreate: boolean;
  onCreate: () => void;
}

export function ProductManagementHeader(props: ProductManagementHeaderProps) {
  return (
    <header className="product-management-header">
      <div>
        <h1 className="text-xl font-bold">
          Products Management - จัดการผลิตภัณฑ์
        </h1>
      </div>
      {props.canCreate && (
        <Button type="primary" icon={<PlusOutlined />} onClick={props.onCreate}>
          Create product
        </Button>
      )}
    </header>
  );
}
