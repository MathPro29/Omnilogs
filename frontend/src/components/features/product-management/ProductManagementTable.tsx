import {
  Alert,
  Button,
  Empty,
  Input,
  Modal,
  Table,
  Tag,
  Tooltip,
  message,
} from "antd";
import {
  DeleteOutlined,
  EyeOutlined,
  ReloadOutlined,
  SearchOutlined,
  ExclamationCircleOutlined,
  EditOutlined,
} from "@ant-design/icons";
import type { ProductOption } from "@/services/product.service";

interface ProductManagementTableProps {
  products: ProductOption[];
  search: string;
  isLoading: boolean;
  isFetching: boolean;
  isError: boolean;
  isPlatformAdmin: boolean;
  selectedRowKeys?: React.Key[];
  isDeleting?: boolean;
  onSearchChange: (value: string) => void;
  onRefresh: () => void;
  onOpen: (product: ProductOption) => void;
  onEdit?: (product: ProductOption) => void;
  onSelectChange?: (selectedRowKeys: React.Key[]) => void;
  onDeleteProduct?: (id: number) => Promise<void>;
  onBulkDeleteProducts?: (ids: number[]) => Promise<void>;
}

export function ProductManagementTable(props: ProductManagementTableProps) {
  const selectedCount = props.selectedRowKeys?.length ?? 0;

  const handleConfirmBulkDelete = () => {
    const selectedIds = (props.selectedRowKeys ?? []) as number[];
    if (selectedIds.length === 0 || !props.onBulkDeleteProducts) return;

    Modal.confirm({
      title: `ลบ Product ที่เลือกจำนวน ${selectedIds.length} รายการ?`,
      icon: <ExclamationCircleOutlined style={{ color: "#ff4d4f" }} />,
      content:
        "การดำเนินการนี้ไม่สามารถยกเลิกได้ ระบบจะลบ Product ที่เลือกทั้งหมดออกจากระบบ",
      okText: "ยืนยันการลบ",
      okType: "danger",
      cancelText: "ยกเลิก",
      onOk: async () => {
        try {
          await props.onBulkDeleteProducts?.(selectedIds);
          message.success(
            `ลบ Product จำนวน ${selectedIds.length} รายการสำเร็จ`,
          );
        } catch {
          message.error("เกิดข้อผิดพลาด ไม่สามารถลบ Product ได้");
        }
      },
    });
  };

  const handleConfirmSingleDelete = (product: ProductOption) => {
    if (!props.onDeleteProduct) return;
    Modal.confirm({
      title: `ลบ Product "${product.name}"?`,
      icon: <ExclamationCircleOutlined style={{ color: "#ff4d4f" }} />,
      content:
        "การดำเนินการนี้ไม่สามารถยกเลิกได้ ข้อมูล Product นี้จะถูกลบถาวร",
      okText: "ยืนยันการลบ",
      okType: "danger",
      cancelText: "ยกเลิก",
      onOk: async () => {
        try {
          await props.onDeleteProduct?.(product.id);
          message.success(`ลบ Product "${product.name}" สำเร็จ`);
        } catch {
          message.error(
            `เกิดข้อผิดพลาด ไม่สามารถลบ Product "${product.name}" ได้`,
          );
        }
      },
    });
  };

  return (
    <section className="product-management-card">
      <div className="product-management-toolbar">
        <Input
          aria-label="Search products"
          prefix={<SearchOutlined />}
          value={props.search}
          onChange={(event) => props.onSearchChange(event.target.value)}
          placeholder="Search name or product ID"
          allowClear
        />
        {selectedCount > 0 && props.onBulkDeleteProducts && (
          <Button
            danger
            type="primary"
            icon={<DeleteOutlined />}
            loading={props.isDeleting}
            onClick={handleConfirmBulkDelete}
          >
            ลบที่เลือก ({selectedCount})
          </Button>
        )}
        <Button
          icon={<ReloadOutlined spin={props.isFetching} />}
          onClick={props.onRefresh}
        >
          Refresh
        </Button>
      </div>
      {props.isError && (
        <Alert
          type="error"
          showIcon
          message="Could not load products"
          description="Your session may not have an active product scope, or the service is unavailable."
          action={<Button onClick={props.onRefresh}>Try again</Button>}
        />
      )}
      <Table<ProductOption>
        rowKey="id"
        loading={props.isLoading}
        dataSource={props.products}
        scroll={{ x: 760 }}
        pagination={{ pageSize: 12, showSizeChanger: false }}
        rowSelection={
          props.onSelectChange
            ? {
                selectedRowKeys: props.selectedRowKeys,
                onChange: (keys) => props.onSelectChange?.(keys),
                getCheckboxProps: (record) => ({
                  disabled:
                    record.id === 0 ||
                    (!record.is_owner && !props.isPlatformAdmin),
                  name: record.name,
                }),
              }
            : undefined
        }
        locale={{
          emptyText: (
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              description={
                props.search
                  ? "No products match this search"
                  : "No products are available in this scope"
              }
            />
          ),
        }}
        columns={[
          {
            title: "Product",
            dataIndex: "name",
            render: (name, record) => (
              <button
                className="product-name-button"
                onClick={() => props.onOpen(record)}
              >
                <strong>{name}</strong>
                <span>{record.description || "No description"}</span>
              </button>
            ),
          },
          {
            title: "ID",
            dataIndex: "id",
            width: 100,
            render: (value) => <code>#{value}</code>,
          },
          {
            title: "Environments",
            dataIndex: "product_environments",
            width: 140,
            render: (value: ProductOption["product_environments"]) =>
              value?.length ?? 0,
          },
          {
            title: "Access",
            dataIndex: "membership_role",
            width: 120,
            render: (value, record) => (
              <Tag
                color={
                  record.is_owner || props.isPlatformAdmin ? "gold" : "blue"
                }
              >
                {record.is_owner || props.isPlatformAdmin
                  ? "Owner"
                  : value || "Member"}
              </Tag>
            ),
          },
          {
            title: "Setup",
            dataIndex: "setup_status",
            width: 150,
            render: (value) => (
              <Tag color={value === "ACTIVE" ? "success" : "default"}>
                {value || "DRAFT"}
              </Tag>
            ),
          },
          {
            title: "Status",
            dataIndex: "is_active",
            width: 110,
            render: (value) => (
              <Tag color={value === false ? "default" : "success"}>
                {value === false ? "Inactive" : "Active"}
              </Tag>
            ),
          },
          {
            title: "",
            key: "actions",
            width: 110,
            render: (_, record) => {
              const canManage =
                record.id > 0 && (record.is_owner || props.isPlatformAdmin);
              return (
                <div style={{ display: "flex", gap: "4px" }}>
                  <Tooltip title="Edit product">
                    <Button
                      type="text"
                      icon={<EditOutlined />}
                      aria-label={`Edit ${record.name}`}
                      onClick={() => props.onEdit?.(record)}
                    />
                  </Tooltip>
                  <Tooltip title="View details">
                    <Button
                      type="text"
                      icon={<EyeOutlined />}
                      aria-label={`View ${record.name}`}
                      onClick={() => props.onOpen(record)}
                    />
                  </Tooltip>
                  {props.onDeleteProduct && canManage && (
                    <Tooltip title="Delete product">
                      <Button
                        type="text"
                        danger
                        icon={<DeleteOutlined />}
                        aria-label={`Delete ${record.name}`}
                        onClick={() => handleConfirmSingleDelete(record)}
                      />
                    </Tooltip>
                  )}
                </div>
              );
            },
          },
        ]}
      />
    </section>
  );
}
