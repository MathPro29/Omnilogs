import { Button, Select, Tag } from "antd";
import {
  SearchOutlined,
  ReloadOutlined,
  SyncOutlined,
} from "@ant-design/icons";
import type { LogExplorerController } from "@/features/log-explorer/hooks/useLogExplorerController";

export function LogExplorerHeader({
  controller,
}: {
  controller: LogExplorerController;
}) {
  const { autoRefreshInterval, explorer } = controller;
  return (
    <header className="log-explorer-title">
      <div className="flex items-center gap-3">
        <h1 className="text-xl font-bold">Log Explorer - ศูนย์รวมล็อก</h1>
        {autoRefreshInterval > 0 && (
          <Tag
            color="#FF3D89"
            className="relative inline-flex items-center font-mono text-xs px-2.5 py-0.5 m-0 rounded-full overflow-visible"
          >
            Live
            <span className="absolute -top-0.5 -right-0.5 flex h-2 w-2">
              <span
                className="animate-ping absolute inline-flex h-full w-full rounded-full opacity-75"
                style={{ background: "#FF3D89" }}
              />
              <span
                className="relative inline-flex rounded-full h-2 w-2"
                style={{ background: "#FF3D89" }}
              />
            </span>
          </Tag>
        )}
      </div>
      <div className="log-explorer-title__actions flex items-center gap-2">
        <Button
          icon={<ReloadOutlined spin={controller.isFetching} />}
          onClick={controller.refresh}
          loading={controller.isLoading}
        >
          รีเฟรช
        </Button>
        <div className="flex items-center gap-1 bg-zinc-100 p-0.5 rounded-md border border-zinc-200 text-xs">
          <Button
            type={autoRefreshInterval > 0 ? "primary" : "default"}
            size="small"
            icon={
              <SyncOutlined
                spin={autoRefreshInterval > 0 && controller.isFetching}
              />
            }
            onClick={() =>
              controller.setAutoRefreshInterval((current) =>
                current > 0 ? 0 : 5000,
              )
            }
            className={
              autoRefreshInterval > 0
                ? "bg-emerald-600 hover:bg-emerald-700 text-white font-medium border-none"
                : "text-zinc-600"
            }
          >
            {autoRefreshInterval > 0
              ? "รีเฟรชอัตโนมัติ: เปิด"
              : "รีเฟรชอัตโนมัติ: ปิด"}
          </Button>
          <Select
            size="small"
            value={autoRefreshInterval}
            onChange={controller.setAutoRefreshInterval}
            className="w-24 font-mono text-xs"
            variant="borderless"
            options={[
              { value: 0, label: "ปิด" },
              { value: 3000, label: "3 วินาที" },
              { value: 5000, label: "5 วินาที" },
              { value: 10000, label: "10 วินาที" },
              { value: 30000, label: "30 วินาที" },
            ]}
          />
        </div>
        {!explorer.inspectorOpen && (
          <Button
            icon={<SearchOutlined />}
            onClick={() => explorer.setInspectorOpen(true)}
          >
            ตรวจสอบ
          </Button>
        )}
      </div>
    </header>
  );
}
