import { apiClient } from '@/api';
import { API_ENDPOINTS } from '@/constants';

export type AuditResult = 'SUCCESS' | 'FAILED' | 'DENIED' | string;

export interface AuditLog {
  audit_id: string;
  actor_user_id?: number;
  product_id?: number;
  action: string;
  resource_type: string;
  resource_id?: string;
  request_id?: string;
  trace_id?: string;
  method?: string;
  path?: string;
  result: AuditResult;
  metadata: Record<string, unknown> | null;
  ip_address?: string;
  user_agent?: string;
  created_at?: string;
}

export interface AuditLogFilters {
  page: number;
  per_page: number;
  keyword?: string;
  action?: string;
  resource_type?: string;
  result?: string;
  date_from?: string;
  date_to?: string;
}

interface AuditListResponse {
  data: AuditLog[];
  meta?: { page: number; perPage: number; total: number; totalPage: number };
}

interface AuditDetailResponse { data: AuditLog }

export const auditService = {
  list: async (filters: AuditLogFilters) => {
    const response = await apiClient.get<AuditListResponse>(API_ENDPOINTS.AUDIT_LOGS.LIST, { params: filters });
    return { data: response.data.data || [], total: response.data.meta?.total || 0 };
  },
  getById: async (auditId: string) => {
    const response = await apiClient.get<AuditDetailResponse>(API_ENDPOINTS.AUDIT_LOGS.DETAIL(auditId));
    return response.data.data;
  },
};