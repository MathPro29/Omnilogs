import axios from 'axios';
import { apiClient } from '@/api';
import { ALL_PERMISSIONS, API_ENDPOINTS } from '@/constants';
import type { ApiResponse, LoginRequest, LoginResponse, User } from '@/types';

interface BackendAuthResponse {
  access_token: string;
  refresh_token?: string;
  role?: string;
  user_id?: number;
}

interface BackendUserResponse {
  id?: number;
  user_id?: number;
  username?: string | null;
  first_name?: string | null;
  last_name?: string | null;
  email: string;
  phone_number?: string | null;
  is_active?: boolean;
  role?: string;
  created_at?: string;
  updated_at?: string;
}

const MOCK_USER: User = {
  id: '1',
  username: 'admin',
  email: 'admin@company.com',
  fullName: 'System Administrator',
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
  permissions: ALL_PERMISSIONS,
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
};

const USE_MOCK = !import.meta.env.VITE_API_BASE_URL;

function derivePermissions(role?: string): string[] {
  const normalizedRole = (role || '').toLowerCase();
  if (['god', 'owner', 'superadmin', 'admin', 'super_admin'].includes(normalizedRole)) {
    return ALL_PERMISSIONS;
  }
  return [];
}

function mapBackendUser(user: BackendUserResponse, fallbackRole?: string): User {
  const roleName = user.role || fallbackRole || 'user';
  const fullName = [user.first_name, user.last_name].filter(Boolean).join(' ').trim() || user.username || user.email;

  return {
    id: String(user.user_id || user.id || ''),
    username: user.username || user.email.split('@')[0],
    email: user.email,
    fullName,
    phone: user.phone_number || undefined,
    status: user.is_active === false ? 'inactive' : 'active',
    roles: [{ id: roleName, name: roleName, permissions: [] }],
    permissions: derivePermissions(roleName),
    createdAt: user.created_at || new Date().toISOString(),
    updatedAt: user.updated_at || new Date().toISOString(),
  };
}

async function fetchCurrentUser(accessToken: string, fallbackRole?: string): Promise<User> {
  const baseURL = import.meta.env.VITE_API_BASE_URL;
  const client = axios.create({
    baseURL,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'Content-Type': 'application/json',
    },
  });

  const response = await client.get<ApiResponse<BackendUserResponse>>(API_ENDPOINTS.AUTH.ME);
  return mapBackendUser(response.data.data, fallbackRole);
}

export const authService = {
  login: async (data: LoginRequest): Promise<LoginResponse> => {
    if (USE_MOCK) {
      await new Promise((resolve) => setTimeout(resolve, 1000));
      if (data.identifier === 'admin' || data.identifier === 'admin@company.com') {
        if (data.password === 'admin123') {
          return {
            accessToken: 'mock-access-token-' + Date.now(),
            refreshToken: 'mock-refresh-token-' + Date.now(),
            user: MOCK_USER,
          };
        }
      }
      throw { status: 401, message: 'ชื่อผู้ใช้/อีเมลหรือรหัสผ่านไม่ถูกต้อง' };
    }

    const response = await apiClient.post<ApiResponse<BackendAuthResponse>>(
      API_ENDPOINTS.AUTH.LOGIN,
      data
    );

    const authData = response.data.data;
    const user = await fetchCurrentUser(authData.access_token, authData.role);

    return {
      accessToken: authData.access_token,
      refreshToken: authData.refresh_token,
      user,
    };
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
    const response = await apiClient.get<ApiResponse<BackendUserResponse>>(API_ENDPOINTS.AUTH.ME);
    return mapBackendUser(response.data.data);
  },
};
