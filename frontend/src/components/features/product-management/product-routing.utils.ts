import type {
  CreateProductLogRoutingRuleInput,
  ProductFeatureOption,
  ProductLogRoutingRule,
  RoutingDiscoveryItem,
} from "@/services/product.service";

export type RoutingFormValues = {
  rule_name: string;
  environment_id?: number;
  service?: string;
  source_project_id?: number;
  request_path: string;
  path_operator: "equals" | "starts_with";
  request_method?: string;
  payload_field?: string;
  payload_operator?: "equals" | "contains" | "starts_with" | "in";
  payload_value?: string;
  target_project_id: number;
  target_category_id?: number;
  priority?: number;
};

export const methodOptions = ["GET", "POST", "PUT", "PATCH", "DELETE"].map(
  (value) => ({ value, label: value }),
);

export function describeConditions(rule: ProductLogRoutingRule) {
  return rule.conditions
    .map(
      (condition) =>
        `${condition.field} ${condition.operator === "starts_with" ? "starts with" : "equals"} ${condition.value}`,
    )
    .join(" · ");
}

export function buildFeatureOptions(features: ProductFeatureOption[]) {
  if (features.length === 0) return [];

  const categories = features.filter(
    (item) => !item.parent_id || item.level === 1,
  );
  const subMap = new Map<number, ProductFeatureOption[]>();

  features.forEach((item) => {
    if (item.parent_id || item.level > 1) {
      const parentId = item.parent_id ?? 0;
      if (!subMap.has(parentId)) subMap.set(parentId, []);
      subMap.get(parentId)!.push(item);
    }
  });

  const groups: Array<{
    label: string;
    options: Array<{ value: number; label: string }>;
  }> = [];

  categories.forEach((category) => {
    const children = subMap.get(category.category_id) ?? [];
    groups.push({
      label: `${category.category_name} (#${category.category_id})`,
      options: [
        {
          value: category.category_id,
          label: `${category.category_name} (Category Main)`,
        },
        ...children.map((child) => ({
          value: child.category_id,
          label: `${child.category_name} (#${child.category_id})`,
        })),
      ],
    });
  });

  subMap.forEach((children, parentId) => {
    if (
      parentId !== 0 &&
      !categories.some((category) => category.category_id === parentId)
    ) {
      groups.push({
        label: `Category #${parentId}`,
        options: children.map((child) => ({
          value: child.category_id,
          label: `${child.category_name} (#${child.category_id})`,
        })),
      });
    }
  });

  return groups;
}

function getCondition(rule: ProductLogRoutingRule, field: string) {
  return rule.conditions.find((item) => item.field === field);
}

export function toRoutingFormValues(
  rule: ProductLogRoutingRule,
): RoutingFormValues {
  const path = getCondition(rule, "request_path");
  const extraCondition = rule.conditions.find(
    (item) =>
      !["request_path", "service", "source_project_id", "request_method"].includes(
        item.field,
      ),
  );

  return {
    rule_name: rule.rule_name,
    environment_id: rule.environment_id ?? undefined,
    service: String(getCondition(rule, "service")?.value ?? "") || undefined,
    source_project_id:
      Number(getCondition(rule, "source_project_id")?.value) || undefined,
    request_method:
      String(getCondition(rule, "request_method")?.value ?? "") || undefined,
    request_path: String(path?.value ?? ""),
    path_operator: path?.operator === "starts_with" ? "starts_with" : "equals",
    payload_field: extraCondition?.field ?? undefined,
    payload_operator: extraCondition?.operator as RoutingFormValues["payload_operator"] ?? "equals",
    payload_value:
      extraCondition?.value !== undefined
        ? String(extraCondition.value)
        : undefined,
    target_project_id: rule.target_project_id,
    target_category_id: rule.target_category_id ?? undefined,
    priority: rule.priority,
  };
}

export function buildRoutingInput(
  values: RoutingFormValues,
): CreateProductLogRoutingRuleInput {
  const conditions: CreateProductLogRoutingRuleInput["conditions"] = [];
  if (values.request_path?.trim()) {
    conditions.push({
      field: "request_path",
      operator: values.path_operator,
      value: values.request_path.trim(),
    });
  }
  if (values.service?.trim()) {
    conditions.unshift({
      field: "service",
      operator: "equals",
      value: values.service.trim(),
    });
  }
  if (values.source_project_id) {
    conditions.unshift({
      field: "source_project_id",
      operator: "equals",
      value: String(values.source_project_id),
    });
  }
  if (values.request_method) {
    conditions.push({
      field: "request_method",
      operator: "equals",
      value: values.request_method,
    });
  }
  if (values.payload_field?.trim() && values.payload_value?.trim()) {
    conditions.push({
      field: values.payload_field.trim(),
      operator: values.payload_operator || "equals",
      value: values.payload_value.trim(),
    });
  }

  return {
    environment_id: values.environment_id,
    rule_name: values.rule_name.trim(),
    priority:
      values.priority ??
      (values.payload_field?.trim() && values.payload_value?.trim() ? 10 : 0),
    conditions,
    target_project_id: values.target_project_id,
    target_category_id: values.target_category_id,
  };
}

export function isDiscoveryItemMapped(
  item: RoutingDiscoveryItem,
  rules: ProductLogRoutingRule[],
) {
  return (
    item.routing_status === "CLASSIFIED" ||
    item.routing_status === "MAPPED" ||
    rules.some((rule) =>
      rule.conditions.some(
        (condition) =>
          condition.field === "request_path" &&
          (condition.value === item.request_path ||
            condition.value === item.route_pattern),
      ),
    )
  );
}
