import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Form, Input, Button, Typography, message, Divider } from 'antd';
import {
  UserIcon,
  LockClosedIcon,
  EnvelopeIcon,
  IdentificationIcon,
  PhoneIcon,
  ArrowLeftIcon,
} from '@heroicons/react/24/outline';
import { motion } from 'motion/react';
import { registerSchema, type RegisterFormData } from '@/schemas';
import { authService } from '@/services';
import { ROUTES } from '@/constants';

const { Title, Text } = Typography;

export function RegisterPage() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();

  const handleSubmit = async (values: RegisterFormData) => {
    // Validate using Zod
    const result = registerSchema.safeParse(values);
    if (!result.success) {
      const errors = result.error.issues;
      errors.forEach((err) => {
        message.error(err.message);
      });
      return;
    }

    setLoading(true);
    try {
      await authService.register(result.data);
      message.success('สมัครสมาชิกสำเร็จแล้ว คุณสามารถเข้าสู่ระบบได้ทันที');
      navigate(ROUTES.LOGIN, { replace: true });
    } catch (error: any) {
      const errMsg = error?.response?.data?.error?.message || error?.message || 'เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง';
      message.error(errMsg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card className="login-card !max-w-lg" bordered={false} style={{ width: '100%' }}>
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
            <UserIcon className="w-8 h-8 text-white" />
          </div>
        </motion.div>
        <Title level={3} style={{ marginBottom: 4 }}>
          สมัครสมาชิก
        </Title>
        <Text type="secondary">สร้างบัญชีผู้ใช้งานระบบ Omnilogs</Text>
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
          <div className="grid grid-cols-1 md:grid-cols-2 gap-x-4">
            <Form.Item
              name="first_name"
              label="ชื่อจริง"
              rules={[{ required: true, message: 'กรุณากรอกชื่อจริง' }]}
            >
              <Input
                prefix={<IdentificationIcon className="w-4 h-4 text-gray-400" />}
                placeholder="สมชาย"
              />
            </Form.Item>

            <Form.Item
              name="last_name"
              label="นามสกุล"
              rules={[{ required: true, message: 'กรุณากรอกนามสกุล' }]}
            >
              <Input
                prefix={<IdentificationIcon className="w-4 h-4 text-gray-400" />}
                placeholder="ใจดี"
              />
            </Form.Item>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-x-4">
            <Form.Item
              name="username"
              label="ชื่อผู้ใช้ (Username)"
            >
              <Input
                prefix={<UserIcon className="w-4 h-4 text-gray-400" />}
                placeholder="somchai_dev"
              />
            </Form.Item>

            <Form.Item
              name="phone_number"
              label="เบอร์โทรศัพท์ (ถ้ามี)"
            >
              <Input
                prefix={<PhoneIcon className="w-4 h-4 text-gray-400" />}
                placeholder="0891234567"
              />
            </Form.Item>
          </div>

          <Form.Item
            name="email"
            label="อีเมล (Email)"
            rules={[
              { required: true, message: 'กรุณากรอกอีเมล' },
              { type: 'email', message: 'รูปแบบอีเมลไม่ถูกต้อง' }
            ]}
          >
            <Input
              prefix={<EnvelopeIcon className="w-4 h-4 text-gray-400" />}
              placeholder="somchai@company.com"
            />
          </Form.Item>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-x-4">
            <Form.Item
              name="password"
              label="รหัสผ่าน"
              rules={[
                { required: true, message: 'กรุณากรอกรหัสผ่าน' },
                { min: 8, message: 'รหัสผ่านต้องมีอย่างน้อย 8 ตัวอักษร' }
              ]}
            >
              <Input.Password
                prefix={<LockClosedIcon className="w-4 h-4 text-gray-400" />}
                placeholder="อย่างน้อย 8 ตัวอักษร"
              />
            </Form.Item>

            <Form.Item
              name="confirm_password"
              label="ยืนยันรหัสผ่าน"
              dependencies={['password']}
              rules={[
                { required: true, message: 'กรุณากรอกยืนยันรหัสผ่าน' },
                ({ getFieldValue }) => ({
                  validator(_, value) {
                    if (!value || getFieldValue('password') === value) {
                      return Promise.resolve();
                    }
                    return Promise.reject(new Error('รหัสผ่านไม่ตรงกัน'));
                  },
                }),
              ]}
            >
              <Input.Password
                prefix={<LockClosedIcon className="w-4 h-4 text-gray-400" />}
                placeholder="กรอกรหัสผ่านอีกครั้ง"
              />
            </Form.Item>
          </div>

          <Form.Item className="mb-2 mt-4">
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
              สมัครสมาชิก
            </Button>
          </Form.Item>
          
          <div className="text-center pt-3">
            <a href={ROUTES.LOGIN} className="inline-flex items-center gap-1.5 text-sm hover:underline text-indigo-600">
              <ArrowLeftIcon className="w-3.5 h-3.5" />
              กลับไปหน้าเข้าสู่ระบบ
            </a>
          </div>
        </Form>
      </motion.div>
    </Card>
  );
}
