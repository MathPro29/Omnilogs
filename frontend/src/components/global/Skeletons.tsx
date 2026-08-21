import { Skeleton, Card, Row, Col, Table } from 'antd';

export function DashboardSkeleton() {
  return (
    <div className="dashboard-skeleton">
      <section className="dashboard-section"><div className="dashboard-section-heading"><div><Skeleton.Input active size="small" style={{ width: 100 }} /><Skeleton.Input active size="large" style={{ width: 220, display: "block", marginTop: 10 }} /></div><Skeleton.Input active size="small" style={{ width: 130 }} /></div><Row gutter={[16, 16]}>{Array.from({ length: 4 }).map((_, index) => <Col xs={24} sm={12} xl={6} key={index}><Card className="dashboard-stat-card"><Skeleton active avatar paragraph={{ rows: 2 }} title={{ width: "56%" }} /></Card></Col>)}</Row></section>
      <section className="dashboard-section"><div className="dashboard-section-heading"><div><Skeleton.Input active size="small" style={{ width: 110 }} /><Skeleton.Input active size="large" style={{ width: 230, display: "block", marginTop: 10 }} /></div></div><Row gutter={[16, 16]}><Col xs={24} xl={16}><Card className="dashboard-chart-card"><Skeleton active title={{ width: "30%" }} paragraph={{ rows: 1 }} /><div className="chart-skeleton-bars">{Array.from({ length: 12 }).map((_, index) => <i key={index} style={{ height: `${28 + ((index * 17) % 62)}%` }} />)}</div></Card></Col><Col xs={24} xl={8}><Card className="dashboard-chart-card"><Skeleton active title={{ width: "48%" }} paragraph={{ rows: 1 }} /><div className="chart-skeleton-donut"><i /></div></Card></Col></Row></section>
      <section className="dashboard-section"><Skeleton.Input active size="small" style={{ width: 170, marginBottom: 14 }} /><Row gutter={[16, 16]}>{Array.from({ length: 3 }).map((_, index) => <Col xs={24} md={8} key={index}><Card><Skeleton active paragraph={{ rows: 2 }} /></Card></Col>)}</Row></section>
      <section className="dashboard-section"><Skeleton.Input active size="small" style={{ width: 180, marginBottom: 14 }} /><Card><Skeleton active paragraph={{ rows: 4 }} /></Card></section>
    </div>
  );
}

export function TableSkeleton() {
  const columns = Array.from({ length: 5 }).map((_, i) => ({ title: <Skeleton.Input active size="small" style={{ width: 80 }} />, key: i, render: () => <Skeleton.Input active size="small" style={{ width: i === 0 ? 120 : 100 }} /> }));
  const dataSource = Array.from({ length: 5 }).map((_, i) => ({ key: i }));
  return <div className="space-y-4"><Card><div className="flex gap-3 flex-wrap"><Skeleton.Input active size="default" style={{ width: 220 }} /><Skeleton.Input active size="default" style={{ width: 150 }} /><Skeleton.Input active size="default" style={{ width: 150 }} /><Skeleton.Button active size="default" style={{ width: 80 }} /></div></Card><Card><Table columns={columns} dataSource={dataSource} pagination={false} size="middle" /></Card></div>;
}

export function FormSkeleton() {
  return <Card><div className="space-y-6 max-w-2xl">{Array.from({ length: 5 }).map((_, i) => <div key={i} className="space-y-2"><Skeleton.Input active size="small" style={{ width: 120 }} /><Skeleton.Input active size="default" style={{ width: '100%' }} /></div>)}<div className="flex gap-3 pt-4"><Skeleton.Button active size="default" style={{ width: 100 }} /><Skeleton.Button active size="default" style={{ width: 80 }} /></div></div></Card>;
}

export function DetailSkeleton() {
  return <Card><div className="flex items-start gap-6 mb-8"><Skeleton.Avatar active size={80} shape="circle" /><div className="flex-1 space-y-3"><Skeleton.Input active size="default" style={{ width: 200 }} /><Skeleton.Input active size="small" style={{ width: 160 }} /></div></div><div className="space-y-4">{Array.from({ length: 6 }).map((_, i) => <div key={i} className="flex gap-4"><Skeleton.Input active size="small" style={{ width: 100 }} /><Skeleton.Input active size="small" style={{ width: 200 }} /></div>)}</div></Card>;
}