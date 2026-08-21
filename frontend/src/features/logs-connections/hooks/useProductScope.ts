import { useEffect, useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { productService } from "@/services/product.service";
import { apiKeyService } from "@/services/apiKey.service";
import { useConnectionStore } from "@/features/logs-connections/store/connection.store";

export function useProductScope() {
  const productId = useConnectionStore((state) => state.productId); const setProductId = useConnectionStore((state) => state.setProductId);
  const productsQuery = useQuery({ queryKey: ["products", "connections"], queryFn: productService.listOptions });
  const products = useMemo(() => productsQuery.data ?? [], [productsQuery.data]);
  useEffect(() => { if (!productId && products.length) setProductId(products[0].id); }, [productId, products, setProductId]);
  const environmentsQuery = useQuery({ queryKey: ["products", productId, "environments"], queryFn: () => apiKeyService.environments(productId!), enabled: Boolean(productId) });
  return { productId, products, productsQuery, environments: environmentsQuery.data ?? [], environmentsQuery };
}

