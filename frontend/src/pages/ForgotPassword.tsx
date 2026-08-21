import { useState } from "react";
import { Card, Form, Input, Button, Typography, message, Steps } from "antd";
import {
  LockClosedIcon,
  EnvelopeIcon,
  ArrowRightIcon,
  ArrowLeftIcon,
} from "@heroicons/react/24/outline";
import { apiClient } from "@/api";
import { ROUTES } from "@/constants";
import { getErrorMessage } from "@/utils/error";
const { Title, Text } = Typography;
type ResetValues = { newPassword: string; confirmPassword: string };
export default function ForgotPassword() {
  const [step, setStep] = useState(0);
  const [loading, setLoading] = useState(false);
  const [email, setEmail] = useState("");
  const [form] = Form.useForm();
  const requestReset = async (v: { email: string }) => {
    setLoading(true);
    try {
      await apiClient.post("/auth/forgot-password", { email: v.email });
      setEmail(v.email);
      setStep(1);
      message.success("Request accepted. Set a new password.");
    } catch (error: unknown) {
      message.error(getErrorMessage(error, "Unable to process request."));
    } finally {
      setLoading(false);
    }
  };
  const resetPassword = async (v: ResetValues) => {
    setLoading(true);
    try {
      await apiClient.post("/auth/reset-password", {
        email,
        new_password: v.newPassword,
        confirm_password: v.confirmPassword,
      });
      message.success("Password changed successfully.");
      window.setTimeout(() => {
        window.location.href = ROUTES.LOGIN;
      }, 1500);
    } catch (error: unknown) {
      message.error(getErrorMessage(error, "Unable to change password."));
    } finally {
      setLoading(false);
    }
  };
  return (
    <Card className="login-card" bordered={false}>
      <div className="text-center pt-4 pb-2">
        <div className="inline-flex items-center justify-center w-14 h-14 rounded-2xl mb-3 bg-[#212121] text-[#1F8457] shadow-md border border-[#3a3a3a]">
          <LockClosedIcon className="w-7 h-7 stroke-[1.75]" />
        </div>
        <Title level={3}>
          {step === 0 ? "Forgot Password" : "Set New Password"}
        </Title>
        <Text type="secondary">
          {step === 0 ? "Enter your email to continue." : `Account: ${email}`}
        </Text>
      </div>
      <div className="my-6">
        <Steps
          size="small"
          current={step}
          items={[{ title: "Email" }, { title: "New Password" }]}
        />
      </div>
      {step === 0 ? (
        <Form
          form={form}
          layout="vertical"
          onFinish={requestReset}
          size="large"
        >
          <Form.Item
            name="email"
            label="Email Address"
            rules={[
              { required: true, message: "Please enter your email." },
              { type: "email", message: "Please enter a valid email." },
            ]}
          >
            <Input
              prefix={<EnvelopeIcon className="w-5 h-5 text-gray-400" />}
              placeholder="email@company.com"
            />
          </Form.Item>
          <Button
            type="primary"
            htmlType="submit"
            loading={loading}
            block
            className="btn-primary"
          >
            Continue <ArrowRightIcon className="inline w-4 h-4 ml-2" />
          </Button>
          <div className="text-center mt-4">
            <a href={ROUTES.LOGIN}>
              <ArrowLeftIcon className="inline w-4 h-4 mr-2" />
              Back to Login
            </a>
          </div>
        </Form>
      ) : (
        <Form layout="vertical" onFinish={resetPassword} size="large">
          <Form.Item
            name="newPassword"
            label="New Password"
            rules={[
              { required: true, message: "Please enter a new password." },
              { min: 8, message: "Password must be at least 8 characters." },
            ]}
          >
            <Input.Password
              prefix={<LockClosedIcon className="w-5 h-5 text-gray-400" />}
            />
          </Form.Item>
          <Form.Item
            name="confirmPassword"
            label="Confirm Password"
            dependencies={["newPassword"]}
            rules={[
              { required: true, message: "Please confirm your password." },
              ({ getFieldValue }) => ({
                validator(_, v) {
                  return !v || getFieldValue("newPassword") === v
                    ? Promise.resolve()
                    : Promise.reject(new Error("Passwords do not match."));
                },
              }),
            ]}
          >
            <Input.Password
              prefix={<LockClosedIcon className="w-5 h-5 text-gray-400" />}
            />
          </Form.Item>
          <Button
            type="primary"
            htmlType="submit"
            loading={loading}
            block
            className="btn-primary"
          >
            Set Password <ArrowRightIcon className="inline w-4 h-4 ml-2" />
          </Button>
          <div className="text-center mt-4">
            <Button type="link" onClick={() => setStep(0)}>
              <ArrowLeftIcon className="inline w-4 h-4 mr-2" />
              Change Email
            </Button>
          </div>
        </Form>
      )}
    </Card>
  );
}
