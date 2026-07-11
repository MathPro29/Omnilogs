import { Skeleton, Card, Row, Col, Table } from 'antd';

/**
 * Skeleton สำหรับ Dashboard page
 */
export function DashboardSkeleton() {
  return (
    <div className="space-y-6">
      {/* Stat Cards */}
      <Row gutter={[16, 16]}>
        {Array.from({ length: 4 }).map((_, i) => (
          <Col xs={24} sm={12} lg={6} key={i}>
            <Card className="stat-card">
              <Skeleton active paragraph={{ rows: 1 }} title={{ width: '40%' }} />
            </Card>
          </Col>
        ))}
      </Row>

      {/* Charts */}
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={14}>
          <Card>
            <Skeleton active paragraph={{ rows: 6 }} title={{ width: '30%' }} />
          </Card>
        </Col>
        <Col xs={24} lg={10}>
          <Card>
            <Skeleton active paragraph={{ rows: 6 }} title={{ width: '30%' }} />
          </Card>
        </Col>
      </Row>

      {/* Activity */}
      <Card>
        <Skeleton active paragraph={{ rows: 5 }} title={{ width: '20%' }} />
      </Card>
    </div>
  );
}

/**
 * Skeleton สำหรับ Table page (Users, Roles, etc.)
 */
export function TableSkeleton() {
  const columns = Array.from({ length: 5 }).map((_, i) => ({
    title: <Skeleton.Input active size="small" style={{ width: 80 }} />,
    key: i,
    render: () => <Skeleton.Input active size="small" style={{ width: i === 0 ? 120 : 100 }} />,
  }));

  const dataSource = Array.from({ length: 5 }).map((_, i) => ({
    key: i,
  }));

  return (
    <div className="space-y-4">
      {/* Filter bar skeleton */}
      <Card>
        <div className="flex gap-3 flex-wrap">
          <Skeleton.Input active size="default" style={{ width: 220 }} />
          <Skeleton.Input active size="default" style={{ width: 150 }} />
          <Skeleton.Input active size="default" style={{ width: 150 }} />
          <Skeleton.Button active size="default" style={{ width: 80 }} />
        </div>
      </Card>

      {/* Table skeleton */}
      <Card>
        <Table
          columns={columns}
          dataSource={dataSource}
          pagination={false}
          size="middle"
        />
      </Card>
    </div>
  );
}

/**
 * Skeleton สำหรับ Form page (Settings, User detail)
 */
export function FormSkeleton() {
  return (
    <Card>
      <div className="space-y-6 max-w-2xl">
        {Array.from({ length: 5 }).map((_, i) => (
          <div key={i} className="space-y-2">
            <Skeleton.Input active size="small" style={{ width: 120 }} />
            <Skeleton.Input active size="default" style={{ width: '100%' }} />
          </div>
        ))}
        <div className="flex gap-3 pt-4">
          <Skeleton.Button active size="default" style={{ width: 100 }} />
          <Skeleton.Button active size="default" style={{ width: 80 }} />
        </div>
      </div>
    </Card>
  );
}

/**
 * Skeleton สำหรับ Detail page
 */
export function DetailSkeleton() {
  return (
    <Card>
      <div className="flex items-start gap-6 mb-8">
        <Skeleton.Avatar active size={80} shape="circle" />
        <div className="flex-1 space-y-3">
          <Skeleton.Input active size="default" style={{ width: 200 }} />
          <Skeleton.Input active size="small" style={{ width: 160 }} />
        </div>
      </div>
      <div className="space-y-4">
        {Array.from({ length: 6 }).map((_, i) => (
          <div key={i} className="flex gap-4">
            <Skeleton.Input active size="small" style={{ width: 100 }} />
            <Skeleton.Input active size="small" style={{ width: 200 }} />
          </div>
        ))}
      </div>
    </Card>
  );
}
