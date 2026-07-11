import { useEffect, useState } from 'react';
import {
  Card,
  Form,
  Input,
  InputNumber,
  Switch,
  Select,
  Button,
  Typography,
  Divider,
  message,
  Tabs,
} from 'antd';
import { Cog6ToothIcon, ShieldCheckIcon } from '@heroicons/react/24/outline';
import { motion } from 'motion/react';
import { settingsSchema, type SettingsFormData } from '@/schemas';
import { useAppStore } from '@/store';
import { usePermission } from '@/hooks';
import { PERMISSIONS, ROUTES, DEFAULT_ROLES } from '@/constants';
import { PageTransition, PermissionGuard, FormSkeleton } from '@/components';

const { Title, Text, Paragraph } = Typography;

// Mock Settings Data
const MOCK_SETTINGS: SettingsFormData = {
  siteName: 'HR Admin Panel',
  siteDescription: 'ระบบจัดการทรัพยากรบุคคล',
  maintenanceMode: false,
  allowRegistration: true,
  defaultRole: DEFAULT_ROLES.VIEWER,
  sessionTimeout: 30,
  maxLoginAttempts: 5,
};

export function SettingsPage() {
  const setBreadcrumbs = useAppStore((state) => state.setBreadcrumbs);
  const { can } = usePermission();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  const canEdit = can(PERMISSIONS.SETTINGS_EDIT);

  useEffect(() => {
    setBreadcrumbs([{ title: 'ตั้งค่าระบบ', path: ROUTES.SETTINGS }]);
  }, [setBreadcrumbs]);

  // จำลองการโหลดข้อมูล
  useEffect(() => {
    const timer = setTimeout(() => {
      form.setFieldsValue(MOCK_SETTINGS);
      setLoading(false);
    }, 800);
    return () => clearTimeout(timer);
  }, [form]);

  const handleSubmit = async (values: SettingsFormData) => {
    const result = settingsSchema.safeParse(values);
    if (!result.success) {
      result.error.issues.forEach((err) => {
        message.error(err.message);
      });
      return;
    }

    setSaving(true);
    try {
      await new Promise((resolve) => setTimeout(resolve, 800));
      message.success('บันทึกการตั้งค่าสำเร็จ');
    } catch {
      message.error('เกิดข้อผิดพลาดในการบันทึก');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <PageTransition>
        <div className="space-y-6">
          <div>
            <Title level={4} className="mb-1">ตั้งค่าระบบ</Title>
            <Text type="secondary">จัดการการตั้งค่าระบบทั่วไป</Text>
          </div>
          <FormSkeleton />
        </div>
      </PageTransition>
    );
  }

  return (
    <PageTransition>
      <div className="space-y-6">
        {/* Page Header */}
        <div>
          <Title level={4} className="mb-1">
            ตั้งค่าระบบ
          </Title>
          <Text type="secondary">จัดการการตั้งค่าระบบทั่วไป</Text>
        </div>

        {/* Settings Form */}
        <motion.div
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1, duration: 0.25 }}
        >
          <Card bordered={false} className="stat-card">
            <Tabs
              items={[
                {
                  key: 'general',
                  label: (
                    <span className="flex items-center gap-2">
                      <Cog6ToothIcon className="w-4 h-4" />
                      ทั่วไป
                    </span>
                  ),
                  children: (
                    <Form
                      form={form}
                      layout="vertical"
                      onFinish={handleSubmit}
                      disabled={!canEdit}
                      className="max-w-2xl"
                    >
                      <Divider plain>
                        ข้อมูลระบบ
                      </Divider>

                      <Form.Item
                        name="siteName"
                        label="ชื่อระบบ"
                        rules={[{ required: true, message: 'กรุณากรอกชื่อระบบ' }]}
                      >
                        <Input placeholder="กรอกชื่อระบบ" />
                      </Form.Item>

                      <Form.Item name="siteDescription" label="คำอธิบาย">
                        <Input.TextArea rows={3} placeholder="คำอธิบายเกี่ยวกับระบบ" />
                      </Form.Item>

                      <Divider plain>
                        การเข้าถึง
                      </Divider>

                      <Form.Item
                        name="maintenanceMode"
                        label="โหมดปิดปรับปรุง"
                        valuePropName="checked"
                      >
                        <Switch />
                      </Form.Item>

                      <Form.Item
                        name="allowRegistration"
                        label="อนุญาตให้ลงทะเบียน"
                        valuePropName="checked"
                      >
                        <Switch />
                      </Form.Item>

                      <Form.Item
                        name="defaultRole"
                        label="บทบาทเริ่มต้น"
                        rules={[{ required: true }]}
                      >
                        <Select
                          options={[
                            { label: 'ผู้ดู', value: DEFAULT_ROLES.VIEWER },
                            { label: 'ผู้จัดการ', value: DEFAULT_ROLES.MANAGER },
                            { label: 'แอดมิน', value: DEFAULT_ROLES.ADMIN },
                          ]}
                        />
                      </Form.Item>

                      <Divider plain>
                        ความปลอดภัย
                      </Divider>

                      <Form.Item
                        name="sessionTimeout"
                        label="หมดเวลาเซสชัน (นาที)"
                        rules={[{ required: true }]}
                      >
                        <InputNumber min={5} max={1440} className="w-full" />
                      </Form.Item>

                      <Form.Item
                        name="maxLoginAttempts"
                        label="จำนวนครั้งที่ login ผิดพลาดสูงสุด"
                        rules={[{ required: true }]}
                      >
                        <InputNumber min={1} max={20} className="w-full" />
                      </Form.Item>

                      <PermissionGuard permission={PERMISSIONS.SETTINGS_EDIT}>
                        <div className="flex gap-3 pt-4">
                          <Button type="primary" htmlType="submit" loading={saving}>
                            บันทึก
                          </Button>
                          <Button onClick={() => form.resetFields()}>
                            ยกเลิก
                          </Button>
                        </div>
                      </PermissionGuard>
                    </Form>
                  ),
                },
                {
                  key: 'security',
                  label: (
                    <span className="flex items-center gap-2">
                      <ShieldCheckIcon className="w-4 h-4" />
                      ความปลอดภัย
                    </span>
                  ),
                  children: (
                    <div className="max-w-2xl py-8 text-center">
                      <ShieldCheckIcon className="w-12 h-12 mx-auto mb-4" style={{ color: '#AE9CFF' }} />
                      <Title level={5}>การตั้งค่าความปลอดภัยเพิ่มเติม</Title>
                      <Paragraph type="secondary">
                        ส่วนนี้สามารถขยายเพิ่มเติมได้ในอนาคต เช่น 2FA, IP Whitelist, Audit Log
                      </Paragraph>
                    </div>
                  ),
                },
              ]}
            />
          </Card>
        </motion.div>
      </div>
    </PageTransition>
  );
}
