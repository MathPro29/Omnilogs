import { useRef } from "react";
import { AnimatePresence, motion } from "motion/react";
import { useFloatingDock } from "@/hooks/useFloatingDock";
import { FloatingDockBar } from "./FloatingDockBar";
import { FloatingDockMenu } from "./FloatingDockMenu";
import type { NavigationItem } from "./navigation";

interface FloatingDockNavbarProps {
  activePath?: string;
  onNavigate?: (item: NavigationItem) => void;
}

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
  const containerRef = useRef<HTMLDivElement>(null);
  const dock = useFloatingDock(containerRef, onNavigate);

  return (
    <>
      <AnimatePresence>
        {dock.isOpen && (
          <motion.button
            type="button"
            aria-label="Close navigation menu"
            className="dock-backdrop"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: dock.shouldReduceMotion ? 0 : 0.2 }}
            onClick={dock.closeDock}
          />
        )}
      </AnimatePresence>
      <div ref={containerRef} className="nav-dock-frame">
        <motion.nav
          layout
          aria-label="Primary navigation"
          className={`nav-dock ${dock.isExpanded ? "nav-dock--open" : ""}`}
          transition={dock.shouldReduceMotion ? { duration: 0 } : menuTransition}
        >
          <FloatingDockBar dock={dock} activePath={activePath} />
          <AnimatePresence initial={false}>
            {dock.isOpen && (
              <FloatingDockMenu activePath={activePath} dock={dock} />
            )}
          </AnimatePresence>
        </motion.nav>
      </div>
    </>
  );
}