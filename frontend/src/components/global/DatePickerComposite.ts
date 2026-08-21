import {
  DateRangePicker,
  SingleDatePicker,
} from "./DatePicker";

export const DatePicker = Object.assign(SingleDatePicker, {
  RangePicker: DateRangePicker,
});

export default DatePicker;
