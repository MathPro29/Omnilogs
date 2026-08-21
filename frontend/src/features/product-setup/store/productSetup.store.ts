import { create } from "zustand";
import type {
  FeatureNode,
  ExistingApiKeyInfo,
  GeneratedApiKeyInfo,
  ProductSetupDraft,
  ProjectInfo,
  TestLogResult,
} from "@/features/product-setup/types/productSetup.types";

interface ProductSetupState {
  currentStepIndex: number; // 0..6
  productId: number | null;
	editingExistingProduct: boolean;
  productDraft: ProductSetupDraft;
  projects: ProjectInfo[];
  features: FeatureNode[];
  generatedApiKey: GeneratedApiKeyInfo | null;
  apiKeySecret: string;
  existingApiKeys: ExistingApiKeyInfo[];
  testLogResult: TestLogResult | null;
  isValidationPassed: boolean;
  validationErrors: string[];

  // Actions
  setStepIndex: (index: number) => void;
  nextStep: () => void;
  prevStep: () => void;
  setProductId: (id: number) => void;
	setEditingExistingProduct: (editing: boolean) => void;
  updateProductDraft: (partial: Partial<ProductSetupDraft>) => void;
  setProjects: (projects: ProjectInfo[]) => void;
  setFeatures: (features: FeatureNode[]) => void;
  setGeneratedApiKey: (key: GeneratedApiKeyInfo | null) => void;
  setApiKeySecret: (secret: string) => void;
  setExistingApiKeys: (keys: ExistingApiKeyInfo[]) => void;
  setTestLogResult: (result: TestLogResult | null) => void;
  setValidationStatus: (passed: boolean, errors?: string[]) => void;
  resetSetup: () => void;
}

const initialDraft: ProductSetupDraft = {
  product_name: "",
  product_code: "",
  description: "",
  is_active: true,
  setup_status: "DRAFT",
  environments: [],
  projects: [],
  features: [],
};

export const useProductSetupStore = create<ProductSetupState>((set) => ({
  currentStepIndex: 0,
  productId: null,
	editingExistingProduct: false,
  productDraft: initialDraft,
  projects: [],
  features: [],
  generatedApiKey: null,
  apiKeySecret: "",
  existingApiKeys: [],
  testLogResult: null,
  isValidationPassed: false,
  validationErrors: [],

  setStepIndex: (index) => set({ currentStepIndex: Math.max(0, Math.min(6, index)) }),
  nextStep: () => set((state) => ({ currentStepIndex: Math.min(6, state.currentStepIndex + 1) })),
  prevStep: () => set((state) => ({ currentStepIndex: Math.max(0, state.currentStepIndex - 1) })),
  setProductId: (id) => set({ productId: id }),
	setEditingExistingProduct: (editingExistingProduct) => set({ editingExistingProduct }),
  updateProductDraft: (partial) =>
    set((state) => ({
      productDraft: { ...state.productDraft, ...partial },
    })),
  setProjects: (projects) => set({ projects }),
  setFeatures: (features) => set({ features }),
  setGeneratedApiKey: (key) => set({ generatedApiKey: key, apiKeySecret: key?.raw_key ?? "" }),
  setApiKeySecret: (apiKeySecret) => set({ apiKeySecret }),
  setExistingApiKeys: (existingApiKeys) => set({ existingApiKeys }),
  setTestLogResult: (result) => set({ testLogResult: result }),
  setValidationStatus: (passed, errors = []) => set({ isValidationPassed: passed, validationErrors: errors }),
  resetSetup: () =>
    set({
      currentStepIndex: 0,
      productId: null,
	editingExistingProduct: false,
      productDraft: initialDraft,
      projects: [],
      features: [],
      generatedApiKey: null,
      apiKeySecret: "",
      existingApiKeys: [],
      testLogResult: null,
      isValidationPassed: false,
      validationErrors: [],
    }),
}));



