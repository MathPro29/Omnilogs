import { useRef, useEffect } from 'react';
import { Routes, Route, Navigate, useLocation } from 'react-router-dom';
import LoadingBar, { type LoadingBarRef } from 'react-top-loading-bar';
import { AuthLayout, AdminLayout } from '@/layouts';
import { ProtectedRoute, GuestRoute } from '@/routes/ProtectedRoute';
import { adminRoutes } from '@/routes/routes';
import { Suspense, lazy } from 'react';
import { Spin } from 'antd';

const LoginPage = lazy(() => import('@/pages/LoginPage').then(module => ({ default: module.LoginPage })));
const RegisterPage = lazy(() => import('@/pages/RegisterPage').then(module => ({ default: module.RegisterPage })));
const NotFoundPage = lazy(() => import('@/pages/NotFoundPage').then(module => ({ default: module.NotFoundPage })));
const ApiTestPage = lazy(() => import('@/pages/ApiTestPage').then(module => ({ default: module.ApiTestPage })));
const ForgotPasswordPage = lazy(() => import('@/pages/ForgotPassword').then(module => ({ default: module.default })));
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

      <Suspense fallback={
        <div className="flex h-screen w-full items-center justify-center">
          <Spin size="large" />
        </div>
      }>
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
          <Route
            path={ROUTES.REGISTER}
            element={
              <GuestRoute>
                <RegisterPage />
              </GuestRoute>
            }
          />
          <Route
            path={ROUTES.FORGOT_PASSWORD}
            element={
              <GuestRoute>
                <ForgotPasswordPage />
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
      </Suspense>
    </>
  );
}
