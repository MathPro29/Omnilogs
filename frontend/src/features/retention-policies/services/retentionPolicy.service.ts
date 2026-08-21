import { apiClient } from "@/api";
import { API_ENDPOINTS } from "@/constants";
import type {
  CreateRetentionPolicyPayload,
  HistoricalResponse,
  CreateArchiveResponse,
  ImportArchiveResponse,
  LogArchive,
  PolicyStats,
  RetentionPolicy,
  RetentionPreviewResponse,
  RestoreArchiveResponse,
  SimulationResult,
  UpdateRetentionPolicyPayload,
  VerifyArchiveResponse,
  ReadyArchiveDownload,
} from "@/features/retention-policies/types/retentionPolicy.types";

export const retentionPolicyService = {
  async getPolicies(
    productId: number | string,
    environmentId: number,
  ): Promise<RetentionPolicy[]> {
    const response = await apiClient.get<RetentionPolicy[]>(
      API_ENDPOINTS.RETENTION_POLICIES.LIST(productId),
      { params: { environment_id: environmentId } },
    );
    return response.data;
  },

  async getPolicyById(
    productId: number | string,
    policyId: number | string,
  ): Promise<RetentionPolicy> {
    const response = await apiClient.get<RetentionPolicy>(
      API_ENDPOINTS.RETENTION_POLICIES.DETAIL(productId, policyId),
    );
    return response.data;
  },

  async createPolicy(
    productId: number | string,
    payload: CreateRetentionPolicyPayload,
  ): Promise<RetentionPolicy> {
    const response = await apiClient.post<RetentionPolicy>(
      API_ENDPOINTS.RETENTION_POLICIES.CREATE(productId),
      payload,
    );
    return response.data;
  },

  async updatePolicy(
    productId: number | string,
    policyId: number | string,
    payload: UpdateRetentionPolicyPayload,
  ): Promise<RetentionPolicy> {
    const response = await apiClient.put<RetentionPolicy>(
      API_ENDPOINTS.RETENTION_POLICIES.UPDATE(productId, policyId),
      payload,
    );
    return response.data;
  },

  async deletePolicy(
    productId: number | string,
    policyId: number | string,
  ): Promise<void> {
    await apiClient.delete(
      API_ENDPOINTS.RETENTION_POLICIES.DELETE(productId, policyId),
    );
  },

  async togglePolicy(
    productId: number | string,
    policyId: number | string,
    isActive: boolean,
  ): Promise<void> {
    await apiClient.patch(
      API_ENDPOINTS.RETENTION_POLICIES.TOGGLE(productId, policyId),
      { is_active: isActive },
    );
  },

  async triggerPolicyNow(
    productId: number | string,
    policyId: number | string,
  ): Promise<RetentionPolicy> {
    const response = await apiClient.post<RetentionPolicy>(
      API_ENDPOINTS.RETENTION_POLICIES.RUN_NOW(productId, policyId),
    );
    return response.data;
  },

  async getStats(
    productId: number | string,
    environmentId: number,
  ): Promise<PolicyStats> {
    const response = await apiClient.get<PolicyStats>(
      API_ENDPOINTS.RETENTION_POLICIES.STATS(productId),
      { params: { environment_id: environmentId } },
    );
    return response.data;
  },

  async simulatePolicy(
    productId: number | string,
    payload: {
      retention_mode: string;
      retention_unit: string;
      retention_value: number;
      folder_structure?: string;
    },
  ): Promise<SimulationResult> {
    const response = await apiClient.post<SimulationResult>(
      API_ENDPOINTS.RETENTION_POLICIES.SIMULATE(productId),
      payload,
    );
    return response.data;
  },

  async getHistorical(
    productId: number | string,
    params: {
      environment_id: number;
      date_from?: string;
      date_to?: string;
      group_by?: "day" | "week" | "month" | "year";
      category_id?: number;
      feature_id?: number;
      sub_feature_id?: number;
    },
  ): Promise<HistoricalResponse> {
    const response = await apiClient.get<HistoricalResponse>(
      API_ENDPOINTS.RETENTION_POLICIES.HISTORICAL(productId),
      { params },
    );
    return response.data;
  },

  async previewPolicy(
    productId: number | string,
    payload: {
      environment_id: number;
      policy_id?: number;
      date_from?: string;
      date_to?: string;
      group_by?: "day" | "week" | "month" | "year";
      category_id?: number;
      feature_id?: number;
      sub_feature_id?: number;
    },
  ): Promise<RetentionPreviewResponse> {
    const response = await apiClient.post<RetentionPreviewResponse>(
      API_ENDPOINTS.RETENTION_POLICIES.PREVIEW(productId),
      payload,
    );
    return response.data;
  },

  async createArchive(
    productId: number | string,
    payload: {
      environment_id: number;
      policy_id: number;
      date_from: string;
      date_to: string;
      backup_type?: "POLICY" | "MANUAL";
      backup_tag?: string;
    },
  ): Promise<CreateArchiveResponse> {
    const response = await apiClient.post<CreateArchiveResponse>(
      API_ENDPOINTS.LOG_ARCHIVES.CREATE(productId),
      payload,
    );
    return response.data;
  },

  async listArchives(
    productId: number | string,
    environmentId: number,
    scope?: {
      category_id?: number;
      feature_id?: number;
      sub_feature_id?: number;
      date_from?: string;
      date_to?: string;
      backup_type?: "POLICY" | "MANUAL";
      backup_tag?: string;
    },
  ): Promise<LogArchive[]> {
    const response = await apiClient.get<LogArchive[]>(
      API_ENDPOINTS.LOG_ARCHIVES.LIST(productId),
      { params: { environment_id: environmentId, ...scope } },
    );
    return response.data;
  },

  async listReadyDownloads(
    productId: number | string,
    environmentId: number,
  ): Promise<ReadyArchiveDownload[]> {
    const response = await apiClient.get<ReadyArchiveDownload[]>(
      API_ENDPOINTS.LOG_ARCHIVES.READY_DOWNLOADS(productId),
      { params: { environment_id: environmentId } },
    );
    return response.data;
  },

  async downloadArchives(
    productId: number | string,
    params: {
      environment_id: number;
      policy_id?: number;
      category_id?: number;
      feature_id?: number;
      sub_feature_id?: number;
      date_from?: string;
      date_to?: string;
    },
  ): Promise<void> {
    const response = await apiClient.get<Blob>(
      API_ENDPOINTS.LOG_ARCHIVES.DOWNLOAD(productId),
      { params, responseType: "blob" },
    );
    const contentDisposition = response.headers["content-disposition"] as string | undefined;
    const filenameMatch = contentDisposition?.match(/filename="?([^";]+)"?/i);
    const filename = filenameMatch?.[1] ?? `omnilogs-archives-${productId}.zip`;
    const url = URL.createObjectURL(response.data);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = filename;
    document.body.appendChild(anchor);
    anchor.click();
    anchor.remove();
    URL.revokeObjectURL(url);
  },

  async downloadReadyArchive(
    productId: number | string,
    environmentId: number,
    downloadId: string,
  ): Promise<void> {
    const response = await apiClient.get<Blob>(
      API_ENDPOINTS.LOG_ARCHIVES.READY_DOWNLOAD(productId, downloadId),
      { params: { environment_id: environmentId }, responseType: "blob" },
    );
    const contentDisposition = response.headers["content-disposition"] as string | undefined;
    const filenameMatch = contentDisposition?.match(/filename="?([^";]+)"?/i);
    const filename = filenameMatch?.[1] ?? `omnilogs-ready-${downloadId}.zip`;
    const url = URL.createObjectURL(response.data);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = filename;
    document.body.appendChild(anchor);
    anchor.click();
    anchor.remove();
    URL.revokeObjectURL(url);
  },

    async importArchive(
      productId: number | string,
      environmentId: number,
      file: File,
      scope?: { category_id?: number; feature_id?: number; sub_feature_id?: number },
    ): Promise<ImportArchiveResponse> {
      const form = new FormData();
      form.append("file", file);
      const response = await apiClient.post<ImportArchiveResponse>(
        API_ENDPOINTS.LOG_ARCHIVES.IMPORT(productId),
        form,
        {
          params: { environment_id: environmentId, ...scope },
          headers: { "Content-Type": "multipart/form-data" },
        },
      );
      return response.data;
    },

  async verifyArchive(
    productId: number | string,
    environmentId: number,
    archiveId: string,
    policyId?: number,
  ): Promise<VerifyArchiveResponse> {
    const response = await apiClient.post<VerifyArchiveResponse>(
      API_ENDPOINTS.LOG_ARCHIVES.VERIFY(productId, archiveId),
      policyId ? { policy_id: policyId } : undefined,
      { params: { environment_id: environmentId } },
    );
    return response.data;
  },

    async restoreArchive(
    productId: number | string,
    environmentId: number,
    archiveId: string,
    conflictStrategy: "SKIP_EXISTING" | "OVERWRITE" = "SKIP_EXISTING",
  ): Promise<RestoreArchiveResponse> {
    const response = await apiClient.post<RestoreArchiveResponse>(
      API_ENDPOINTS.LOG_ARCHIVES.RESTORE(productId, archiveId),
      { mode: "RESTORE_TO_ACTIVE", conflict_strategy: conflictStrategy },
      { params: { environment_id: environmentId } },
    );
      return response.data;
    },

    async restoreSnapshotForSearch(
      productId: number | string,
      environmentId: number,
      archiveId: string,
    ): Promise<RestoreArchiveResponse> {
      const response = await apiClient.post<RestoreArchiveResponse>(
        API_ENDPOINTS.LOG_ARCHIVES.RESTORE(productId, archiveId),
        { mode: "SNAPSHOT_SEARCH_ONLY" },
        { params: { environment_id: environmentId } },
      );
      return response.data;
    },
  };
