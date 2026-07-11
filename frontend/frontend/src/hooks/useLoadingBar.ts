import { useRef, useCallback } from 'react';
import LoadingBar, { type LoadingBarRef } from 'react-top-loading-bar';

/**
 * Hook สำหรับใช้ Top Loading Bar
 * ใช้ตอน navigate ระหว่างหน้า หรือ fetch ข้อมูลสำคัญ
 */
export function useLoadingBar() {
  const ref = useRef<LoadingBarRef>(null);

  const start = useCallback(() => {
    ref.current?.continuousStart(0, 500);
  }, []);

  const complete = useCallback(() => {
    ref.current?.complete();
  }, []);

  const increase = useCallback((value: number) => {
    ref.current?.increase(value);
  }, []);

  return { ref, start, complete, increase, LoadingBar };
}
