import { create } from "zustand";

interface ConnectionState { productId?: number; environmentId?: number; search: string; selectedKeyId?: number; generateOpen: boolean; setProductId: (value?: number) => void; setEnvironmentId: (value?: number) => void; setSearch: (value: string) => void; selectKey: (value?: number) => void; setGenerateOpen: (value: boolean) => void; }
export const useConnectionStore = create<ConnectionState>((set) => ({
  search: "", generateOpen: false,
  setProductId: (productId) => set({ productId, environmentId: undefined, selectedKeyId: undefined }), setEnvironmentId: (environmentId) => set({ environmentId }),
  setSearch: (search) => set({ search }), selectKey: (selectedKeyId) => set({ selectedKeyId }), setGenerateOpen: (generateOpen) => set({ generateOpen }),
}));
