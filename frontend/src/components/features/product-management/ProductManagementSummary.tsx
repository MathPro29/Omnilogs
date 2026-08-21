import type { ProductOption } from "@/services/product.service";

export function ProductManagementSummary({
  products,
}: {
  products: ProductOption[];
}) {
  return (
    <section
      className="product-management-summary"
      aria-label="Product summary"
    >
      <div>
        <span>Visible products</span>
        <strong>{products.length}</strong>
      </div>
      <div>
        <span>Active</span>
        <strong>
          {products.filter((product) => product.is_active !== false).length}
        </strong>
      </div>
      <div>
        <span>Setup complete</span>
        <strong>
          {products.filter((product) => product.setup_status === "ACTIVE").length}
        </strong>
      </div>
    </section>
  );
}
