import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Space, Select, Table, Card, Typography, Modal, Button } from "antd";
import { dashboardService, productAdminService } from "@/services";

const { Title, Text } = Typography;

const AuditsPage = () => {
    const [selectedProductsId, setSelectedProductsId] = useState<number | undefined>(undefined);
    const [selectedProjectsId, setSelectedProjectsId] = useState<number | undefined>(undefined);
    const [selectedFeatureId, setSelectedFeatureId] = useState<number | undefined>(undefined);
    const [page, setPage] = useState(1);
    const [pageSize, setPageSize] = useState(10);
    const [selectedAudit, setSelectedAudit] = useState<any | null>(null);

    // ดึงรายชื่อ Products เพื่อนำมาใช้ใน Dropdown
    const { data: products } = useQuery({
        queryKey: ['products'],
        queryFn: productAdminService.listProducts,
    });

    // ดึงรายชื่อ Projects ของ Product ที่เลือก
    const { data: projects } = useQuery({
        queryKey: ['projects', selectedProductsId],
        queryFn: () => productAdminService.listProjects(selectedProductsId!),
        enabled: !!selectedProductsId,
    });

    // ดึงรายชื่อ Feature ของ Project ที่เลือก
    const { data: features } = useQuery({
        queryKey: ['features', selectedProjectsId],
        queryFn: () => productAdminService.listFeatures(selectedProductsId!, selectedProjectsId!),
        enabled: !!selectedProjectsId && !!selectedProductsId,
    });

    // ดึงข้อมูล Audit Logs จริงจาก Backend
    const { data: auditLogsData, isLoading } = useQuery({
        queryKey: ['auditLogs', selectedProductsId, selectedProjectsId, selectedFeatureId, page, pageSize],
        queryFn: () =>
            dashboardService.getAuditLogs({
                product_id: selectedProductsId,
                project_id: selectedProjectsId,
                feature_id: selectedFeatureId,
                limit: pageSize,
                offset: (page - 1) * pageSize,
            }),
    });

    const columns = [
        {
            title: 'เวลาบันทึก',
            dataIndex: 'created_at',
            key: 'created_at',
            render: (text: string) => (text ? new Date(text).toLocaleString('th-TH') : '-'),
        },
        {
            title: 'การกระทำ (Action)',
            dataIndex: 'action',
            key: 'action',
        },
        {
            title: 'ผู้ใช้ (Actor User ID)',
            dataIndex: 'actor_user_id',
            key: 'actor_user_id',
            render: (val: any) => val ?? 'System / Guest',
        },
        {
            title: 'Product ID',
            dataIndex: 'product_id',
            key: 'product_id',
            render: (val: any) => val ?? '-',
        },
        {
            title: 'Project ID',
            key: 'project_id',
            render: (_: any, record: any) => {
                const queryProj = record.metadata?.query?.project_id;
                const requestProj = record.metadata?.request?.project_id;
                if (queryProj) return queryProj;
                if (requestProj) return requestProj;
                const match = record.path?.match(/\/projects\/(\d+)/);
                return match ? match[1] : '-';
            },
        },
        {
            title: 'HTTP Method',
            dataIndex: 'method',
            key: 'method',
            render: (val: string) => <span style={{ fontWeight: 'bold' }}>{val || '-'}</span>,
        },
        {
            title: 'Path',
            dataIndex: 'path',
            key: 'path',
        },
        {
            title: 'ผลลัพธ์ (Result)',
            dataIndex: 'result',
            key: 'result',
            render: (val: string) => {
                let color = 'blue';
                if (val === 'SUCCESS') color = 'green';
                if (val === 'DENIED') color = 'orange';
                if (val === 'FAILED') color = 'red';
                return <span style={{ color, fontWeight: 'bold' }}>{val}</span>;
            },
        },
        {
            title: 'ดูรายละเอียด',
            key: 'actions',
            render: (_: any, record: any) => (
                <Button size="small" onClick={() => setSelectedAudit(record)}>
                    ดูข้อมูล JSON
                </Button>
            ),
        },
    ];

    return (
        <div style={{ padding: '20px', background: '#f5f5f5', minHeight: '100vh' }}>
            <Title level={3}>Audit Logs</Title>
            <Text type="secondary">ประวัติและข้อมูลการบันทึกกิจกรรมทั้งหมดภายในระบบ</Text>

            <Space direction="vertical" size="large" style={{ width: '100%', marginTop: '20px' }}>
                {/* ส่วนตัวกรอง */}
                <Card title="กรองข้อมูลกิจกรรม">
                    <Space size="middle">
                        <Space>
                            <span>เลือก Product:</span>
                            <Select
                                style={{ width: 250 }}
                                placeholder="ทั้งหมด"
                                allowClear
                                value={selectedProductsId}
                                onChange={(value) => {
                                    setSelectedProductsId(value);
                                    setSelectedProjectsId(undefined);
                                    setPage(1);
                                }}
                                options={
                                    products?.map((p) => ({
                                        value: p.productId,
                                        label: `[ID: ${p.productId}] ${p.productName} (${p.productCode})`,
                                    })) || []
                                }
                            />
                        </Space>

                        <Space>
                            <span>เลือก Project:</span>
                            <Select
                                style={{ width: 250 }}
                                placeholder={selectedProductsId ? "ทั้งหมด" : "กรุณาเลือก Product ก่อน"}
                                disabled={!selectedProductsId}
                                allowClear
                                value={selectedProjectsId}
                                onChange={(value) => {
                                    setSelectedProjectsId(value);
                                    setPage(1);
                                }}
                                options={
                                    projects?.map((p) => ({
                                        value: p.projectId,
                                        label: `[ID: ${p.projectId}] ${p.projectName} (${p.projectCode})`,
                                    })) || []
                                }
                            />
                        </Space>

                        <Space>
                            <span>เลือก Feature:</span>
                            <Select
                                style={{ width: 250 }}
                                placeholder={selectedProjectsId ? "ทั้งหมด" : "กรุณาเลือก Feature ก่อน"}
                                disabled={!selectedProjectsId}
                                allowClear
                                value={selectedFeatureId}
                                onChange={(value) => {
                                    setSelectedFeatureId(value);
                                    setPage(1);
                                }}
                                options={
                                    features?.map((f) => ({
                                        value: f.categoryId,
                                        label: `[ID: ${f.categoryId}] ${f.categoryName} (${f.categoryCode})`,
                                    })) || []
                                }
                            />
                        </Space>

                    </Space>
                </Card>

                {/* ตารางแสดงผล */}
                <Card title="ข้อมูลกิจกรรมในระบบ">
                    <Table
                        dataSource={auditLogsData?.data || []}
                        columns={columns}
                        rowKey="audit_id"
                        loading={isLoading}
                        pagination={{
                            current: page,
                            pageSize: pageSize,
                            total: auditLogsData?.total || 0,
                            showSizeChanger: true,
                            onChange: (p, ps) => {
                                setPage(p);
                                setPageSize(ps);
                            },
                        }}
                    />
                </Card>
            </Space>

            {/* Modal แสดงข้อมูลรายละเอียด JSON */}
            <Modal
                title="รายละเอียด System Audit Log"
                open={!!selectedAudit}
                onCancel={() => setSelectedAudit(null)}
                footer={[
                    <Button key="close" onClick={() => setSelectedAudit(null)}>
                        ปิด
                    </Button>,
                ]}
                width={700}
            >
                {selectedAudit && (
                    <pre
                        style={{
                            background: '#f5f5f5',
                            padding: '15px',
                            borderRadius: '5px',
                            overflowX: 'auto',
                            whiteSpace: 'pre-wrap',
                        }}
                    >
                        {JSON.stringify(selectedAudit, null, 2)}
                    </pre>
                )}
            </Modal>
        </div>
    );
};

export default AuditsPage;
