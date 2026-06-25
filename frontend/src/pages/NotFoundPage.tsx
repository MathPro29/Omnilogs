import { useNavigate } from 'react-router-dom';
import { Button, Typography } from 'antd';
import { ExclamationTriangleIcon } from '@heroicons/react/24/outline';
import { motion } from 'motion/react';
import { ROUTES } from '@/constants';

const { Text } = Typography;

/**
 * 404 Not Found Page
 * แสดงเป็นหน้าเต็มจอ ไม่ใช้ Admin Layout
 */
export function NotFoundPage() {
  const navigate = useNavigate();

  return (
    <div
      className="flex items-center justify-center"
      style={{
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #f5f5f5 0%, #F0EDFF 100%)',
      }}
    >
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4, ease: 'easeOut' }}
        style={{
          background: '#ffffff',
          borderRadius: 16,
          padding: '48px 40px',
          boxShadow: '0 8px 32px rgba(102, 71, 240, 0.08)',
          maxWidth: 480,
          width: '100%',
          margin: '0 16px',
          textAlign: 'center',
        }}
      >
        <motion.div
          initial={{ scale: 0.8 }}
          animate={{ scale: 1 }}
          transition={{ delay: 0.1, duration: 0.3 }}
        >
          <ExclamationTriangleIcon
            className="w-20 h-20 mx-auto"
            style={{ color: '#D4B106' }}
          />
        </motion.div>

        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.2, duration: 0.3 }}
        >
          <h1
            style={{
              fontSize: 24,
              fontWeight: 700,
              color: '#000000D9',
              marginTop: 24,
              marginBottom: 8,
            }}
          >
            404 - ไม่พบหน้าที่ต้องการ
          </h1>
          <Text type="secondary" style={{ fontSize: 15 }}>
            หน้าที่คุณกำลังมองหาอาจถูกลบ เปลี่ยนชื่อ หรือไม่มีอยู่ในระบบ
          </Text>
        </motion.div>

        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.3, duration: 0.3 }}
          className="flex gap-3 justify-center"
          style={{ marginTop: 32 }}
        >
          <Button
            type="primary"
            size="large"
            onClick={() => navigate(ROUTES.DASHBOARD)}
          >
            กลับหน้าแดชบอร์ด
          </Button>
          <Button size="large" onClick={() => navigate(-1)}>
            ย้อนกลับ
          </Button>
        </motion.div>
      </motion.div>
    </div>
  );
}
