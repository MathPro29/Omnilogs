import axios from 'axios';
import type { InternalAxiosRequestConfig, AxiosResponse, AxiosError } from 'axios';
import { useAuthStore } from '@/store';
import { ROUTES } from '@/constants';

// ===== Axios Instance =====
const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// ===== Request Interceptor =====
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const { accessToken } = useAuthStore.getState();
    if (accessToken) {
      config.headers.Authorization = `Bearer ${accessToken}`;
    }
    return config;
  },
  (error: AxiosError) => {
    return Promise.reject(error);
  }
);

// ===== Response Interceptor =====
apiClient.interceptors.response.use(
  (response: AxiosResponse) => {
    return response;
  },
  (error: AxiosError) => {
    const status = error.response?.status;

    if (status === 401) {
      // Token หมดอายุหรือไม่ถูกต้อง -> ล้าง auth state แล้ว redirect ไป login
      useAuthStore.getState().clearAuth();
      window.location.href = ROUTES.LOGIN;
    }

    if (status === 403) {
      // ไม่มีสิทธิ์เข้าถึง -> redirect ไปหน้า 403
      window.location.href = ROUTES.FORBIDDEN;
    }

    // Normalize error response
    const message =
      (error.response?.data as { message?: string })?.message ||
      error.message ||
      'เกิดข้อผิดพลาดในการเชื่อมต่อ';

    return Promise.reject({
      status,
      message,
      data: error.response?.data,
    });
  }
);

export default apiClient;
