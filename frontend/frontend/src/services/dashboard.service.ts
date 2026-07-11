import type { DashboardSummary, ApiResponse } from '@/types';
import { apiClient } from '@/api';
import { API_ENDPOINTS } from '@/constants';

const MOCK_SUMMARY: DashboardSummary = {
  totalUsers: 156,
  activeUsers: 142,
  newUsersThisMonth: 12,
  departments: 8,
  recentActivities: [
    {
      id: '1',
      action: 'เข้าสู่ระบบ',
      user: 'สมชาย ใจดี',
      target: 'ระบบ',
      timestamp: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
    },
    {
      id: '2',
      action: 'แก้ไขข้อมูลผู้ใช้',
      user: 'ศิริพร มั่นคง',
      target: 'วิชัย สุขสันต์',
      timestamp: new Date(Date.now() - 15 * 60 * 1000).toISOString(),
    },
    {
      id: '3',
      action: 'สร้างผู้ใช้ใหม่',
      user: 'System Administrator',
      target: 'ณัฐยา เจริญสุข',
      timestamp: new Date(Date.now() - 30 * 60 * 1000).toISOString(),
    },
    {
      id: '4',
      action: 'เปลี่ยนสถานะ',
      user: 'สมชาย ใจดี',
      target: 'ประสิทธิ์ รุ่งเรือง',
      timestamp: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
    },
    {
      id: '5',
      action: 'อัปเดตการตั้งค่า',
      user: 'System Administrator',
      target: 'ตั้งค่าระบบ',
      timestamp: new Date(Date.now() - 5 * 60 * 60 * 1000).toISOString(),
    },
  ],
  usersByDepartment: [
    { label: 'IT', value: 35 },
    { label: 'HR', value: 22 },
    { label: 'Finance', value: 28 },
    { label: 'Sales', value: 30 },
    { label: 'Marketing', value: 18 },
    { label: 'Operations', value: 23 },
  ],
  userGrowth: [
    { label: 'ม.ค.', value: 120 },
    { label: 'ก.พ.', value: 128 },
    { label: 'มี.ค.', value: 132 },
    { label: 'เม.ย.', value: 138 },
    { label: 'พ.ค.', value: 145 },
    { label: 'มิ.ย.', value: 156 },
  ],
};

const USE_MOCK = !import.meta.env.VITE_API_BASE_URL;

export const dashboardService = {
  getSummary: async (): Promise<DashboardSummary> => {
    if (USE_MOCK) {
      await new Promise((resolve) => setTimeout(resolve, 1200));
      return MOCK_SUMMARY;
    }

    const response = await apiClient.get<ApiResponse<DashboardSummary>>(
      API_ENDPOINTS.DASHBOARD.SUMMARY
    );
    return response.data.data;
  },
  getAuditLogs: async (params?: { product_id?: number; project_id?: number; feature_id?: number; limit?: number; offset?: number }): Promise<{ total: number; limit: number; offset: number; data: any[] }> => {
    const response = await apiClient.get<ApiResponse<{ total: number; limit: number; offset: number; data: any[] }>>(
      '/dashboard/audit-logs',
      { params }
    );
    return response.data.data;
  },
};
