import type {
  GeneratedApiKey,
  ProductEnvironment,
} from "@/types/apiKey.types";
import type { ProductOption } from "@/services/product.service";

export interface GeneratedSecretState {
  value: GeneratedApiKey;
  product: ProductOption;
  environment?: ProductEnvironment;
}
