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
      staleTime: 5 * 60 * 1000, // 5 นาที
      retry: 1,
      refetchOnWindowFocus: false,
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
