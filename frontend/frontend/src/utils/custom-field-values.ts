import dayjs from 'dayjs';
import type { CustomField } from '@/services/custom-field.service';

export function getCustomFieldInitialValues(fields: CustomField[]): Record<string, unknown> {
  return fields.reduce<Record<string, unknown>>((values, field) => {
    let value: unknown = field.default_value;
    if (field.data_type === 'enum') {
      value = field.default_value || field.enum_options?.find((option) => option.is_active && option.is_default)?.option_value;
    } else if (field.data_type === 'boolean') {
      value = field.default_value === 'true';
    } else if (field.data_type === 'number' && field.default_value != null) {
      value = Number(field.default_value);
    } else if (field.data_type === 'date' && field.default_value) {
      value = dayjs(field.default_value);
    }
    if (value !== undefined && value !== null && value !== '') values[field.field_key] = value;
    return values;
  }, {});
}

export function normalizeCustomFieldValues(
  fields: CustomField[],
  values?: Record<string, unknown>,
): Record<string, unknown> | undefined {
  if (!values) return undefined;
  const normalized: Record<string, unknown> = {};
  for (const field of fields) {
    const value = values[field.field_key];
    if (value === undefined || value === null || value === '') continue;
    if ((field.data_type === 'object' || field.data_type === 'array') && typeof value === 'string') {
      normalized[field.field_key] = JSON.parse(value);
    } else if (field.data_type === 'date' && dayjs.isDayjs(value)) {
      normalized[field.field_key] = value.toISOString();
    } else {
      normalized[field.field_key] = value;
    }
  }
  return Object.keys(normalized).length > 0 ? normalized : undefined;
}
