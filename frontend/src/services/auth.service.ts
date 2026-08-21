import { apiClient } from '@/api';
import { API_ENDPOINTS } from '@/constants';
import type { LoginRequest, LoginResponse, User, ApiResponse } from '@/types';
import type { RegisterFormData } from '@/schemas';

export const authService = {
  login: async (data: LoginRequest): Promise<LoginResponse> => {
    const response = await apiClient.post<ApiResponse<{
      access_token: string;
      refresh_token?: string;
      role: string;
      user_id: number;
    }>>(
      API_ENDPOINTS.AUTH.LOGIN,
      {
        identifier: data.username,
        password: data.password,
      }
    );

    const tokens = response.data.data;

    // ดึงข้อมูลผู้ใช้จาก /me เพื่อสร้าง User Object
    const meResponse = await apiClient.get<ApiResponse<{
      id?: number;
      user_id?: number;
      username?: string;
      first_name?: string;
      last_name?: string;
      email: string;
      phone_number?: string;
      is_active: boolean;
      role?: string;
      platform_role_name?: string;
      permissions?: string[];
    }>>(API_ENDPOINTS.AUTH.ME, {
      headers: {
        Authorization: `Bearer ${tokens.access_token}`
      }
    });

    const meData = meResponse.data.data;
    const roleName = meData.role || meData.platform_role_name || tokens.role || 'admin';
    const userPermissions = Array.isArray(meData.permissions) ? meData.permissions : [];

    // แปลงข้อมูลจากหลังบ้านเข้าสู่ Format ที่หน้าบ้านใช้งาน
    const user: User = {
      id: String(meData.id ?? meData.user_id ?? tokens.user_id ?? '1'),
      username: meData.username || meData.email,
      email: meData.email,
      fullName: `${meData.first_name || ''} ${meData.last_name || ''}`.trim() || 'User',
      phone: meData.phone_number || '',
      department: 'Platform',
      position: roleName,
      status: meData.is_active ? 'active' : 'inactive',
      roles: [
        {
          id: 'platform-role',
          name: roleName,
          permissions: userPermissions,
        }
      ],
      permissions: userPermissions,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };

    return {
      accessToken: tokens.access_token,
      refreshToken: tokens.refresh_token,
      user: user,
    };
  },

  register: async (data: RegisterFormData): Promise<void> => {
    await apiClient.post(API_ENDPOINTS.AUTH.REGISTER, {
      username: data.username,
      first_name: data.first_name,
      last_name: data.last_name,
      email: data.email,
      phone_number: data.phone_number || undefined,
      password: data.password,
      confirm_password: data.confirm_password,
    });
  },

  logout: async (): Promise<void> => {
    await apiClient.post(API_ENDPOINTS.AUTH.LOGOUT);
  },

  getMe: async (): Promise<User> => {
    const response = await apiClient.get<ApiResponse<User>>(API_ENDPOINTS.AUTH.ME);
    return response.data.data;
  },
};
