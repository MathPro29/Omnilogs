import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { AuthState, LoginResponse } from "@/types";

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      currentUser: null,
      roles: [],
      permissions: [],
      isAuthenticated: false,
      accessToken: null,

      setAuth: (data: LoginResponse) => {
        set({
          currentUser: data.user,
          roles: (data.user.roles ?? [])
            .map((r) => r?.name ? r.name.trim().toLowerCase().replace(/\s+/g, "_") : "")
            .filter(Boolean),
          permissions: data.user.permissions ?? [],
          isAuthenticated: true,
          accessToken: data.accessToken,
        });
      },

      clearAuth: () => {
        set({
          currentUser: null,
          roles: [],
          permissions: [],
          isAuthenticated: false,
          accessToken: null,
        });
      },

      updatePermissions: (permissions: string[]) => {
        set({ permissions });
      },
    }),
    {
      name: "hr-auth-storage",
      partialize: (state) => ({
        currentUser: state.currentUser,
        roles: state.roles,
        permissions: state.permissions,
        isAuthenticated: state.isAuthenticated,
        accessToken: state.accessToken,
      }),
    },
  ),
);
