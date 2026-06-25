import { useEffect, useMemo } from 'react';
import { Card, Row, Col, Typography, Tag, Timeline } from 'antd';
import {
  UsersIcon,
  UserPlusIcon,
  BuildingOfficeIcon,
  CheckCircleIcon,
} from '@heroicons/react/24/outline';
import { useQuery } from '@tanstack/react-query';
import { motion } from 'motion/react';
import { dashboardService } from '@/services';
import { useAppStore } from '@/store';
import { QUERY_KEYS, ROUTES } from '@/constants';
import { formatNumber, timeAgo } from '@/utils';
import { PageTransition, DashboardSkeleton } from '@/components';

const { Title, Text, Paragraph } = Typography;

// ===== Stat Card Component =====
interface StatCardProps {
  title: string;
  value: number;
  icon: React.ReactNode;
  color: string;
  bgColor: string;
  index: number;
}

function StatCard({ title, value, icon, color, bgColor, index }: StatCardProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 16 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.08, duration: 0.3 }}
    >
      <Card className="stat-card" bordered={false}>
        <div className="flex items-center justify-between">
          <div>
            <Text type="secondary" className="text-sm">
              {title}
            </Text>
            <Title level={2} className="mb-0 mt-1" style={{ color }}>
              {formatNumber(value)}
            </Title>
          </div>
          <div
            className="flex items-center justify-center w-12 h-12 rounded-xl"
            style={{ backgroundColor: bgColor }}
          >
            <div style={{ color }}>{icon}</div>
          </div>
        </div>
      </Card>
    </motion.div>
  );
}

// ===== Simple Bar Chart =====
function BarChart({ data }: { data: { label: string; value: number }[] }) {
  const maxValue = Math.max(...data.map((d) => d.value));

  return (
    <div className="flex items-end gap-3 h-48 mt-4">
      {data.map((item, index) => (
        <div key={item.label} className="flex-1 flex flex-col items-center gap-2">
          <Text className="text-xs" style={{ color: '#00000073' }}>
            {item.value}
          </Text>
          <motion.div
            className="w-full rounded-t-lg"
            style={{
              background: 'linear-gradient(180deg, var(--color-primary), var(--color-purple-250))',
              minHeight: 4,
            }}
            initial={{ height: 0 }}
            animate={{ height: `${(item.value / maxValue) * 140}px` }}
            transition={{ delay: index * 0.1, duration: 0.5, ease: 'easeOut' }}
          />
          <Text className="text-xs" style={{ color: '#00000073' }}>
            {item.label}
          </Text>
        </div>
      ))}
    </div>
  );
}

// ===== Dashboard Page =====
export function DashboardPage() {
  const setBreadcrumbs = useAppStore((state) => state.setBreadcrumbs);

  useEffect(() => {
    setBreadcrumbs([{ title: 'แดชบอร์ด', path: ROUTES.DASHBOARD }]);
  }, [setBreadcrumbs]);

  const { data, isLoading } = useQuery({
    queryKey: QUERY_KEYS.DASHBOARD_SUMMARY,
    queryFn: dashboardService.getSummary,
  });

  if (isLoading || !data) {
    return (
      <PageTransition>
        <DashboardSkeleton />
      </PageTransition>
    );
  }

  // useMemo: statCards สร้างใหม่เฉพาะเมื่อ data เปลี่ยน
  const statCards = useMemo<Omit<StatCardProps, 'index'>[]>(() => [
    {
      title: 'ผู้ใช้ทั้งหมด',
      value: data.totalUsers,
      icon: <UsersIcon className="w-6 h-6" />,
      color: 'var(--color-primary)',
      bgColor: 'var(--color-purple-50)',
    },
    {
      title: 'ผู้ใช้ที่ใช้งาน',
      value: data.activeUsers,
      icon: <CheckCircleIcon className="w-6 h-6" />,
      color: 'var(--color-success)',
      bgColor: '#E6F9F3',
    },
    {
      title: 'ผู้ใช้ใหม่เดือนนี้',
      value: data.newUsersThisMonth,
      icon: <UserPlusIcon className="w-6 h-6" />,
      color: 'var(--color-info)',
      bgColor: '#E6F4FF',
    },
    {
      title: 'แผนก',
      value: data.departments,
      icon: <BuildingOfficeIcon className="w-6 h-6" />,
      color: 'var(--color-warning)',
      bgColor: '#FFFBE6',
    },
  ], [data.totalUsers, data.activeUsers, data.newUsersThisMonth, data.departments]);

  return (
    <PageTransition>
      <div className="space-y-6">
        {/* Page Title */}
        <div>
          <Title level={4} className="mb-1">
            แดชบอร์ด
          </Title>
          <Text type="secondary">ภาพรวมของระบบจัดการทรัพยากรบุคคล</Text>
        </div>

        {/* Stat Cards */}
        <Row gutter={[16, 16]}>
          {statCards.map((card, index) => (
            <Col xs={24} sm={12} lg={6} key={card.title}>
              <StatCard {...card} index={index} />
            </Col>
          ))}
        </Row>

        {/* Charts Row */}
        <Row gutter={[16, 16]}>
          {/* User Growth */}
          <Col xs={24} lg={14}>
            <motion.div
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.3, duration: 0.3 }}
            >
              <Card bordered={false} className="stat-card">
                <Title level={5}>การเติบโตผู้ใช้งาน</Title>
                <BarChart data={data.userGrowth} />
              </Card>
            </motion.div>
          </Col>

          {/* User by Department */}
          <Col xs={24} lg={10}>
            <motion.div
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.4, duration: 0.3 }}
            >
              <Card bordered={false} className="stat-card">
                <Title level={5}>ผู้ใช้ตามแผนก</Title>
                <div className="space-y-3 mt-4">
                  {data.usersByDepartment.map((dept) => (
                    <div key={dept.label} className="flex items-center justify-between">
                      <Text>{dept.label}</Text>
                      <div className="flex items-center gap-3 flex-1 mx-4">
                        <div className="flex-1 h-2 rounded-full bg-gray-100 overflow-hidden">
                          <motion.div
                            className="h-full rounded-full"
                            style={{ background: 'linear-gradient(90deg, var(--color-primary), var(--color-purple-250))' }}
                            initial={{ width: 0 }}
                            animate={{
                              width: `${(dept.value / Math.max(...data.usersByDepartment.map((d) => d.value))) * 100}%`,
                            }}
                            transition={{ duration: 0.6, ease: 'easeOut' }}
                          />
                        </div>
                        <Tag color="purple">{dept.value}</Tag>
                      </div>
                    </div>
                  ))}
                </div>
              </Card>
            </motion.div>
          </Col>
        </Row>

        {/* Recent Activities */}
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.5, duration: 0.3 }}
        >
          <Card bordered={false} className="stat-card">
            <Title level={5} className="mb-4">
              กิจกรรมล่าสุด
            </Title>
            <Timeline
              items={useMemo(() => data.recentActivities.map((activity) => ({
                color: 'var(--color-primary)',
                children: (
                  <div>
                    <Paragraph className="mb-0">
                      <Text strong>{activity.user}</Text>{' '}
                      <Text>{activity.action}</Text>{' '}
                      <Text type="secondary">{activity.target}</Text>
                    </Paragraph>
                    <Text type="secondary" className="text-xs">
                      {timeAgo(activity.timestamp)}
                    </Text>
                  </div>
                ),
              })), [data.recentActivities])}
            />
          </Card>
        </motion.div>
      </div>
    </PageTransition>
  );
}
