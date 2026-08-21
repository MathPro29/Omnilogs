function withoutTrailingSlash(value: string): string {
  return value.replace(/\/+$/, "");
}

export function getOmniLogsBaseUrl(): string {
  const configured = import.meta.env.VITE_INGEST_BASE_URL?.trim();
  if (configured) return withoutTrailingSlash(configured);

  // In local development the Vite UI runs on a different port from the API.
  // Reuse an absolute API base URL so ingestion is sent to the backend rather
  // than to the frontend dev server.
  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL?.trim();
  if (apiBaseUrl && /^https?:\/\//i.test(apiBaseUrl)) {
    const url = new URL(apiBaseUrl);
    url.pathname = url.pathname.replace(/\/api(?:\/v1)?\/?$/, "") || "/";
    return withoutTrailingSlash(url.toString());
  }

  if (typeof window !== "undefined") return withoutTrailingSlash(window.location.origin);
  return "";
}

export function getOmniLogsIngestUrl(): string {
  return `${getOmniLogsBaseUrl()}/api/v1/ingest/logs`;
}
