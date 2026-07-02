import { useRef, useEffect } from 'react';
import { Routes, Route, Navigate, useLocation } from 'react-router-dom';
import LoadingBar, { type LoadingBarRef } from 'react-top-loading-bar';
import { AuthLayout, AdminLayout } from '@/layouts';
import { ProtectedRoute, GuestRoute } from '@/routes/ProtectedRoute';
import { adminRoutes } from '@/routes/routes';
import { LoginPage, NotFoundPage, ApiTestPage } from '@/pages';
import { ROUTES } from '@/constants';

// ===== Route Change Loading Bar =====
function RouteChangeHandler({ loadingBarRef }: { loadingBarRef: React.RefObject<LoadingBarRef | null> }) {
  const location = useLocation();

  useEffect(() => {
    loadingBarRef.current?.continuousStart(0, 300);
    const timer = setTimeout(() => {
      loadingBarRef.current?.complete();
    }, 400);
    return () => clearTimeout(timer);
  }, [location.pathname, loadingBarRef]);

  return null;
}

// ===== App Routes =====
export function AppRoutes() {
  const loadingBarRef = useRef<LoadingBarRef>(null);

  return (
    <>
      <LoadingBar
        ref={loadingBarRef}
        color="var(--color-primary)"
        height={3}
        shadow={true}
        transitionTime={200}
      />
      <RouteChangeHandler loadingBarRef={loadingBarRef} />

      <Routes>
        {/* API Test Route */}
        <Route path={ROUTES.API_TEST} element={<ApiTestPage />} />

        {/* Public Routes */}
        <Route element={<AuthLayout />}>
          <Route
            path={ROUTES.LOGIN}
            element={
              <GuestRoute>
                <LoginPage />
              </GuestRoute>
            }
          />
        </Route>

        {/* Protected Admin Routes - generate จาก config */}
        <Route
          element={
            <ProtectedRoute>
              <AdminLayout />
            </ProtectedRoute>
          }
        >
          {adminRoutes.map((route) => (
            <Route
              key={route.path}
              path={route.path}
              element={
                route.requiredPermissions ? (
                  <ProtectedRoute
                    requiredPermissions={route.requiredPermissions}
                    requireAll={route.requireAll}
                  >
                    {route.element}
                  </ProtectedRoute>
                ) : (
                  route.element
                )
              }
            />
          ))}
        </Route>

        {/* 404 - standalone (ไม่ใช้ admin layout) */}
        <Route path={ROUTES.NOT_FOUND} element={<NotFoundPage />} />

        {/* Redirect root to dashboard */}
        <Route path="/" element={<Navigate to={ROUTES.DASHBOARD} replace />} />

        {/* Global catch-all -> 404 standalone */}
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </>
  );
}
