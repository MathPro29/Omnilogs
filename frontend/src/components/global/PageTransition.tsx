import type { ReactNode } from "react";
import { motion } from "motion/react";

interface PageTransitionProps {
  children: ReactNode;
}

/**
 * Page transition wrapper ด้วย Motion
 * ครอบทุก page component เพื่อให้มี animation ตอนเปลี่ยนหน้า
 */
export function PageTransition({ children }: PageTransitionProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -8 }}
      transition={{
        duration: 0.25,
        ease: [0.25, 0.46, 0.45, 0.94],
      }}
    >
      {children}
    </motion.div>
  );
}
