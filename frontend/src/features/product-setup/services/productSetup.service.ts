import { apiClient } from "@/api";
import type {
  EnvironmentInfo,
  FeatureNode,
  GeneratedApiKeyInfo,
  ProjectInfo,
  SetupStatusData,
  TestLogPayload,
} from "@/features/product-setup/types/productSetup.types";
import { getOmniLogsIngestUrl } from "@/utils/omnilogs-url";

interface Envelope<T> {
  data: T;
  message?: string;
  success?: boolean;
}

type ProductSetupCreateResult =
  | { product_id: number; id?: number }
  | { product_id?: number; id: number };

interface FeatureRecord extends Omit<FeatureNode, "id"> {
  category_id: number;
  id?: string;
}

interface ApiKeyWire {
  key_id?: number;
  KeyID?: number;
  key_name?: string;
  KeyName?: string;
  raw_key?: string;
  rawKey?: string;
  api_key?: string;
  key?: string;
  key_prefix?: string;
  KeyPrefix?: string;
  product_code?: string;
  expires_at?: string;
  ExpiresAt?: string;
}

const endpoints = {
  product: (productId: number) => `/v1/products/${productId}`,
  projects: (productId: number) => `/v1/products/${productId}/projects`,
  features: (productId: number, projectId: number) =>
    `/v1/products/${productId}/projects/${projectId}/features`,
};

export const productSetupService = {
  async createProductSetup(payload: {
    product_name: string;
    product_code?: string;
    description?: string;
    owner_id?: number;
    environments?: Array<{ environment_code: string; environment_name: string; environment_type?: string }>;  }): Promise<ProductSetupCreateResult> {
    const response = await apiClient.post<Envelope<ProductSetupCreateResult>>(
      "/v1/products/setup",
      payload,
    );
    return response.data.data;
  },

  async listEnvironments(productId: number): Promise<EnvironmentInfo[]> { const response = await apiClient.get<Envelope<EnvironmentInfo[]>>("/v1/products/" + productId + "/environments"); return response.data.data ?? []; },

  async createEnvironment(productId: number, environment: Pick<EnvironmentInfo, "environment_code" | "environment_name" | "environment_type">): Promise<EnvironmentInfo> { const response = await apiClient.post<Envelope<EnvironmentInfo>>("/v1/products/" + productId + "/environments", environment); return response.data.data; },

  async deleteEnvironment(productId: number, environmentId: number): Promise<void> { await apiClient.delete("/v1/products/" + productId + "/environments/" + environmentId); },

  async getSetupStatus(productId: number): Promise<SetupStatusData> {
    const response = await apiClient.get<Envelope<SetupStatusData>>(
      `${endpoints.product(productId)}/setup-status`,
    );
    return response.data.data;
  },

  async updateProduct(
    productId: number,
    payload: {
      product_name?: string;
      product_code?: string;
      description?: string;
      setup_status?: string;
    },
  ): Promise<unknown> {
    const response = await apiClient.patch<Envelope<unknown>>(
      `${endpoints.product(productId)}/setup/product`,
      payload,
    );
    return response.data.data;
  },

  async getProjects(productId: number): Promise<ProjectInfo[]> {
    const response = await apiClient.get<Envelope<ProjectInfo[]>>(
      endpoints.projects(productId),
    );
    return response.data.data ?? [];
  },

  async createProject(
    productId: number,
    payload: {
      project_name: string;
      project_code?: string;
      description?: string;
    },
  ): Promise<ProjectInfo> {
    const response = await apiClient.post<Envelope<ProjectInfo>>(
      endpoints.projects(productId),
      payload,
    );
    return response.data.data;
  },

  async updateProject(
    productId: number,
    projectId: number,
    payload: {
      project_name?: string;
      project_code?: string;
      description?: string;
      is_active?: boolean;
    },
  ): Promise<ProjectInfo> {
    const response = await apiClient.patch<Envelope<ProjectInfo>>(
      `${endpoints.projects(productId)}/${projectId}`,
      payload,
    );
    return response.data.data;
  },

  async deleteProject(productId: number, projectId: number): Promise<void> {
    await apiClient.delete(`${endpoints.projects(productId)}/${projectId}`);
  },

  async getFeatures(
    productId: number,
    projectId: number,
  ): Promise<FeatureRecord[]> {
    const response = await apiClient.get<Envelope<FeatureRecord[]>>(
      endpoints.features(productId, projectId),
    );
    return response.data.data ?? [];
  },

  async createFeature(
    productId: number,
    projectId: number,
    payload: {
      category_name: string;
      category_code?: string;
      description?: string;
      parent_id?: number | null;
      category_type?: string;
    },
  ): Promise<FeatureRecord> {
    const response = await apiClient.post<Envelope<FeatureRecord>>(
      endpoints.features(productId, projectId),
      payload,
    );
    return response.data.data;
  },

  async updateFeature(
    productId: number,
    projectId: number,
    featureId: number,
    payload: {
      category_name?: string;
      category_code?: string;
      description?: string;
      parent_id?: number | null;
      is_active?: boolean;
    },
  ): Promise<FeatureRecord> {
    const response = await apiClient.patch<Envelope<FeatureRecord>>(
      `${endpoints.features(productId, projectId)}/${featureId}`,
      payload,
    );
    return response.data.data;
  },

  async deleteFeature(
    productId: number,
    projectId: number,
    featureId: number,
  ): Promise<void> {
    await apiClient.delete(
      `${endpoints.features(productId, projectId)}/${featureId}`,
    );
  },

  async validateStructure(
    productId: number,
  ): Promise<{ valid: boolean; validation_errors?: string[] }> {
    const response = await apiClient.post<
      Envelope<{ valid: boolean; validation_errors?: string[] }>
    >(`${endpoints.product(productId)}/setup/validate`);
    return response.data.data;
  },

  async createApiKey(
    productId: number,
    payload: {
      key_name: string;
      environment_id?: number;
      environment_code?: string;
      default_project_id?: number;
      default_category_id?: number;
      expiration?: string;
      description?: string;
    },
  ): Promise<GeneratedApiKeyInfo> {
    const response = await apiClient.post<Envelope<ApiKeyWire>>(
      `${endpoints.product(productId)}/api-keys`,
      { ...payload, permissions: ["LOG_INGEST_CREATE"] },
    );
    const data = response.data.data;
    return {
      key_id: data.key_id ?? data.KeyID ?? 0,
      key_name: data.key_name ?? data.KeyName ?? payload.key_name,
      raw_key:
        data.raw_key ?? data.rawKey ?? data.api_key ?? data.key ?? "",
      key_prefix: data.key_prefix ?? data.KeyPrefix ?? "",
      environment_code: payload.environment_code ?? "",
      product_code: data.product_code ?? "",
      expires_at: data.expires_at ?? data.ExpiresAt,
      permissions: ["LOG_INGEST_CREATE"],
    };
  },

  async sendTestLog(
    apiKeySecret: string,
    payload: TestLogPayload,
  ): Promise<Envelope<{ item_id?: string; batch_id?: string }>> {
    const productCode =
      typeof payload.custom_fields?.product_code === "string"
        ? payload.custom_fields.product_code
        : typeof payload.metadata.product_code === "string"
          ? payload.metadata.product_code
          : "";
    const environmentCode =
      typeof payload.custom_fields?.environment_code === "string"
        ? payload.custom_fields.environment_code
        : typeof payload.metadata.environment_code === "string"
          ? payload.metadata.environment_code
          : "";
    const queueEnvironmentCode =
      environmentCode.trim().toUpperCase() || "UNSPECIFIED";

    // สร้าง Batch Ingest Request Payload ตามสัญญา IngestLogBatchRequest ของ Backend
    const batchPayload = {
      queue_key: `${productCode || "OMNILOGS"}_${queueEnvironmentCode}_QUEUE`,
      source_type: "SETUP_WIZARD",
      source_platform: "WEB",
      logs: [
        {
          sequence_no: 1,
          source_type: "SETUP_WIZARD",
          source_platform: "WEB",
          data: payload,
        },
      ],
    };

    const response = await apiClient.post<
      Envelope<{ item_id?: string; batch_id?: string }>
    >(
      getOmniLogsIngestUrl(),
      batchPayload,
      {
        headers: {
          "X-API-Key": apiKeySecret,
          "X-Product-Code": productCode,
          "X-Environment-Code": environmentCode,
        },
      },
    );
    if (!response.data.data?.batch_id) {
      throw new Error("ระบบรับคำขอแล้ว แต่ไม่ได้สร้างหมายเลข batch สำหรับ test log");
    }
    return response.data;
  },

  async completeSetup(productId: number): Promise<unknown> {
    const response = await apiClient.post<Envelope<unknown>>(
      `${endpoints.product(productId)}/setup/complete`,
    );
    return response.data.data;
  },
};
