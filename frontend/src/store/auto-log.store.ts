import { create } from 'zustand';
import type { ProjectFeature } from '@/types';

export interface AutoLogConfig {
  productId: number;
  environmentId: number;
  projectId?: number;
  categoryId?: number;
  featureFullPath?: string | null;
  featurePathIds?: string | null;
  eventType: string;
  features?: ProjectFeature[];
}

interface AutoLogState {
  isRunning: boolean;
  config: AutoLogConfig | null;
  start: (config: AutoLogConfig) => void;
  stop: () => void;
}

// เก็บสถานะไว้นอกหน้า Dashboard เพื่อให้การเปลี่ยน Route ไม่ทำให้ Process หยุด
export const useAutoLogStore = create<AutoLogState>((set) => ({
  isRunning: false,
  config: null,
  start: (config) => set({ isRunning: true, config }),
  stop: () => set({ isRunning: false, config: null }),
}));
