import { apiClient } from '@/api';
import type { ApiResponse } from '@/types';

export interface SensitiveAccessRequest {
  request_id: string; user_id: number; product_id: number;
  log_id?: string; secret_id?: string; field_path?: string; reason?: string;
  approval_status: 'PENDING' | 'APPROVED' | 'REJECTED';
  approved_by?: number; approved_at?: string; expires_at?: string; created_at?: string;
}

export interface SensitiveSecretOption {
  secret_id: string;
  log_id: string;
  product_id: number;
  field_key: string;
  field_path: string;
  source_section: string;
  created_at?: string;
}

export const sensitiveLogService = {
  async createRequest(payload: Record<string, unknown> & { product_id: number; reason: string }) {
    const response = await apiClient.post<ApiResponse<SensitiveAccessRequest>>(`/products/${payload.product_id}/sensitive-logs/requests`, payload);
    return response.data.data;
  },
  async listRequests(productId: number) {
    const response = await apiClient.get<ApiResponse<SensitiveAccessRequest[]>>(`/products/${productId}/sensitive-logs/requests`);
    return response.data.data || [];
  },
  async listSecrets(productId: number) {
    const response = await apiClient.get<ApiResponse<SensitiveSecretOption[]>>(`/products/${productId}/sensitive-logs/secrets`);
    return response.data.data || [];
  },
  async reviewRequest(productId: number, requestId: string, approval_status: 'APPROVED' | 'REJECTED') {
    const response = await apiClient.post<ApiResponse<SensitiveAccessRequest>>(`/products/${productId}/sensitive-logs/requests/${requestId}/review`, { approval_status });
    return response.data.data;
  },
  async listAuditRequests(auditId: string) {
    const response = await apiClient.get<ApiResponse<any[]>>(`/audit-logs/${auditId}/secrets/requests`);
    return response.data.data || [];
  },
  async listAuditLogs() {
    const response = await apiClient.get<ApiResponse<any[]>>('/audit-logs', { params: { page: 1, per_page: 100 } });
    return response.data.data || [];
  },
  async listAuditSecrets(auditId: string) {
    const response = await apiClient.get<ApiResponse<any[]>>(`/audit-logs/${auditId}/secrets`);
    return response.data.data || [];
  },
  async createAuditRequest(auditId: string, secret_id: string, reason: string) {
    const response = await apiClient.post<ApiResponse<any>>(`/audit-logs/${auditId}/secrets/requests`, { secret_id, reason });
    return response.data.data;
  },
  async listPendingAuditRequests() {
    const response = await apiClient.get<ApiResponse<any[]>>('/audit-logs/secrets/requests');
    return response.data.data || [];
  },
  async reviewAuditRequest(auditId: string, requestId: string, approval_status: 'APPROVED' | 'REJECTED') {
    const response = await apiClient.post<ApiResponse<any>>(`/audit-logs/${auditId}/secrets/requests/${requestId}/review`, { approval_status });
    return response.data.data;
  },
  async revealAuditSecret(auditId: string, request_id: string, password: string) {
    const response = await apiClient.post<ApiResponse<any>>(`/audit-logs/${auditId}/secrets/reveal`, { request_id, password });
    return response.data.data;
  },
};
