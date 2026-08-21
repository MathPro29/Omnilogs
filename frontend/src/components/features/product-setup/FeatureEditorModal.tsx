import { Input, Modal } from "antd";
import type { FeatureHierarchyEditor } from "@/features/product-setup/hooks/useFeatureHierarchyEditor";
import { toCode } from "@/features/product-setup/utils/hierarchyValidator";

export function FeatureEditorModal({
  editor,
}: {
  editor: FeatureHierarchyEditor;
}) {
  return (
    <Modal
      title={
        editor.editingNode
          ? "แก้ไข Feature"
          : editor.parentTargetId
            ? "เพิ่ม Sub-feature"
            : "เพิ่ม Root feature"
      }
      open={editor.isModalOpen}
      onOk={editor.saveModal}
      onCancel={() => editor.setIsModalOpen(false)}
      okText={editor.editingNode ? "อัปเดต Feature" : "สร้าง Feature"}
      confirmLoading={editor.isSaving}
      cancelButtonProps={{ disabled: editor.isSaving }}
      okButtonProps={{ className: "bg-zinc-900 hover:bg-zinc-800" }}
    >
      {" "}
      <div className="space-y-4 py-2">
        <div>
          <label className="text-xs font-semibold text-zinc-700 block mb-1">
            ชื่อ Feature *
          </label>
          <Input
            placeholder="เช่น Authentication หรือ Login"
            value={editor.draft.name}
            onChange={(event) =>
              editor.setDraft((current) => ({
                ...current,
                name: event.target.value,
                code: toCode(event.target.value),
              }))
            }
            autoFocus
          />
        </div>
      </div>{" "}
    </Modal>
  );
}

