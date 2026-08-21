import { useState } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { Card, Form, Input, Button, Typography, message, Divider } from "antd";
import {
  UserIcon,
  LockClosedIcon,
  ArrowRightIcon,
} from "@heroicons/react/24/outline";
import { motion } from "motion/react";
import { loginSchema, type LoginFormData } from "@/schemas";
import { authService } from "@/services";
import { useAuthStore } from "@/store";
import { ROUTES } from "@/constants";

const { Title, Text } = Typography;

export function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const setAuth = useAuthStore((state) => state.setAuth);
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();

  const from =
    (location.state as { from?: { pathname: string } })?.from?.pathname ||
    ROUTES.DASHBOARD;

  const handleSubmit = async (values: LoginFormData) => {
    // Validateด้วย Zod
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
      message.success("เข้าสู่ระบบสำเร็จ");
      navigate(from, { replace: true });
    } catch (error: unknown) {
      const err = error as { message?: string };
      message.error(err?.message || "เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card className="login-card" bordered={false}>
      {/* Header */}
      <div className="text-center pt-2 pb-1">
        <motion.div
          initial={{ scale: 0.85, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          transition={{ delay: 0.05, duration: 0.35 }}
        >
          <div className="inline-flex items-center justify-center w-14 h-14 rounded-2xl mb-3 bg-[#212121] text-[#1F8457] shadow-md border border-[#3a3a3a]">
            <LockClosedIcon className="w-7 h-7 stroke-[1.75]" />
          </div>
        </motion.div>
        <Title
          level={3}
          className="!text-zinc-900 !font-bold !tracking-tight !mb-0.5"
        >
          เข้าสู่ระบบ
        </Title>
        <Text className="text-xs uppercase tracking-widest text-zinc-500 font-semibold">
          Omnilogs Admin Panel
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
          <Form.Item
            name="username"
            label={
              <span className="text-zinc-700 font-medium text-xs uppercase tracking-wider">
                ชื่อผู้ใช้ / อีเมล
              </span>
            }
            rules={[{ required: true, message: "กรุณากรอกชื่อผู้ใช้" }]}
            className="mb-4"
          >
            <Input
              prefix={<UserIcon className="w-4 h-4 text-zinc-400 mr-1" />}
              placeholder="กรอกชื่อผู้ใช้"
              className="!rounded-xl !border-zinc-200 hover:!border-zinc-400 focus:!border-zinc-950 !h-11"
            />
          </Form.Item>

          <Form.Item
            name="password"
            label={
              <span className="text-zinc-700 font-medium text-xs uppercase tracking-wider">
                รหัสผ่าน
              </span>
            }
            rules={[{ required: true, message: "กรุณากรอกรหัสผ่าน" }]}
            className="mb-4"
          >
            <Input.Password
              prefix={<LockClosedIcon className="w-4 h-4 text-zinc-400 mr-1" />}
              placeholder="กรอกรหัสผ่าน"
              className="!rounded-xl !border-zinc-200 hover:!border-zinc-400 focus:!border-zinc-950 !h-11"
            />
          </Form.Item>

          <Form.Item className="mb-3 mt-5">
            <Button
              type="primary"
              htmlType="submit"
              loading={loading}
              block
              className="!bg-zinc-950 hover:!bg-zinc-800 active:!bg-black !text-white !border-zinc-950 !h-12 !text-sm !font-semibold !rounded-xl transition-all shadow-sm flex items-center justify-center gap-2"
            >
              <span>เข้าสู่ระบบ</span>
              <ArrowRightIcon className="w-4 h-4 stroke-[2]" />
            </Button>
          </Form.Item>

          <div className="flex items-center justify-between text-xs pt-1">
            <a
              href={ROUTES.FORGOT_PASSWORD}
              className="text-zinc-500 hover:text-zinc-900 transition-colors font-medium"
            >
              ลืมรหัสผ่าน?
            </a>
            <a
              href={ROUTES.REGISTER}
              className="text-zinc-950 hover:text-black font-semibold transition-colors hover:underline"
            >
              สมัครสมาชิก
            </a>
          </div>
        </Form>

        {/* Demo credentials */}
        <div className="mt-5 p-3.5 rounded-xl bg-zinc-50 border border-zinc-200/80 text-left">
          <div className="flex items-center justify-between mb-2">
            <span className="text-[10px] font-bold uppercase tracking-widest text-zinc-400">
              Demo Credentials
            </span>
            <span className="text-[10px] bg-zinc-200 text-zinc-700 font-semibold px-2 py-0.5 rounded-md">
              Testing
            </span>
          </div>
          <div className="space-y-1 text-xs font-mono">
            <div className="flex justify-between items-center text-zinc-600">
              <span className="text-zinc-400 font-sans text-[11px]">
                Email:
              </span>
              <span className="font-semibold text-zinc-900 select-all">
                godmode@godmail.com
              </span>
            </div>
            <div className="flex justify-between items-center text-zinc-600">
              <span className="text-zinc-400 font-sans text-[11px]">
                Password:
              </span>
              <span className="font-semibold text-zinc-900 select-all">
                OhMyGod999@!
              </span>
            </div>
          </div>
        </div>
      </motion.div>
    </Card>
  );
}
