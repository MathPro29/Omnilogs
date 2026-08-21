import { useEffect, useState } from "react";
import { Button, Card, Col, Empty, Progress, Row, Space, Tag } from "antd";
import {
  ArrowUpOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  DatabaseOutlined,
  ReloadOutlined,
  WarningOutlined,
} from "@ant-design/icons";
import { PageTransition, DashboardSkeleton } from "@/components";
import { useQuery } from "@tanstack/react-query";
import { productService } from "@/services/product.service";
import "@/styles/pages/dashboard.css";

const volumeBars = [38, 52, 44, 68, 58, 76, 64, 86, 72, 92, 78, 88];
const activities = [
  {
    title: "Production ingestion is healthy",
    detail: "API key om-prod-•••• received logs",
    time: "2 minutes ago",
    tone: "success",
  },
  {
    title: "New product environment added",
    detail: "Checkout / Production",
    time: "18 minutes ago",
    tone: "info",
  },
  {
    title: "Elevated error rate detected",
    detail: "Payment service · 5xx responses",
    time: "42 minutes ago",
    tone: "warning",
  },
  {
    title: "Retention policy updated",
    detail: "Commerce product · 30 days",
    time: "1 hour ago",
    tone: "neutral",
  },
];

const mockLoad = (setLoading: (value: boolean) => void) => {
  setLoading(true);
  window.setTimeout(() => setLoading(false), 650);
};

export function DashboardPage() {
  const productsQuery = useQuery({
    queryKey: ["products", "dashboard-access"],
    queryFn: productService.listOptions,
  });
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    const timer = window.setTimeout(() => setLoading(false), 450);
    return () => window.clearTimeout(timer);
  }, []);

  return (
    <PageTransition>
      <main className="dashboard-page">
        <header className="dashboard-header">
          <div>
            <span className="dashboard-eyebrow">Workspace overview</span>
            <h1>Overview</h1>
            <p>
              Monitor log ingestion, product health, and recent activity from
              one place.
            </p>
          </div>
          <Space>
            <Tag color="blue">All products</Tag>
            <Button
              icon={<ReloadOutlined spin={loading} />}
              onClick={() => mockLoad(setLoading)}
            >
              Refresh
            </Button>
          </Space>
        </header>

        {productsQuery.isLoading || loading ? (
          <DashboardSkeleton />
        ) : !productsQuery.data?.length ? (
          <Card>
            <Empty description="No data is available yet. Ask an administrator to add you as a product member." />
          </Card>
        ) : (
          <>
            <section
              className="dashboard-section"
              aria-labelledby="summary-heading"
            >
              <div className="dashboard-section-heading">
                <div>
                  <span className="dashboard-section-kicker">At a glance</span>
                  <h2 id="summary-heading">Operational summary</h2>
                </div>
                <span className="dashboard-updated">Updated just now</span>
              </div>
              <Row gutter={[16, 16]}>
                <Col xs={24} sm={12} xl={6}>
                  <Card className="dashboard-stat-card">
                    <div className="dashboard-stat-icon purple">
                      <DatabaseOutlined />
                    </div>
                    <span>Total logs today</span>
                    <strong>2.48M</strong>
                    <small className="positive">
                      <ArrowUpOutlined /> 18.4% vs yesterday
                    </small>
                  </Card>
                </Col>
                <Col xs={24} sm={12} xl={6}>
                  <Card className="dashboard-stat-card">
                    <div className="dashboard-stat-icon blue">
                      <ClockCircleOutlined />
                    </div>
                    <span>Ingestion rate</span>
                    <strong>
                      1,240 <em>/ min</em>
                    </strong>
                    <small>Peak 2,860 / min</small>
                  </Card>
                </Col>
                <Col xs={24} sm={12} xl={6}>
                  <Card className="dashboard-stat-card">
                    <div className="dashboard-stat-icon green">
                      <CheckCircleOutlined />
                    </div>
                    <span>Healthy products</span>
                    <strong>
                      12 <em>/ 14</em>
                    </strong>
                    <Progress
                      percent={86}
                      showInfo={false}
                      strokeColor="#22c55e"
                      trailColor="#dcfce7"
                      size="small"
                    />
                  </Card>
                </Col>
                <Col xs={24} sm={12} xl={6}>
                  <Card className="dashboard-stat-card">
                    <div className="dashboard-stat-icon orange">
                      <WarningOutlined />
                    </div>
                    <span>Error rate</span>
                    <strong>0.82%</strong>
                    <small className="warning">2 products need attention</small>
                  </Card>
                </Col>
              </Row>
            </section>

            <section
              className="dashboard-section"
              aria-labelledby="traffic-heading"
            >
              <div className="dashboard-section-heading">
                <div>
                  <span className="dashboard-section-kicker">
                    Last 12 hours
                  </span>
                  <h2 id="traffic-heading">Log volume and health</h2>
                </div>
                <Space size={6}>
                  <span className="dashboard-legend">
                    <i className="legend-dot purple" /> Logs received
                  </span>
                  <span className="dashboard-legend">
                    <i className="legend-dot orange" /> Errors
                  </span>
                </Space>
              </div>
              <Row gutter={[16, 16]}>
                <Col xs={24} xl={16}>
                  <Card className="dashboard-chart-card">
                    <div className="dashboard-card-title">
                      <div>
                        <strong>Logs received</strong>
                        <span>Volume across all active environments</span>
                      </div>
                      <Tag color="success">Live</Tag>
                    </div>
                    <div
                      className="volume-chart"
                      role="img"
                      aria-label="Log volume bar chart"
                    >
                      {volumeBars.map((height, index) => (
                        <div
                          className="volume-column"
                          key={`${height}-${index}`}
                        >
                          <div
                            className="volume-bar"
                            style={{ height: `${height}%` }}
                          />
                          <span>
                            {index % 2 === 0 ? `${index + 8}:00` : ""}
                          </span>
                        </div>
                      ))}
                    </div>
                  </Card>
                </Col>
                <Col xs={24} xl={8}>
                  <Card className="dashboard-chart-card error-distribution-card">
                    <div className="dashboard-card-title">
                      <div>
                        <strong>Error distribution</strong>
                        <span>By response status</span>
                      </div>
                    </div>
                    <div className="donut-row">
                      <div
                        className="donut-chart"
                        aria-label="Error distribution chart"
                      >
                        <span>
                          0.82%<small>error rate</small>
                        </span>
                      </div>
                      <div className="donut-legend">
                        <span>
                          <i className="legend-dot red" /> 5xx <strong>42%</strong>
                        </span>
                        <span>
                          <i className="legend-dot orange" /> 4xx{" "}
                          <strong>35%</strong>
                        </span>
                        <span>
                          <i className="legend-dot gray" /> Other{" "}
                          <strong>23%</strong>
                        </span>
                      </div>
                    </div>
                  </Card>
                </Col>
              </Row>
            </section>

            <section
              className="dashboard-section"
              aria-labelledby="health-heading"
            >
              <div className="dashboard-section-heading">
                <div>
                  <span className="dashboard-section-kicker">
                    System status
                  </span>
                  <h2 id="health-heading">Service health</h2>
                </div>
              </div>
              <Row gutter={[16, 16]}>
                {[
                  [
                    "Log ingestion",
                    "All pipelines operational",
                    "success",
                    "99.98%",
                  ],
                  [
                    "Search cluster",
                    "Elasticsearch responding normally",
                    "success",
                    "99.95%",
                  ],
                  [
                    "Queue processing",
                    "3 items waiting for retry",
                    "warning",
                    "98.40%",
                  ],
                ].map(([title, detail, tone, uptime]) => (
                  <Col xs={24} md={8} key={title}>
                    <Card className="health-card">
                      <div className={`health-status ${tone}`}>
                        <span className="health-pulse" />
                        {tone === "success"
                          ? "Operational"
                          : "Needs attention"}
                      </div>
                      <strong>{title}</strong>
                      <span>{detail}</span>
                      <div className="health-footer">
                        <small>Uptime</small>
                        <b>{uptime}</b>
                      </div>
                    </Card>
                  </Col>
                ))}
              </Row>
            </section>

            <section
              className="dashboard-section"
              aria-labelledby="activity-heading"
            >
              <div className="dashboard-section-heading">
                <div>
                  <span className="dashboard-section-kicker">
                    What changed
                  </span>
                  <h2 id="activity-heading">Recent activity</h2>
                </div>
                <Button type="link">View all activity</Button>
              </div>
              <Card className="activity-card">
                {activities.map((activity) => (
                  <div className="activity-item" key={activity.title}>
                    <span className={`activity-marker ${activity.tone}`} />
                    <div>
                      <strong>{activity.title}</strong>
                      <span>{activity.detail}</span>
                    </div>
                    <time>{activity.time}</time>
                  </div>
                ))}
              </Card>
            </section>
          </>
        )}
      </main>
    </PageTransition>
  );
}
