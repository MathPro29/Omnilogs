import { useMemo } from "react";
import { Empty, Tag, Tree, Typography } from "antd";
import type { DataNode } from "antd/es/tree";
import type {
  ProductFeatureOption,
  ProductProjectOption,
} from "@/services/product.service";

const { Text } = Typography;

interface Props {
  projects?: ProductProjectOption[];
  features?: ProductFeatureOption[];
  value?: number[];
  onChange?: (featureIds: number[]) => void;
  disabled?: boolean;
}

function ancestorsOf(
  id: number,
  byId: Map<number, ProductFeatureOption>,
): number[] {
  const result: number[] = [];
  let current = byId.get(id);
  while (current?.parent_id) {
    result.push(current.parent_id);
    current = byId.get(current.parent_id);
  }
  return result;
}

function descendantsOf(id: number, features: ProductFeatureOption[]): number[] {
  const result: number[] = [];
  const visit = (parentId: number) => {
    features
      .filter((feature) => feature.parent_id === parentId)
      .forEach((feature) => {
        result.push(feature.category_id);
        visit(feature.category_id);
      });
  };
  visit(id);
  return result;
}

export function FeatureScopeTree({
  projects = [],
  features = [],
  value = [],
  onChange,
  disabled,
}: Props) {
  const safeValue = useMemo(() => value ?? [], [value]);
  const safeFeatures = useMemo(() => features ?? [], [features]);
  const safeProjects = useMemo(() => projects ?? [], [projects]);

  const byId = useMemo(
    () => new Map(safeFeatures.map((feature) => [feature.category_id, feature])),
    [safeFeatures],
  );
  const inheritedIds = useMemo(
    () =>
      new Set(
        safeValue
          .flatMap((id) => ancestorsOf(id, byId))
          .filter((id) => !safeValue.includes(id)),
      ),
    [byId, safeValue],
  );
  const visualIds = useMemo(
    () => [...new Set([...safeValue, ...inheritedIds])],
    [inheritedIds, safeValue],
  );
  const childrenByParent = useMemo(() => {
    const result = new Map<string, ProductFeatureOption[]>();
    safeFeatures.forEach((feature) => {
      const key = `${feature.project_id}:${feature.parent_id ?? "root"}`;
      result.set(key, [...(result.get(key) ?? []), feature]);
    });
    return result;
  }, [safeFeatures]);
  const treeData = useMemo<DataNode[]>(() => {
    const walk = (projectId: number, parentId?: number): DataNode[] =>
      (childrenByParent.get(`${projectId}:${parentId ?? "root"}`) ?? []).map(
        (feature) => ({
          key: `feature:${feature.category_id}`,
          title: (
            <span className="feature-scope-title">
              <span>{feature.category_name}</span>
              {feature.parent_id && (
                <Text type="secondary">
                  Parent:{" "}
                  {byId.get(feature.parent_id)?.category_name ??
                    feature.parent_id}
                </Text>
              )}
              {inheritedIds.has(feature.category_id) && (
                <Tag color="blue">Auto-selected parent</Tag>
              )}
            </span>
          ),
          children: walk(projectId, feature.category_id),
        }),
      );
    return safeProjects.map((project) => ({
      key: `project:${project.project_id}`,
      title: <strong>{project.project_name}</strong>,
      disableCheckbox: true,
      selectable: false,
      children: walk(project.project_id),
    }));
  }, [byId, childrenByParent, inheritedIds, safeProjects]);

  if (!treeData.length)
    return (
      <Empty
        image={Empty.PRESENTED_IMAGE_SIMPLE}
        description="ไม่พบ Feature ใน Product นี้"
      />
    );

  return (
    <div className="feature-scope-tree">
      <Tree
        checkable
        checkStrictly
        blockNode
        defaultExpandAll
        disabled={disabled}
        treeData={treeData}
        checkedKeys={{
          checked: visualIds.map((id) => `feature:${id}`),
          halfChecked: [],
        }}
        onCheck={(_, info) => {
          const key = String(info.node.key);
          if (!key.startsWith("feature:")) return;
          const id = Number(key.slice("feature:".length));
          if (info.checked) {
            const descendants = new Set(descendantsOf(id, safeFeatures));
            onChange?.([
              ...safeValue.filter((selected) => !descendants.has(selected)),
              id,
            ]);
            return;
          }
          const descendants = new Set(descendantsOf(id, safeFeatures));
          onChange?.(
            safeValue.filter(
              (selected) => selected !== id && !descendants.has(selected),
            ),
          );
        }}
      />
      <Text type="secondary">
        Feature ที่มีป้ายสีน้ำเงินถูกเลือกเพื่อให้เข้าถึงเส้นทางด้านบนเท่านั้น
        ระบบจะบันทึกเฉพาะ Scope ที่เลือกจริงและไม่เปิดสิทธิ์ให้ Feature พี่น้อง
      </Text>
    </div>
  );
}
