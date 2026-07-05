import { BrowserRouter } from 'react-router-dom';
import { ConfigProvider } from 'antd';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { antdThemeConfig, antdLocale } from '@/app/theme';
import { AppRoutes } from '@/routes';
import { AutoLogProcessor } from '@/components/AutoLogProcessor';

// ===== Query Client =====
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000,
      retry: 1,
      refetchOnWindowFocus: false,
      refetchInterval: 3000, // Real-time update every 3 seconds globally
    },
  },
});

// ===== App Entry Point =====
export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ConfigProvider theme={antdThemeConfig} locale={antdLocale}>
        <BrowserRouter>
          <AutoLogProcessor />
          <AppRoutes />
        </BrowserRouter>
      </ConfigProvider>
    </QueryClientProvider>
  );
}
