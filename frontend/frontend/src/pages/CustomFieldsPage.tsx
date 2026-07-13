import { useEffect, useMemo, useState } from 'react';
import {
  Alert, Button, Card, Col, Collapse, Empty, Form, Input, InputNumber, Modal, Row, Select,
  Space, Switch, Table, Tag, Typography, message,
} from 'antd';
import { PencilIcon, PlusIcon, TrashIcon } from '@heroicons/react/24/outline';
import { useQuery } from '@tanstack/react-query';
import { PageTransition } from '@/components';
import { ROUTES } from '@/constants';
import { productAdminService } from '@/services';
import {
  customFieldService,
  type CustomField,
  type CustomFieldOption,
} from '@/services/custom-field.service';
import { useAppStore } from '@/store';

const { Title, Text } = Typography;

type EnumOptionFormValue = Partial<CustomFieldOption> & {
  option_key: string;
  option_label: string;
  option_value: string;
};

type CustomFieldFormValue = Partial<CustomField> & {
  field_key: string;
  data_type: string;
  enum_options?: EnumOptionFormValue[];
};

export function CustomFieldsPage() {
  const setBreadcrumbs = useAppStore((state) => state.setBreadcrumbs);
  const [productId, setProductId] = useState<number>();
  const [editing, setEditing] = useState<CustomField | null>(null);
  const [open, setOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const [activeCollapseKeys, setActiveCollapseKeys] = useState<string[]>([]);
  const [form] = Form.useForm<CustomFieldFormValue>();
  const projectId = Form.useWatch('project_id', form);
  const dataType = Form.useWatch('data_type', form);

  const productsQuery = useQuery({ queryKey: ['custom-field-products'], queryFn: productAdminService.listProducts });
  const projectsQuery = useQuery({
    queryKey: ['custom-field-projects', productId],
    queryFn: () => productAdminService.listProjects(productId!),
    enabled: !!productId,
  });
  const featuresQuery = useQuery({
    queryKey: ['custom-field-categories', productId, projectId],
    queryFn: () => productAdminService.listFeatures(productId!, projectId!),
    enabled: !!productId && !!projectId,
  });
  const fieldsQuery = useQuery({
    queryKey: ['custom-fields', productId],
    queryFn: () => customFieldService.list(productId),
    enabled: !!productId,
  });
  const product = useMemo(
    () => (productsQuery.data || []).find((item) => item.productId === productId),
    [productsQuery.data, productId],
  );

  useEffect(() => {
    setBreadcrumbs([{ title: 'ผลิตภัณฑ์ที่จัดการ', path: ROUTES.PRODUCTS }, { title: 'Custom Fields' }]);
  }, [setBreadcrumbs]);
  

  const closeModal = () => {
    setOpen(false);
    setEditing(null);
    form.resetFields();
    setActiveCollapseKeys([]);
    setBreadcrumbs([{ title: 'ผลิตภัณฑ์ที่จัดการ', path: ROUTES.PRODUCTS }, { title: 'Custom Fields' }]);
  };

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({
      data_type: 'string',
      is_visible: true,
      is_searchable: true,
      is_filterable: true,
      is_sortable: true,
      is_aggregatable: true,
      is_required: false,
      is_sensitive: false,
      mask_before_index: false,
      encrypt_before_archive: false,
      enum_options: [],
    });
    setActiveCollapseKeys([]);
    setOpen(true);
  };

  const openEdit = (field: CustomField) => {
    setEditing(field);
    form.resetFields();
    form.setFieldsValue({
      ...field,
      config_json: field.config_json ? JSON.stringify(field.config_json, null, 2) : undefined,
      enum_options: (field.enum_options || []).filter((option) => option.is_active),
    });
    if (field.config_json) {
      setActiveCollapseKeys(['advanced']);
      setBreadcrumbs([
        { title: 'ผลิตภัณฑ์ที่จัดการ', path: ROUTES.PRODUCTS },
        { title: 'Custom Fields' },
        { title: 'Advanced' }
      ]);
    } else {
      setActiveCollapseKeys([]);
    }
    setOpen(true);
  };

  const syncOptions = async (fieldId: number, previous: CustomFieldOption[], next: EnumOptionFormValue[]) => {
    const nextIDs = new Set(next.map((option) => option.option_id).filter((id): id is number => !!id));
    await Promise.all(
      previous
        .filter((option) => option.is_active && !nextIDs.has(option.option_id))
        .map((option) => customFieldService.removeOption(fieldId, option.option_id)),
    );
    for (const [index, option] of next.entries()) {
      const payload = {
        option_key: option.option_key.trim(),
        option_label: option.option_label.trim(),
        option_value: option.option_value.trim(),
        color_code: option.color_code?.trim() || null,
        description: option.description?.trim() || null,
        display_order: option.display_order ?? index,
        is_default: !!option.is_default,
        is_active: true,
      };
      if (option.option_id) {
        await customFieldService.updateOption(fieldId, option.option_id, payload);
      } else {
        await customFieldService.createOption(fieldId, payload);
      }
    }
  };

  const save = async (values: CustomFieldFormValue) => {
    if (!productId) return;
    const enumOptions = values.data_type === 'enum' ? values.enum_options || [] : [];
    if (values.data_type === 'enum' && enumOptions.length === 0) {
      message.error('Enum ต้องมีอย่างน้อย 1 ตัวเลือก');
      return;
    }
    if (enumOptions.filter((option) => option.is_default).length > 1) {
      message.error('กำหนดค่าเริ่มต้นของ Enum ได้เพียง 1 ตัวเลือก');
      return;
    }

    setSaving(true);
    try {
      const fieldValues = { ...values };
      delete fieldValues.enum_options;
      
      let parsedConfigJson = undefined;
      if (typeof fieldValues.config_json === 'string' && fieldValues.config_json.trim()) {
        try {
          parsedConfigJson = JSON.parse(fieldValues.config_json);
        } catch (e) {
          message.error('Config JSON ไม่ถูกต้อง (ต้องเป็น JSON ที่ valid)');
          setSaving(false);
          return;
        }
      }

      const fieldPath = fieldValues.field_path?.trim() || `custom_fields.${fieldValues.field_key}`;
      const payload = {
        ...fieldValues,
        config_json: parsedConfigJson,
        product_id: productId,
        field_path: fieldPath,
        source_section: 'payload',
        elastic_field_name: fieldValues.elastic_field_name?.trim() || `payload.custom_fields.${fieldValues.field_key}`,
        value_source_type: fieldValues.data_type === 'enum' ? 'ENUM' : 'NONE',
      };
      let fieldId = editing?.field_definition_id;
      if (fieldId) {
        await customFieldService.update(fieldId, payload);
      } else {
        const created = await customFieldService.create(payload);
        fieldId = created.field_definition_id;
      }
      await syncOptions(fieldId, editing?.enum_options || [], enumOptions);
      message.success(editing ? 'อัปเดต Custom Field สำเร็จ' : 'สร้าง Custom Field สำเร็จ');
      closeModal();
      await fieldsQuery.refetch();
    } catch (error: unknown) {
      const apiError = error as { response?: { data?: { error?: string } }; message?: string };
      message.error(apiError.response?.data?.error || apiError.message || 'บันทึกไม่สำเร็จ');
    } finally {
      setSaving(false);
    }
  };

  const columns = [
    {
      title: 'Field',
      render: (_: unknown, row: CustomField) => (
        <Space direction="vertical" size={0}>
          <Text strong>{row.display_name || row.field_key}</Text>
          <Text type="secondary">{row.field_key}</Text>
        </Space>
      ),
    },
    { title: 'Scope', render: (_: unknown, row: CustomField) => <Text>Project {row.project_id ? `#${row.project_id}` : 'ทั้งหมด'} / Category {row.category_id ? `#${row.category_id}` : 'ทั้งหมด'}</Text> },
    { title: 'Field Path', dataIndex: 'field_path', render: (value: string | null) => <Text code>{value || '—'}</Text> },
    {
      title: 'Type / Options',
      render: (_: unknown, row: CustomField) => (
        <Space direction="vertical" size={4}>
          <Tag color="blue">{row.data_type}</Tag>
          {row.data_type === 'enum' && <Space wrap>{(row.enum_options || []).map((option) => <Tag key={option.option_id}>{option.option_label} = {option.option_value}</Tag>)}</Space>}
        </Space>
      ),
    },
    {
      title: 'Capabilities',
      render: (_: unknown, row: CustomField) => (
        <Space wrap>
          {row.is_visible && <Tag>แสดงผล</Tag>}
          {row.is_searchable && <Tag>ค้นหา</Tag>}
          {row.is_filterable && <Tag>Filter</Tag>}
          {row.is_sortable && <Tag>Sort</Tag>}
          {row.is_aggregatable && <Tag>สถิติ</Tag>}
          {row.is_required && <Tag color="red">จำเป็น</Tag>}
        </Space>
      ),
    },
    {
      title: 'จัดการ',
      render: (_: unknown, row: CustomField) => (
        <Space>
          <Button icon={<PencilIcon className="w-4 h-4" />} onClick={() => openEdit(row)}>แก้ไข</Button>
          <Button
            danger
            icon={<TrashIcon className="w-4 h-4" />}
            onClick={() => Modal.confirm({
              title: 'ปิดใช้งาน Custom Field?',
              content: `${row.display_name || row.field_key} จะไม่แสดงในฟอร์มและตัวกรองใหม่`,
              onOk: async () => {
                await customFieldService.remove(row.field_definition_id);
                await fieldsQuery.refetch();
              },
            })}
          >
            ปิดใช้
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <PageTransition>
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <div>
          <Title level={2}>Custom Fields</Title>
          <Text type="secondary">กำหนดข้อมูลเพิ่มเติมของ Log พร้อมชนิดข้อมูล ตัวเลือก Enum และขอบเขตการใช้งาน</Text>
        </div>
        <Card>
          <Row gutter={16} align="bottom">
            <Col xs={24} md={10}>
              <Text strong>Product</Text>
              <Select
                style={{ width: '100%', marginTop: 8 }}
                placeholder="เลือก Product"
                value={productId}
                loading={productsQuery.isLoading}
                onChange={(value) => { setProductId(value); closeModal(); }}
                options={(productsQuery.data || []).map((item) => ({ value: item.productId, label: `${item.productName} (#${item.productId})` }))}
              />
            </Col>
            <Col><Button type="primary" disabled={!productId} icon={<PlusIcon className="w-4 h-4" />} onClick={openCreate}>เพิ่ม Custom Field</Button></Col>
          </Row>
        </Card>
        {productId && (
          <Card title={`Fields ของ ${product?.productName || 'Product'}`}>
            <Alert
              className="mb-4"
              type="info"
              showIcon
              message="Field ที่เปิดแสดงผลจะปรากฏใน Log Generator และ Logs Explorer"
              description="Enum จะใช้ option_value เป็นค่าที่บันทึก และแสดง option_label ให้ผู้ใช้เลือก/อ่านผล"
            />
            <Table
              rowKey="field_definition_id"
              loading={fieldsQuery.isLoading}
              columns={columns}
              dataSource={fieldsQuery.data || []}
              locale={{ emptyText: <Empty description="ยังไม่มี Custom Field" /> }}
              scroll={{ x: 1200 }}
            />
          </Card>
        )}
        <Modal
          title={editing ? 'แก้ไข Custom Field' : 'สร้าง Custom Field'}
          open={open}
          width={900}
          confirmLoading={saving}
          onCancel={closeModal}
          onOk={() => form.submit()}
          destroyOnHidden
        >
          <Form 
            form={form} 
            layout="vertical" 
            onFinish={save}
            onValuesChange={(changedValues, allValues) => {
              if (changedValues.field_key) {
                const rawKey = changedValues.field_key.trim();
                const key = rawKey.toLowerCase().replace(/[^a-z0-9_]/g, '_');
                
                if (!allValues.display_name) {
                  const displayName = rawKey
                    .replace(/_/g, ' ')
                    .replace(/\b\w/g, (c) => c.toUpperCase());
                  form.setFieldsValue({ display_name: displayName });
                }
                
                if (!allValues.field_path || allValues.field_path.startsWith('custom_fields.')) {
                  form.setFieldsValue({ field_path: `custom_fields.${key}` });
                }
                
                if (!allValues.elastic_field_name || allValues.elastic_field_name.startsWith('payload.custom_fields.')) {
                  form.setFieldsValue({ elastic_field_name: `payload.custom_fields.${key}` });
                }
              }
            }}
          >
            <Row gutter={16}>
              <Col xs={24} md={12}>
                <Form.Item name="project_id" label="Project"><Select allowClear disabled={!!editing} placeholder="ทุก Project" options={(projectsQuery.data || []).map((item) => ({ value: item.projectId, label: item.projectName }))} /></Form.Item>
              </Col>
              <Col xs={24} md={12}>
                <Form.Item name="category_id" label="Category / Feature"><Select allowClear disabled={!projectId || !!editing} placeholder="ทุก Category" options={(featuresQuery.data || []).map((item) => ({ value: item.categoryId, label: item.categoryName }))} /></Form.Item>
              </Col>
              <Col xs={24} md={12}>
                <Form.Item name="field_key" label="Field key" rules={[{ required: true }, { pattern: /^[a-z][a-z0-9_]*$/, message: 'ใช้ตัวพิมพ์เล็ก ตัวเลข และ underscore เท่านั้น' }]}><Input placeholder="customer_tier" disabled={!!editing} /></Form.Item>
              </Col>
              <Col xs={24} md={12}><Form.Item name="display_name" label="ชื่อที่แสดง" rules={[{ required: true }]}><Input placeholder="Customer Tier" /></Form.Item></Col>
              <Col xs={24} md={12}><Form.Item name="data_type" label="Data type" rules={[{ required: true }]}><Select options={['string', 'number', 'boolean', 'datetime', 'json', 'array', 'object', 'enum'].map((value) => ({ value, label: value.toUpperCase() }))} /></Form.Item></Col>
              <Col xs={24} md={12}><Form.Item name="display_order" label="ลำดับแสดงผล"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item></Col>
              <Col xs={24} md={12}><Form.Item name="field_path" label="Field Path" extra="เว้นว่างเพื่อใช้ custom_fields.{field_key}"><Input placeholder="custom_fields.customer_tier" /></Form.Item></Col>
              <Col xs={24} md={12}><Form.Item name="elastic_field_name" label="Elasticsearch Field" extra="เว้นว่างเพื่อสร้างอัตโนมัติ"><Input placeholder="payload.custom_fields.customer_tier" /></Form.Item></Col>
            </Row>

            <Form.List name="enum_options">
              {(fields, { add, remove }) => dataType === 'enum' ? (
                <Card size="small" title="Enum options" style={{ marginBottom: 16 }} extra={<Button type="dashed" onClick={() => add({ is_default: false })}>เพิ่มตัวเลือก</Button>}>
                  {fields.length === 0 && <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="เพิ่มอย่างน้อย 1 ตัวเลือก" />}
                  {fields.map(({ key, name, ...restField }) => (
                    <Row key={key} gutter={8} align="middle">
                      <Form.Item {...restField} name={[name, 'option_id']} hidden><Input /></Form.Item>
                      <Col xs={24} md={4}><Form.Item {...restField} name={[name, 'option_key']} label="Key" rules={[{ required: true }]}><Input placeholder="premium" /></Form.Item></Col>
                      <Col xs={24} md={5}><Form.Item {...restField} name={[name, 'option_label']} label="Label" rules={[{ required: true }]}><Input placeholder="Premium" /></Form.Item></Col>
                      <Col xs={24} md={5}><Form.Item {...restField} name={[name, 'option_value']} label="ค่าที่บันทึก" rules={[{ required: true }]}><Input placeholder="PREMIUM" /></Form.Item></Col>
                      <Col xs={12} md={3}><Form.Item {...restField} name={[name, 'color_code']} label="สี (Hex)"><Input type="color" className="w-full" style={{ padding: '0 4px' }} /></Form.Item></Col>
                      <Col xs={12} md={4}><Form.Item {...restField} name={[name, 'is_default']} label="ค่าเริ่มต้น" valuePropName="checked"><Switch /></Form.Item></Col>
                      <Col xs={24} md={24}><Form.Item {...restField} name={[name, 'description']} label="คำอธิบายเพิ่มเติม (Option)"><Input placeholder="ใช้สำหรับ..." /></Form.Item></Col>
                      <Col xs={24} md={24} className="text-right mb-4 border-b pb-4"><Button danger onClick={() => remove(name)}>ลบตัวเลือกนี้</Button></Col>
                    </Row>
                  ))}
                </Card>
              ) : null}
            </Form.List>

            <Collapse 
              ghost 
              activeKey={activeCollapseKeys}
              onChange={(activeKeys) => {
                const keys = Array.isArray(activeKeys) ? activeKeys : [activeKeys];
                setActiveCollapseKeys(keys);
                if (keys.includes('advanced')) {
                  setBreadcrumbs([
                    { title: 'ผลิตภัณฑ์ที่จัดการ', path: ROUTES.PRODUCTS },
                    { title: 'Custom Fields' },
                    { title: 'Advanced' }
                  ]);
                } else {
                  setBreadcrumbs([
                    { title: 'ผลิตภัณฑ์ที่จัดการ', path: ROUTES.PRODUCTS },
                    { title: 'Custom Fields' }
                  ]);
                }
              }}
            >
              <Collapse.Panel header="Advance" key="advanced" forceRender>
                <Form.Item name="config_json" label="Dynamic Config JSON (Advanced)">
                  <Input.TextArea 
                    rows={4} 
                    placeholder={`{\n  "validation": { "required": true },\n  "transform": [ { "type": "UPPERCASE" } ]\n}`} 
                  />
                </Form.Item>
              </Collapse.Panel>
            </Collapse>
          </Form>
        </Modal>
      </Space>
    </PageTransition>
  );
}
