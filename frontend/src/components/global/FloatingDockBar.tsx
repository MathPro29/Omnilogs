import { AnimatePresence, motion } from "motion/react";
import { Menu, X } from "lucide-react";
import { Avatar, Dropdown } from "antd";
import { getInitials } from "@/utils";
import type { FloatingDockController } from "@/hooks/useFloatingDock";
import { isNavigationPathActive } from "./navigation";

export function FloatingDockBar({
  dock,
  activePath,
}: {
  dock: FloatingDockController;
  activePath: string;
}) {
  const isAllowed = () => true;

  return (
    <>
      <motion.div layout="position" className="nav-dock__bar">
        <button
          type="button"
          className="nav-dock__brand"
          onClick={dock.toggleDock}
          aria-expanded={dock.isOpen}
          aria-controls={dock.menuId}
        >
          <span
            className={`nav-dock__brand-copy ${dock.isExpanded ? "nav-dock__brand-copy--open" : ""}`}
          >
            <span className="nav-dock__brand-title font-extrabold">
              <a href="/">OmniLogs</a>
            </span>
          </span>
        </button>
        {!dock.isExpanded && (
          <span className="nav-dock__divider" aria-hidden="true" />
        )}
        <div className="nav-dock__actions">
          <div
            className="nav-dock__quick-actions"
            aria-label="Quick navigation"
          >
            <button
              type="button"
              className={`nav-dock__quick-link ${isNavigationPathActive(activePath, "/") ? "nav-dock__quick-link--active" : ""}`}
              aria-current={
                isNavigationPathActive(activePath, "/") ? "page" : undefined
              }
              onClick={() => dock.navigateTo("/")}
            >
              แดชบอร์ด
            </button>
            {isAllowed() && (
              <button
                type="button"
                className={`nav-dock__quick-link ${isNavigationPathActive(activePath, "/logs-explorer") ? "nav-dock__quick-link--active" : ""}`}
                aria-current={
                  isNavigationPathActive(activePath, "/logs-explorer")
                    ? "page"
                    : undefined
                }
                onClick={() => dock.navigateTo("/logs-explorer")}
              >
                Logs Explorer
              </button>
            )}
            {isAllowed() && (
              <button
                type="button"
                className={`nav-dock__quick-link ${isNavigationPathActive(activePath, "/audit-logs") ? "nav-dock__quick-link--active" : ""}`}
                aria-current={
                  isNavigationPathActive(activePath, "/audit-logs")
                    ? "page"
                    : undefined
                }
                onClick={() => dock.navigateTo("/audit-logs")}
              >
                Audit Logs
              </button>
            )}
            <button
              type="button"
              className={`nav-dock__quick-link ${isNavigationPathActive(activePath, "/retention-policies") ? "nav-dock__quick-link--active" : ""}`}
              aria-current={
                isNavigationPathActive(activePath, "/retention-policies")
                  ? "page"
                  : undefined
              }
              onClick={() => dock.navigateTo("/retention-policies")}
            >
              จัดเก็บ Logs
            </button>
          </div>
          <Dropdown
            menu={{ items: dock.userMenuItems }}
            placement="bottomRight"
            trigger={["click"]}
          >
            <button
              type="button"
              className="nav-dock__profile"
              aria-label="Open profile menu"
            >
              <Avatar size={{ xs: 32, sm: 36 }} className="nav-dock__avatar">
                {dock.currentUser
                  ? getInitials(dock.currentUser.fullName)
                  : "U"}
              </Avatar>
            </button>
          </Dropdown>
          <button
            type="button"
            onClick={dock.toggleDock}
            className="nav-dock__toggle"
            aria-label={dock.isOpen ? "Close menu" : "Open menu"}
            aria-expanded={dock.isOpen}
            aria-controls={dock.menuId}
          >
            <AnimatePresence mode="wait" initial={false}>
              <motion.span
                key={dock.isOpen ? "close" : "open"}
                initial={
                  dock.shouldReduceMotion
                    ? false
                    : { opacity: 0, rotate: -45, scale: 0.7 }
                }
                animate={{ opacity: 1, rotate: 0, scale: 1 }}
                exit={
                  dock.shouldReduceMotion
                    ? undefined
                    : { opacity: 0, rotate: 45, scale: 0.7 }
                }
                transition={{ duration: 0.15 }}
              >
                {dock.isOpen ? <X size={20} /> : <Menu size={20} />}
              </motion.span>
            </AnimatePresence>
          </button>
        </div>
      </motion.div>
    </>
  );
}
