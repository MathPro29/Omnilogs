import type { ProductSetupWizardController } from "@/features/product-setup/hooks/useProductSetupWizard";
import { ApiKeyStep } from "./ApiKeyStep";
import { ConnectionGuideStep } from "./ConnectionGuideStep";
import { FeatureHierarchyStep } from "./FeatureHierarchyStep";
import { ProductInformationStep } from "./ProductInformationStep";
import { ProductReviewStep } from "./ProductReviewStep";
import { ProjectBuilderStep } from "./ProjectBuilderStep";
import { SetupCompleteStep } from "./SetupCompleteStep";

export function ProductSetupStepContent({
  wizard,
  navigation,
}: {
  wizard: ProductSetupWizardController;
  navigation?: React.ReactNode;
}) {
  const { store } = wizard;
  const selectedEnvironmentCode =
    store.generatedApiKey?.environment_code ||
    store.productDraft.environments.find((environment) => environment.is_default)
      ?.environment_code ||
    store.productDraft.environments[0]?.environment_code ||
    "";

  switch (store.currentStepIndex) {
    case 0:
      return (
        <ProductInformationStep
          initialValues={{
            product_name: store.productDraft.product_name,
            product_code: store.productDraft.product_code,
            description: store.productDraft.description,
            owner_id: store.productDraft.owner_id,
            is_active: store.productDraft.is_active,
            environments: store.productDraft.environments,
          }}
          onChange={wizard.updateProductDraft}
          productNameError={wizard.productNameError}
          onProductNameBlur={() => void wizard.validateProductNameAvailability()}
          isCreated={wizard.isProductCreated}
          navigation={navigation}
        />
      );
    case 1:
      return (
        <ProjectBuilderStep
          productName={store.productDraft.product_name}
          productCode={store.productDraft.product_code}
          projects={store.projects}
          onChange={store.setProjects}
        />
      );
    case 2:
      return (
        <FeatureHierarchyStep
          projects={store.projects}
          features={store.features}
          onChange={store.setFeatures}
          onCreateFeature={wizard.createFeature}
          hasUnsavedChanges={wizard.hasUnsavedChangesForCurrentStep}
        />
      );
    case 3:
      return (
        <ProductReviewStep
          draft={store.productDraft}
          projects={store.projects}
          features={store.features}
          validation={{
            valid: wizard.validation.isStructureValid,
            errors: wizard.validation.structureErrors,
          }}
          onJumpToStep={store.setStepIndex}
        />
      );
    case 4:
      return (
        <ApiKeyStep
          productName={store.productDraft.product_name}
          productCode={store.productDraft.product_code}
          generatedKey={store.generatedApiKey}
          existingApiKeys={store.existingApiKeys}
          projects={store.projects}
          features={store.features}
          environments={store.productDraft.environments}
          onDeleteExistingApiKey={wizard.deleteExistingApiKey}
          onGenerate={wizard.generateApiKey}
          isGenerating={wizard.isSubmitting}
        />
      );
    case 5:
      return (
        <ConnectionGuideStep
          productId={store.productId ?? undefined}
          environmentCode={selectedEnvironmentCode}
          projects={store.projects}
          features={store.features}
          apiKey={store.generatedApiKey}
          testLogResult={store.testLogResult}
          apiKeySecret={store.apiKeySecret}
          onApiKeySecretChange={store.setApiKeySecret}
          onTestResult={store.setTestLogResult}
        />
      );
    case 6:
      return (
        <SetupCompleteStep
          draft={store.productDraft}
          projects={store.projects}
          features={store.features}
          onComplete={wizard.completeSetup}
        />
      );
    default:
      return null;
  }
}

