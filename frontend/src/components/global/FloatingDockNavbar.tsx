import { AnimatePresence, motion, useReducedMotion } from "motion/react";
import { ChevronDown, Menu, X, User as UserIcon, LogOut } from "lucide-react";
import { useEffect, useId, useRef, useState } from "react";
import { useAuthStore } from "@/store";
import { useNavigate } from "react-router-dom";
import { Dropdown, Avatar } from "antd";
import type { MenuProps } from "antd";
import { getInitials } from "@/utils";

import { navigationGroups, type NavigationItem } from "./navigation";

type FloatingDockNavbarProps = {
  activePath?: string;
  onNavigate?: (item: NavigationItem) => void;
};

const menuTransition = {
  type: "spring" as const,
  stiffness: 380,
  damping: 34,
  mass: 0.8,
};

export function FloatingDockNavbar({
  activePath = "/dashboard",
  onNavigate,
}: FloatingDockNavbarProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [isExpanded, setIsExpanded] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const collapseTimeoutRef = useRef<number | null>(null);
  const menuId = useId();
  const shouldReduceMotion = useReducedMotion();

  const currentUser = useAuthStore((state) => state.currentUser);
  const clearAuth = useAuthStore((state) => state.clearAuth);
  const navigate = useNavigate();

  const handleLogout = () => {
    clearAuth();
    navigate("/login");
  };

  const userMenuItems: MenuProps["items"] = [
    {
      key: "profile",
      icon: <UserIcon size={16} />,
      label: "โปรไฟล์",
    },
    { type: "divider" },
    {
      key: "logout",
      icon: <LogOut size={16} className="text-red-500" />,
      label: "ออกจากระบบ",
      danger: true,
      onClick: handleLogout,
    },
  ];

  function clearCollapseTimer() {
    if (collapseTimeoutRef.current !== null) {
      window.clearTimeout(collapseTimeoutRef.current);
      collapseTimeoutRef.current = null;
    }
  }

  function openDock() {
    clearCollapseTimer();
    setIsExpanded(true);
    setIsOpen(true);
  }

  function closeDock() {
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
  }

  function toggleDock() {
    if (isOpen) {
      closeDock();
      return;
    }

    openDock();
  }

  useEffect(() => {
    function handlePointerDown(event: PointerEvent) {
      if (
        containerRef.current &&
        !containerRef.current.contains(event.target as Node)
      ) {
        closeDock();
      }
    }

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        closeDock();
      }
    }

    document.addEventListener("pointerdown", handlePointerDown);
    document.addEventListener("keydown", handleKeyDown);

    return () => {
      document.removeEventListener("pointerdown", handlePointerDown);
      document.removeEventListener("keydown", handleKeyDown);
      clearCollapseTimer();
    };
  }, [shouldReduceMotion]);

  function handleNavigate(item: NavigationItem) {
    closeDock();

    if (onNavigate) {
      onNavigate(item);
      return;
    }

    window.location.href = item.href;
  }

  return (
    <>
      <AnimatePresence>
        {isOpen && (
          <motion.button
            type="button"
            aria-label="Close navigation menu"
            className="dock-backdrop"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: shouldReduceMotion ? 0 : 0.2 }}
            onClick={() => setIsOpen(false)}
          />
        )}
      </AnimatePresence>

      <div ref={containerRef} className="nav-dock-frame">
        <motion.nav
          layout
          aria-label="Primary navigation"
          className={`nav-dock ${isExpanded ? "nav-dock--open" : ""}`}
          transition={shouldReduceMotion ? { duration: 0 } : menuTransition}
        >
          <motion.div layout="position" className="nav-dock__bar">
            <button
              type="button"
              className="nav-dock__brand"
              onClick={toggleDock}
              aria-expanded={isOpen}
              aria-controls={menuId}
            >
              <span className="nav-dock__logo">O</span>
              <span
                className={`nav-dock__brand-copy ${
                  isExpanded ? "nav-dock__brand-copy--open" : ""
                }`}
              >
                <span className="nav-dock__brand-title">OmniLogs</span>
              </span>
            </button>

            <div className="nav-dock__actions" style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
               {/* li */}
                <a href="/">
                  Home
                </a>
                <a href="/logs">
                  Explore Logs
                </a>
                <a href="">
                  Page A
                </a>
                <a href="">
                  Page B
                </a>
                
              <Dropdown menu={{ items: userMenuItems }} placement="bottomRight" trigger={["click"]}>
                <div style={{ cursor: "pointer", display: "flex", alignItems: "center" }}>
                  <Avatar
                    size={36}
                    style={{
                      backgroundColor: "var(--color-primary)",
                      fontWeight: 600,
                      fontSize: "0.875rem",
                    }}
                  >
                    {currentUser ? getInitials(currentUser.fullName) : "U"}
                  </Avatar>
                </div>
              </Dropdown>

              <button
                type="button"
                onClick={toggleDock}
                className="nav-dock__toggle"
                aria-label={isOpen ? "Close menu" : "Open menu"}
                aria-expanded={isOpen}
                aria-controls={menuId}
              >
                <AnimatePresence mode="wait" initial={false}>
                  <motion.span
                    key={isOpen ? "close" : "open"}
                    initial={
                      shouldReduceMotion
                        ? false
                        : { opacity: 0, rotate: -45, scale: 0.7 }
                    }
                    animate={{ opacity: 1, rotate: 0, scale: 1 }}
                    exit={
                      shouldReduceMotion
                        ? undefined
                        : { opacity: 0, rotate: 45, scale: 0.7 }
                    }
                    transition={{ duration: 0.15 }}
                  >
                    {isOpen ? <X size={20} /> : <Menu size={20} />}
                  </motion.span>
                </AnimatePresence>
              </button>
            </div>
          </motion.div>

          <AnimatePresence initial={false}>
            {isOpen && (
              <motion.div
                id={menuId}
                className="nav-dock__menu"
                initial={
                  shouldReduceMotion ? false : { opacity: 0, height: 0, y: -12 }
                }
                animate={{ opacity: 1, height: "auto", y: 0 }}
                exit={
                  shouldReduceMotion
                    ? undefined
                    : { opacity: 0, height: 0, y: -8 }
                }
                transition={
                  shouldReduceMotion
                    ? { duration: 0 }
                    : {
                        height: menuTransition,
                        opacity: { duration: 0.2 },
                        y: menuTransition,
                      }
                }
              >
                <div 
                  className="nav-dock__grid"
                  style={{
                    gridTemplateColumns: `repeat(${Math.min(3, navigationGroups.length)}, minmax(0, 1fr))`,
                    justifyContent: 'center',
                    maxWidth: navigationGroups.length < 3 ? '720px' : 'none',
                    margin: '0 auto'
                  }}
                >
                  {navigationGroups.map((group, groupIndex) => (
                    <motion.section
                      key={group.id}
                      className="nav-dock__group"
                      initial={
                        shouldReduceMotion ? false : { opacity: 0, y: 10 }
                      }
                      animate={{ opacity: 1, y: 0 }}
                      transition={{
                        delay: shouldReduceMotion ? 0 : groupIndex * 0.05,
                      }}
                    >
                      <h2>{group.title}</h2>

                      <div className="nav-dock__items">
                        {group.items.map((item) => {
                          const Icon = item.icon;
                          const isActive = activePath === item.href;

                          return (
                            <button
                              key={item.id}
                              type="button"
                              onClick={() => handleNavigate(item)}
                              className={`nav-dock__item ${
                                isActive ? "nav-dock__item--active" : ""
                              }`}
                            >
                              {isActive && (
                                <motion.span
                                  layoutId="active-navigation"
                                  className="nav-dock__active-bg"
                                  transition={menuTransition}
                                />
                              )}

                              <span className="nav-dock__item-icon">
                                <Icon className="nav-dock__icon-svg" />
                              </span>

                              <span className="nav-dock__item-label">
                                {item.label}
                              </span>

                              {item.badge && (
                                <span className="nav-dock__badge">
                                  {item.badge}
                                </span>
                              )}

                              <ChevronDown className="nav-dock__chevron" />
                            </button>
                          );
                        })}
                      </div>
                    </motion.section>
                  ))}
                </div>
              </motion.div>
            )}
          </AnimatePresence>
        </motion.nav>
      </div>
    </>
  );
}
