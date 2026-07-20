import { apiClient } from '@/api';
import { API_ENDPOINTS } from '@/constants';
import type { LoginRequest, LoginResponse, User, ApiResponse } from '@/types';
import type { RegisterFormData } from '@/schemas';

// ===== Mock data สำหรับ development =====
const MOCK_USER: User = {
  id: '1',
  username: 'admin',
  email: 'admin@company.com',
  fullName: 'Systemx Administrator',
  phone: '0891234567',
  department: 'IT',
  position: 'System Admin',
  status: 'active',
  roles: [
    {
      id: '1',
      name: 'super_admin',
      description: 'Super Administrator',
      permissions: [],
    },
  ],
  permissions: [
    'dashboard:view',
    'user:view',
    'user:create',
    'user:edit',
    'user:delete',
    'user:export',
    'role:view',
    'role:create',
    'role:edit',
    'role:delete',
    'settings:view',
    'settings:edit',
    'approve:user',
  ],
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
};

const USE_MOCK = !import.meta.env.VITE_API_BASE_URL;

export const authService = {
  login: async (data: LoginRequest): Promise<LoginResponse> => {
    if (USE_MOCK) {
      // จำลอง delay เหมือน API จริง
      await new Promise((resolve) => setTimeout(resolve, 1000));
      if (data.username === 'admin' && data.password === 'admin123') {
        return {
          accessToken: 'mock-access-token-' + Date.now(),
          refreshToken: 'mock-refresh-token-' + Date.now(),
          user: MOCK_USER,
        };
      }
      throw { status: 401, message: 'ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง' };
    }
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
      user_id: number;
      username: string;
      first_name: string;
      last_name: string;
      email: string;
      phone_number?: string;
      is_active: boolean;
      platform_role_name: string;
    }>>(API_ENDPOINTS.AUTH.ME, {
      headers: {
        Authorization: `Bearer ${tokens.access_token}`
      }
    });

    const meData = meResponse.data.data;

    // แปลงข้อมูลจากหลังบ้านเข้าสู่ Format ที่หน้าบ้านใช้งาน
    const user: User = {
      id: String(meData.user_id),
      username: meData.username || meData.email,
      email: meData.email,
      fullName: `${meData.first_name || ''} ${meData.last_name || ''}`.trim() || 'Admin User',
      phone: meData.phone_number || '',
      department: 'Platform',
      position: meData.platform_role_name,
      status: meData.is_active ? 'active' : 'inactive',
      roles: [
        {
          id: 'platform-role',
          name: meData.platform_role_name,
          permissions: [],
        }
      ],
      permissions: [
        'dashboard:view',
        'user:view',
        'user:create',
        'user:edit',
        'user:delete',
        'user:export',
        'role:view',
        'role:create',
        'role:edit',
        'role:delete',
        'settings:view',
        'settings:edit',
        'approve:user',
      ],
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
    if (USE_MOCK) {
      await new Promise((resolve) => setTimeout(resolve, 700));
      return;
    }
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
    if (USE_MOCK) {
      await new Promise((resolve) => setTimeout(resolve, 300));
      return;
    }
    await apiClient.post(API_ENDPOINTS.AUTH.LOGOUT);
  },

  getMe: async (): Promise<User> => {
    if (USE_MOCK) {
      await new Promise((resolve) => setTimeout(resolve, 500));
      return MOCK_USER;
    }
    const response = await apiClient.get<ApiResponse<User>>(API_ENDPOINTS.AUTH.ME);
    return response.data.data;
  },
};
