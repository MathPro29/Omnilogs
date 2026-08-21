import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Card, Form, Input, Button, Typography, message, Divider } from "antd";
import {
  UserIcon,
  LockClosedIcon,
  EnvelopeIcon,
  IdentificationIcon,
  PhoneIcon,
  ArrowLeftIcon,
} from "@heroicons/react/24/outline";
import { motion } from "motion/react";
import { registerSchema, type RegisterFormData } from "@/schemas";
import { authService } from "@/services";
import { ROUTES } from "@/constants";
import { getErrorMessage } from "@/utils/error";

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
      message.success("สมัครสมาชิกสำเร็จแล้ว คุณสามารถเข้าสู่ระบบได้ทันที");
      navigate(ROUTES.LOGIN, { replace: true });
    } catch (error: unknown) {
      const errMsg = getErrorMessage(
        error,
        "Unable to register. Please try again.",
      );
      message.error(errMsg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card
      className="login-card !max-w-xl"
      bordered={false}
      style={{ width: "100%" }}
    >
      {/* Header */}
      <div className="text-center pt-2 pb-1">
        <motion.div
          initial={{ scale: 0.85, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          transition={{ delay: 0.05, duration: 0.35 }}
        >
          <div className="inline-flex items-center justify-center w-14 h-14 rounded-2xl mb-3 bg-[#212121] text-[#1F8457] shadow-md border border-[#3a3a3a]">
            <UserIcon className="w-7 h-7 stroke-[1.75]" />
          </div>
        </motion.div>
        <Title
          level={3}
          className="!text-zinc-900 !font-bold !tracking-tight !mb-0.5"
        >
          สมัครสมาชิก
        </Title>
        <Text className="text-xs uppercase tracking-widest text-zinc-500 font-semibold">
          สร้างบัญชีผู้ใช้งานระบบ Omnilogs
        </Text>
      </div>

      <Divider className="!my-4 !border-zinc-100" />

      {/* Form */}
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.15, duration: 0.3 }}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          autoComplete="off"
          size="large"
          requiredMark={false}
        >
          <div className="grid grid-cols-1 md:grid-cols-2 gap-x-4">
            <Form.Item
              name="first_name"
              label={
                <span className="text-zinc-700 font-medium text-xs uppercase tracking-wider">
                  ชื่อจริง
                </span>
              }
              rules={[{ required: true, message: "กรุณากรอกชื่อจริง" }]}
              className="mb-3.5"
            >
              <Input
                prefix={
                  <IdentificationIcon className="w-4 h-4 text-zinc-400 mr-1" />
                }
                placeholder="สมชาย"
                className="!rounded-xl !border-zinc-200 hover:!border-zinc-400 focus:!border-zinc-950 !h-11"
              />
            </Form.Item>

            <Form.Item
              name="last_name"
              label={
                <span className="text-zinc-700 font-medium text-xs uppercase tracking-wider">
                  นามสกุล
                </span>
              }
              rules={[{ required: true, message: "กรุณากรอกนามสกุล" }]}
              className="mb-3.5"
            >
              <Input
                prefix={
                  <IdentificationIcon className="w-4 h-4 text-zinc-400 mr-1" />
                }
                placeholder="ใจดี"
                className="!rounded-xl !border-zinc-200 hover:!border-zinc-400 focus:!border-zinc-950 !h-11"
              />
            </Form.Item>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-x-4">
            <Form.Item
              name="username"
              label={
                <span className="text-zinc-700 font-medium text-xs uppercase tracking-wider">
                  ชื่อผู้ใช้ (Username)
                </span>
              }
              className="mb-3.5"
            >
              <Input
                prefix={<UserIcon className="w-4 h-4 text-zinc-400 mr-1" />}
                placeholder="somchai_dev"
                className="!rounded-xl !border-zinc-200 hover:!border-zinc-400 focus:!border-zinc-950 !h-11"
              />
            </Form.Item>

            <Form.Item
              name="phone_number"
              label={
                <span className="text-zinc-700 font-medium text-xs uppercase tracking-wider">
                  เบอร์โทรศัพท์ (ถ้ามี)
                </span>
              }
              className="mb-3.5"
            >
              <Input
                prefix={<PhoneIcon className="w-4 h-4 text-zinc-400 mr-1" />}
                placeholder="0891234567"
                className="!rounded-xl !border-zinc-200 hover:!border-zinc-400 focus:!border-zinc-950 !h-11"
              />
            </Form.Item>
          </div>

          <Form.Item
            name="email"
            label={
              <span className="text-zinc-700 font-medium text-xs uppercase tracking-wider">
                อีเมล (Email)
              </span>
            }
            rules={[
              { required: true, message: "กรุณากรอกอีเมล" },
              { type: "email", message: "รูปแบบอีเมลไม่ถูกต้อง" },
            ]}
            className="mb-3.5"
          >
            <Input
              prefix={<EnvelopeIcon className="w-4 h-4 text-zinc-400 mr-1" />}
              placeholder="somchai@company.com"
              className="!rounded-xl !border-zinc-200 hover:!border-zinc-400 focus:!border-zinc-950 !h-11"
            />
          </Form.Item>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-x-4">
            <Form.Item
              name="password"
              label={
                <span className="text-zinc-700 font-medium text-xs uppercase tracking-wider">
                  รหัสผ่าน
                </span>
              }
              rules={[
                { required: true, message: "กรุณากรอกรหัสผ่าน" },
                { min: 8, message: "รหัสผ่านต้องมีอย่างน้อย 8 ตัวอักษร" },
              ]}
              className="mb-3.5"
            >
              <Input.Password
                prefix={
                  <LockClosedIcon className="w-4 h-4 text-zinc-400 mr-1" />
                }
                placeholder="อย่างน้อย 8 ตัวอักษร"
                className="!rounded-xl !border-zinc-200 hover:!border-zinc-400 focus:!border-zinc-950 !h-11"
              />
            </Form.Item>

            <Form.Item
              name="confirm_password"
              label={
                <span className="text-zinc-700 font-medium text-xs uppercase tracking-wider">
                  ยืนยันรหัสผ่าน
                </span>
              }
              dependencies={["password"]}
              rules={[
                { required: true, message: "กรุณากรอกยืนยันรหัสผ่าน" },
                ({ getFieldValue }) => ({
                  validator(_, value) {
                    if (!value || getFieldValue("password") === value) {
                      return Promise.resolve();
                    }
                    return Promise.reject(new Error("รหัสผ่านไม่ตรงกัน"));
                  },
                }),
              ]}
              className="mb-3.5"
            >
              <Input.Password
                prefix={
                  <LockClosedIcon className="w-4 h-4 text-zinc-400 mr-1" />
                }
                placeholder="กรอกรหัสผ่านอีกครั้ง"
                className="!rounded-xl !border-zinc-200 hover:!border-zinc-400 focus:!border-zinc-950 !h-11"
              />
            </Form.Item>
          </div>

          <Form.Item className="mb-2 mt-5">
            <Button
              type="primary"
              htmlType="submit"
              loading={loading}
              block
              className="!bg-zinc-950 hover:!bg-zinc-800 active:!bg-black !text-white !border-zinc-950 !h-12 !text-sm !font-semibold !rounded-xl transition-all shadow-sm flex items-center justify-center gap-2"
            >
              <span>สมัครสมาชิก</span>
            </Button>
          </Form.Item>

          <div className="text-center pt-2">
            <a
              href={ROUTES.LOGIN}
              className="inline-flex items-center gap-1.5 text-xs text-zinc-500 hover:text-zinc-900 font-medium transition-colors"
            >
              <ArrowLeftIcon className="w-3.5 h-3.5 stroke-[2]" />
              กลับไปหน้าเข้าสู่ระบบ
            </a>
          </div>
        </Form>
      </motion.div>
    </Card>
  );
}
