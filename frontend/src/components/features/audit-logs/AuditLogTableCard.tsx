import { Button, Card as AntCard, Empty, Table, type TableColumnsType } from "antd";
import type { AuditLog } from "@/features/audit-logs/services/audit.service";

export interface AuditTableCardProps {
  columns: TableColumnsType<AuditLog>;
  audits: {
    data?: { data: AuditLog[]; total: number };
    isLoading: boolean;
    isError: boolean;
    refetch: () => void;
  };
  page: number;
  pageSize: number;
  setPage: (page: number) => void;
  setPageSize: (pageSize: number) => void;
  setSelectedAuditId: (id: string) => void;
}

const AuditLogTableCard = ({
  columns,
  audits,
  page,
  pageSize,
  setPage,
  setPageSize,
  setSelectedAuditId,
}: AuditTableCardProps) => {
  return (
    <AntCard className="audit-table-card" bordered={false}>
      <Table<AuditLog>
        className="audit-table"
        rowKey="audit_id"
        columns={columns}
        dataSource={audits.data?.data || []}
        loading={audits.isLoading}
        scroll={{ x: 980 }}
        locale={{
          emptyText: audits.isError ? (
            <Empty description="ไม่สามารถโหลด Audit Logs ได้">
              <Button onClick={() => audits.refetch()}>ลองอีกครั้ง</Button>
            </Empty>
          ) : (
            <Empty description="ไม่พบ Activity" />
          ),
        }}
        pagination={{
          current: page,
          pageSize,
          total: audits.data?.total || 0,
          showSizeChanger: true,
          pageSizeOptions: [10, 20, 50, 100],
          showTotal: (total, range) =>
            `${range[0]}–${range[1]} จาก ${total} รายการ`,
          onChange: (nextPage, nextSize) => {
            setPage(nextSize !== pageSize ? 1 : nextPage);
            setPageSize(nextSize);
          },
        }}
        onRow={(record) => ({
          onDoubleClick: () => setSelectedAuditId(record.audit_id),
        })}
      />
    </AntCard>
  );
};

export default AuditLogTableCard;
