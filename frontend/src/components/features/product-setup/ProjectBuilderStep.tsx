import { useState } from "react";
import { Button, Input, Card, Tag, Popconfirm, Empty, Tooltip } from "antd";
import {
  PlusOutlined,
  DeleteOutlined,
  CopyOutlined,
  CheckOutlined,
  CloseOutlined,
  AppstoreAddOutlined,
} from "@ant-design/icons";
import type { ProjectInfo } from "@/features/product-setup/types/productSetup.types";
import { toCode } from "@/features/product-setup/utils/hierarchyValidator";

interface ProjectBuilderStepProps {
  productName: string;
  productCode: string;
  projects: ProjectInfo[];
  onChange: (projects: ProjectInfo[]) => void;
}

export function ProjectBuilderStep({
  productName,
  productCode,
  projects,
  onChange,
}: ProjectBuilderStepProps) {
  const [newProjectName, setNewProjectName] = useState("");
  const [newDescription, setNewDescription] = useState("");
  const [isAdding, setIsAdding] = useState(false);
  const nextProjectId = () => `proj_${crypto.randomUUID()}`;

  const handleAddDefaultProject = () => {
    const code = toCode(`${productCode}_MAIN`);
    const defaultProject: ProjectInfo = {
      id: nextProjectId(),
      project_name: productName || "Main Project",
      project_code: code,
      description: "สร้าง Project เริ่มต้นจากชื่อ Product",
      is_active: true,
      display_order: projects.length + 1,
    };
    onChange([...projects, defaultProject]);
  };

  const handleSaveNew = () => {
    if (!newProjectName.trim()) return;
    const code = toCode(newProjectName);
    const newProj: ProjectInfo = {
      id: nextProjectId(),
      project_name: newProjectName.trim(),
      project_code: code,
      description: newDescription.trim(),
      is_active: true,
      display_order: projects.length + 1,
    };
    onChange([...projects, newProj]);
    setNewProjectName("");

    setNewDescription("");
    setIsAdding(false);
  };

  const handleKeyDownNew = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      e.preventDefault();
      handleSaveNew();
    } else if (e.key === "Escape") {
      setIsAdding(false);
      setNewProjectName("");

      setNewDescription("");
    }
  };

  const handleDelete = (index: number) => {
    const next = [...projects];
    next.splice(index, 1);
    onChange(next);
  };

  const handleDuplicate = (proj: ProjectInfo) => {
    const dup: ProjectInfo = {
      id: nextProjectId(),
      project_name: `${proj.project_name} (Copy)`,
      project_code: toCode(`${proj.project_code}_COPY`),
      description: proj.description,
      is_active: proj.is_active,
      display_order: projects.length + 1,
    };
    onChange([...projects, dup]);
  };

  const handleToggleStatus = (index: number) => {
    const next = [...projects];
    next[index].is_active = !next[index].is_active;
    onChange(next);
  };

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <div className="text-xs font-mono font-semibold uppercase text-zinc-500 tracking-wider">
            Product:{" "}
            <span className="text-zinc-900 font-bold">
              {productName || "ยังไม่มีชื่อ Product"}
            </span>
          </div>
          <h2 className="text-xl font-bold text-zinc-900 mt-0.5">
            ขั้นตอนที่ 2: สร้าง Projects
          </h2>
          <p className="text-sm text-zinc-500 mt-1">
            กำหนด Software projects ภายใน Product (เช่น Backend API, Web
            App, Mobile App) ต้องมีอย่างน้อย 1 Project
          </p>
        </div>

        {projects.length === 0 && (
          <Button
            icon={<AppstoreAddOutlined />}
            onClick={handleAddDefaultProject}
            className="border-zinc-300 hover:border-zinc-900 font-medium"
          >
            สร้าง Default Project ({productName})
          </Button>
        )}
      </div>

      <div className="space-y-3">
        {projects.map((proj, idx) => (
          <Card
            key={proj.project_id || proj.id || idx}
            className="border-zinc-200 shadow-sm rounded-lg hover:border-zinc-300 transition-colors"
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded bg-zinc-100 flex items-center justify-center font-mono text-xs font-bold text-zinc-600">
                  #{idx + 1}
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <span className="font-semibold text-zinc-900 text-base">
                      {proj.project_name}
                    </span>
                    {proj.project_id ? (
                      <Tag color="blue" className="font-mono text-xs">
                        Project ID: {proj.project_id}
                      </Tag>
                    ) : null}
                    {!proj.is_active && <Tag color="error">ปิดใช้งาน</Tag>}
                  </div>
                  {proj.description && (
                    <p className="text-xs text-zinc-500 mt-0.5">
                      {proj.description}
                    </p>
                  )}
                </div>
              </div>

              <div className="flex items-center gap-1">
                <Tooltip
                  title={proj.is_active ? "ปิดใช้งาน Project" : "เปิดใช้งาน Project"}
                >
                  <Button
                    type="text"
                    size="small"
                    onClick={() => handleToggleStatus(idx)}
                    className={
                      proj.is_active
                        ? "text-zinc-500"
                        : "text-emerald-600 font-medium"
                    }
                  >
                    {proj.is_active ? "ปิดใช้งาน" : "เปิดใช้งาน"}
                  </Button>
                </Tooltip>

                <Tooltip title="ทำสำเนา Project">
                  <Button
                    type="text"
                    icon={<CopyOutlined />}
                    size="small"
                    onClick={() => handleDuplicate(proj)}
                    className="text-zinc-500 hover:text-zinc-900"
                  />
                </Tooltip>

                <Popconfirm
                  title="ลบ Project หรือไม่?"
                  description="Project นี้จะถูกลบออกจากโครงสร้าง Product"
                  onConfirm={() => handleDelete(idx)}
                  okText="ใช่ ลบ"
                  cancelText="ยกเลิก"
                >
                  <Button
                    type="text"
                    danger
                    icon={<DeleteOutlined />}
                    size="small"
                  />
                </Popconfirm>
              </div>
            </div>
          </Card>
        ))}{" "}
        {isAdding ? (
          <Card className="border-2 border-dashed border-zinc-300 rounded-lg p-4 bg-zinc-50/50">
            <div className="space-y-3">
              <div>
                <label className="text-xs font-semibold text-zinc-700 block mb-1">
                  ชื่อ Project *
                </label>
                <Input
                  placeholder="เช่น Backend API"
                  value={newProjectName}
                  onChange={(e) => setNewProjectName(e.target.value)}
                  onKeyDown={handleKeyDownNew}
                  autoFocus
                />
              </div>
              <div>
                <label className="text-xs font-semibold text-zinc-700 block mb-1">
                  คำอธิบาย (ไม่บังคับ)
                </label>
                <Input
                  placeholder="บริการ REST API หลักของ Backend"
                  value={newDescription}
                  onChange={(e) => setNewDescription(e.target.value)}
                  onKeyDown={handleKeyDownNew}
                />
              </div>
              <div className="flex items-center justify-end gap-2 pt-2">
                <Button
                  size="small"
                  onClick={() => setIsAdding(false)}
                  icon={<CloseOutlined />}
                >
                  ยกเลิก (Esc)
                </Button>
                <Button
                  type="primary"
                  size="small"
                  onClick={handleSaveNew}
                  disabled={!newProjectName.trim()}
                  icon={<CheckOutlined />}
                  className="bg-zinc-900 hover:bg-zinc-800"
                >
                  บันทึก Project (Enter)
                </Button>
              </div>
            </div>
          </Card>
        ) : (
          <Button
            type="dashed"
            block
            icon={<PlusOutlined />}
            onClick={() => setIsAdding(true)}
            className="h-12 border-zinc-300 hover:border-zinc-900 text-zinc-600 hover:text-zinc-900 font-medium rounded-lg"
          >
            เพิ่ม Project ในหน้านี้
          </Button>
        )}
        {projects.length === 0 && !isAdding && (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description="ยังไม่มี Project"
          />
        )}
      </div>
    </div>
  );
}

