import { useMemo, useState } from "react";
import { Select } from "antd";
import type {
  ProductEnvironmentOption,
  ProductFeatureOption,
  ProductOption,
  ProductProjectOption,
} from "@/services/product.service";
import { AddProductModal } from "./AddProductModal";

interface SelectProductFieldProps {
  productId?: number;
  environmentId?: number;
  projectId?: number;
  categoryId?: number;
  subFeatureId?: number;
  products: ProductOption[];
  projects?: ProductProjectOption[];
  environments?: ProductEnvironmentOption[];
  categories?: ProductFeatureOption[];
  subFeatures?: ProductFeatureOption[];
  onProductChange?: (value: number) => void;
  onEnvironmentChange?: (value?: number) => void;
  onProjectChange?: (value?: number) => void;
  onCategoryChange?: (value?: number) => void;
  onSubFeatureChange?: (value?: number) => void;
}

const SelectProductField = (props: SelectProductFieldProps) => {
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);

  const subFeatureOptions = useMemo(() => {
    if (!props.subFeatures || props.subFeatures.length === 0) return [];

    const categoryMap = new Map<number, string>();
    (props.categories ?? []).forEach((cat) => {
      categoryMap.set(cat.category_id, cat.category_name);
    });

    const groupsMap = new Map<
      number | "uncategorized",
      ProductFeatureOption[]
    >();

    props.subFeatures.forEach((sub) => {
      const parentId = sub.parent_id ?? "uncategorized";
      if (!groupsMap.has(parentId)) {
        groupsMap.set(parentId, []);
      }
      groupsMap.get(parentId)!.push(sub);
    });

    const options: Array<{
      label: string;
      options: Array<{ value: number; label: string }>;
    }> = [];

    groupsMap.forEach((items, parentId) => {
      let groupLabel = "Other Sub Features";
      if (parentId !== "uncategorized") {
        const catName = categoryMap.get(parentId as number);
        groupLabel = catName
          ? `${catName} (#${parentId})`
          : `Category #${parentId}`;
      }

      options.push({
        label: groupLabel,
        options: items.map((item) => ({
          value: item.category_id,
          label: item.full_path
            ? `${item.full_path} (#${item.category_id})`
            : `${item.category_name} (#${item.category_id})`,
        })),
      });
    });

    return options;
  }, [props.subFeatures, props.categories]);

  return (
    <div className="log-filter-stack">
      <label>
        <Select
          showSearch
          optionFilterProp="label"
          value={props.productId}
          onChange={props.onProductChange}
          placeholder="Choose a product"
          options={props.products.map((product) => ({
            value: product.id,
            label: product.product_code
              ? `${product.name} (${product.product_code})`
              : product.name,
          }))}
        />
      </label>
      <label>
        Environment
        <Select
          allowClear
          showSearch
          optionFilterProp="label"
          value={props.environmentId}
          onChange={props.onEnvironmentChange}
          placeholder="All environments"
          loading={
            props.productId !== undefined && props.environments === undefined
          }
          options={(props.environments ?? []).map((item) => ({
            value: item.environment_id,
            label: `${item.environment_name} (${item.environment_code})`,
          }))}
        />
      </label>
      <label>
        Project
        <Select
          allowClear
          showSearch
          optionFilterProp="label"
          value={props.projectId}
          onChange={props.onProjectChange}
          placeholder="All projects"
          loading={
            props.productId !== undefined && props.projects === undefined
          }
          options={(props.projects ?? []).map((item) => ({
            value: item.project_id,
            label: `${item.project_name} (Project ID: ${item.project_id})`,
          }))}
        />
      </label>
      <label>
        Category
        <Select
          allowClear
          showSearch
          optionFilterProp="label"
          value={props.categoryId}
          onChange={props.onCategoryChange}
          placeholder="All categories"
          loading={
            props.productId !== undefined && props.categories === undefined
          }
          options={(props.categories ?? []).map((item) => ({
            value: item.category_id,
            label: `${item.category_name} (#${item.category_id})`,
          }))}
        />
      </label>
      <label>
        Sub feature
        <Select
          allowClear
          showSearch
          optionFilterProp="label"
          value={props.subFeatureId}
          onChange={props.onSubFeatureChange}
          placeholder="All sub features"
          loading={
            props.productId !== undefined && props.subFeatures === undefined
          }
          options={subFeatureOptions}
        />
      </label>
      <AddProductModal
        open={isAddModalOpen}
        onClose={() => setIsAddModalOpen(false)}
        onSuccess={(newId) => {
          if (props.onProductChange) props.onProductChange(newId);
        }}
      />
    </div>
  );
};

export default SelectProductField;
