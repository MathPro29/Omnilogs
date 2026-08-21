import { useEffect, useRef } from "react";
import type { CSSProperties, PointerEvent as ReactPointerEvent } from "react";
import { SearchPanel } from "./SearchPanel";
import { LogList } from "./LogList";
import { RawLogInspector } from "./RawLogInspector";
import type { LogExplorerController } from "@/features/log-explorer/hooks/useLogExplorerController";

function startInspectorResize(
  event: ReactPointerEvent<HTMLDivElement>,
  controller: LogExplorerController,
) {
  const startX = event.clientX;
  const startWidth = controller.explorer.inspectorWidth;
  let frameId: number | null = null;
  const move = (moveEvent: PointerEvent) => {
    if (frameId !== null) return;
    frameId = requestAnimationFrame(() => {
      frameId = null;
      controller.explorer.setInspectorWidth(
        Math.max(360, Math.min(760, startWidth + startX - moveEvent.clientX)),
      );
    });
  };
  const stop = () => {
    if (frameId !== null) cancelAnimationFrame(frameId);
    window.removeEventListener("pointermove", move);
    window.removeEventListener("pointerup", stop);
  };
  window.addEventListener("pointermove", move);
  window.addEventListener("pointerup", stop);
}

export function LogExplorerWorkspace({
  controller,
}: {
  controller: LogExplorerController;
}) {
  const { explorer } = controller;
  const inspectorRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!explorer.inspectorOpen || explorer.inspectorFullscreen) return;

    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Element | null;
      if (!target) return;

      if (!document.body.contains(target)) return;
      if (inspectorRef.current?.contains(target)) return;
      if (target.closest(".log-row")) return;
      if (
        target.closest(
          ".ant-popover, .ant-select-dropdown, .ant-picker-dropdown, .ant-modal-root, .ant-tooltip, .ant-dropdown, .ant-dropdown-menu, .ant-dropdown-trigger, [class*='ant-dropdown'], .ant-typography, .ant-typography-ellipsis, .ant-message, .ant-notification",
        )
      ) {
        return;
      }

      explorer.setInspectorOpen(false);
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [
    explorer,
    explorer.inspectorOpen,
    explorer.inspectorFullscreen,
    explorer.setInspectorOpen,
  ]);

  return (
    <div
      className={`log-explorer-shell ${!explorer.filtersOpen ? "filters-collapsed" : ""} ${!explorer.inspectorOpen ? "inspector-collapsed" : ""}`}
      style={
        {
          "--inspector-width": `${explorer.inspectorWidth}px`,
        } as CSSProperties
      }
    >
      <div className="log-search-panel-wrapper">
        <SearchPanel
          query={controller.query}
          productId={controller.selectedProductId}
          projectId={controller.projectId}
          environmentId={controller.environmentId}
          categoryId={controller.categoryId}
          subFeatureId={controller.subFeatureId}
          products={controller.products}
          projects={controller.projects}
          environments={controller.environments}
          categories={controller.categories}
          subFeatures={controller.subFeatures}
          timeRange={controller.timeRange}
          rules={controller.rules}
          loading={controller.isLoading}
          dirty={controller.filtersDirty}
          onQueryChange={controller.setQuery}
          onProductChange={controller.changeProduct}
          onProjectChange={controller.setProjectId}
          onEnvironmentChange={controller.setEnvironmentId}
          onCategoryChange={controller.setCategoryId}
          onSubFeatureChange={controller.setSubFeatureId}
          onTimeRangeChange={controller.setTimeRange}
          onRulesChange={controller.setRules}
          onSearch={controller.applyFilters}
          onClear={controller.clearFilters}
        />
      </div>
      <LogList
        records={controller.records}
        total={controller.total}
        loading={controller.isLoading || controller.isFetchingNextPage}
        hasMore={controller.hasNextPage}
        query={controller.submittedQuery}
        activeId={explorer.selectedId}
        selectedIds={explorer.selectedIds}
        columns={explorer.columns}
        favorites={controller.favoriteFields}
        onFavoriteReorder={controller.reorderFavoriteFields}
        onSelect={(id) => {
          explorer.select(id);
          explorer.setInspectorOpen(true);
          explorer.setInspectorFullscreen(false);
        }}
        onToggle={explorer.toggleSelected}
        onColumnsChange={explorer.setColumns}
        onLoadMore={controller.loadMore}
        onOpen={() => {
          explorer.setInspectorOpen(true);
          explorer.setInspectorFullscreen(false);
          explorer.setInspectorFullscreen(true);
        }}
      />
      {explorer.inspectorOpen && !explorer.inspectorFullscreen && (
        <div
          className="log-inspector-backdrop"
          onClick={() => explorer.setInspectorOpen(false)}
          aria-hidden="true"
        />
      )}
      <div className={`log-inspector-wrap ${explorer.inspectorFullscreen ? "is-fullscreen-wrap" : ""} ${!controller.selectedRecord ? "is-empty-wrap" : ""}`} ref={inspectorRef}>
        <div
          className="log-inspector-resizer"
          role="separator"
          aria-orientation="vertical"
          onPointerDown={(event) => startInspectorResize(event, controller)}
        />
        <RawLogInspector
          record={controller.selectedRecord}
          productId={controller.selectedProductId}
          fullscreen={explorer.inspectorFullscreen}
          onFullscreen={explorer.setInspectorFullscreen}
          onClose={() => explorer.setInspectorOpen(false)}
          onAddRule={controller.addRule}
        />
      </div>
    </div>
  );
}


