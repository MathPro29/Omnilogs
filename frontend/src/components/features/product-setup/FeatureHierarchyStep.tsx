import { Button, Empty, Input, Select, Tag, Tree } from "antd";
import type { TreeDataNode } from "antd";
import {
  ApartmentOutlined,
  CheckCircleOutlined,
  CompressOutlined,
  CopyOutlined,
  ExpandOutlined,
  ExclamationCircleOutlined,
  FolderOpenOutlined,
  InfoCircleOutlined,
  PlusOutlined,
  SearchOutlined,
} from "@ant-design/icons";
import type {
  FeatureNode,
  ProjectInfo,
} from "@/features/product-setup/types/productSetup.types";
import { useFeatureHierarchyEditor } from "@/features/product-setup/hooks/useFeatureHierarchyEditor";
import { FeatureEditorModal } from "./FeatureEditorModal";
import { FeatureTreeNodeTitle } from "./FeatureTreeNodeTitle";
import { ImportFeatureModal } from "./ImportFeatureModal";

interface FeatureHierarchyStepProps {
  projects: ProjectInfo[];
  features: FeatureNode[];
  onChange: (features: FeatureNode[]) => void;
  onCreateFeature?: (feature: FeatureNode) => Promise<boolean>;
  hasUnsavedChanges?: boolean;
}

export function FeatureHierarchyStep({
  projects,
  features,
  onChange,
  onCreateFeature,
  hasUnsavedChanges = false,
}: FeatureHierarchyStepProps) {
  const editor = useFeatureHierarchyEditor(
    projects,
    features,
    onChange,
    onCreateFeature,
  );

  const hasOtherProjectsWithFeatures = projects.some((project) => {
    const projectId = String(project.project_id ?? project.id);
    return (
      projectId !== String(editor.selectedProjectId) &&
      features.some((feature) => String(feature.project_id) === projectId)
    );
  });

  const buildTreeData = (
    nodes: FeatureNode[],
    parentId: string | number | null = null,
  ): TreeDataNode[] =>
    nodes
      .filter(
        (node) => String(node.parent_id ?? "") === String(parentId ?? ""),
      )
      .filter(
        (node) =>
          !editor.searchQuery ||
          node.category_name
            .toLowerCase()
            .includes(editor.searchQuery.toLowerCase()),
      )
      .map((node) => {
        const nodeId = node.category_id ?? node.id;
        const children = buildTreeData(nodes, nodeId);
        return {
          key: nodeId,
          title: (
            <FeatureTreeNodeTitle
              node={node}
              nodeId={nodeId}
              hasChildren={children.length > 0}
              onAdd={editor.openAddModal}
              onEdit={editor.openEditModal}
              onDelete={editor.deleteNode}
            />
          ),
          children: children.length ? children : undefined,
        };
      });

  const treeData = buildTreeData(editor.currentProjectFeatures);
  const rootFeatureCount = editor.currentProjectFeatures.filter(
    (feature) => feature.parent_id === null || feature.parent_id === undefined,
  ).length;
  const currentProjectName = editor.currentProject?.project_name || "No project";
  const currentProjectCode = editor.currentProject?.project_code || "—";

  return (
    <div className="feature-hierarchy-page">
      <div className="feature-hierarchy-intro">
        <div>
          <span className="feature-hierarchy-intro__eyebrow">
            ขั้นตอนที่ 3 · Feature hierarchy
          </span>
          <h2>สร้าง Feature hierarchy</h2>
          <p>
            จัดกลุ่ม Feature และ Sub-feature ตาม Project เพื่อให้ Log ที่เข้ามา
            ถูกจัดเส้นทางไปยังตำแหน่งที่ถูกต้อง
          </p>
        </div>
        <div
          className={
            hasUnsavedChanges
              ? "hierarchy-save-state hierarchy-save-state--dirty"
              : "hierarchy-save-state"
          }
        >
          {hasUnsavedChanges ? (
            <ExclamationCircleOutlined />
          ) : (
            <CheckCircleOutlined />
          )}
          <span>{hasUnsavedChanges ? "มีการเปลี่ยนแปลงที่ยังไม่บันทึก" : "บันทึกการเปลี่ยนแปลงแล้ว"}</span>
        </div>
      </div>

      <div className="hierarchy-layout">
        <aside className="hierarchy-scope-card">
          <div className="hierarchy-scope-card__heading">
            <span className="hierarchy-scope-card__icon">
              <FolderOpenOutlined />
            </span>
            <div>
              <span className="hierarchy-section-label">ขอบเขตที่กำลังแก้ไข</span>
              <strong>Project</strong>
            </div>
          </div>

          <Select
            value={String(editor.selectedProjectId)}
            onChange={editor.setSelectedProjectId}
            className="hierarchy-project-select"
            options={projects.map((project) => ({
              value: String(project.project_id ?? project.id),
              label: project.project_name || "Unnamed project",
            }))}
            aria-label="เลือก Project"
          />

          <div className="hierarchy-scope-card__project">
            <span className="hierarchy-section-label">Project ปัจจุบัน</span>
            <strong>{currentProjectName}</strong>
            <code>{currentProjectCode}</code>
          </div>

          <div className="hierarchy-scope-stats">
            <div>
              <strong>{rootFeatureCount}</strong>
              <span>Root features</span>
            </div>
            <div>
              <strong>{editor.currentProjectFeatures.length}</strong>
              <span>รายการทั้งหมด</span>
            </div>
          </div>

          <div className="hierarchy-scope-card__tip">
            <InfoCircleOutlined />
            <span>
              ใช้ปุ่มด้านขวาของแต่ละรายการเพื่อเพิ่ม Child, แก้ไข หรือลบ
              Branch
            </span>
          </div>
        </aside>

        <section className="hierarchy-editor-card">
          <div className="hierarchy-editor-card__header">
            <div>
              <span className="hierarchy-section-label">ตัวแก้ไขโครงสร้าง</span>
              <h3>{currentProjectName}</h3>
              <p>
                Project <span>›</span> Feature <span>›</span> Sub-feature
              </p>
            </div>
            <Tag color={hasUnsavedChanges ? "gold" : "green"}>
              {hasUnsavedChanges ? "รอบันทึก" : "บันทึกแล้ว"}
            </Tag>
          </div>

          <div className="hierarchy-editor-toolbar">
            <Input
              prefix={<SearchOutlined />}
              placeholder="ค้นหาด้วยชื่อ Feature"
              value={editor.searchQuery}
              onChange={(event) => editor.setSearchQuery(event.target.value)}
              allowClear
              aria-label="ค้นหา Feature"
            />
            <div className="hierarchy-editor-toolbar__actions">
              <Button
                icon={<ExpandOutlined />}
                onClick={editor.expandAll}
                aria-label="ขยาย Feature ทั้งหมด"
              >
                ขยายทั้งหมด
              </Button>
              <Button
                icon={<CompressOutlined />}
                onClick={editor.collapseAll}
                aria-label="ย่อ Feature ทั้งหมด"
              >
                ย่อทั้งหมด
              </Button>
              {hasOtherProjectsWithFeatures && (
                <Button
                  icon={<CopyOutlined />}
                  onClick={editor.openImportModal}
                >
                  นำเข้า
                </Button>
              )}
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => editor.openAddModal(null)}
                className="hierarchy-editor-toolbar__primary"
              >
                เพิ่ม Root feature
              </Button>
            </div>
          </div>

          <div className="hierarchy-editor-card__body">
            {treeData.length ? (
              <Tree
                treeData={treeData}
                expandedKeys={editor.expandedKeys}
                onExpand={editor.setExpandedKeys}
                className="feature-tree"
                showLine={{ showLeafIcon: false }}
                blockNode
              />
            ) : (
              <Empty
                image={<ApartmentOutlined className="hierarchy-empty-icon" />}
                description={
                  <div className="hierarchy-empty-copy">
                    <strong>ยังไม่มี Feature ใน Project นี้</strong>
                    <span>
                      เริ่มด้วย Root feature หรือเลือกนำเข้าโครงสร้างจาก
                      Project อื่น
                    </span>
                  </div>
                }
              >
                <div className="hierarchy-empty-actions">
                  {hasOtherProjectsWithFeatures && (
                    <Button
                      icon={<CopyOutlined />}
                      onClick={editor.openImportModal}
                    >
                      นำเข้าจาก Project อื่น
                    </Button>
                  )}
                  <Button
                    type="primary"
                    icon={<PlusOutlined />}
                    onClick={() => editor.openAddModal(null)}
                    className="hierarchy-editor-toolbar__primary"
                  >
                    เพิ่ม Feature แรก
                  </Button>
                </div>
              </Empty>
            )}
          </div>

          <div className="hierarchy-editor-card__footer">
            <InfoCircleOutlined />
            <span>
              {hasUnsavedChanges
                ? "เมื่อพร้อมแล้ว ให้กดบันทึกการเปลี่ยนแปลง Hierarchy ด้านล่าง"
                : "การเปลี่ยนแปลงในหน้านี้ถูกบันทึกแล้ว กดไปต่อเมื่อพร้อม"}
            </span>
          </div>
        </section>
      </div>

      <FeatureEditorModal editor={editor} />
      <ImportFeatureModal
        editor={editor}
        projects={projects}
        features={features}
      />
    </div>
  );
}




