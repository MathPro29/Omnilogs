import type { SetupStepInfo, SetupStepKey, StepStatus } from "../types/productSetup.types";

export const SETUP_STEPS: SetupStepInfo[] = [
  { key: "product", stepNumber: 1, title: "Product", subtitle: "รายละเอียด Product", status: "not_started" },
  { key: "projects", stepNumber: 2, title: "Projects", subtitle: "กำหนด Projects", status: "not_started" },
  { key: "features", stepNumber: 3, title: "Features", subtitle: "Feature hierarchy", status: "not_started" },
  { key: "review", stepNumber: 4, title: "Review", subtitle: "ตรวจสอบโครงสร้าง", status: "not_started" },
  { key: "api_key", stepNumber: 5, title: "API Key", subtitle: "สร้าง API Key", status: "not_started" },
  { key: "connect_logs", stepNumber: 6, title: "Connect Logs", subtitle: "ทดสอบการรับ Log", status: "not_started" },
  { key: "complete", stepNumber: 7, title: "เสร็จสิ้นการตั้งค่า", subtitle: "ตั้งค่าเสร็จสิ้น", status: "not_started" },
];

export function computeStepStatuses(
  currentStepIndex: number,
  hasProduct: boolean,
  hasProjects: boolean,
  hasFeatures: boolean,
  isReviewValid: boolean,
  hasApiKey: boolean,
  hasTestLog: boolean,
  hasFailedTestLog: boolean,
  isCompleted: boolean,
): Record<SetupStepKey, StepStatus> {
  return {
    product: isCompleted || hasProduct ? "completed" : currentStepIndex === 0 ? "in_progress" : "not_started",
    projects: isCompleted || (hasProduct && hasProjects) ? "completed" : currentStepIndex === 1 ? "in_progress" : hasProduct ? "needs_attention" : "not_started",
    features: isCompleted || (hasProjects && hasFeatures) ? "completed" : currentStepIndex === 2 ? "in_progress" : hasProjects ? "needs_attention" : "not_started",
    review: isCompleted || isReviewValid ? "completed" : currentStepIndex === 3 ? "in_progress" : hasFeatures ? "needs_attention" : "not_started",
    api_key: isCompleted || hasApiKey ? "completed" : currentStepIndex === 4 ? "in_progress" : isReviewValid ? "needs_attention" : "not_started",
    // Having an API key only means this step is ready to start; it is not an
    // error that needs attention. Show the warning state only after a real
    // test-log attempt has failed.
    connect_logs: isCompleted || hasTestLog ? "completed" : hasFailedTestLog ? "needs_attention" : currentStepIndex === 5 ? "in_progress" : "not_started",
    complete: isCompleted ? "completed" : currentStepIndex === 6 ? "in_progress" : "not_started",
  };
}

