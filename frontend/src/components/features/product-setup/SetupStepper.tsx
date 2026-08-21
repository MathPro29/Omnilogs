import type { SetupStepKey, StepStatus } from "@/features/product-setup/types/productSetup.types";
import { SETUP_STEPS } from "@/features/product-setup/utils/setupProgress";
import {
  CheckCircleFilled,
  ExclamationCircleFilled,
  LoadingOutlined,
} from "@ant-design/icons";

interface SetupStepperProps {
  currentStepIndex: number;
  stepStatuses: Record<SetupStepKey, StepStatus>;
  onSelectStep: (stepIndex: number) => void;
}

export function SetupStepper({
  currentStepIndex,
  stepStatuses,
  onSelectStep,
}: SetupStepperProps) {
  const renderBadge = (status: StepStatus) => {
    if (status === "completed") {
      return (
        <span className="inline-flex items-center gap-1 text-xs px-2 py-0.5 rounded bg-emerald-50 text-emerald-700 border border-emerald-200">
          <CheckCircleFilled className="text-emerald-600" /> เสร็จแล้ว
        </span>
      );
    }
    if (status === "in_progress") {
      return (
        <span className="inline-flex items-center gap-1 text-xs px-2 py-0.5 rounded bg-zinc-900 text-white font-medium">
          <LoadingOutlined className="animate-spin text-xs" /> กำลังดำเนินการ
        </span>
      );
    }
    if (status === "needs_attention") {
      return (
        <span className="inline-flex items-center gap-1 text-xs px-2 py-0.5 rounded bg-amber-50 text-amber-700 border border-amber-200">
          <ExclamationCircleFilled className="text-amber-500" /> ต้องตรวจสอบ
        </span>
      );
    }
    return (
      <span className="inline-flex items-center text-xs px-2 py-0.5 rounded bg-zinc-100 text-zinc-500 border border-zinc-200">
        ยังไม่เริ่ม
      </span>
    );
  };

  return (
    <div className="w-full bg-white border-b border-zinc-200 px-6 py-4">
      <div className="max-w-7xl mx-auto">
        <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-7 gap-2">
          {SETUP_STEPS.map((step, idx) => {
            const status = stepStatuses[step.key];
            const isActive = idx === currentStepIndex;
            const isClickable =
              status === "completed" ||
              status === "in_progress" ||
              status === "needs_attention" ||
              idx < currentStepIndex;

            return (
              <button
                key={step.key}
                type="button"
                disabled={!isClickable}
                onClick={() => isClickable && onSelectStep(idx)}
                className={`text-left p-3 rounded-lg border transition-all text-sm ${
                  isActive
                    ? "border-zinc-900 bg-zinc-50/80 shadow-sm"
                    : isClickable
                      ? "border-zinc-200 bg-white hover:border-zinc-300 hover:bg-zinc-50/50 cursor-pointer"
                      : "border-zinc-100 bg-zinc-50/30 opacity-60 cursor-not-allowed"
                }`}
              >
                <div className="flex items-center justify-between gap-1 mb-1.5">
                  <span className="font-mono text-xs font-semibold text-zinc-400">
                    0{step.stepNumber}
                  </span>
                  {renderBadge(status)}
                </div>
                <div className="font-semibold text-zinc-900 truncate">
                  {step.title}
                </div>
                <div className="text-xs text-zinc-500 truncate mt-0.5">
                  {step.subtitle}
                </div>
              </button>
            );
          })}
        </div>
      </div>
    </div>
  );
}

