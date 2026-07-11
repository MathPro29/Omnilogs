/**
 * Components - Public API
 *
 * แบ่งเป็น 2 กลุ่ม:
 * - global/   -> ใช้ได้ทุกหน้า (PageTransition, PermissionGuard, Skeletons)
 * - features/ -> ผูกกับ feature/module เฉพาะ (UserAvatar, DashboardChart, ...)
 */

// Global Components
export {
  PermissionGuard,
  PageTransition,
  DashboardSkeleton,
  TableSkeleton,
  FormSkeleton,
  DetailSkeleton,
  FloatingDockNavbar,
} from './global';

// Feature Components
// export { ... } from './features';
