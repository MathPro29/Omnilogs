import { useNavigate } from 'react-router-dom';
import { Result, Button } from 'antd';
import { ShieldExclamationIcon } from '@heroicons/react/24/outline';
import { motion } from 'motion/react';
import { ROUTES } from '@/constants';

/**
 * 403 Forbidden Page
 */
export function ForbiddenPage() {
  const navigate = useNavigate();

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.96 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ duration: 0.3 }}
      className="flex items-center justify-center min-h-[60vh]"
    >
      <Result
        icon={
          <ShieldExclamationIcon
            className="w-20 h-20 mx-auto"
            style={{ color: '#FF3D89' }}
          />
        }
        title="403 - ไม่มีสิทธิ์เข้าถึง"
        subTitle="คุณไม่มีสิทธิ์ในการเข้าถึงหน้านี้ กรุณาติดต่อผู้ดูแลระบบ"
        extra={[
          <Button
            type="primary"
            key="dashboard"
            onClick={() => navigate(ROUTES.DASHBOARD)}
          >
            กลับหน้าแดชบอร์ด
          </Button>,
          <Button key="back" onClick={() => navigate(-1)}>
            ย้อนกลับ
          </Button>,
        ]}
      />
    </motion.div>
  );
}
