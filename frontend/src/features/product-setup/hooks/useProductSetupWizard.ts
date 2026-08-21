import { useRef, useState } from "react";
import axios from "axios";
import { message } from "antd";
import { useQueryClient } from "@tanstack/react-query";
import { useProductSetupStore } from "@/features/product-setup/store/productSetup.store";
import { productSetupService } from "@/features/product-setup/services/productSetup.service";
import { productService } from "@/services/product.service";
import { apiKeyService } from "@/services/apiKey.service";
import { computeStepStatuses } from "@/features/product-setup/utils/setupProgress";
import { getSetupErrorMessage } from "@/features/product-setup/utils/setupError";
import {
  featureSnapshot,
  projectSnapshot,
} from "@/features/product-setup/utils/setupDirtyState";
import type { ProjectInfo } from "@/features/product-setup/types/productSetup.types";
import { useSetupValidation } from "./useSetupValidation";

interface ProductConflictResponse {
  message?: string;
  details?: {
    field?: string;
    value?: string;
    scope?: string;
  };
}

const normalizeProductName = (value: string) =>
  value.normalize("NFKC").trim().toLocaleLowerCase();
export interface GenerateSetupApiKeyValues {
  key_name: string;
  environment_code: string;
  default_project_id?: number;
  default_category_id?: number;
  expiration: string;
  description?: string;
}

export function useProductSetupWizard() {
  const queryClient = useQueryClient();
  const store = useProductSetupStore();

  const invalidateFeatureQueries = () => {
    void queryClient.invalidateQueries({ queryKey: ["products"] });
    void queryClient.invalidateQueries({ queryKey: ["features"] });
    void queryClient.invalidateQueries({ queryKey: ["log-routing-features"] });
    void queryClient.invalidateQueries({ queryKey: ["log-explorer-parent-features"] });
    void queryClient.invalidateQueries({ queryKey: ["log-explorer-routing-features"] });
    void queryClient.invalidateQueries({ queryKey: ["features-options"] });
  };
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [productNameError, setProductNameError] = useState<string>();
  const submissionLock = useRef(false);
  const [savedProjectsSnapshot, setSavedProjectsSnapshot] = useState(() =>
    projectSnapshot(store.projects),
  );
  const [savedFeaturesSnapshot, setSavedFeaturesSnapshot] = useState(() =>
    featureSnapshot(store.features),
  );
  const [isProductCreated, setIsProductCreated] = useState(
    Boolean(store.productId),
  );
  const hasUnsavedProjectChanges =
    savedProjectsSnapshot !== projectSnapshot(store.projects);
  const hasUnsavedFeatureChanges =
    savedFeaturesSnapshot !== featureSnapshot(store.features);
  const validation = useSetupValidation(
    store.currentStepIndex,
    store.productDraft,
    store.projects,
    store.features,
    Boolean(store.generatedApiKey) || store.existingApiKeys.some((key) => key.is_active),
    store.testLogResult?.status === "received",
  );
  const stepStatuses = computeStepStatuses(
    store.currentStepIndex,
    isProductCreated,
    store.projects.length > 0,
    store.features.length > 0,
    validation.isStructureValid,
    Boolean(store.generatedApiKey) || store.existingApiKeys.some((key) => key.is_active),
    store.testLogResult?.status === "received",
    store.testLogResult?.status === "failed",
    store.productDraft.setup_status === "ACTIVE",
  );

  const updateProductDraft = (partial: Partial<typeof store.productDraft>) => {
    if (Object.hasOwn(partial, "product_name")) {
      setProductNameError(undefined);
    }
    store.updateProductDraft(partial);
  };

  const findDuplicateProductName = async () => {
    const candidate = normalizeProductName(store.productDraft.product_name);
    if (!candidate) return undefined;
    try {
      const products = await productService.listOptions();
      return products.find(
        (product) =>
          product.id !== store.productId &&
          normalizeProductName(product.name) === candidate,
      );
    } catch {
      // The create/update endpoint remains the authoritative conflict check.
      return undefined;
    }
  };

  const validateProductNameAvailability = async (): Promise<boolean> => {
    setProductNameError(undefined);
    const duplicate = await findDuplicateProductName();
    if (!duplicate) return true;
    setProductNameError(
      `Product name already exists (Product ID #${duplicate.id}). Open the existing Product or use a different name.`,
    );
    return false;
  };
  const saveProduct = async (advance = true): Promise<boolean> => {
    const existingProductId = store.productId;
    setProductNameError(undefined);
    setIsSubmitting(true);
    try {
      if (!(await validateProductNameAvailability())) {
        return false;
      }

      if (!store.productId) {
        const created = await productSetupService.createProductSetup({
          product_name: store.productDraft.product_name.trim(),
          description: store.productDraft.description,
          owner_id: store.productDraft.owner_id,
          environments: store.productDraft.environments.map((environment) => ({
            environment_code: environment.environment_code,
            environment_name: environment.environment_name,
            environment_type: environment.environment_type,
          })),
        });
        const createdId = created.product_id ?? created.id;
        if (!createdId) throw new Error("ไม่พบ product_id จากข้อมูลตอบกลับของการสร้าง Product");
        store.setProductId(createdId);
        setIsProductCreated(true);
        message.success("สร้าง Product สำเร็จ");
      } else {
        await productSetupService.updateProduct(store.productId, {
          product_name: store.productDraft.product_name.trim(),
          description: store.productDraft.description,
        });
      }

      if (existingProductId) {
        const existingEnvironments = await productSetupService.listEnvironments(existingProductId);
        const selectedCodes = new Set(
          store.productDraft.environments.map((environment) =>
            environment.environment_code.toUpperCase(),
          ),
        );
        for (const environment of store.productDraft.environments) {
          if (!environment.environment_id) {
            await productSetupService.createEnvironment(existingProductId, {
              environment_code: environment.environment_code,
              environment_name: environment.environment_name,
              environment_type: environment.environment_type,
            });
          }
        }
        for (const environment of existingEnvironments) {
          if (
            environment.environment_id &&
            !selectedCodes.has(environment.environment_code.toUpperCase())
          ) {
            await productSetupService.deleteEnvironment(
              existingProductId,
              environment.environment_id,
            );
          }
        }
      }
      if (advance) store.nextStep();
      return true;
    } catch (error) {
      const isProductConflict =
        axios.isAxiosError<ProductConflictResponse>(error) &&
        error.response?.status === 409;
      if (isProductConflict) {
        setProductNameError(
          error.response?.data?.message ||
            "มี Product ชื่อนี้อยู่แล้ว กรุณาเปิด Product เดิมหรือใช้ชื่ออื่น",
        );
      } else {
        message.error(
          getSetupErrorMessage(error, "บันทึกข้อมูล Product ไม่สำเร็จ"),
        );
      }
      return false;
    } finally {
      setIsSubmitting(false);
    }
  };

  const saveProjects = async (advance = true) => {
    if (!store.productId || isSubmitting) return;
    if (!hasUnsavedProjectChanges) {
      if (advance) store.nextStep();
      return;
    }
    setIsSubmitting(true);
    try {
      const existing = await productSetupService.getProjects(store.productId);
      const existingIds = new Set(existing.map((project) => project.project_id));
      const retainedIds = new Set<number>();
      const localIdToProjectServerId = new Map<string, number>();

      for (const project of store.projects) {
        const localId = String(project.project_id ?? project.id ?? "");
        let targetProjectId =
          project.project_id && existingIds.has(project.project_id)
            ? project.project_id
            : undefined;

        if (!targetProjectId) {
          const normName = normalizeProductName(project.project_name);
          const matchedByServerName = existing.find(
            (item) =>
              normalizeProductName(item.project_name) === normName &&
              item.project_id !== undefined &&
              !retainedIds.has(item.project_id),
          );
          if (matchedByServerName?.project_id) {
            targetProjectId = matchedByServerName.project_id;
          }
        }

        if (targetProjectId) {
          const updated = await productSetupService.updateProject(
            store.productId,
            targetProjectId,
            {
              project_name: project.project_name,
              is_active: project.is_active,
            },
          );
          const serverId = updated.project_id!;
          retainedIds.add(serverId);
          if (localId) localIdToProjectServerId.set(localId, serverId);
        } else {
          const created = await productSetupService.createProject(
            store.productId,
            {
              project_name: project.project_name,
              description: project.description,
            },
          );
          const serverId = created.project_id!;
          retainedIds.add(serverId);
          if (localId) localIdToProjectServerId.set(localId, serverId);
        }
      }

      for (const project of existing) {
        if (project.project_id && !retainedIds.has(project.project_id)) {
          await productSetupService.deleteProject(
            store.productId,
            project.project_id,
          );
        }
      }

      const updatedProjects = await productSetupService.getProjects(
        store.productId,
      );
      const activeProjectIds = new Set(
        updatedProjects.map((project) => String(project.project_id)),
      );

      const updatedFeatures = store.features
        .map((feature) => {
          const locallrojId = String(feature.project_id);
          const mappedServerlrojId = localIdToProjectServerId.get(locallrojId);
          const finallrojId = mappedServerlrojId ?? feature.project_id;
          return { ...feature, project_id: finallrojId };
        })
        .filter((feature) => activeProjectIds.has(String(feature.project_id)));

      store.setProjects(
        updatedProjects.map<ProjectInfo>((project) => ({
          project_id: project.project_id,
          project_name: project.project_name,
          project_code: project.project_code,
          description: project.description,
          is_active: project.is_active ?? true,
          display_order: project.display_order ?? 0,
        })),
      );
      store.setFeatures(updatedFeatures);
      setSavedProjectsSnapshot(projectSnapshot(updatedProjects));
      message.success("บันทึกโครงสร้าง Project เรียบร้อยแล้ว");
      if (advance) store.nextStep();
    } catch (error) {
      message.error(
        getSetupErrorMessage(
          error,
          "บันทึกโครงสร้าง Project ไม่สำเร็จ ข้อมูลบนหน้านี้ยังอยู่และสามารถลองใหม่ได้",
        ),
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  const saveFeatures = async (advance = true) => {
    if (!store.productId || isSubmitting) return;
    if (!hasUnsavedFeatureChanges) {
      if (advance) store.nextStep();
      return;
    }
    setIsSubmitting(true);
    try {
      for (const project of store.projects) {
        if (!project.project_id) continue;

        const projectFeatures = store.features.filter(
          (feature) => String(feature.project_id) === String(project.project_id),
        );
        const existing = await productSetupService.getFeatures(
          store.productId,
          project.project_id,
        );
        const existingIds = new Set(
          existing.map((feature) => feature.category_id),
        );
        const serverIdByLocalId = new Map<string, number>();

        for (const feature of projectFeatures) {
          if (feature.category_id && existingIds.has(feature.category_id)) {
            const localId = String(feature.category_id ?? feature.id);
            serverIdByLocalId.set(localId, feature.category_id);
            if (feature.id)
              serverIdByLocalId.set(String(feature.id), feature.category_id);
          }
        }

        const persistedIds = new Set<number>();
        const pending = [...projectFeatures];
        while (pending.length > 0) {
          let progressed = false;
          for (let index = pending.length - 1; index >= 0; index -= 1) {
            const feature = pending[index];
            const localId = String(feature.category_id ?? feature.id);
            const parentLocalId =
              feature.parent_id === null || feature.parent_id === undefined
                ? null
                : String(feature.parent_id);
            const parentServerId = parentLocalId
              ? serverIdByLocalId.get(parentLocalId)
              : undefined;

            if (parentLocalId && !parentServerId) continue;

            let targetCategoryId =
              feature.category_id && existingIds.has(feature.category_id)
                ? feature.category_id
                : undefined;

            if (!targetCategoryId) {
              const matchedServerFeature = existing.find((item) => {
                if (persistedIds.has(item.category_id)) return false;
                const sameCode =
                  Boolean(feature.category_code) &&
                  item.category_code.toUpperCase() ===
                    feature.category_code.toUpperCase();
                const sameName =
                  item.category_name.trim().toLowerCase() ===
                  feature.category_name.trim().toLowerCase();
                const samelarent =
                  (item.parent_id ?? null) === (parentServerId ?? null);
                return sameCode || (sameName && samelarent);
              });
              if (matchedServerFeature) {
                targetCategoryId = matchedServerFeature.category_id;
              }
            }

            if (targetCategoryId) {
              await productSetupService.updateFeature(
                store.productId,
                project.project_id,
                targetCategoryId,
                {
                  category_name: feature.category_name,
                  description: feature.description,
                  parent_id: parentServerId,
                  is_active: feature.is_active,
                },
              );
              serverIdByLocalId.set(localId, targetCategoryId);
              if (feature.id)
                serverIdByLocalId.set(String(feature.id), targetCategoryId);
              persistedIds.add(targetCategoryId);
            } else {
              const created = await productSetupService.createFeature(
                store.productId,
                project.project_id,
                {
                  category_name: feature.category_name,
                  description: feature.description,
                  parent_id: parentServerId ?? null,
                },
              );
              serverIdByLocalId.set(localId, created.category_id);
              if (feature.id)
                serverIdByLocalId.set(String(feature.id), created.category_id);
              persistedIds.add(created.category_id);
            }

            pending.splice(index, 1);
            progressed = true;
          }

          if (!progressed) {
            const unresolved = pending
              .map((feature) => feature.category_name)
              .join(", ");
            throw new Error(
              `ไม่สามารถบันทึก Feature ได้ เนื่องจากไม่พบ Parent: ${unresolved}`,
            );
          }
        }

        const remainingDeleteIds = new Set(
          existing
            .map((feature) => feature.category_id)
            .filter((id) => !persistedIds.has(id)),
        );
        while (remainingDeleteIds.size > 0) {
          const leafId = Array.from(remainingDeleteIds).find(
            (candidate) =>
              !existing.some(
                (feature) =>
                  feature.parent_id === candidate &&
                  remainingDeleteIds.has(feature.category_id),
              ),
          );
          if (!leafId) {
            throw new Error(
              "ไม่สามารถลบ Feature ได้ เนื่องจากโครงสร้าง Parent ไม่ถูกต้อง",
            );
          }
          await productSetupService.deleteFeature(
            store.productId,
            project.project_id,
            leafId,
          );
          remainingDeleteIds.delete(leafId);
        }
      }

      const refreshedFeatures = (
        await Promise.all(
          store.projects
            .filter(
              (project): project is ProjectInfo & { project_id: number } =>
                Boolean(project.project_id),
            )
            .map(async (project) => {
              const list = await productSetupService.getFeatures(
                store.productId!,
                project.project_id,
              );
              return list.map((feature) => ({
                ...feature,
                project_id: feature.project_id ?? project.project_id,
                id: String(feature.category_id),
                parent_id: feature.parent_id ?? null,
              }));
            }),
        )
      ).flat();

      store.setFeatures(refreshedFeatures);
      setSavedFeaturesSnapshot(featureSnapshot(refreshedFeatures));
      invalidateFeatureQueries();
      message.success("บันทึก Product hierarchy เรียบร้อยแล้ว");
      if (advance) store.nextStep();
    } catch (error) {
      message.error(
        getSetupErrorMessage(
          error,
          "บันทึก Product hierarchy ไม่สำเร็จ ข้อมูลบนหน้านี้ยังอยู่และสามารถลองใหม่ได้",
        ),
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  const createFeature = async (feature: typeof store.features[number]) => {
    if (!store.productId) {
      message.error("กรุณาบันทึก Product ก่อนสร้าง Feature");
      return false;
    }

    const projectId = Number(feature.project_id);
    if (!Number.isInteger(projectId) || projectId <= 0) {
      message.error("กรุณาบันทึก Project ก่อนสร้าง Feature");
      return false;
    }

    setIsSubmitting(true);
    try {
      const created = await productSetupService.createFeature(
        store.productId,
        projectId,
        {
          category_name: feature.category_name,
          category_code: feature.category_code,
          description: feature.description,
          parent_id:
            feature.parent_id === null ? null : Number(feature.parent_id),
        },
      );
      const persistedFeature = {
        ...feature,
        ...created,
        id: String(created.category_id),
        project_id: created.project_id ?? projectId,
        parent_id: created.parent_id ?? null,
        category_code: created.category_code || feature.category_code,
        is_active: created.is_active ?? feature.is_active,
        display_order: created.display_order ?? feature.display_order,
      };
      const updatedFeatures = [...store.features, persistedFeature];
      store.setFeatures(updatedFeatures);
      setSavedFeaturesSnapshot(featureSnapshot(updatedFeatures));
      invalidateFeatureQueries();
      return true;
    } catch (error) {
      message.error(getSetupErrorMessage(error, "สร้าง Feature ไม่สำเร็จ"));
      return false;
    } finally {
      setIsSubmitting(false);
    }
  };
  const validateStructure = async () => {
    if (!store.productId) return;
    setIsSubmitting(true);
    try {
      const result = await productSetupService.validateStructure(
        store.productId,
      );
      if (result.valid) {
        store.nextStep();
      } else {
        message.error(
          result.validation_errors?.join("; ") ||
            "กรุณาแก้ไขประเด็นของโครงสร้างก่อนดำเนินการต่อ",
        );
      }
    } catch (error) {
      message.error(getSetupErrorMessage(error, "ตรวจสอบโครงสร้างไม่สำเร็จ"));
    } finally {
      setIsSubmitting(false);
    }
  };

  const generateApiKey = async (values: GenerateSetupApiKeyValues) => {
    if (!store.productId) return;
    setIsSubmitting(true);
    try {
      store.setGeneratedApiKey(
        await productSetupService.createApiKey(store.productId, values),
      );
    } catch (error) {
      message.error(getSetupErrorMessage(error, "สร้าง API Key ไม่สำเร็จ"));
    } finally {
      setIsSubmitting(false);
    }
  };

  const continueSetup = async () => {
    if (submissionLock.current || isSubmitting) return;
    const handler = [
      saveProduct,
      saveProjects,
      saveFeatures,
      validateStructure,
      async () => store.nextStep(),
      async () => store.nextStep(),
    ][store.currentStepIndex];
    if (!handler) return;

    submissionLock.current = true;
    try {
      if (store.editingExistingProduct && store.currentStepIndex === 2) {
        await saveFeatures(false);
      } else {
        await handler();
      }
    } finally {
      submissionLock.current = false;
    }
  };  const deleteExistingApiKey = async (keyId: number) => {
    if (!store.productId) return;
    try {
      await apiKeyService.revoke(store.productId, keyId);
      store.setExistingApiKeys(
        store.existingApiKeys.filter((key) => key.key_id !== keyId),
      );
      message.success("ลบ API Key แล้ว");
    } catch (error) {
      message.error(getSetupErrorMessage(error, "ลบ API Key ไม่สำเร็จ"));
    }
  };
  const saveDraft = async () => {
    if (submissionLock.current || isSubmitting) return;
    submissionLock.current = true;
    try {
      const productSaved = await saveProduct(false);
      if (!productSaved) return;
      if (store.currentStepIndex >= 1) await saveProjects(false);
      if (store.currentStepIndex >= 2) await saveFeatures(false);
      if (store.productId) {
        await productSetupService.updateProduct(store.productId, { setup_status: "DRAFT" });
        store.updateProductDraft({ setup_status: "DRAFT" });
      }
      message.success("บันทึก Draft แล้ว สามารถกลับมาแก้ไขต่อได้");
    } catch (error) {
      message.error(getSetupErrorMessage(error, "บันทึก Draft ไม่สำเร็จ"));
    } finally {
      submissionLock.current = false;
    }
  };
  const completeSetup = async () => {
    if (!store.productId) return;
    await productSetupService.completeSetup(store.productId);
    store.updateProductDraft({ setup_status: "ACTIVE" });
  };

  const currentStepRequiresSave =
    store.currentStepIndex === 1 || store.currentStepIndex === 2;
  const hasUnsavedChangesForCurrentStep =
    store.currentStepIndex === 1
      ? hasUnsavedProjectChanges
      : store.currentStepIndex === 2
        ? hasUnsavedFeatureChanges
        : true;
  const saveDisabledReason =
    currentStepRequiresSave && !hasUnsavedChangesForCurrentStep
      ? "No changes to save"
      : undefined;

  return {
    store,
    validation,
    stepStatuses,
    isSubmitting,
    hasUnsavedChangesForCurrentStep,
    saveDisabledReason,
    isProductCreated,
    productNameError,
    updateProductDraft,
    validateProductNameAvailability,
    generateApiKey,
    createFeature,
    deleteExistingApiKey,
    continueSetup,
    saveDraft,
    completeSetup,
  };
}

export type ProductSetupWizardController = ReturnType<
  typeof useProductSetupWizard
>;









