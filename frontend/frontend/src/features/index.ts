/**
 * Features Directory - Feature-Based Modules
 *
 * โฟลเดอร์นี้เก็บ business logic ที่ผูกกับ feature/module เฉพาะ
 * ใช้เมื่อ logic นั้นเกี่ยวข้องกับ 1 module เท่านั้น ไม่ได้ใช้ร่วมกับ module อื่น
 *
 * === โครงสร้างของแต่ละ feature ===
 *
 * features/
 *   users/
 *     components/        -> UI components เฉพาะ user module
 *       UserForm.tsx
 *       UserDetailCard.tsx
 *     hooks/             -> Custom hooks เฉพาะ user module
 *       useUserFilters.ts
 *       useUserMutation.ts
 *     utils/             -> Utility functions เฉพาะ user module
 *       userValidation.ts
 *     types.ts           -> Types เฉพาะ feature (ถ้ามี)
 *     index.ts           -> Public API ของ feature
 *
 * === หลักการแบ่ง Global vs Feature ===
 *
 * ถามตัวเอง: "ถ้าลบ feature นี้ออก ยังมีคนใช้ของชิ้นนี้ไหม?"
 *   ใช่  -> ใส่ไว้ใน src/utils/, src/hooks/, src/components/global/
 *   ไม่  -> ใส่ไว้ใน src/features/<feature-name>/
 *
 * === ตัวอย่าง ===
 *
 * Global (src/utils/):
 *   - formatDate(), formatNumber()    -> ทุก module ใช้
 *   - can(), canAny(), canAll()       -> ไม่ผูกกับ feature
 *   - debounce()                      -> utility ทั่วไป
 *
 * Feature (src/features/payroll/utils/):
 *   - calculateOvertimePay()          -> เฉพาะ payroll
 *   - generatePayslipPDF()            -> เฉพาะ payroll
 *
 * Global (src/hooks/):
 *   - usePermission()                 -> ทุก module ใช้
 *   - useLoadingBar()                 -> ไม่ผูกกับ feature
 *
 * Feature (src/features/users/hooks/):
 *   - useUserFilters()                -> เฉพาะ user management
 *   - useUserMutation()               -> เฉพาะ user CRUD
 *
 * Global (src/components/global/):
 *   - PageTransition                  -> ทุกหน้า
 *   - PermissionGuard                 -> ทุกหน้า
 *   - Skeletons                       -> ทุกหน้า
 *
 * Feature (src/features/leave/components/):
 *   - LeaveRequestForm                -> เฉพาะลางาน
 *   - LeaveCalendar                   -> เฉพาะลางาน
 */

export {};
