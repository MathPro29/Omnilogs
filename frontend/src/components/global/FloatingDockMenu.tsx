import { motion } from "motion/react";
import { ChevronDown } from "lucide-react";
import { isNavigationPathActive, navigationGroups } from "./navigation";
import type { FloatingDockController } from "@/hooks/useFloatingDock";
import { useAuthStore } from "@/store";

const menuTransition = {
  type: "spring" as const,
  stiffness: 380,
  damping: 34,
  mass: 0.8,
};

interface FloatingDockMenuProps {
  activePath: string;
  dock: FloatingDockController;
}

export function FloatingDockMenu({
  activePath,
  dock,
}: FloatingDockMenuProps) {  const roles = useAuthStore((state) => state.roles) || [];
  const normalizedRoles = roles.map((role) => role.trim().toLowerCase());
  const isItemAllowed = (item: (typeof navigationGroups)[number]["items"][number]) =>
    !item.hiddenForRoles?.some((role) => normalizedRoles.includes(role));

  return (
    <motion.div
      id={dock.menuId}
      className="nav-dock__menu"
      initial={dock.shouldReduceMotion ? false : { opacity: 0, height: 0, y: -12 }}
      animate={{ opacity: 1, height: "auto", y: 0 }}
      exit={dock.shouldReduceMotion ? undefined : { opacity: 0, height: 0, y: -8 }}
      transition={dock.shouldReduceMotion ? { duration: 0 } : { height: menuTransition, opacity: { duration: 0.2 }, y: menuTransition }}
    >
      <div className="nav-dock__grid">
        {navigationGroups.map((group, groupIndex) => {
          const visibleItems = group.items.filter(isItemAllowed);
          if (visibleItems.length === 0) return null;

          return (
            <motion.section
              key={group.id}
              className="nav-dock__group"
              initial={dock.shouldReduceMotion ? false : { opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: dock.shouldReduceMotion ? 0 : groupIndex * 0.05 }}
            >
              <h2>{group.title}</h2>
              <div className="nav-dock__items">
                {visibleItems.map((item) => {
                const Icon = item.icon;
                const isActive = isNavigationPathActive(activePath, item.href);
                const isHighlighted = isActive;
                return (
                  <button
                    key={item.id}
                    type="button"
                    aria-current={isActive ? "page" : undefined}
                    onClick={() => dock.handleNavigate(item)}
                    className={`nav-dock__item ${isHighlighted ? "nav-dock__item--active" : ""}`}
                  >
                    {isHighlighted && (
                      <motion.span
                        layoutId="active-navigation"
                        className="nav-dock__active-bg"
                        transition={menuTransition}
                      />
                    )}
                    <span className="nav-dock__item-icon"><Icon className="nav-dock__icon-svg" /></span>
                    <span className="nav-dock__item-label">{item.label}</span>
                    {item.badge && <span className="nav-dock__badge">{item.badge}</span>}
                    <ChevronDown className="nav-dock__chevron" />
                  </button>
                );
              })}
            </div>
          </motion.section>
        );
      })}
      </div>
    </motion.div>
  );
}