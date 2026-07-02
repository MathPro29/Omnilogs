import { apiClient } from '@/api';
import { API_ENDPOINTS } from '@/constants';
import type { User, PaginatedResponse, UserFilterParams, ApiResponse } from '@/types';

// ===== Mock Users =====
const MOCK_USERS: User[] = [
  {
    id: '1',
    username: 'admin',
    email: 'admin@company.com',
    fullName: 'System Administrator',
    phone: '0891234567',
    department: 'IT',
    position: 'System Admin',
    status: 'active',
    roles: [{ id: '1', name: 'super_admin', description: 'Super Admin', permissions: [] }],
    permissions: [],
    createdAt: '2024-01-15T08:00:00Z',
    updatedAt: '2024-06-20T10:30:00Z',
  },
  {
    id: '2',
    username: 'somchai',
    email: 'somchai@company.com',
    fullName: 'สมชาย ใจดี',
    phone: '0812345678',
    department: 'HR',
    position: 'HR Manager',
    status: 'active',
    roles: [{ id: '2', name: 'admin', description: 'Admin', permissions: [] }],
    permissions: [],
    createdAt: '2024-02-01T09:00:00Z',
    updatedAt: '2024-07-15T14:20:00Z',
  },
  {
    id: '3',
    username: 'siriporn',
    email: 'siriporn@company.com',
    fullName: 'ศิริพร มั่นคง',
    phone: '0823456789',
    department: 'Finance',
    position: 'Accountant',
    status: 'active',
    roles: [{ id: '3', name: 'manager', description: 'Manager', permissions: [] }],
    permissions: [],
    createdAt: '2024-03-10T10:00:00Z',
    updatedAt: '2024-08-01T11:45:00Z',
  },
  {
    id: '4',
    username: 'wichai',
    email: 'wichai@company.com',
    fullName: 'วิชัย สุขสันต์',
    phone: '0834567890',
    department: 'Sales',
    position: 'Sales Executive',
    status: 'inactive',
    roles: [{ id: '4', name: 'viewer', description: 'Viewer', permissions: [] }],
    permissions: [],
    createdAt: '2024-04-05T08:30:00Z',
    updatedAt: '2024-09-10T16:00:00Z',
  },
  {
    id: '5',
    username: 'nattaya',
    email: 'nattaya@company.com',
    fullName: 'ณัฐยา เจริญสุข',
    phone: '0845678901',
    department: 'Marketing',
    position: 'Marketing Specialist',
    status: 'active',
    roles: [{ id: '3', name: 'manager', description: 'Manager', permissions: [] }],
    permissions: [],
    createdAt: '2024-05-20T09:15:00Z',
    updatedAt: '2024-10-05T13:30:00Z',
  },
  {
    id: '6',
    username: 'prasit',
    email: 'prasit@company.com',
    fullName: 'ประสิทธิ์ รุ่งเรือง',
    phone: '0856789012',
    department: 'IT',
    position: 'Developer',
    status: 'active',
    roles: [{ id: '4', name: 'viewer', description: 'Viewer', permissions: [] }],
    permissions: [],
    createdAt: '2024-06-01T08:00:00Z',
    updatedAt: '2024-11-12T09:00:00Z',
  },
  {
    id: '7',
    username: 'kannika',
    email: 'kannika@company.com',
    fullName: 'กรรณิการ์ สว่างจิต',
    phone: '0867890123',
    department: 'HR',
    position: 'HR Specialist',
    status: 'pending',
    roles: [{ id: '4', name: 'viewer', description: 'Viewer', permissions: [] }],
    permissions: [],
    createdAt: '2024-07-15T10:00:00Z',
    updatedAt: '2024-12-01T15:20:00Z',
  },
  {
    id: '8',
    username: 'surasak',
    email: 'surasak@company.com',
    fullName: 'สุรศักดิ์ ทองดี',
    phone: '0878901234',
    department: 'Operations',
    position: 'Operations Manager',
    status: 'active',
    roles: [{ id: '2', name: 'admin', description: 'Admin', permissions: [] }],
    permissions: [],
    createdAt: '2024-08-01T08:30:00Z',
    updatedAt: '2025-01-10T10:00:00Z',
  },
];

const USE_MOCK = import.meta.env.VITE_USE_MOCK === 'true';

interface BackendAdminUser {
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

function mapBackendUser(user: BackendAdminUser): User {
  const fullName = [user.first_name, user.last_name].filter(Boolean).join(' ').trim() || user.username || user.email;
  const roleName = user.role || 'user';

  return {
    id: String(user.user_id || user.id || ''),
    username: user.username || user.email.split('@')[0],
    email: user.email,
    fullName,
    phone: user.phone_number || undefined,
    position: roleName,
    status: user.is_active === false ? 'inactive' : 'active',
    roles: [{ id: roleName, name: roleName, permissions: [] }],
    permissions: [],
    createdAt: user.created_at || new Date().toISOString(),
    updatedAt: user.updated_at || new Date().toISOString(),
  };
}

export const userService = {
  getUsers: async (params: UserFilterParams): Promise<PaginatedResponse<User>> => {
    if (USE_MOCK) {
      await new Promise((resolve) => setTimeout(resolve, 800));

      let filtered = [...MOCK_USERS];

      // Apply search filter
      if (params.search) {
        const search = params.search.toLowerCase();
        filtered = filtered.filter(
          (u) =>
            u.fullName.toLowerCase().includes(search) ||
            u.username.toLowerCase().includes(search) ||
            u.email.toLowerCase().includes(search)
        );
      }

      // Apply status filter
      if (params.status) {
        filtered = filtered.filter((u) => u.status === params.status);
      }

      // Apply department filter
      if (params.department) {
        filtered = filtered.filter((u) => u.department === params.department);
      }

      const total = filtered.length;
      const start = (params.page - 1) * params.pageSize;
      const data = filtered.slice(start, start + params.pageSize);

      return {
        data,
        total,
        page: params.page,
        pageSize: params.pageSize,
        totalPages: Math.ceil(total / params.pageSize),
      };
    }

    const response = await apiClient.get<ApiResponse<BackendAdminUser[]>>(API_ENDPOINTS.ADMIN.USERS);
    const users = response.data.data.map(mapBackendUser);

    let filtered = [...users];
    if (params.search) {
      const search = params.search.toLowerCase();
      filtered = filtered.filter(
        (u) =>
          u.fullName.toLowerCase().includes(search) ||
          u.username.toLowerCase().includes(search) ||
          u.email.toLowerCase().includes(search)
      );
    }
    if (params.status) {
      filtered = filtered.filter((u) => u.status === params.status);
    }
    if (params.department) {
      filtered = filtered.filter((u) => u.department === params.department);
    }

    const total = filtered.length;
    const start = (params.page - 1) * params.pageSize;
    const data = filtered.slice(start, start + params.pageSize);

    return {
      data,
      total,
      page: params.page,
      pageSize: params.pageSize,
      totalPages: Math.max(1, Math.ceil(total / params.pageSize)),
    };
  },

  getUserById: async (id: string): Promise<User> => {
    if (USE_MOCK) {
      await new Promise((resolve) => setTimeout(resolve, 500));
      const user = MOCK_USERS.find((u) => u.id === id);
      if (!user) throw { status: 404, message: 'ไม่พบผู้ใช้งาน' };
      return user;
    }

    const response = await apiClient.get<ApiResponse<User>>(API_ENDPOINTS.USERS.DETAIL(id));
    return response.data.data;
  },

  createUser: async (data: Partial<User>): Promise<User> => {
    if (USE_MOCK) {
      await new Promise((resolve) => setTimeout(resolve, 800));
      return {
        id: String(MOCK_USERS.length + 1),
        username: data.username || '',
        email: data.email || '',
        fullName: data.fullName || '',
        status: 'active',
        roles: [],
        permissions: [],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        ...data,
      } as User;
    }

    const response = await apiClient.post<ApiResponse<User>>(API_ENDPOINTS.USERS.CREATE, data);
    return response.data.data;
  },

  updateUser: async (id: string, data: Partial<User>): Promise<User> => {
    if (USE_MOCK) {
      await new Promise((resolve) => setTimeout(resolve, 800));
      const user = MOCK_USERS.find((u) => u.id === id);
      if (!user) throw { status: 404, message: 'ไม่พบผู้ใช้งาน' };
      return { ...user, ...data, updatedAt: new Date().toISOString() };
    }

    const response = await apiClient.put<ApiResponse<User>>(API_ENDPOINTS.USERS.UPDATE(id), data);
    return response.data.data;
  },

  deleteUser: async (id: string): Promise<void> => {
    if (USE_MOCK) {
      await new Promise((resolve) => setTimeout(resolve, 500));
      return;
    }

    await apiClient.delete(API_ENDPOINTS.USERS.DELETE(id));
  },
};
