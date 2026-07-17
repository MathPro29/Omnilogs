import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { AnimatePresence } from 'motion/react';
import { FloatingDockNavbar } from '@/components';
import type { NavigationItem } from '@/components/global/navigation';

/**
 * AdminLayout - Layout หลักของระบบ admin โดยใช้ FloatingDockNavbar และ CSS ดั้งเดิม
 */
export function AdminLayout() {
  const navigate = useNavigate();
  const location = useLocation();

  return (
    <main className="app-shell">
      <FloatingDockNavbar
        activePath={location.pathname}
        onNavigate={(item: NavigationItem) => {
          if (item.href.startsWith('http://') || item.href.startsWith('https://')) {
            window.open(item.href, '_blank', 'noopener,noreferrer');
          } else {
            navigate(item.href);
          }
        }}
      />
      <section className="page-panel" aria-live="polite">
        <AnimatePresence mode="wait">
          <Outlet />
        </AnimatePresence>
      </section>
    </main>
  );
}
