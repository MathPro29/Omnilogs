import dayjs from 'dayjs';
import 'dayjs/locale/th';
import buddhistEra from 'dayjs/plugin/buddhistEra';
import relativeTime from 'dayjs/plugin/relativeTime';
import { DATE_FORMAT, STATUS_LABELS, STATUS_COLORS } from '@/constants';

// Setup dayjs plugins
dayjs.locale('th');
dayjs.extend(buddhistEra);
dayjs.extend(relativeTime);

/**
 * Format วันที่สำหรับแสดงผล
 */
export function formatDate(date: string | Date, format: string = DATE_FORMAT.DISPLAY): string {
  if (!date) return '-';
  return dayjs(date).format(format);
}

/**
 * Format วันที่พร้อมเวลา
 */
export function formatDateTime(date: string | Date): string {
  return formatDate(date, DATE_FORMAT.DISPLAY_TIME);
}

/**
 * เวลาที่ผ่านมา เช่น "3 นาทีที่แล้ว"
 */
export function timeAgo(date: string | Date): string {
  return dayjs(date).fromNow();
}

/**
 * Format ตัวเลขพร้อม comma
 */
export function formatNumber(value: number): string {
  return new Intl.NumberFormat('th-TH').format(value);
}

/**
 * ตัด text ถ้ายาวเกินไป
 */
export function truncateText(text: string, maxLength: number = 50): string {
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength) + '...';
}

/**
 * ดึง label ของ status
 */
export function getStatusLabel(status: string): string {
  return STATUS_LABELS[status] || status;
}

/**
 * ดึง color ของ status
 */
export function getStatusColor(status: string): string {
  return STATUS_COLORS[status] || 'default';
}

/**
 * สร้าง initials จากชื่อ
 */
export function getInitials(name: string): string {
  return name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .substring(0, 2);
}

/**
 * Debounce function
 */
export function debounce<T extends (...args: Parameters<T>) => void>(
  fn: T,
  delay: number
): (...args: Parameters<T>) => void {
  let timeoutId: ReturnType<typeof setTimeout>;
  return (...args: Parameters<T>) => {
    clearTimeout(timeoutId);
    timeoutId = setTimeout(() => fn(...args), delay);
  };
}
