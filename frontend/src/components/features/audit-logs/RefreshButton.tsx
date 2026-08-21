import { Button } from "antd";
import { ReloadOutlined } from "@ant-design/icons";
import { auditService, type AuditLogFilters } from "@/features/audit-logs/services/audit.service";
import { useQuery } from "@tanstack/react-query";

const RefreshButton = ({ filters }: { filters: AuditLogFilters }) => {
  const audits = useQuery({
    queryKey: ["audit-logs"],
    queryFn: () => auditService.list(filters),
  });
  return (
    <Button
      icon={<ReloadOutlined />}
      onClick={() => audits.refetch()}
      loading={audits.isFetching}
    >
      รีเฟรช
    </Button>
  );
};

export default RefreshButton;
