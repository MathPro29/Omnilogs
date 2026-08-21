import type { DashboardSummary, ApiResponse } from '@/types';
import { apiClient } from '@/api';
import { API_ENDPOINTS } from '@/constants';

export const dashboardService = {
  getSummary: async (): Promise<DashboardSummary> => {
    const response = await apiClient.get<ApiResponse<DashboardSummary>>(
      API_ENDPOINTS.DASHBOARD.SUMMARY
    );
    return response.data.data;
  },
};
