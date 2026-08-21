import { Button } from "antd";
import { ArrowLeftOutlined } from "@ant-design/icons";
import { PageTransition } from "@/components";
import { GeneratedSecretDialog } from "@/components/features/logs-connections/GeneratedSecretDialog";
import { ProductDetailsDrawer } from "@/components/features/product-management/ProductDetailsDrawer";
import { ProductManagementHeader } from "@/components/features/product-management/ProductManagementHeader";
import { ProductManagementSummary } from "@/components/features/product-management/ProductManagementSummary";
import { ProductManagementTable } from "@/components/features/product-management/ProductManagementTable";
import { ProductSetupWizard } from "@/components/features/product-setup/ProductSetupWizard";
import { useProductManagement } from "@/features/product-management/hooks/useProductManagement";
import "@/styles/pages/product-management.css";

export function ProductCreatePage() {
  const management = useProductManagement();
  const displayProducts = management.products;

  if (management.setupOpen) {
    return (
      <PageTransition>
        <div className="product-setup-shell">
          <Button icon={<ArrowLeftOutlined />} onClick={management.closeSetup}>
            กลับหน้าหลัก
          </Button>
          <ProductSetupWizard />
        </div>
      </PageTransition>
    );
  }

  const canManageSelected = Boolean(
    management.selected &&
    (management.selected.is_owner || management.isPlatformAdmin),
  );

  return (
    <PageTransition>
      <main className="product-management-page">
        <ProductManagementHeader
          roleLabel={management.roleLabel}
          canCreate={management.canCreateProduct}
          onCreate={management.startCreate}
        />
        <ProductManagementSummary products={management.query.data ?? []} />
        <ProductManagementTable
          products={displayProducts}
          search={management.search}
          isLoading={management.query.isLoading}
          isFetching={management.query.isFetching}
          isError={management.query.isError}
          isPlatformAdmin={management.isPlatformAdmin}
          selectedRowKeys={management.selectedRowKeys}
          isDeleting={management.isDeleting}
          onSearchChange={management.setSearch}
          onRefresh={() => void management.query.refetch()}
          onOpen={management.openProduct}
          onEdit={management.continueSetup}
          onSelectChange={management.setSelectedRowKeys}
          onDeleteProduct={management.deleteProduct}
          onBulkDeleteProducts={management.bulkDeleteProducts}
        />
        <ProductDetailsDrawer
          open={management.productDrawerOpen}
          product={management.selected}
          canManage={canManageSelected}
          onClose={management.closeProduct}
          onAfterOpenChange={management.afterProductDrawerChange}
          onContinueSetup={management.continueSetup}
          onEdit={management.continueSetup}
          onDeleteProduct={management.deleteProduct}
          onGenerated={management.handleGenerated}
        />
        <GeneratedSecretDialog
          value={management.visibleSecret?.value ?? null}
          product={management.visibleSecret?.product}
          environment={management.visibleSecret?.environment}
          onClose={management.closeSecret}
        />
      </main>
    </PageTransition>
  );
}
