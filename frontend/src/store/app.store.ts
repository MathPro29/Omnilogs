import { create } from 'zustand';
import type { AppState, BreadcrumbItem } from '@/types';

export const useAppStore = create<AppState>()((set) => ({
  pageTitle: '',
  breadcrumbs: [],

  setPageTitle: (title: string) => {
    set({ pageTitle: title });
  },

  setBreadcrumbs: (items: BreadcrumbItem[]) => {
    set({ breadcrumbs: items });
  },
}));
