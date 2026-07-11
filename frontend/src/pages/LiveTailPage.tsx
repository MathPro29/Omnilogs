import { useEffect, useMemo, useRef, useState } from 'react';
import apiClient from '@/api/client';
import { ROUTES } from '@/constants';
import { useAppStore, useAuthStore } from '@/store';

type ProductEnvironment = {
  environmentId: number;
  environmentCode: string;
  environmentName: string;
};

type Product = {
  productId: number;
  productCode: string;
  productName: string;
  productEnvironments?: ProductEnvironment[];
};

type LiveLog = {
  logId: string;
  timestamp?: string | null;
  level?: string | null;
  logType?: string | null;
  message?: string | null;
  method?: string | null;
  path?: string | null;
  traceId?: string | null;
  requestId?: string | null;
  statusCode?: number | null;
  latencyMs?: number | null;
  raw?: Record<string, unknown>;
};

const MAX_LOGS = 200;

function normalizeProduct(value: any): Product {
  return {
    productId: value.product_id ?? value.productId,
    productCode: value.product_code ?? value.productCode,
    productName: value.product_name ?? value.productName,
    productEnvironments: (value.environments ?? value.product_environments ?? value.productEnvironments ?? []).map((env: any) => ({
      environmentId: env.environment_id ?? env.environmentId,
      environmentCode: env.environment_code ?? env.environmentCode,
      environmentName: env.environment_name ?? env.environmentName,
    })),
  };
}

function normalizeLiveLog(value: any): LiveLog {
  return {
    logId: value.log_id ?? value.logId ?? `${Date.now()}-${Math.random()}`,
    timestamp: value.timestamp ?? value['@timestamp'] ?? null,
    level: value.level ?? value.payload?.log_level ?? null,
    logType: value.log_type ?? value.payload?.event_type ?? null,
    message: value.message ?? value.payload?.message ?? null,
    method: value.method ?? value.payload?.request_method ?? null,
    path: value.path ?? value.payload?.request_path ?? null,
    traceId: value.trace_id ?? value.payload?.trace_id ?? null,
    requestId: value.request_id ?? value.payload?.source_request_id ?? null,
    statusCode: value.status_code ?? value.payload?.status_code ?? null,
    latencyMs: value.latency_ms ?? value.payload?.duration_ms ?? null,
    raw: value.raw ?? value,
  };
}

function levelClass(level?: string | null) {
  switch ((level || '').toUpperCase()) {
    case 'ERROR':
    case 'FATAL':
      return 'bg-red-100 text-red-700';
    case 'WARN':
    case 'WARNING':
      return 'bg-amber-100 text-amber-700';
    case 'DEBUG':
      return 'bg-sky-100 text-sky-700';
    default:
      return 'bg-emerald-100 text-emerald-700';
  }
}

export function LiveTailPage() {
  const setBreadcrumbs = useAppStore((state) => state.setBreadcrumbs);
  const accessToken = useAuthStore((state) => state.accessToken);
  const eventSourceRef = useRef<EventSource | null>(null);

  const [products, setProducts] = useState<Product[]>([]);
  const [selectedProductId, setSelectedProductId] = useState<number | null>(null);
  const [selectedEnvironmentId, setSelectedEnvironmentId] = useState<number | null>(null);
  const [logs, setLogs] = useState<LiveLog[]>([]);
  const [isLoadingProducts, setIsLoadingProducts] = useState(false);
  const [isLive, setIsLive] = useState(false);
  const [isPaused, setIsPaused] = useState(false);
  const [connectionState, setConnectionState] = useState<'idle' | 'connecting' | 'connected' | 'error'>('idle');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  useEffect(() => {
    setBreadcrumbs([
      { title: 'Dashboard', path: ROUTES.DASHBOARD },
      { title: 'Live Tail' },
    ]);
  }, [setBreadcrumbs]);

  useEffect(() => {
    let active = true;
    setIsLoadingProducts(true);

    apiClient
      .get('/v1/products')
      .then((response) => {
        if (!active) return;
        const items = ((response.data?.data ?? []) as any[]).map(normalizeProduct);
        setProducts(items);
      })
      .catch((error) => {
        if (!active) return;
        setErrorMessage(error?.message || 'Failed to load products');
      })
      .finally(() => {
        if (active) {
          setIsLoadingProducts(false);
        }
      });

    return () => {
      active = false;
    };
  }, []);

  const selectedProduct = useMemo(
    () => products.find((item) => item.productId === selectedProductId) ?? null,
    [products, selectedProductId]
  );

  const environments = selectedProduct?.productEnvironments ?? [];

  useEffect(() => {
    setSelectedEnvironmentId(environments[0]?.environmentId ?? null);
  }, [selectedProductId, environments]);

  useEffect(() => {
    return () => {
      eventSourceRef.current?.close();
    };
  }, []);

  useEffect(() => {
    eventSourceRef.current?.close();
    eventSourceRef.current = null;

    if (!isLive || !selectedProductId || !accessToken) {
      setConnectionState('idle');
      return;
    }

    const apiBase = import.meta.env.VITE_API_BASE_URL || '/api';
    let url = `${apiBase}/v1/logs/live?product_id=${selectedProductId}&token=${encodeURIComponent(accessToken)}`;
    if (selectedEnvironmentId) {
      url += `&environment_id=${selectedEnvironmentId}`;
    }

    setConnectionState('connecting');
    setErrorMessage(null);

    const source = new EventSource(url);
    eventSourceRef.current = source;

    source.onopen = () => {
      setConnectionState('connected');
    };

    source.onmessage = (event) => {
      if (isPaused) {
        return;
      }

      try {
        const payload = JSON.parse(event.data);
        const list = (Array.isArray(payload) ? payload : [payload]).map(normalizeLiveLog);
        setLogs((current) => [...list, ...current].slice(0, MAX_LOGS));
      } catch (error) {
        setConnectionState('error');
        setErrorMessage(error instanceof Error ? error.message : 'Failed to parse live logs');
      }
    };

    source.onerror = () => {
      setConnectionState('error');
      setErrorMessage('Live stream disconnected or unavailable');
    };

    return () => {
      source.close();
    };
  }, [accessToken, isLive, isPaused, selectedEnvironmentId, selectedProductId]);

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 rounded-2xl border border-slate-200 bg-white p-6 shadow-sm lg:flex-row lg:items-center lg:justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-slate-900">Live Tail</h1>
          <p className="mt-1 text-sm text-slate-500"></p>
        </div>

        <div className="flex flex-wrap items-center gap-3">
          <span className={`rounded-full px-3 py-1 text-xs font-semibold ${
            connectionState === 'connected'
              ? 'bg-emerald-100 text-emerald-700'
              : connectionState === 'connecting'
                ? 'bg-sky-100 text-sky-700'
                : connectionState === 'error'
                  ? 'bg-red-100 text-red-700'
                  : 'bg-slate-100 text-slate-700'
          }`}>
            {connectionState.toUpperCase()}
          </span>
          <button
            type="button"
            onClick={() => setIsLive((value) => !value)}
            disabled={!selectedProductId}
            className="rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
          >
            {isLive ? 'Stop' : 'Start'}
          </button>
          <button
            type="button"
            onClick={() => setIsPaused((value) => !value)}
            disabled={!isLive}
            className="rounded-lg border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 disabled:cursor-not-allowed disabled:text-slate-400"
          >
            {isPaused ? 'Resume' : 'Pause'}
          </button>
          <button
            type="button"
            onClick={() => setLogs([])}
            className="rounded-lg border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700"
          >
            Clear
          </button>
        </div>
      </div>

      <div className="grid gap-4 rounded-2xl border border-slate-200 bg-white p-6 shadow-sm md:grid-cols-3">
        <label className="space-y-2 text-sm">
          <span className="font-medium text-slate-700">Product</span>
          <select
            value={selectedProductId ?? ''}
            onChange={(event) => setSelectedProductId(event.target.value ? Number(event.target.value) : null)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-slate-900"
          >
            <option value="">{isLoadingProducts ? 'Loading products...' : 'Select product'}</option>
            {products.map((product) => (
              <option key={product.productId} value={product.productId}>
                {product.productName} ({product.productCode})
              </option>
            ))}
          </select>
        </label>

        <label className="space-y-2 text-sm">
          <span className="font-medium text-slate-700">Environment</span>
          <select
            value={selectedEnvironmentId ?? ''}
            onChange={(event) => setSelectedEnvironmentId(event.target.value ? Number(event.target.value) : null)}
            disabled={!selectedProductId || environments.length === 0}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-slate-900 disabled:bg-slate-100"
          >
            <option value="">All / Default</option>
            {environments.map((environment) => (
              <option key={environment.environmentId} value={environment.environmentId}>
                {environment.environmentName} ({environment.environmentCode})
              </option>
            ))}
          </select>
        </label>

        <div className="space-y-2 text-sm">
          <span className="font-medium text-slate-700">Window</span>
          <div className="rounded-lg border border-dashed border-slate-300 px-3 py-2 text-slate-500">
            Showing latest {logs.length} / {MAX_LOGS} logs
          </div>
        </div>
      </div>

      {errorMessage ? (
        <div className="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {errorMessage}
        </div>
      ) : null}

      <div className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
        {logs.length === 0 ? (
          <div className="px-6 py-16 text-center text-sm text-slate-500">
            {selectedProductId ? 'No logs yet' : 'Select a product and start live tail'}
          </div>
        ) : (
          <div className="max-h-[70vh] overflow-auto">
            {logs.map((log) => (
              <div key={`${log.logId}-${log.timestamp ?? ''}`} className="border-b border-slate-100 px-6 py-4 last:border-b-0">
                <div className="flex flex-wrap items-center gap-2 text-xs">
                  <span className="rounded bg-slate-100 px-2 py-1 font-mono text-slate-700">
                    {log.timestamp ? new Date(log.timestamp).toLocaleString() : '-'}
                  </span>
                  <span className={`rounded px-2 py-1 font-semibold ${levelClass(log.level)}`}>
                    {(log.level || 'INFO').toUpperCase()}
                  </span>
                  {log.logType ? <span className="rounded bg-slate-100 px-2 py-1 text-slate-700">{log.logType}</span> : null}
                  {log.method || log.path ? (
                    <span className="text-slate-500">{`${log.method || 'LOG'} ${log.path || ''}`.trim()}</span>
                  ) : null}
                </div>

                <div className="mt-2 whitespace-pre-wrap break-words font-mono text-sm text-slate-900">
                  {log.message || JSON.stringify(log.raw)}
                </div>

                <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-slate-500">
                  {log.traceId ? <span>trace: {log.traceId}</span> : null}
                  {log.requestId ? <span>request: {log.requestId}</span> : null}
                  {typeof log.statusCode === 'number' ? <span>status: {log.statusCode}</span> : null}
                  {typeof log.latencyMs === 'number' ? <span>latency: {log.latencyMs} ms</span> : null}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
