import { useEffect, useRef } from 'react';
import { productAdminService } from '@/services';
import { useAutoLogStore } from '@/store/auto-log.store';

const NASA_LOG_SEQUENCE = [
  { targetCategory: 'Head_logs', level: 'INFO', message: 'เริ่มกระบวนการตรวจสอบระบบยานอวกาศส่วนหัว (Head) ก่อนปล่อยตัว' },
  { targetCategory: 'Engine_logs', level: 'DEBUG', message: 'กำลังตรวจสอบสถานะเซ็นเซอร์ระบบขับเคลื่อนทั้งหมด (Engine)' },
  { targetCategory: 'Temp_logs', level: 'WARN', message: 'พบความผิดปกติเล็กน้อยที่เซ็นเซอร์อุณหภูมิ (Temp)' },
  { targetCategory: 'Temp_logs', level: 'INFO', message: 'เปิดใช้งานระบบทำความเย็นสำรองเพื่อลดอุณหภูมิ (Temp)' },
  { targetCategory: 'Compass_logs', level: 'DEBUG', message: 'กำลังปรับเทียบระบบนำทางและทิศทางเข้าสู่วงโคจร (Compass)' },
  { targetCategory: 'Rader_logs', level: 'ERROR', message: 'การเชื่อมต่อเรดาร์สัญญาณกับสถานีภาคพื้นดินขัดข้องชั่วคราว (Rader)' },
  { targetCategory: 'Rader_logs', level: 'INFO', message: 'ระบบสลับไปใช้เรดาร์ช่องสัญญาณสำรองสำเร็จ กู้คืนการเชื่อมต่อ (Rader)' },
  { targetCategory: 'Passenger_logs', level: 'INFO', message: 'ผู้โดยสารเข้าสู่ที่นั่งและรัดเข็มขัดนิรภัยเรียบร้อย (Passenger)' },
  { targetCategory: 'Wings_logs', level: 'WARN', message: 'กระแสไฟฟ้าปีกซ้ายลดลง 5% จากระดับมาตรฐาน (Wings)' },
  { targetCategory: 'Head_logs', level: 'INFO', message: 'อัปเดตแพทช์ระบบควบคุมการบินอัตโนมัติสำเร็จ (Head)' },
];

// Component นี้ Mount อยู่ระดับ App จึงไม่ถูกทำลายเมื่อผู้ใช้เปลี่ยนหน้าภายในระบบ
export function AutoLogProcessor() {
  const isRunning = useAutoLogStore((state) => state.isRunning);
  const config = useAutoLogStore((state) => state.config);
  const sendingRef = useRef(false);
  const sequenceIndexRef = useRef(0);

  useEffect(() => {
    if (!isRunning || !config) {
      // รีเซ็ต index เมื่อหยุดส่ง
      sequenceIndexRef.current = 0;
      return;
    }

    const sendLog = async () => {
      // ป้องกัน Interval รอบใหม่ยิงซ้อน หาก Request รอบก่อนยังไม่เสร็จ
      if (sendingRef.current) return;
      sendingRef.current = true;

      const currentLog = NASA_LOG_SEQUENCE[sequenceIndexRef.current];
      
      // หา Feature ปลายทางตามที่กำหนดใน Sequence (ถ้ามีข้อมูล Features ให้ค้นหา)
      let dynamicCategoryId = config.categoryId;
      let dynamicFeatureFullPath = config.featureFullPath;
      let dynamicFeaturePathIds = config.featurePathIds;

      if (config.features && config.features.length > 0) {
        const targetFeature = config.features.find((f: any) => f.categoryName === currentLog.targetCategory);
        if (targetFeature) {
          dynamicCategoryId = targetFeature.categoryId;
          dynamicFeatureFullPath = targetFeature.fullPath;
          dynamicFeaturePathIds = targetFeature.pathIds;
        }
      }

      try {
        await productAdminService.importLogs({
          ...config,
          categoryId: dynamicCategoryId,
          featureFullPath: dynamicFeatureFullPath,
          featurePathIds: dynamicFeaturePathIds,
          logLevel: currentLog.level,
          message: `${currentLog.message} (seq: ${sequenceIndexRef.current + 1}, time: ${new Date().toLocaleTimeString()})`,
        });
        
        // ขยับไป index ถัดไปวนลูป
        sequenceIndexRef.current = (sequenceIndexRef.current + 1) % NASA_LOG_SEQUENCE.length;
      } catch (error) {
        // Error ถูกเก็บไว้ที่ Console แต่ Process ยังทำงานต่อในรอบถัดไป
        console.error('Auto log sending failed:', error);
      } finally {
        sendingRef.current = false;
      }
    };

    const interval = window.setInterval(sendLog, 2000);
    return () => window.clearInterval(interval);
  }, [isRunning, config]);

  return null;
}
