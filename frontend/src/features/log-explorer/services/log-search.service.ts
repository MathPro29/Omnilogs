import { apiClient } from "@/api";

export type SearchOperator = "eq" | "neq" | "contains" | "gt" | "gte" | "lt" | "lte" | "exists" | "not_exists";


export interface SearchRule {
  id: string;
  field: string;
  label: string;
  type: "text" | "number" | "keyword";
  operator: SearchOperator;
  value?: string | number;
  enabled: boolean;
}

export interface SearchModel {
  version: 1;
  scope: { product_id: number; environment_id?: number; project_id?: number; category_id?: number; archive_id?: string };
  time_range: { field: "@timestamp"; type: "relative"; value: string };
  root: {
    type: "group";
    operator: "AND" | "OR";
    children: Array<{
      type: "condition";
      field: string;
      data_type: SearchRule["type"];
      operator: SearchOperator;
      value?: string | number;
      enabled: boolean;
    }>;
  };
  page_size: number;
}

export interface SearchRecord {
  id?: string;
  log_id?: string;
  "@timestamp"?: string;
  timestamp?: string;
  level?: string;
  service?: string;  project_id?: number; source_project_id?: number;
  product_id?: number;
  environment_id?: number;
  method?: string;
  path?: string;
  route_pattern?: string;
  category_id?: number;
  feature_path_ids?: string;
  feature_full_path?: string;
  routing_status?: string;
  routing_method?: string;
  message?: string;
  payload?: Record<string, unknown>;
  [key: string]: unknown;
}

interface Envelope<T> {
  data: T;
  meta?: { total?: number; page?: number; perPage?: number; totalPage?: number };
}

export const logSearchService = {
  execute: async (model: SearchModel, page = 1, perPage = 20, signal?: AbortSignal) => {
    // Inject pagination into model
    const payload = {
      ...model,
      page: page,
      page_size: perPage
    };

    const response = await apiClient.post<Envelope<SearchRecord[]>>("/v1/logs/search", payload, {
      signal,
    });
    return {
      records: response.data.data ?? [],
      total: response.data.meta?.total ?? response.data.data?.length ?? 0,
    };
  },
};
