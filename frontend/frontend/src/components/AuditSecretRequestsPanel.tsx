import { useEffect, useState } from 'react';
import { Button, Card, Form, Input, Modal, Select, Space, Table, Tag, Typography, message } from 'antd';
import { sensitiveLogService } from '@/services';
import { useAuthStore } from '@/store';

const { Text } = Typography;

export function AuditSecretRequestsPanel({ canReview }: { canReview: boolean }) {
  const currentUser = useAuthStore((state) => state.currentUser);
  const [audits, setAudits] = useState<any[]>([]);
  const [auditId, setAuditId] = useState<string>();
  const [secrets, setSecrets] = useState<any[]>([]);
  const [pending, setPending] = useState<any[]>([]);
  const [loadingSecrets, setLoadingSecrets] = useState(false);
  const [loadingPending, setLoadingPending] = useState(false);
  const [revealRequest, setRevealRequest] = useState<any | null>(null);
  const [password, setPassword] = useState('');
  const [revealedValue, setRevealedValue] = useState<any | null>(null);
  const [revealing, setRevealing] = useState(false);

  const loadPending = async () => {
    if (!canReview) return;
    setLoadingPending(true);
    try { setPending(await sensitiveLogService.listPendingAuditRequests()); }
    catch (error: any) { message.error(error.message || 'Unable to load pending audit requests'); }
    finally { setLoadingPending(false); }
  };

  useEffect(() => {
    sensitiveLogService.listAuditLogs().then(setAudits).catch(() => message.error('Unable to load audit logs'));
  }, []);
  useEffect(() => { loadPending(); }, [canReview]);

  const selectAudit = async (value: string) => {
    setAuditId(value); setSecrets([]); setLoadingSecrets(true);
    try { setSecrets(await sensitiveLogService.listAuditSecrets(value)); }
    catch (error: any) { message.error(error.message || 'Unable to load protected fields'); }
    finally { setLoadingSecrets(false); }
  };
  const submit = async (values: { secretId: string; reason: string }) => {
    if (!auditId) return;
    try {
      await sensitiveLogService.createAuditRequest(auditId, values.secretId, values.reason);
      message.success('Audit secret request submitted');
      await loadPending();
    } catch (error: any) { message.error(error.message || 'Unable to submit audit request'); }
  };
  const review = async (record: any, status: 'APPROVED' | 'REJECTED') => {
    try { await sensitiveLogService.reviewAuditRequest(record.audit_id, record.request_id, status); message.success(`Request ${status.toLowerCase()}`); await loadPending(); }
    catch (error: any) { message.error(error.message || 'Unable to review audit request'); }
  };
  const reveal = async () => {
    if (!revealRequest || !password) return;
    setRevealing(true);
    try { setRevealedValue(await sensitiveLogService.revealAuditSecret(revealRequest.audit_id, revealRequest.request_id, password)); }
    catch (error: any) { message.error(error.message || 'Unable to reveal audit secret'); }
    finally { setRevealing(false); }
  };
  const canReveal = (record: any) => record.approval_status === 'APPROVED' && String(record.user_id) === String(currentUser?.id);

  return <Card title="Audit secret access" extra={<Button onClick={loadPending} loading={loadingPending}>Refresh requests</Button>}>
    <Form layout="vertical" onFinish={submit} className="mb-6">
      <Form.Item label="Audit event" required><Select value={auditId} onChange={selectAudit} loading={!audits.length} placeholder="Choose an audit event" options={audits.map((audit) => ({ value: audit.audit_id, label: `${audit.action || 'Audit event'} - ${audit.path || audit.audit_id.slice(0, 8)}` }))} /></Form.Item>
      <Form.Item name="secretId" label="Protected field" rules={[{ required: true, message: 'Choose a protected field' }]}><Select disabled={!auditId} loading={loadingSecrets} placeholder="Choose a protected field" options={secrets.map((secret) => ({ value: secret.secret_id, label: `${secret.field_key} (${secret.source_section})` }))} /></Form.Item>
      <Form.Item name="reason" label="Reason" rules={[{ required: true }]}><Input.TextArea rows={3} placeholder="Explain why you need this value" /></Form.Item>
      <Button type="primary" htmlType="submit" disabled={!auditId}>Submit audit secret request</Button>
    </Form>
    <Table rowKey="request_id" loading={loadingPending} dataSource={pending} scroll={{ x: 700 }} columns={[{ title: 'Audit event', dataIndex: 'audit_id', render: (v: string) => <Text copyable={{ text: v }}>{v.slice(0, 8)}...</Text> }, { title: 'Protected field', dataIndex: 'secret_id', render: (v: string) => <Text copyable={{ text: v }}>{v.slice(0, 8)}...</Text> }, { title: 'Requester', dataIndex: 'user_id' }, { title: 'Reason', dataIndex: 'reason', ellipsis: true }, { title: 'Status', dataIndex: 'approval_status', render: (v: string) => <Tag color={v === 'APPROVED' ? 'green' : v === 'REJECTED' ? 'red' : 'gold'}>{v}</Tag> }, { title: 'Action', render: (_: unknown, record: any) => record.approval_status === 'PENDING' && canReview ? <Space><Button size="small" type="primary" onClick={() => review(record, 'APPROVED')}>Approve</Button><Button size="small" danger onClick={() => review(record, 'REJECTED')}>Reject</Button></Space> : canReveal(record) ? <Button size="small" onClick={() => { setRevealRequest(record); setPassword(''); setRevealedValue(null); }}>Reveal value</Button> : <Text type="secondary">{record.approval_status === 'PENDING' ? 'Awaiting review' : 'Reviewed'}</Text> }]} />
    <Modal title="Reveal audit secret" open={!!revealRequest} onCancel={() => setRevealRequest(null)} onOk={reveal} okText="Reveal" confirmLoading={revealing} destroyOnClose>
      {!revealedValue ? <Input.Password value={password} onChange={(e) => setPassword(e.target.value)} placeholder="Confirm your password" /> : <pre className="rounded bg-slate-950 p-3 text-xs text-slate-100 overflow-auto">{JSON.stringify(revealedValue.value, null, 2)}</pre>}
    </Modal>
  </Card>;
}
