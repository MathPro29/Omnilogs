import { DatePicker, Form, Input, InputNumber, Select, Switch, Typography } from 'antd';
import type { CustomField } from '@/services/custom-field.service';

const { Text } = Typography;

type Props = {
  fields: CustomField[];
  namePrefix?: string;
};

function jsonValidator(expectedType: 'array' | 'object') {
  return async (_: unknown, value?: string) => {
    if (!value?.trim()) return;
    try {
      const parsed = JSON.parse(value);
      const valid = expectedType === 'array'
        ? Array.isArray(parsed)
        : parsed !== null && typeof parsed === 'object' && !Array.isArray(parsed);
      if (!valid) throw new Error('invalid type');
    } catch {
      throw new Error(expectedType === 'array' ? 'กรุณากรอก JSON array ที่ถูกต้อง' : 'กรุณากรอก JSON object ที่ถูกต้อง');
    }
  };
}

export function CustomFieldInputs({ fields, namePrefix = 'customFields' }: Props) {
  const visibleFields = fields.filter((field) => field.is_active && field.is_visible);
  if (visibleFields.length === 0) return null;

  return (
    <div style={{ borderTop: '1px solid #f0f0f0', paddingTop: 16, marginTop: 8 }}>
      <Text strong>Custom Fields</Text>
      <div style={{ marginTop: 12 }}>
        {visibleFields.map((field) => {
          const name = [namePrefix, field.field_key];
          const label = field.display_name || field.field_key;
          const requiredRule = field.is_required
            ? [{ required: true, message: `กรุณาระบุ ${label}` }]
            : [];

          if (field.data_type === 'enum') {
            return (
              <Form.Item key={field.field_definition_id} name={name} label={label} rules={requiredRule} extra={field.description}>
                <Select
                  allowClear={!field.is_required}
                  placeholder={`เลือก ${label}`}
                  options={(field.enum_options || []).filter((option) => option.is_active).map((option) => ({
                    value: option.option_value,
                    label: option.option_label,
                  }))}
                />
              </Form.Item>
            );
          }

          if (field.data_type === 'number') {
            return (
              <Form.Item key={field.field_definition_id} name={name} label={label} rules={requiredRule} extra={field.description}>
                <InputNumber style={{ width: '100%' }} placeholder={`กรอก ${label}`} />
              </Form.Item>
            );
          }

          if (field.data_type === 'boolean') {
            return (
              <Form.Item key={field.field_definition_id} name={name} label={label} valuePropName="checked" extra={field.description}>
                <Switch checkedChildren="ใช่" unCheckedChildren="ไม่ใช่" />
              </Form.Item>
            );
          }

          if (field.data_type === 'date') {
            return (
              <Form.Item key={field.field_definition_id} name={name} label={label} rules={requiredRule} extra={field.description}>
                <DatePicker showTime style={{ width: '100%' }} />
              </Form.Item>
            );
          }

          if (field.data_type === 'object' || field.data_type === 'array') {
            return (
              <Form.Item
                key={field.field_definition_id}
                name={name}
                label={label}
                rules={[...requiredRule, { validator: jsonValidator(field.data_type) }]}
                extra={field.description || `รูปแบบ JSON ${field.data_type}`}
              >
                <Input.TextArea rows={3} placeholder={field.data_type === 'array' ? '["value"]' : '{"key":"value"}'} />
              </Form.Item>
            );
          }

          return (
            <Form.Item key={field.field_definition_id} name={name} label={label} rules={requiredRule} extra={field.description}>
              <Input placeholder={`กรอก ${label}`} />
            </Form.Item>
          );
        })}
      </div>
    </div>
  );
}
