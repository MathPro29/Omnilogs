import { SaveOutlined } from "@ant-design/icons";
import { PageTransition } from "@/components";
import { ProductSetupStepContent } from "@/components/features/product-setup/ProductSetupStepContent";
import { SetupNavigation } from "@/components/features/product-setup/SetupNavigation";
import { SetupStepper } from "@/components/features/product-setup/SetupStepper";
import { useProductSetupWizard } from "@/features/product-setup/hooks/useProductSetupWizard";
import "@/styles/pages/product-setup.css";

export function ProductSetupWizard() {
  const wizard = useProductSetupWizard();
  const { store } = wizard;
  const isEditingHierarchy =
    store.editingExistingProduct && store.currentStepIndex === 2;
  const canSubmit =
    wizard.validation.canContinue &&
    (!isEditingHierarchy || wizard.hasUnsavedChangesForCurrentStep);

  const primaryAction = (() => {
    switch (store.currentStepIndex) {
      case 0:
        return {
          label: "บันทึก Product และไปต่อ",
          icon: <SaveOutlined />,
          showArrow: true,
          hint: "บันทึกรายละเอียด Product แล้วไปกำหนด Projects",
        };
      case 1:
        return {
          label: "บันทึก Projects และไปต่อ",
          icon: <SaveOutlined />,
          showArrow: true,
          hint: "บันทึกโครงสร้าง Projects แล้วสร้าง Feature hierarchy",
        };
      case 2:
        return isEditingHierarchy
          ? {
              label: "บันทึกการเปลี่ยนแปลง Hierarchy",
              icon: <SaveOutlined />,
              showArrow: false,
              hint: "อัปเดต Product hierarchy และอยู่ที่หน้านี้",
            }
          : {
              label: "บันทึก Hierarchy และไปต่อ",
              icon: <SaveOutlined />,
              showArrow: true,
              hint: "บันทึก Hierarchy แล้วตรวจสอบโครงสร้างทั้งหมด",
            };
      case 3:
        return {
          label: "ไปต่อเพื่อสร้าง API Key",
          showArrow: true,
          hint: "โครงสร้างพร้อมแล้ว ขั้นตอนถัดไปคือสร้าง Ingestion API Key",
        };
      case 4:
        return {
          label: "ไปต่อเพื่อทดสอบการเชื่อมต่อ",
          showArrow: true,
          hint: "ตรวจสอบ Key ด้วยการส่ง Test log รายการแรก",
        };
      case 5:
        return {
          label: "เสร็จสิ้นการตั้งค่า",
          showArrow: true,
          hint: "ตั้งค่าให้เสร็จหลังจากทดสอบการเชื่อมต่อสำเร็จ",
        };
      default:
        return {
          label: "ไปต่อ",
          showArrow: true,
          hint: undefined,
        };
    }
  })();

  const disabledReason =
    wizard.validation.disabledReason ||
    (isEditingHierarchy && !wizard.hasUnsavedChangesForCurrentStep
      ? "There are no hierarchy changes to save."
      : wizard.saveDisabledReason);

  const navNode = (
    <SetupNavigation
      currentStepIndex={store.currentStepIndex}
      canContinue={canSubmit}
      disabledReason={disabledReason}
      isSubmitting={wizard.isSubmitting}
      onBack={store.prevStep}
      onContinue={wizard.continueSetup}
      primaryAction={primaryAction}
      actionHint={primaryAction.hint}
    />
  );

  return (
    <PageTransition>
      <div className="product-setup-wizard">
        <header className="product-setup-header">
          <div className="product-setup-header__copy">
            <span className="product-setup-header__eyebrow">
              {store.editingExistingProduct ? "การจัดการ Product" : "การตั้งค่า Product"}
            </span>
            <h1>
              {store.editingExistingProduct
                ? "อัปเดต Product hierarchy"
                : "ตั้งค่า Product ของคุณ"}
            </h1>
            <p>
              {store.editingExistingProduct
                ? "ตรวจสอบและอัปเดตโครงสร้าง Project และ Feature ของ Product นี้"
                : "สร้างโครงสร้าง สร้าง Ingestion API Key และตรวจสอบ Log รายการแรก"}
            </p>
          </div>
          <div className="product-setup-header__progress">
            <span className="product-setup-header__progress-label">
              ขั้นตอนปัจจุบัน
            </span>
            <strong>{store.currentStepIndex + 1} / 7</strong>
          </div>
        </header>

        <SetupStepper
          currentStepIndex={store.currentStepIndex}
          stepStatuses={wizard.stepStatuses}
          onSelectStep={store.setStepIndex}
        />

        <main className="product-setup-content">
          <ProductSetupStepContent wizard={wizard} navigation={navNode} />
        </main>

        {store.currentStepIndex !== 0 && navNode}
      </div>
    </PageTransition>
  );
}







