import { create } from "zustand";
import { useAuthStore } from "@/store/auth.store";

// แสดง Data ใน Log Explorer
export type BuiltInLogColumn = "timestamp" | "level" | "product" | "hierarchy" | "message";
export type FavoriteLogColumn = `favorite:${number}`;
export type LogColumn = BuiltInLogColumn | FavoriteLogColumn;

export const isFavoriteLogColumn = (column: LogColumn): column is FavoriteLogColumn => column.startsWith("favorite:");
export const favoriteLogColumnKey = (fieldDefinitionId: number): FavoriteLogColumn => `favorite:${fieldDefinitionId}`;
export const favoriteFieldDefinitionId = (column: FavoriteLogColumn) => Number(column.slice("favorite:".length));

export const DEFAULT_LOG_COLUMNS: LogColumn[] = [
  "timestamp",
  "level",
  "product",
  "hierarchy",
  "message",
];

const STORAGE_KEY_PREFIX = "omnilogs_displayed_columns_";

export function getSavedUserColumns(userId: string | undefined): LogColumn[] {
  if (!userId) return DEFAULT_LOG_COLUMNS;
  try {
    const raw = localStorage.getItem(`${STORAGE_KEY_PREFIX}${userId}`);
    if (!raw) return DEFAULT_LOG_COLUMNS;
    const parsed = JSON.parse(raw);
    if (Array.isArray(parsed) && parsed.length > 0 && parsed.every((item) => typeof item === "string")) {
      return parsed as LogColumn[];
    }
  } catch {
    // Ignore parse error
  }
  return DEFAULT_LOG_COLUMNS;
}

export function saveUserColumns(userId: string | undefined, columns: LogColumn[]): void {
  if (!userId) return;
  try {
    localStorage.setItem(`${STORAGE_KEY_PREFIX}${userId}`, JSON.stringify(columns));
  } catch {
    // Ignore storage write error
  }
}

interface LogExplorerState {
  selectedId: string | null; selectedIds: string[]; filtersOpen: boolean; inspectorOpen: boolean;
  inspectorFullscreen: boolean; inspectorWidth: number; columns: LogColumn[];
  select: (id: string | null) => void; toggleSelected: (id: string) => void; clearSelected: () => void;
  setFiltersOpen: (open: boolean) => void; setInspectorOpen: (open: boolean) => void;
  setInspectorFullscreen: (open: boolean) => void; setInspectorWidth: (width: number) => void;
  setColumns: (columns: LogColumn[]) => void;
}

export const useLogExplorerStore = create<LogExplorerState>((set) => ({
  selectedId: null, selectedIds: [], filtersOpen: true, inspectorOpen: false, inspectorFullscreen: false, inspectorWidth: 440,
  columns: DEFAULT_LOG_COLUMNS,
  select: (selectedId) => set({ selectedId }),
  toggleSelected: (id) => set((state) => ({ selectedIds: state.selectedIds.includes(id) ? state.selectedIds.filter((item) => item !== id) : [...state.selectedIds, id] })),
  clearSelected: () => set({ selectedIds: [] }), setFiltersOpen: (filtersOpen) => set({ filtersOpen }),
  setInspectorOpen: (inspectorOpen) => set({ inspectorOpen }), setInspectorFullscreen: (inspectorFullscreen) => set({ inspectorFullscreen }),
  setInspectorWidth: (inspectorWidth) => set({ inspectorWidth }),
  setColumns: (columns) => {
    set({ columns });
    const userId = useAuthStore.getState().currentUser?.id;
    saveUserColumns(userId, columns);
  },
}));

const FAVORITES_KEY_PREFIX = "omnilogs_user_favorites_";

export function getUserFavoriteIds(userId: string | undefined, productId: number | undefined): number[] | null {
  if (!userId || !productId) return null;
  try {
    const raw = localStorage.getItem(`${FAVORITES_KEY_PREFIX}${userId}_${productId}`);
    if (raw !== null) {
      const parsed = JSON.parse(raw);
      if (Array.isArray(parsed)) return parsed.map(Number).filter((n) => !isNaN(n));
    }
  } catch {
    // Ignore error
  }
  return null;
}

export function saveUserFavoriteIds(userId: string | undefined, productId: number | undefined, ids: number[]): void {
  if (!userId || !productId) return;
  try {
    localStorage.setItem(`${FAVORITES_KEY_PREFIX}${userId}_${productId}`, JSON.stringify(ids));
  } catch {
    // Ignore error
  }
}

export function addUserFavoriteId(userId: string | undefined, productId: number | undefined, id: number): void {
  if (!userId || !productId) return;
  const current = getUserFavoriteIds(userId, productId) ?? [];
  if (!current.includes(id)) {
    saveUserFavoriteIds(userId, productId, [...current, id]);
  }
}

export function removeUserFavoriteId(userId: string | undefined, productId: number | undefined, id: number): void {
  if (!userId || !productId) return;
  const current = getUserFavoriteIds(userId, productId) ?? [];
  saveUserFavoriteIds(userId, productId, current.filter((item) => item !== id));
}


