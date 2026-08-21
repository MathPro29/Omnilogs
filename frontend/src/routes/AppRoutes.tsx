import { useRef, useEffect } from 'react';
import { Routes, Route, Navigate, useLocation } from 'react-router-dom';
import LoadingBar, { type LoadingBarRef } from 'react-top-loading-bar';
import { AuthLayout, AdminLayout } from '@/layouts';
import { ProtectedRoute, GuestRoute } from '@/routes/ProtectedRoute';
import { adminRoutes } from '@/routes/routes';
import { LoginPage, RegisterPage, ForgotPasswordPage, NotFoundPage, GenericBlankPage } from '@/pages';
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
          <Route path={ROUTES.REGISTER} element={<GuestRoute><RegisterPage /></GuestRoute>} />
          <Route path={ROUTES.FORGOT_PASSWORD} element={<GuestRoute><ForgotPasswordPage /></GuestRoute>} />
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
                route.requiredPermissions || route.forbiddenRoles ? (
                  <ProtectedRoute
                    requiredPermissions={route.requiredPermissions}
                    requireAll={route.requireAll}
                    forbiddenRoles={route.forbiddenRoles}
                  >
                    {route.element}
                  </ProtectedRoute>
                ) : (
                  route.element
                )
              }
            />
          ))}
          <Route path="/analytics" element={<GenericBlankPage />} />
          <Route path="/pricing" element={<GenericBlankPage />} />
          <Route path="/subscriptions" element={<GenericBlankPage />} />
          <Route path="/billing" element={<GenericBlankPage />} />
          <Route path="/integrations" element={<GenericBlankPage />} />
          <Route path="/documentation" element={<GenericBlankPage />} />
          <Route path="/changelog" element={<GenericBlankPage />} />

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
