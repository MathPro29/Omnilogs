import { Outlet } from 'react-router-dom';
import { motion } from 'motion/react';

/**
 * AuthLayout - Layout สำหรับหน้า login / register
 * พื้นหลัง gradient สีม่วง
 */
export function AuthLayout() {
  return (
    <div className="login-container">
      <motion.div
        initial={{ opacity: 0, scale: 0.96 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ duration: 0.4, ease: [0.25, 0.46, 0.45, 0.94] }}
      >
        <Outlet />
      </motion.div>
    </div>
  );
}
