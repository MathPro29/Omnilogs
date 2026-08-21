import { Button, Tooltip } from "antd";
import {
  ArrowLeftOutlined,
  ArrowRightOutlined,
  SaveOutlined,
} from "@ant-design/icons";
import type { ReactNode } from "react";

interface SetupNavigationProps {
  currentStepIndex: number; // 0..6
  canContinue: boolean;
  disabledReason?: string;
  isSubmitting?: boolean;
  onBack: () => void;
  onContinue: () => void;
  primaryAction: {
    label: string;
    icon?: ReactNode;
    showArrow?: boolean;
  };
  actionHint?: string;
}

export function SetupNavigation({
  currentStepIndex,
  canContinue,
  disabledReason,
  isSubmitting = false,
  onBack,
  onContinue,
  primaryAction,
  actionHint,
}: SetupNavigationProps) {
  const isFirstStep = currentStepIndex === 0;
  const isLastStep = currentStepIndex === 6;

  if (isLastStep) return null;

  return (
    <div className="setup-navigation">
      <div className="setup-navigation__context" aria-live="polite">
        <span className="setup-navigation__context-icon">
          <SaveOutlined />
        </span>
        <span>{actionHint || "ระบบจะบันทึกข้อมูลเมื่อกดไปต่อ"}</span>
      </div>

      <div className="setup-navigation__actions">
        {!isFirstStep && (
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={onBack}
            disabled={isSubmitting}
            className="setup-navigation__back"
          >
            ย้อนกลับ
          </Button>
        )}

        {canContinue ? (
          <Button
            type="primary"
            onClick={onContinue}
            loading={isSubmitting}
            icon={primaryAction.icon}
            className="setup-navigation__primary"
          >
            {primaryAction.label}
            {primaryAction.showArrow && <ArrowRightOutlined />}
          </Button>
        ) : (
          <Tooltip
            title={
              disabledReason ||
              "กรุณากรอกข้อมูลที่จำเป็นให้ครบก่อนดำเนินการต่อ"
            }
          >
            <span>
              <Button
                type="primary"
                disabled
                icon={primaryAction.icon}
                className="setup-navigation__primary setup-navigation__primary--disabled"
              >
                {primaryAction.label}
                {primaryAction.showArrow && <ArrowRightOutlined />}
              </Button>
            </span>
          </Tooltip>
        )}
      </div>
    </div>
  );
}




