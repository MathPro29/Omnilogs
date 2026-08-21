import { useMemo, useState } from "react";
import { message } from "antd";
import { useQuery } from "@tanstack/react-query";
import { useSetupPermissions } from "@/features/product-setup/hooks/useSetupPermissions";
import { useProductSetupStore } from "@/features/product-setup/store/productSetup.store";
import { productSetupService } from "@/features/product-setup/services/productSetup.service";
import { productService, type ProductOption } from "@/services/product.service";
import { apiKeyService } from "@/services/apiKey.service";
import { useAuthStore } from "@/store";
import type { GeneratedApiKey, ProductEnvironment } from "@/types/apiKey.types";
import type { GeneratedSecretState } from "@/features/product-management/types/productManagement.types";
import {
  formatPlatformRoles,
  toProductSetupDraft,
} from "@/features/product-management/utils/productManagement";

export function useProductManagement() {
  const [setupOpen, setSetupOpen] = useState(false);
  const [selected, setSelected] = useState<ProductOption>();
  const [productDrawerOpen, setProductDrawerOpen] = useState(false);
  const [search, setSearch] = useState("");
  const [pendingSecret, setPendingSecret] = useState<GeneratedSecretState>();
  const [visibleSecret, setVisibleSecret] = useState<GeneratedSecretState>();
  const { canCreateProduct, isPlatformAdmin } = useSetupPermissions();
  const roles = useAuthStore((state) => state.roles);
  const setupStore = useProductSetupStore();
  const query = useQuery({
    queryKey: ["products", "management"],
    queryFn: productService.listOptions,
    staleTime: 15_000,
  });
  const products = useMemo(
    () =>
      (query.data ?? []).filter((product) =>
        `${product.name} ${product.id}`
          .toLowerCase()
          .includes(search.toLowerCase()),
      ),
    [query.data, search],
  );

  const openProduct = (product: ProductOption) => {
    setSelected(product);
    setProductDrawerOpen(true);
  };
  
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [editingProduct, setEditingProduct] = useState<ProductOption>();
  
  const openEditProduct = (product: ProductOption) => {
    setEditingProduct(product);
    setEditModalOpen(true);
  };
  
  const closeEditProduct = () => {
    setEditModalOpen(false);
  };
  
  const handleEditSuccess = () => {
    setEditModalOpen(false);
    void query.refetch();
  };

  const startCreate = () => {
    setupStore.resetSetup();
    setSetupOpen(true);
  };
  const continueSetup = async (product: ProductOption) => {
    setupStore.resetSetup();
    setupStore.setProductId(product.id);
    setupStore.setEditingExistingProduct(true);
	setupStore.setStepIndex(2);
    setupStore.updateProductDraft(toProductSetupDraft(product));

    try {
      const projects = await productSetupService.getProjects(product.id);
      const [existingKeys, environments] = await Promise.all([
        apiKeyService.list(product.id).catch(() => []),
        productSetupService.listEnvironments(product.id).catch(() => []),
      ]);
      const features = (
        await Promise.all(
          projects
            .filter((project) => project.project_id !== undefined)
            .map(async (project) => {
              const list = await productSetupService.getFeatures(
                product.id,
                project.project_id!,
              ).catch(() => []);
              return list.map((feature) => ({
                ...feature,
                project_id: feature.project_id ?? project.project_id!,
                id:
                  feature.id ||
                  (feature.category_id !== undefined
                    ? String(feature.category_id)
                    : `feat_${Math.random().toString(36).substring(2, 9)}`),
                parent_id: feature.parent_id ?? null,
              }));
            }),
        )
      ).flat();

      setupStore.updateProductDraft({ environments });
      setupStore.setProjects(projects);
      setupStore.setFeatures(features);
      setupStore.setExistingApiKeys(existingKeys.map((key) => ({ key_id: key.keyId, key_name: key.name, key_prefix: key.prefix, environment_id: key.environmentId, is_active: key.active, created_at: key.createdAt })));
    } catch (error) {
      message.error("Could not load the existing project and feature structure.");
      console.error("Failed to restore product setup structure:", error);
    }

    setSelected(undefined);
    setProductDrawerOpen(false);
    setSetupOpen(true);
  };
  const closeSetup = () => {
    setSetupOpen(false);
    void query.refetch();
  };
  const closeProduct = () => setProductDrawerOpen(false);
  const afterProductDrawerChange = (open: boolean) => {
    if (open) return;
    setSelected(undefined);
    if (pendingSecret) {
      setVisibleSecret(pendingSecret);
      setPendingSecret(undefined);
    }
  };
  const handleGenerated = (
    value: GeneratedApiKey,
    environment?: ProductEnvironment,
  ) => {
    if (!selected) return;
    setPendingSecret({ value, product: selected, environment });
    closeProduct();
  };

  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [isDeleting, setIsDeleting] = useState(false);

  const deleteProduct = async (id: number) => {
    try {
      setIsDeleting(true);
      await productService.deleteProduct(id);
      setSelectedRowKeys((prev) => prev.filter((k) => k !== id));
      void query.refetch();
    } catch (err: unknown) {
      console.error("Failed to delete product:", err);
      throw err;
    } finally {
      setIsDeleting(false);
    }
  };

  const bulkDeleteProducts = async (idsToDelete?: number[]) => {
    const ids = idsToDelete ?? (selectedRowKeys as number[]);
    if (ids.length === 0) return;
    try {
      setIsDeleting(true);
      await productService.bulkDeleteProducts(ids);
      setSelectedRowKeys([]);
      void query.refetch();
    } catch (err: unknown) {
      console.error("Failed to bulk delete products:", err);
      throw err;
    } finally {
      setIsDeleting(false);
    }
  };

  return {
    setupOpen,
    selected,
    productDrawerOpen,
    search,
    visibleSecret,
    products,
    query,
    selectedRowKeys,
    setSelectedRowKeys,
    isDeleting,
    deleteProduct,
    bulkDeleteProducts,
    editModalOpen,
    editingProduct,
    openEditProduct,
    closeEditProduct,
    handleEditSuccess,
    roleLabel: formatPlatformRoles(roles),
    canCreateProduct,
    isPlatformAdmin,
    setSearch,
    openProduct,
    startCreate,
    continueSetup,
    closeSetup,
    closeProduct,
    afterProductDrawerChange,
    handleGenerated,
    closeSecret: () => setVisibleSecret(undefined),
  };
}





