import { Alert, Button } from "antd";
import { PageTransition } from "@/components";
import { LogExplorerHeader } from "@/components/features/log-explorer/LogExplorerHeader";
import { LogExplorerWorkspace } from "@/components/features/log-explorer/LogExplorerWorkspace";
import { useLogExplorerController } from "@/features/log-explorer/hooks/useLogExplorerController";
import "@/styles/pages/visual-log-search.css";

export function VisualLogSearchBuilder() {
  const controller = useLogExplorerController();

  return (
    <PageTransition>
      <main className="log-explorer-page rounded-4xl">
        <LogExplorerHeader controller={controller} />
        {controller.isError && (
          <Alert
            className="log-search-error"
            type="error"
            showIcon
            message="Could not search logs"
            description="Check your scope and filters, then try again."
            action={
              <Button size="small" onClick={controller.applyFilters}>
                Try again
              </Button>
            }
          />
        )}
        <LogExplorerWorkspace controller={controller} />
      </main>
    </PageTransition>
  );
}
