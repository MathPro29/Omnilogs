import { useAuthStore } from "@/store";

export function useSetupPermissions() {
  const store = useAuthStore();
  const roles = store.roles ?? [];
  const currentUser = store.currentUser;
  
  // รวบรวม role จากทั้ง store.roles และ currentUser.roles / position
  const roleStrings: string[] = [];
  
  if (Array.isArray(roles)) {
    roles.forEach((r) => {
      if (typeof r === "string" && r) roleStrings.push(r);
    });
  }
  
  if (currentUser?.roles && Array.isArray(currentUser.roles)) {
    currentUser.roles.forEach((r) => {
      if (r?.name) roleStrings.push(r.name);
    });
  }

  if (currentUser?.position) {
    roleStrings.push(currentUser.position);
  }

  // ดึง role จาก JWT accessToken หากใน store เป็น null/ว่างเปล่า
  if (store.accessToken) {
    try {
      const parts = store.accessToken.split(".");
      if (parts.length === 3) {
        const payload = JSON.parse(atob(parts[1]));
        if (payload?.role) {
          roleStrings.push(payload.role);
        }
      }
    } catch {
      // Ignore parse error
    }
  }

  const normalizedRoles = roleStrings
    .map((role) => role.trim().toLowerCase().replace(/\s+/g, "_"))
    .filter(Boolean);

  const isPlatformAdmin = normalizedRoles.some((role) =>
    ["god", "owner", "superadmin", "super_admin", "admin"].includes(role),
  );

  const canCreateProduct = isPlatformAdmin || normalizedRoles.includes("owner");
  const canCreateApiKey = canCreateProduct;

  return {
    isPlatformAdmin,
    canCreateProduct,
    canCreateApiKey,
  };
}

