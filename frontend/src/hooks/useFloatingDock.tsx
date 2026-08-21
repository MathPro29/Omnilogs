import {
  useCallback,
  useEffect,
  useId,
  useMemo,
  useRef,
  useState,
} from "react";
import type { RefObject } from "react";
import { useReducedMotion } from "motion/react";
import { LogOut, User as UserIcon } from "lucide-react";
import { useNavigate } from "react-router-dom";
import type { MenuProps } from "antd";
import { useAuthStore } from "@/store";
import { authService } from "@/services/auth.service";
import type { NavigationItem } from "@/components/global/navigation";

export function useFloatingDock(
  containerRef: RefObject<HTMLDivElement | null>,
  onNavigate?: (item: NavigationItem) => void,
) {
  const [isOpen, setIsOpen] = useState(false);
  const [isExpanded, setIsExpanded] = useState(false);
  const collapseTimeoutRef = useRef<number | null>(null);
  const menuId = useId();
  const shouldReduceMotion = useReducedMotion();
  const currentUser = useAuthStore((state) => state.currentUser);
  const clearAuth = useAuthStore((state) => state.clearAuth);
  const navigate = useNavigate();

  const clearCollapseTimer = useCallback(() => {
    if (collapseTimeoutRef.current !== null) {
      window.clearTimeout(collapseTimeoutRef.current);
      collapseTimeoutRef.current = null;
    }
  }, []);
  const closeDock = useCallback(() => {
    clearCollapseTimer();
    setIsOpen(false);
    if (shouldReduceMotion) {
      setIsExpanded(false);
      return;
    }
    collapseTimeoutRef.current = window.setTimeout(() => {
      setIsExpanded(false);
      collapseTimeoutRef.current = null;
    }, 630);
  }, [clearCollapseTimer, shouldReduceMotion]);
  const openDock = useCallback(() => {
    clearCollapseTimer();
    setIsExpanded(true);
    setIsOpen(true);
  }, [clearCollapseTimer]);
  const toggleDock = useCallback(() => {
    if (isOpen) closeDock();
    else openDock();
  }, [closeDock, isOpen, openDock]);
  const handleNavigate = useCallback(
    (item: NavigationItem) => {
      closeDock();
      if (onNavigate) onNavigate(item);
      else navigate(item.href);
    },
    [closeDock, navigate, onNavigate],
  );
  const navigateTo = useCallback(
    (path: string) => {
      closeDock();
      navigate(path);
    },
    [closeDock, navigate],
  );
  const handleLogout = useCallback(() => {
    void authService.logout().finally(() => {
      clearAuth();

      navigate("/login");
    });
  }, [clearAuth, navigate]);
  const userMenuItems = useMemo<MenuProps["items"]>(
    () => [
      {
        key: "profile",
        icon: <UserIcon size={16} />,
        label: "Profile",
      },
      { type: "divider" },
      {
        key: "logout",
        icon: (
          <LogOut
            size={16}
            className="text-red-500 transition-colors duration-300 [.ant-dropdown-menu-item:hover_&]:text-white [.ant-dropdown-menu-item-danger:hover_&]:text-white"
          />
        ),
        label: <span className="transition-colors duration-300">Log out</span>,
        danger: true,
        onClick: handleLogout,
      },
    ],
    [handleLogout],
  );

  useEffect(() => {
    const handlePointerDown = (event: PointerEvent) => {
      if (
        containerRef.current &&
        !containerRef.current.contains(event.target as Node)
      ) {
        closeDock();
      }
    };
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") closeDock();
    };
    document.addEventListener("pointerdown", handlePointerDown);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("pointerdown", handlePointerDown);
      document.removeEventListener("keydown", handleKeyDown);
      clearCollapseTimer();
    };
  }, [clearCollapseTimer, closeDock, containerRef]);

  return {
    isOpen,
    isExpanded,
    menuId,
    shouldReduceMotion,
    currentUser,
    userMenuItems,
    closeDock,
    openDock,
    toggleDock,
    handleNavigate,
    navigateTo,
  };
}

export type FloatingDockController = ReturnType<typeof useFloatingDock>;
