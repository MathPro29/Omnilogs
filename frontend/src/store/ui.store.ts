import { create } from 'zustand';
import type { UIState } from '@/types';

export const useUIStore = create<UIState>()((set) => ({
  sidebarCollapsed: false,

  toggleSidebar: () => {
    set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed }));
  },

  setSidebarCollapsed: (collapsed: boolean) => {
    set({ sidebarCollapsed: collapsed });
  },
}));
