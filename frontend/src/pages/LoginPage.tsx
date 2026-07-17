import { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { Card, Form, Input, Button, Typography, message, Divider } from 'antd';
import {
  UserIcon,
  LockClosedIcon,
  ArrowRightIcon,
} from '@heroicons/react/24/outline';
import { motion } from 'motion/react';
import { loginSchema, type LoginFormData } from '@/schemas';
import { authService } from '@/services';
import { useAuthStore } from '@/store';
import { ROUTES } from '@/constants';

const { Title, Text, Paragraph } = Typography;

export function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const setAuth = useAuthStore((state) => state.setAuth);
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();

  const from = (location.state as { from?: { pathname: string } })?.from?.pathname || ROUTES.DASHBOARD;

  const handleSubmit = async (values: LoginFormData) => {
    // Validate ด้วย Zod
    const result = loginSchema.safeParse(values);
    if (!result.success) {
      const errors = result.error.issues;
      errors.forEach((err) => {
        message.error(err.message);
      });
      return;
    }

    setLoading(true);
    try {
      const response = await authService.login(result.data);
      setAuth(response);
      message.success('เข้าสู่ระบบสำเร็จ');
      navigate(from, { replace: true });
    } catch (error: unknown) {
      const err = error as { message?: string };
      message.error(err?.message || 'เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card className="login-card" bordered={false}>
      {/* Header */}
      <div className="text-center pt-4 pb-2">
        <motion.div
          initial={{ scale: 0.8, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          transition={{ delay: 0.1, duration: 0.4 }}
        >
          <div
            className="inline-flex items-center justify-center w-16 h-16 rounded-2xl mb-4"
            style={{ background: 'linear-gradient(135deg, var(--color-primary), var(--color-purple-250))' }}
          >
            <LockClosedIcon className="w-8 h-8 text-white" />
          </div>
        </motion.div>
        <Title level={3} style={{ marginBottom: 4 }}>
          เข้าสู่ระบบ
        </Title>
        <Text type="secondary">HR Admin Panel</Text>
      </div>

      <Divider />

      {/* Form */}
      <motion.div
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.2, duration: 0.3 }}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          autoComplete="off"
          size="large"
        >
          <Form.Item
            name="username"
            label="ชื่อผู้ใช้"
            rules={[{ required: true, message: 'กรุณากรอกชื่อผู้ใช้' }]}
          >
            <Input
              prefix={<UserIcon className="w-4 h-4 text-gray-400" />}
              placeholder="กรอกชื่อผู้ใช้"
            />
          </Form.Item>

          <Form.Item
            name="password"
            label="รหัสผ่าน"
            rules={[{ required: true, message: 'กรุณากรอกรหัสผ่าน' }]}
          >
            <Input.Password
              prefix={<LockClosedIcon className="w-4 h-4 text-gray-400" />}
              placeholder="กรอกรหัสผ่าน"
            />
          </Form.Item>

          <Form.Item className="mb-2">
            <Button
              type="primary"
              htmlType="submit"
              loading={loading}
              block
              style={{
                height: 48,
                fontSize: '1rem',
                fontWeight: 600,
                borderRadius: 10,
              }}
            >
              <span className="flex items-center justify-center gap-2">
                เข้าสู่ระบบ
                <ArrowRightIcon className="w-4 h-4" />
              </span>
            </Button>
          </Form.Item>
          <div className="flex items-center justify-between text-sm">
            <a href={ROUTES.FORGOT_PASSWORD}>ลืมรหัสผ่าน?</a>
            <a href={ROUTES.REGISTER} className="font-medium">สมัครสมาชิก</a>
          </div>
        </Form>

        {/* Demo credentials */}
        <div
          className="mt-4 p-3 rounded-lg text-center"
          style={{ background: 'var(--color-purple-50)' }}
        >
          <Paragraph
            className="mb-0 text-xs"
            style={{ color: 'var(--color-text-secondary)' }}
          >
            ทดสอบเข้าสู่ระบบ: admin / admin123
          </Paragraph>
        </div>
      </motion.div>
    </Card>
  );
}
