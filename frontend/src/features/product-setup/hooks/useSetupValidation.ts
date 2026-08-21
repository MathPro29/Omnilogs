import { useMemo } from "react";
import type { FeatureNode, ProductSetupDraft, ProjectInfo } from "@/features/product-setup/types/productSetup.types";
import { productInformationSchema } from "@/features/product-setup/schemas/productSetup.schema";

export function useSetupValidation(
  stepIndex: number,
  draft: ProductSetupDraft,
  projects: ProjectInfo[],
  features: FeatureNode[],
  hasApiKey: boolean,
  hasTestLog: boolean
) {
  return useMemo(() => {
    // Step 0: ตรวจสอบข้อมูล Product
    const productValidation = productInformationSchema.safeParse({
      product_name: draft.product_name.trim(),
      description: draft.description,
      owner_id: draft.owner_id,
      is_active: draft.is_active,
    });
    const isProductValid = productValidation.success;

    // Step 1: ตรวจสอบ Projects
    const hasProjects = projects.length > 0;
    const isProjectsValid = hasProjects;

    // Step 2: ตรวจสอบ Features
    const hasFeatures = features.length > 0;
    const isFeaturesValid = hasFeatures;

    const hasSelectedEnvironment = draft.environments.some(
      (environment) => environment.environment_code.trim() !== "",
    );

    // Step 3: ตรวจสอบโครงสร้างทั้งหมด
    const errors: string[] = [];
    if (!productValidation.success) {
      for (const issue of productValidation.error.issues) {
        const field = "Product name";
        errors.push(`${field}: ${issue.message}. กรุณาไปขั้นตอนที่ 1 เพื่อแก้ไขข้อมูลนี้`);
      }
    }
    if (!hasProjects) errors.push("ยังไม่มี Project กรุณาไปขั้นตอนที่ 2 และเพิ่มอย่างน้อย 1 Project");
    if (!hasFeatures) errors.push("ยังไม่มี Feature กรุณาไปขั้นตอนที่ 3 และเพิ่มอย่างน้อย 1 Feature");
    if (!hasSelectedEnvironment) {
      errors.push("ต้องเลือก Environment อย่างน้อย 1 รายการในขั้นตอนที่ 1");
    }

    const isStructureValid = errors.length === 0;

    // Determine if user can proceed from current step
    let canContinue = false;
    let disabledReason = "";

    switch (stepIndex) {
      case 0:
        canContinue = isProductValid && hasSelectedEnvironment;
        if (!isProductValid) disabledReason = "กรุณากรอกชื่อ Product";
        else if (!hasSelectedEnvironment) disabledReason = "กรุณาเลือก Environment อย่างน้อย 1 รายการ";
        break;
      case 1:
        canContinue = isProjectsValid;
        if (!hasProjects) disabledReason = "ต้องมีอย่างน้อย 1 Project";
        break;
      case 2:
        canContinue = isFeaturesValid;
        if (!hasFeatures) disabledReason = "ต้องมีอย่างน้อย 1 Feature";
        break;
      case 3:
        canContinue = isStructureValid;
        if (!isStructureValid) disabledReason = "กรุณาแก้ไขประเด็นของโครงสร้าง";
        break;
      case 4:
        canContinue = hasApiKey;
        if (!hasApiKey) disabledReason = "กรุณาสร้าง Ingestion API Key";
        break;
      case 5:
        canContinue = hasTestLog;
        if (!hasTestLog) disabledReason = "กรุณาส่ง Test log ที่สำเร็จเพื่อยืนยันการเชื่อมต่อ";
        break;
      case 6:
        canContinue = true;
        break;
      default:
        canContinue = false;
    }

    return {
      isProductValid,
      isProjectsValid,
      isFeaturesValid,
      hasSelectedEnvironment,
      isStructureValid,
      structureErrors: errors,
      canContinue,
      disabledReason,
    };
  }, [stepIndex, draft, projects, features, hasApiKey, hasTestLog]);
}


