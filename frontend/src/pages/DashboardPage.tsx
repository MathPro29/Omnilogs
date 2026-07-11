import { PageTransition } from '@/components';

export function DashboardPage() {
  return (
    <PageTransition>
      <div style={{ textAlign: 'center', padding: '60px 20px' }}>
        <p style={{ color: '#64748b', fontSize: '14px', margin: 0 }}>Current page</p>
        <h1 style={{ color: '#020617', fontSize: '32px', fontWeight: 700, margin: '8px 0 0' }}>
          Overview
        </h1>
      </div>
    </PageTransition>
  );
}
