import type { ProductSetupDraft } from "@/features/product-setup/types/productSetup.types";
import type { ProductOption } from "@/services/product.service";

const setupStatuses: ProductSetupDraft["setup_status"][] = [
  "DRAFT",
  "STRUCTURE_INCOMPLETE",
  "READY_FOR_API_KEY",
  "WAITING_FOR_FIRST_LOG",
  "ACTIVE",
];

export function toProductSetupDraft(
  product: ProductOption,
): Partial<ProductSetupDraft> {
  return {
    product_name: product.name,
    product_code: product.product_code ?? "",
    description: product.description ?? "",
    is_active: product.is_active ?? true,
    setup_status: setupStatuses.includes(
      product.setup_status as ProductSetupDraft["setup_status"],
    )
      ? (product.setup_status as ProductSetupDraft["setup_status"])
      : "DRAFT",
  };
}

export function formatPlatformRoles(roles?: unknown[]): string {
  return (
    (roles ?? [])
      .filter((role): role is string => typeof role === "string" && Boolean(role))
      .map((role) => role.replace(/_/g, " "))
      .join(", ") || "product scoped"
  );
}
