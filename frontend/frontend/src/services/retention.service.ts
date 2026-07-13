import { apiClient } from '@/api';

export interface RetentionPolicy {
  elastic_policy_id: number;
  product_id: number;
  environment_id?: number | null;
  project_id?: number | null;
  category_id?: number | null;
  index_prefix: string;
  retention_days?: number | null;
  is_active: boolean;
}

export interface LogArchiveRecord {
  archive_id: string;
  product_id: number;
  environment_id?: number | null;
  archive_year: number;
  archive_month: number;
  date_from?: string | null;
  date_to?: string | null;
  storage_provider: string;
  bucket_name?: string | null;
  file_path?: string | null;
  file_format?: string | null;
  total_logs?: number | null;
  compressed_size_bytes?: number | null;
  status?: string | null;
  exported_at?: string | null;
  restored_at?: string | null;
  deleted_at?: string | null;
  purged_at?: string | null;
  created_at?: string | null;
}

export interface CreateRetentionPayload {
  product_id: number;
  environment_id?: number;
  project_id?: number;
  category_id?: number;
  index_prefix: string;
  rollover_type: 'age' | 'size';
  retention_days: number;
  number_of_shards?: number;
  number_of_replicas?: number;
}

export const retentionService = {
  createPolicy: async (productId: number, payload: CreateRetentionPayload): Promise<RetentionPolicy> => {
    const response = await apiClient.post<RetentionPolicy>(`/products/${productId}/elastic-index-policies`, payload);
    return response.data;
  },
  listPolicies: async (productId: number, environmentId?: number): Promise<RetentionPolicy[]> => {
    const response = await apiClient.get<RetentionPolicy[]>(`/products/${productId}/elastic-index-policies`, {
      params: environmentId ? { environment_id: environmentId } : undefined,
    });
    return response.data;
  },
  listArchives: async (productId: number, environmentId?: number): Promise<LogArchiveRecord[]> => {
    const response = await apiClient.get<LogArchiveRecord[]>(`/products/${productId}/log-archives`, {
      params: environmentId ? { environment_id: environmentId } : undefined,
    });
    return response.data;
  },
  updatePolicy: async (productId: number, policyId: number, payload: any): Promise<RetentionPolicy> => {
    const response = await apiClient.patch<RetentionPolicy>(`/products/${productId}/elastic-index-policies/${policyId}`, {
      elastic_policy_id: policyId,
      ...payload,
    });
    return response.data;
  },
  deletePolicy: async (productId: number, policyId: number, environmentId?: number | null): Promise<void> => {
    await apiClient.delete(`/products/${productId}/elastic-index-policies/${policyId}`, {
      data: {
        product_id: productId,
        elastic_policy_id: policyId,
        environment_id: environmentId || undefined,
      },
    });
  },
  pushToArchives: async (productId: number, policyId: number, environmentId: number | null | undefined, fromDate: string, toDate: string): Promise<{ success_count: number; failed_count: number }> => {
    const response = await apiClient.post<{ success_count: number; failed_count: number }>(`/products/${productId}/elastic-index-policies/${policyId}/push-to-archives`, {
      product_id: productId,
      elastic_policy_id: policyId,
      environment_id: environmentId || undefined,
      from_date: fromDate,
      to_date: toDate,
    });
    return response.data;
  },
  clearAllLogs: async (productId: number): Promise<void> => {
    await apiClient.delete(`/products/${productId}/elastic-index-policies/clear-logs`);
  },
  restoreArchive: async (productId: number, archiveId: string): Promise<void> => {
    await apiClient.post(`/products/${productId}/log-archives/${archiveId}/restore`);
  },
};
