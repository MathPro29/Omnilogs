import { useState } from "react";
import type { Key } from "react";
import { message } from "antd";
import type { FeatureNode, ProjectInfo } from "@/features/product-setup/types/productSetup.types";
import { toCode } from "@/features/product-setup/utils/hierarchyValidator";

interface FeatureDraft {
  name: string;
  code: string;
  description: string;
}

const emptyDraft: FeatureDraft = { name: "", code: "", description: "" };
const entityId = (value: { category_id?: number; id: string }) =>
  String(value.category_id ?? value.id);
const projectId = (value: ProjectInfo) => String(value.project_id ?? value.id);

export function useFeatureHierarchyEditor(
  projects: ProjectInfo[],
  features: FeatureNode[],
  onChange: (features: FeatureNode[]) => void,
  onCreate?: (feature: FeatureNode) => Promise<boolean>,
) {
  const [selectedProjectId, setSelectedProjectId] = useState<string>(
    projects[0] ? projectId(projects[0]) : "",
  );
  const [searchQuery, setSearchQuery] = useState("");
  const [expandedKeys, setExpandedKeys] = useState<Key[]>([]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingNode, setEditingNode] = useState<FeatureNode | null>(null);
  const [parentTargetId, setParentTargetId] = useState<string | number | null>(null);
  const [draft, setDraft] = useState<FeatureDraft>(emptyDraft);
  const [isSaving, setIsSaving] = useState(false);


  const activeProjectId = projects.some(
    (project) => projectId(project) === selectedProjectId,
  )
    ? selectedProjectId
    : projects[0]
      ? projectId(projects[0])
      : "";
  const currentProject =
    projects.find((project) => projectId(project) === activeProjectId) ?? projects[0];
  const currentProjectFeatures = features.filter(
    (feature) => String(feature.project_id) === activeProjectId,
  );

  const uniqueCode = (name: string, ignoredId?: string) => {
    const base = toCode(name) || "FEATURE";
    const used = new Set(
      currentProjectFeatures
        .filter((feature) => entityId(feature) !== ignoredId)
        .map((feature) => feature.category_code.toUpperCase()),
    );
    let code = base;
    let suffix = 2;
    while (used.has(code)) {
      code = `${base}_${suffix}`;
      suffix += 1;
    }
    return code;
  };

  const openAddModal = (parentId: string | number | null = null) => {
    setEditingNode(null);
    setParentTargetId(parentId);
    setDraft(emptyDraft);
    setIsModalOpen(true);
  };
  const openEditModal = (node: FeatureNode) => {
    setEditingNode(node);
    setParentTargetId(node.parent_id);
    setDraft({
      name: node.category_name,
      code: node.category_code,
      description: node.description ?? "",
    });
    setIsModalOpen(true);
  };
  const saveModal = async () => {
    const name = draft.name.trim();
    if (!name || !currentProject) {
      message.error("กรุณาระบุชื่อ Feature");
      return;
    }
    const editingId = editingNode ? entityId(editingNode) : undefined;
    const code = uniqueCode(name, editingId);
    if (editingNode && editingId) {
      onChange(
        features.map((feature) =>
          entityId(feature) === editingId
            ? {
                ...feature,
                category_name: name,
                category_code: code,
                description: draft.description.trim(),
                parent_id: parentTargetId,
              }
            : feature,
        ),
      );
      message.success("แก้ไข Feature แล้ว");
    } else {
      const id = `feat_${crypto.randomUUID()}`;
      const feature: FeatureNode = {
        id,
        project_id: currentProject.project_id ?? currentProject.id ?? activeProjectId,
        parent_id: parentTargetId,
        category_name: name,
        category_code: code,
        description: draft.description.trim(),
        is_active: true,
        display_order: currentProjectFeatures.length + 1,
      };
      if (onCreate) {
        setIsSaving(true);
        try {
          if (!(await onCreate(feature))) return;
        } finally {
          setIsSaving(false);
        }
      } else {
        onChange([...features, feature]);
      }
      setExpandedKeys((keys) =>
        parentTargetId === null || keys.map(String).includes(String(parentTargetId))
          ? keys
          : [...keys, parentTargetId],
      );
      message.success("สร้าง Feature และบันทึกเรียบร้อยแล้ว");
    }
    setIsModalOpen(false);
  };
  const deleteNode = (node: FeatureNode) => {
    const id = entityId(node);
    const childIds = new Set<string>();
    const collectChildren = (parentId: string) => {
      features.forEach((feature) => {
        if (String(feature.parent_id ?? "") === parentId) {
          const childId = entityId(feature);
          if (!childIds.has(childId)) {
            childIds.add(childId);
            collectChildren(childId);
          }
        }
      });
    };
    collectChildren(id);
    onChange(
      features.filter((feature) => {
        const featureId = entityId(feature);
        return featureId !== id && !childIds.has(featureId);
      }),
    );
    message.info("นำ Feature และ Sub-feature ออกจากรายการแล้ว");
  };

  const [isImportModalOpen, setIsImportModalOpen] = useState(false);
  const [importSourceProjectId, setImportSourceProjectId] = useState<string>("");

  const openImportModal = () => {
    const defaultSource = projects.find(
      (p) => projectId(p) !== activeProjectId
    );
    setImportSourceProjectId(defaultSource ? projectId(defaultSource) : "");
    setIsImportModalOpen(true);
  };

  const importFeaturesFromProject = (sourceProjId: string) => {
    if (!sourceProjId || !currentProject) return;
    const sourceFeatures = features.filter(
      (f) => String(f.project_id) === String(sourceProjId)
    );
    if (sourceFeatures.length === 0) {
      message.warning("ไม่มี Feature ใน Project ต้นทางที่เลือก");
      return;
    }

    const currentProjId = currentProject.project_id ?? currentProject.id ?? activeProjectId;
    const idMap = new Map<string, string>();
    const newFeatures: FeatureNode[] = [];

    const pending = [...sourceFeatures];
    let addedCount = 0;

    while (pending.length > 0) {
      let progressed = false;
      for (let i = pending.length - 1; i >= 0; i--) {
        const item = pending[i];
        const oldId = entityId(item);
        const oldParentId = item.parent_id !== null && item.parent_id !== undefined ? String(item.parent_id) : null;

        if (oldParentId && !idMap.has(oldParentId)) {
          const parentInSource = sourceFeatures.some((f) => entityId(f) === oldParentId);
          if (parentInSource) continue;
        }

        const newId = `feat_${crypto.randomUUID()}`;
        idMap.set(oldId, newId);
        const newParentId = oldParentId ? (idMap.get(oldParentId) ?? null) : null;

        newFeatures.push({
          id: newId,
          project_id: currentProjId,
          parent_id: newParentId,
          category_name: item.category_name,
          category_code: uniqueCode(item.category_name),
          description: item.description ?? "",
          is_active: item.is_active ?? true,
          display_order: currentProjectFeatures.length + addedCount + 1,
        });

        addedCount++;
        pending.splice(i, 1);
        progressed = true;
      }

      if (!progressed && pending.length > 0) {
        for (const item of pending) {
          const oldId = entityId(item);
          const newId = `feat_${crypto.randomUUID()}`;
          idMap.set(oldId, newId);
          newFeatures.push({
            id: newId,
            project_id: currentProjId,
            parent_id: null,
            category_name: item.category_name,
            category_code: uniqueCode(item.category_name),
            description: item.description ?? "",
            is_active: item.is_active ?? true,
            display_order: currentProjectFeatures.length + addedCount + 1,
          });
          addedCount++;
        }
        break;
      }
    }

    onChange([...features, ...newFeatures]);
    setIsImportModalOpen(false);
    message.success(`นำเข้า ${newFeatures.length} Feature เรียบร้อยแล้ว`);
  };

  return {
    selectedProjectId,
    searchQuery,
    expandedKeys,
    isModalOpen,
    isImportModalOpen,
    importSourceProjectId,
    editingNode,
    parentTargetId,
    draft,
    isSaving,
    currentProject,
    currentProjectFeatures,
    setSelectedProjectId: (value: string | number) => setSelectedProjectId(String(value)),
    setSearchQuery,
    setExpandedKeys,
    setIsModalOpen,
    setIsImportModalOpen,
    setImportSourceProjectId: (value: string | number) => setImportSourceProjectId(String(value)),
    setDraft,
    openAddModal,
    openEditModal,
    openImportModal,
    importFeaturesFromProject,
    saveModal,
    deleteNode,
    expandAll: () => setExpandedKeys(currentProjectFeatures.map(entityId)),
    collapseAll: () => setExpandedKeys([]),
  };
}

export type FeatureHierarchyEditor = ReturnType<typeof useFeatureHierarchyEditor>;
