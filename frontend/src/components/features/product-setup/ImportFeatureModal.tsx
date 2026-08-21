import { Modal, Select, Button, Typography } from "antd";
import { CopyOutlined } from "@ant-design/icons";
import type { FeatureHierarchyEditor } from "@/features/product-setup/hooks/useFeatureHierarchyEditor";
import type { ProjectInfo, FeatureNode } from "@/features/product-setup/types/productSetup.types";

interface ImportFeatureModalProps {
  editor: FeatureHierarchyEditor;
  projects: ProjectInfo[];
  features: FeatureNode[];
}

const projectId = (value: ProjectInfo) => String(value.project_id ?? value.id);

export function ImportFeatureModal({
  editor,
  projects,
  features,
}: ImportFeatureModalProps) {
  const otherProjects = projects.filter(
    (p) => projectId(p) !== editor.selectedProjectId
  );

  const selectedSourceFeatures = features.filter(
    (f) => String(f.project_id) === String(editor.importSourceProjectId)
  );

  const rootFeaturesCount = selectedSourceFeatures.filter(
    (f) => !f.parent_id
  ).length;

  return (
    <Modal
      title={
        <div className="flex items-center gap-2 text-zinc-900 font-semibold">
          <CopyOutlined className="text-blue-600" />
          <span>นำเข้า Feature จาก Project อื่น</span>
        </div>
      }
      open={editor.isImportModalOpen}
      onCancel={() => editor.setIsImportModalOpen(false)}
      footer={[
        <Button
          key="cancel"
          onClick={() => editor.setIsImportModalOpen(false)}
        >
          ยกเลิก
        </Button>,
        <Button
          key="import"
          type="primary"
          icon={<CopyOutlined />}
          disabled={!editor.importSourceProjectId || selectedSourceFeatures.length === 0}
          onClick={() =>
            editor.importFeaturesFromProject(editor.importSourceProjectId)
          }
          className="bg-zinc-900 hover:bg-zinc-800"
        >
          นำเข้า Feature ({selectedSourceFeatures.length} รายการ)
        </Button>,
      ]}
      destroyOnClose
    >
      <div className="py-3 space-y-4">
        <Typography.Text type="secondary" className="text-xs">
          คัดลอกโครงสร้าง Feature และ Sub-feature ทั้งหมดจาก Project ต้นทางมายัง Project ปัจจุบัน (
          <strong>{editor.currentProject?.project_name}</strong>)
        </Typography.Text>

        <div className="space-y-1.5">
          <label className="text-xs font-semibold text-zinc-700 block">
            เลือก Project ต้นทาง:
          </label>
          <Select
            value={editor.importSourceProjectId || undefined}
            onChange={editor.setImportSourceProjectId}
            placeholder="เลือก Project ต้นทาง..."
            className="w-full"
            options={otherProjects.map((p) => {
              const pFeatures = features.filter(
                (f) => String(f.project_id) === projectId(p)
              );
              return {
                value: projectId(p),
                label: `${p.project_name || "ไม่มีชื่อ Project"} (${pFeatures.length} Features)`,
              };
            })}
          />
        </div>

        {editor.importSourceProjectId && (
          <div className="p-3 bg-zinc-50 border border-zinc-200 rounded-md text-xs space-y-1 text-zinc-600">
            <div>
              <strong>จำนวน Feature ที่จะคัดลอก:</strong>{" "}
              {selectedSourceFeatures.length} รายการ (Root features: {rootFeaturesCount})
            </div>
            {selectedSourceFeatures.length === 0 && (
              <div className="text-amber-600 font-medium">
                ⚠ Project ต้นทางนี้ไม่มี Feature ให้คัดลอก
              </div>
            )}
          </div>
        )}
      </div>
    </Modal>
  );
}


