import { apiClient } from "@/api";

export interface FavoriteField {
  field_definition_id: number;
  product_id?: number;
  field_key: string;
  display_name?: string;
  field_path?: string;
  data_type: string;
  is_favorite: boolean;
  display_order?: number;
}

export interface AddFavoriteFieldInput {
  product_id: number;
  field_path: string;
  display_name?: string;
  sample_value?: unknown;
  detected_type: string;
}

export interface ReorderFavoriteFieldsInput {
  product_id: number;
  items: Array<{ field_definition_id: number; display_order: number }>;
}

export const customFieldFavoriteService = {
  list: async (productId: number) => {
    const response = await apiClient.get<FavoriteField[]>("/v1/custom-fields/favorites", { params: { product_id: productId } });
    return response.data ?? [];
  },
  add: async (input: AddFavoriteFieldInput) => {
    const response = await apiClient.post<FavoriteField>("/v1/custom-fields/favorites", input);
    return response.data;
  },
  reorder: async (input: ReorderFavoriteFieldsInput) => {
    await apiClient.put("/v1/custom-fields/favorites/reorder", input);
  },
  remove: async (fieldDefinitionId: number) => {
    await apiClient.delete(`/v1/custom-fields/favorites/${fieldDefinitionId}`);
  },
};
