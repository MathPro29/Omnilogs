import { apiClient } from '@/api';
import { API_ENDPOINTS } from '@/constants';
import type { User, PaginatedResponse, UserFilterParams, ApiResponse } from '@/types';
import { normalizeUsersResponse } from '@/utils/userNormalizer';

type UserMutationInput = Partial<User> & { password?: string; role?: string };

export const userService = {
  getUsers: async (params: UserFilterParams): Promise<PaginatedResponse<User>> => {
    const response = await apiClient.get<unknown>(API_ENDPOINTS.USERS.LIST, {
      params,
    });
    const data = normalizeUsersResponse(response.data);
    return {
      data,
      total: data.length,
      page: params.page || 1,
      pageSize: params.pageSize || 100,
      totalPages: 1,
    };
  },

  getUserById: async (id: string): Promise<User> => {
    const response = await apiClient.get<ApiResponse<unknown>>(API_ENDPOINTS.USERS.DETAIL(id));
    const user = normalizeUsersResponse([response.data.data])[0];
    if (!user) throw new Error("ไม่พบผู้ใช้งาน");
    return user;
  },

  createUser: async (data: UserMutationInput): Promise<User> => {
    const response = await apiClient.post<ApiResponse<User>>(API_ENDPOINTS.USERS.CREATE, data);
    return response.data.data;
  },

  updateUser: async (id: string, data: UserMutationInput): Promise<User> => {
    const response = await apiClient.put<ApiResponse<User>>(API_ENDPOINTS.USERS.UPDATE(id), data);
    return response.data.data;
  },

  deleteUser: async (id: string): Promise<void> => {
    await apiClient.delete(API_ENDPOINTS.USERS.DELETE(id));
  },
};
