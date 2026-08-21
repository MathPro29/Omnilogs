import type { CSSProperties } from "react";
import { DatePicker as AntDatePicker } from "antd";
import dayjs, { type Dayjs } from "dayjs";

export type DateRange = [Dayjs | null, Dayjs | null] | null;

const { RangePicker: AntRangePicker } = AntDatePicker;

export interface DateRangePickerProps {
  value?: DateRange;
  onChange?: (value: DateRange) => void;
  showTime?: boolean | object;
  format?: string;
  className?: string;
  style?: CSSProperties;
  allowClear?: boolean;
  placeholder?: [string, string];
  disabled?: boolean;
}

export function DateRangePicker({
  value,
  onChange,
  showTime = {
    defaultValue: [
      dayjs("00:00:00", "HH:mm:ss"),
      dayjs("23:59:59", "HH:mm:ss"),
    ],
  },
  format = "YYYY-MM-DD HH:mm:ss",
  className,
  style,
  allowClear = true,
  placeholder = ["Start Date", "End Date"],
  disabled = false,
}: DateRangePickerProps) {
  return (
    <AntRangePicker
      value={value}
      onChange={(val) => onChange?.(val as DateRange)}
      showTime={showTime}
      format={format}
      className={className}
      style={{ width: "100%", ...style }}
      allowClear={allowClear}
      placeholder={placeholder}
      disabled={disabled}
    />
  );
}

export interface SingleDatePickerProps {
  value?: Dayjs | null;
  onChange?: (value: Dayjs | null) => void;
  showTime?: boolean | object;
  format?: string;
  className?: string;
  style?: CSSProperties;
  allowClear?: boolean;
  placeholder?: string;
  disabled?: boolean;
}

export function SingleDatePicker({
  value,
  onChange,
  showTime = true,
  format = "YYYY-MM-DD HH:mm:ss",
  className,
  style,
  allowClear = true,
  placeholder = "Select Date",
  disabled = false,
}: SingleDatePickerProps) {
  return (
    <AntDatePicker
      value={value}
      onChange={(val) => onChange?.(val)}
      showTime={showTime}
      format={format}
      className={className}
      style={{ width: "100%", ...style }}
      allowClear={allowClear}
      placeholder={placeholder}
      disabled={disabled}
    />
  );
}

