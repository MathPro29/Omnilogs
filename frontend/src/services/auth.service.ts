import { apiClient } from '@/api';
import { API_ENDPOINTS } from '@/constants';
import type { LoginRequest, LoginResponse, User, ApiResponse } from '@/types';

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
    const response = await apiClient.post<ApiResponse<LoginResponse>>(
      API_ENDPOINTS.AUTH.LOGIN,
      data
    );
    return response.data.data;
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
