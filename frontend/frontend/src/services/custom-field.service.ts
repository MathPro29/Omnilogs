import { apiClient } from '@/api';

export type CustomField = {
  field_definition_id: number;
  product_id?: number | null;
  project_id?: number | null;
  category_id?: number | null;
  field_key: string;
  display_name?: string | null;
  description?: string | null;
  source_section: string;
  field_path?: string | null;
  elastic_field_name: string;
  data_type: string;
  value_source_type: string;
  value_source_key?: string | null;
  is_required: boolean;
  is_sensitive: boolean;
  mask_before_index: boolean;
  encrypt_before_archive: boolean;
  is_visible: boolean;
  is_searchable: boolean;
  is_filterable: boolean;
  is_sortable: boolean;
  is_aggregatable: boolean;
  display_order?: number | null;
  default_value?: string | null;
  is_active: boolean;
  schema_version?: number;
  field_type?: string | null;
  config_json?: any; // Represents FieldConfigJSON
  is_favorite: boolean;
  sample_value?: unknown;
  detected_type?: string | null;
  sample_path_found: boolean;
  enum_options: CustomFieldOption[];
};

export type JSONFieldNode = {
  field_name: string;
  json_path: string;
  sample_value: unknown;
  data_type: string;
  is_leaf: boolean;
  children?: JSONFieldNode[];
};

export type ParsedJSONFields = {
  nodes: JSONFieldNode[];
  field_count: number;
  max_nodes: number;
};

export type CustomFieldOption = {
  option_id: number;
  field_definition_id: number;
  option_key: string;
  option_label: string;
  option_value: string;
  display_order?: number | null;
  color_code?: string | null;
  description?: string | null;
  is_default: boolean;
  is_active: boolean;
};

export type CustomFieldDetail = {
  field: CustomField;
  options: CustomFieldOption[];
};

const endpoint = '/custom-fields';
export const customFieldService = {
  list: async (productId?: number, projectId?: number, categoryId?: number): Promise<CustomField[]> => {
    const response = await apiClient.get<CustomField[]>(`${endpoint}/`, { params: { product_id: productId, project_id: projectId, category_id: categoryId, active_only: true } });
    return response.data;
  },
  create: async (payload: Record<string, unknown>) => {
    const response = await apiClient.post<CustomField>(`${endpoint}/`, payload);
    return response.data;
  },
  get: async (id: number): Promise<CustomFieldDetail> => {
    const response = await apiClient.get<CustomFieldDetail>(`${endpoint}/${id}`);
    return response.data;
  },
  update: async (id: number, payload: Record<string, unknown>): Promise<CustomFieldDetail> => {
    const response = await apiClient.put<CustomFieldDetail>(`${endpoint}/${id}`, payload);
    return response.data;
  },
  remove: async (id: number) => { await apiClient.delete(`${endpoint}/${id}`); },
  bulkDelete: async (ids: number[]): Promise<void> => {
    await apiClient.delete(`${endpoint}/bulk/delete`, { data: { field_definition_ids: ids } });
  },
  parseJSON: async (sampleJSON: string): Promise<ParsedJSONFields> =>
    (await apiClient.post(`${endpoint}/parse-json`, { sample_json: JSON.parse(sampleJSON) })).data,
  listFavorites: async (productId: number, projectId?: number, categoryId?: number): Promise<CustomField[]> =>
    (await apiClient.get(`${endpoint}/favorites`, { params: { product_id: productId, project_id: projectId, category_id: categoryId } })).data,
  favorite: async (payload: {
    product_id: number;
    project_id?: number;
    category_id?: number;
    field_path: string;
    display_name?: string;
    sample_value?: unknown;
    detected_type: string;
  }): Promise<CustomField> => (await apiClient.post(`${endpoint}/favorites`, payload)).data,
  reorderFavorites: async (productId: number, items: Array<{ field_definition_id: number; display_order: number }>): Promise<void> => {
    await apiClient.put(`${endpoint}/favorites/reorder`, { product_id: productId, items });
  },
  removeFavorite: async (id: number): Promise<void> => {
    await apiClient.delete(`${endpoint}/favorites/${id}`);
  },
  listOptions: async (id: number, activeOnly = false): Promise<CustomFieldOption[]> =>
    (await apiClient.get(`${endpoint}/${id}/options`, { params: { active_only: activeOnly || undefined } })).data,
  createOption: async (id: number, payload: Partial<CustomFieldOption>): Promise<CustomFieldOption> =>
    (await apiClient.post(`${endpoint}/${id}/options`, payload)).data,
  updateOption: async (fieldId: number, optionId: number, payload: Partial<CustomFieldOption>): Promise<CustomFieldOption> =>
    (await apiClient.put(`${endpoint}/${fieldId}/options/${optionId}`, payload)).data,
  removeOption: async (fieldId: number, optionId: number): Promise<void> => {
    await apiClient.delete(`${endpoint}/${fieldId}/options/${optionId}`);
  },
};
