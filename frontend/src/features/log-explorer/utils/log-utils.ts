import type { SearchRecord } from "@/features/log-explorer/services/log-search.service";

export const getRecordId = (record: SearchRecord, index = 0) =>
  String(record.id ?? record.log_id ?? `${record["@timestamp"] ?? record.timestamp ?? "log"}-${index}`);

export const getValue = (record: SearchRecord, ...paths: string[]) => {
  const raw = record.raw && typeof record.raw === "object" ? (record.raw as Record<string, unknown>) : undefined;
  const payload = record.payload && typeof record.payload === "object" ? (record.payload as Record<string, unknown>) : undefined;
  const data = record.data && typeof record.data === "object" ? (record.data as Record<string, unknown>) : undefined;
  const rawPayload = raw?.payload && typeof raw.payload === "object" ? (raw.payload as Record<string, unknown>) : undefined;
  const rawData = raw?.data && typeof raw.data === "object" ? (raw.data as Record<string, unknown>) : undefined;

  const sources = [record, payload, data, raw, rawPayload, rawData].filter(Boolean) as Record<string, unknown>[];

  for (const path of paths) {
    for (const source of sources) {
      const value = path.split(".").reduce<unknown>((current, key) => {
        if (current && typeof current === "object") return (current as Record<string, unknown>)[key];
        return undefined;
      }, source);
      if (value !== undefined && value !== null && value !== "") return String(value);
    }
  }
  return "-";
};


const isGenericCompletionMessage = (value: string) =>
  /^(?:http\s+)?request completed\.?$/i.test(value.trim());

const findReason = (value: unknown): string | undefined => {
  if (typeof value === "string") {
    const trimmed = value.trim();
    if (!trimmed || isGenericCompletionMessage(trimmed)) return undefined;
    try {
      return findReason(JSON.parse(trimmed)) ?? trimmed;
    } catch {
      return trimmed;
    }
  }
  if (!value || typeof value !== "object") return undefined;
  if (Array.isArray(value)) {
    const parts = value.map((item) => findReason(item)).filter((item): item is string => Boolean(item));
    return parts.length ? parts.join("; ") : undefined;
  }
  const object = value as Record<string, unknown>;
  const priorityKeys = ["error_message", "errorMessage", "failure_reason", "reason", "Message", "message", "error", "err", "warning", "warn", "detail", "details", "fail"];
  for (const key of priorityKeys) {
    const candidate = object[key];
    if (typeof candidate === "string" && candidate.trim() && !isGenericCompletionMessage(candidate)) return candidate;
    if (candidate && typeof candidate === "object") {
      const nested = findReason(candidate);
      if (nested) return nested;
    }
  }
  for (const child of Object.values(object)) {
    const nested = findReason(child);
    if (nested) return nested;
  }
  return undefined;
};
export interface LogAuthor {
  id?: string;
  name?: string;
  role?: string;
}

// Normalize common author shapes so old and new log sources render consistently.
export const getLogAuthor = (record: SearchRecord): LogAuthor => {
  const raw = record.raw && typeof record.raw === "object" ? (record.raw as Record<string, unknown>) : {};
  const payload = record.payload && typeof record.payload === "object" ? (record.payload as Record<string, unknown>) : {};
  const data = record.data && typeof record.data === "object" ? (record.data as Record<string, unknown>) : {};
  const rawPayload = raw.payload && typeof raw.payload === "object" ? (raw.payload as Record<string, unknown>) : {};
  const rawData = raw.data && typeof raw.data === "object" ? (raw.data as Record<string, unknown>) : {};
  const metadata = record.metadata && typeof record.metadata === "object" ? (record.metadata as Record<string, unknown>) : {};
  const sources = [record, payload, data, raw, rawPayload, rawData, metadata] as Record<string, unknown>[];
  const actorValues = sources.map((source) => source.actor).filter((value) => value !== undefined && value !== null && value !== "");
  const nested = sources.flatMap((source) =>
    [source.author, source.actor, source.user, (source.metadata as Record<string, unknown> | undefined)?.actor].filter(
      (value): value is Record<string, unknown> => Boolean(value) && typeof value === "object" && !Array.isArray(value),
    ),
  );
  const value = (keys: string[]) => {
    for (const source of [...nested, ...sources]) {
      for (const key of keys) {
        const item = source[key];
        if (item !== undefined && item !== null && item !== "") return String(item);
      }
    }
    return undefined;
  };

  return {
    id: value(["id", "user_id", "userId", "author_id", "authorId", "actor_id", "actorId", "session_actor_id", "sessionActorId"]) ?? (actorValues.length ? String(actorValues[0]) : undefined),
    name: value(["name", "username", "userName", "full_name", "fullName", "display_name", "displayName", "author_name", "authorName", "actor_name", "actorName"]),
    role: value(["role", "role_name", "roleName", "user_role", "userRole", "author_role", "authorRole", "actor_role", "actorRole", "actor_type", "actorType", "type"]),
  };
};

export const compactRoutePattern = (value: string) => {
  const path = value.trim().replace(/^https?:\/\/[^/]+/i, "").split("?")[0];
  if (!path || path === "-") return value;

  const segments = path.split("/").filter(Boolean);
  if (segments[0] === "api") segments.shift();
  if (segments.length <= 2) return segments.join("/");
  return `${segments[0]}/${segments.at(-1)}`;
};

export const getCustomFieldValue = (record: SearchRecord, fieldPath: string) => {
  const normalizedPath = fieldPath.replace(/^\$\.?/, "").replace(/^payload\.custom_fields\.?/, "").replace(/^custom_fields\.?/, "");
  if (!normalizedPath) return getValue(record, "custom_fields", "payload.custom_fields", "data.custom_fields");
  return getValue(
    record,
    `custom_fields.${normalizedPath}`,
    `payload.custom_fields.${normalizedPath}`,
    `data.custom_fields.${normalizedPath}`,
    normalizedPath,
    `payload.${normalizedPath}`,
    `data.${normalizedPath}`,
  );
};

const sensitivePattern = /(password|passwd|secret|token|authorization|api[_-]?key|cookie|session)/i;

export const maskSensitive = (value: unknown, key = ""): unknown => {
  if (sensitivePattern.test(key)) return "********";
  if (Array.isArray(value)) return value.map((item) => maskSensitive(item));
  if (value && typeof value === "object") {
    return Object.fromEntries(
      Object.entries(value as Record<string, unknown>).map(([childKey, childValue]) => [
        childKey,
        maskSensitive(childValue, childKey),
      ]),
    );
  }
  return value;
};

export interface LogChangeItem {
  path: string;
  before: unknown;
  after: unknown;
}

const changeObject = (value: unknown): Record<string, unknown> | undefined =>
  value && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : undefined;

const flattenChangeValue = (value: unknown, prefix = "", result = new Map<string, unknown>()) => {
  if (Array.isArray(value)) {
    if (!value.length) result.set(prefix || "$", value);
    value.forEach((child, index) => flattenChangeValue(child, `${prefix}[${index}]`, result));
    return result;
  }
  const object = changeObject(value);
  if (!object) {
    result.set(prefix || "$", value);
    return result;
  }
  const entries = Object.entries(object);
  if (!entries.length) result.set(prefix || "$", value);
  for (const [key, child] of entries) {
    const path = prefix ? `${prefix}.${key}` : key;
    if (changeObject(child)) flattenChangeValue(child, path, result);
    else result.set(path, child);
  }
  return result;
};

const equalChangeValue = (left: unknown, right: unknown) =>
  JSON.stringify(left) === JSON.stringify(right);

const compareChangePair = (before: unknown, after: unknown): LogChangeItem[] => {
  const beforeFields = flattenChangeValue(before);
  const afterFields = flattenChangeValue(after);
  const paths = new Set([...beforeFields.keys(), ...afterFields.keys()]);
  return [...paths]
    .filter((path) => !equalChangeValue(beforeFields.get(path), afterFields.get(path)))
    .sort()
    .map((path) => ({ path, before: beforeFields.get(path), after: afterFields.get(path) }));
};

const extractLogChanges = (...sources: unknown[]): LogChangeItem[] => {
  const result: LogChangeItem[] = [];
  const pairAliases = [
    ["before", "after"],
    ["old", "new"],
    ["old_data", "new_data"],
    ["before_data", "after_data"],
    ["before_update", "after_update"],
    ["before_value", "after_value"],
    ["beforeValue", "afterValue"],
    ["old_value", "new_value"],
    ["oldValue", "newValue"],
    ["previous", "current"],
    ["original", "updated"],
    ["from", "to"],
  ] as const;
  const visited = new Set<Record<string, unknown>>();

  const inspect = (source: unknown) => {
    const object = changeObject(source);
    if (!object || visited.has(object)) return;
    visited.add(object);

    for (const [beforeKey, afterKey] of pairAliases) {
      if (beforeKey in object && afterKey in object) {
        result.push(...compareChangePair(object[beforeKey], object[afterKey]));
      }
    }

    const changes = changeObject(
      object.changes ?? object.change_list ?? object.diff ?? object.modifications,
    );
    if (changes) {
      for (const [path, value] of Object.entries(changes)) {
        const item = changeObject(value);
        if (!item) continue;
        const before =
          item.before ??
          item.old ??
          item.old_value ??
          item.before_value ??
          item.beforeValue ??
          item.from ??
          item.original;
        const after =
          item.after ??
          item.new ??
          item.new_value ??
          item.after_value ??
          item.afterValue ??
          item.to ??
          item.updated;
        if (!equalChangeValue(before, after)) result.push({ path, before, after });
      }
    }

    for (const key of [
      "metadata",
      "audit",
      "change",
      "changes",
      "diff",
      "request",
      "response",
      "body",
      "data",
      "payload",
      "custom_fields",
    ]) {
      inspect(object[key]);
    }
  };

  sources.forEach(inspect);
  const unique = new Map<string, LogChangeItem>();
  for (const item of result) {
    unique.set(`${item.path}|${JSON.stringify(item.before)}|${JSON.stringify(item.after)}`, item);
  }
  return [...unique.values()];
};

export const fieldGroups = (record: SearchRecord) => {
  const raw = ((record.raw ?? {}) as Record<string, unknown>) || {};
  const parseValue = (value: unknown): unknown => {
    if (typeof value === "string") {
      try {
        return parseValue(JSON.parse(value));
      } catch {
        return value;
      }
    }
    if (Array.isArray(value)) return value.map(parseValue);
    if (value && typeof value === "object") {
      return Object.fromEntries(
        Object.entries(value as Record<string, unknown>).map(([key, child]) => [
          key,
          parseValue(child),
        ]),
      );
    }
    return value;
  };
  const asObject = (value: unknown): Record<string, unknown> | undefined => {
    const parsed = parseValue(value);
    return parsed && typeof parsed === "object" && !Array.isArray(parsed)
      ? (parsed as Record<string, unknown>)
      : undefined;
  };
  const compact = (value: Record<string, unknown>) =>
    Object.fromEntries(
      Object.entries(value)
        .filter(([, item]) => item !== undefined && item !== null && item !== "")
        .map(([key, item]) => [key, parseValue(item)]),
    );
  const merge = (...values: Array<Record<string, unknown> | undefined>) =>
    Object.assign({}, ...values.filter(Boolean));
  const omit = (value: Record<string, unknown>, keys: string[]) =>
    Object.fromEntries(Object.entries(value).filter(([key]) => !keys.includes(key)));

  const payload =
    asObject(record.payload) ??
    asObject(record.data) ??
    asObject(raw.data) ??
    asObject(raw.payload) ??
    {};
  const request = merge(
    compact({
      request_id: record.request_id ?? payload.source_request_id,
      method: record.method ?? payload.request_method,
      path: record.path ?? payload.request_path,
      url: record.url ?? payload.url,
      headers: record.request_headers ?? payload.request_headers,
      query: record.query_params ?? payload.query_params,
      body: record.request_payload ?? record.request_body ?? payload.request_payload ?? payload.request_body,
    }),
    asObject(payload.request),
    asObject(record.request),
  );
  const response = merge(
    compact({
      status_code: record.status_code ?? payload.status_code,
      duration_ms: record.latency_ms ?? payload.duration_ms,
      headers: record.response_headers ?? payload.response_headers,
      body: record.response_payload ?? record.response_body ?? payload.response_payload ?? payload.response_body,
      error_code: record.error_code ?? payload.error_code,
      error_message: record.error_message ?? payload.error_message,
    }),
    asObject(payload.response),
    asObject(record.response),
  );

  const requestOverview = omit(request, ["headers", "query", "query_params", "body", "bodyText"]);
  const responseOverview = omit(response, ["headers", "body", "bodyText", "error", "error_code", "error_message", "stack_trace"]);
  const responseError = compact({
    error: response.error,
    error_code: response.error_code ?? record.error_code ?? payload.error_code,
    error_message: response.error_message ?? record.error_message ?? payload.error_message,
    stack_trace: response.stack_trace ?? record.stack_trace ?? payload.stack_trace,
  });
  const changes = extractLogChanges(payload, record, request, response);
  const author = getLogAuthor(record);

  return {
    payload: parseValue(payload),
    request: requestOverview,
    requestHeaders: parseValue(request.headers ?? record.request_headers ?? payload.request_headers ?? {}),
    requestQuery: parseValue(request.query ?? request.query_params ?? record.query_params ?? payload.query_params ?? {}),
    requestBody: parseValue(request.body ?? record.request_payload ?? record.request_body ?? payload.request_payload ?? payload.request_body ?? payload.body ?? {}),
    response: responseOverview,
    responseHeaders: parseValue(response.headers ?? record.response_headers ?? payload.response_headers ?? {}),
    responseBody: parseValue(response.body ?? record.response_payload ?? record.response_body ?? payload.response_payload ?? payload.response_body ?? {}),
    responseError: parseValue(responseError),
    changes,
    author: compact({ ...author }),
    metadata: parseValue(payload.metadata ?? record.metadata ?? compact({ product_id: record.product_id ?? raw.product_id, environment_id: record.environment_id ?? raw.environment_id, source_id: record.source_id ?? raw.source_id, log_type: record.log_type ?? payload.event_type })),
    trace: parseValue(payload.trace ?? record.trace ?? compact({ request_id: record.request_id ?? payload.source_request_id, trace_id: record.trace_id ?? payload.trace_id, correlation_id: payload.correlation_id, span_id: payload.span_id, parent_span_id: payload.parent_span_id })),
    custom: parseValue(record.custom_fields ?? payload.custom_fields ?? {}),
  };
};

const httpStatusMessages: Record<number, string> = {
  400: "Bad Request",
  401: "Unauthorized",
  402: "Payment Required",
  403: "Forbidden",
  404: "Not Found",
  405: "Method Not Allowed",
  406: "Not Acceptable",
  408: "Request Timeout",
  409: "Conflict",
  410: "Gone",
  415: "Unsupported Media Type",
  422: "Unprocessable Entity",
  429: "Too Many Requests",
  500: "Internal Server Error",
  501: "Not Implemented",
  502: "Bad Gateway",
  503: "Service Unavailable",
  504: "Gateway Timeout",
};

export function getHttpStatusText(statusCode: number): string {
  return httpStatusMessages[statusCode] || (statusCode >= 500 ? "Server Error" : statusCode >= 400 ? "Client Error" : "OK");
}

export interface LogErrorInfo {
  severity: "error" | "warn" | "info";
  statusCode?: number;
  statusText?: string;
  level?: string;
  errorTitle?: string;
  errorMessage?: string;
}

export function getLogErrorInfo(record: SearchRecord): LogErrorInfo {
  const groups = fieldGroups(record);
  const resp = groups.response as Record<string, unknown>;

  const rawStatusCode =
    resp.status_code ??
    record.status_code ??
    (record.payload as Record<string, unknown> | undefined)?.status_code ??
    (record.payload as Record<string, unknown> | undefined)?.status;

  const statusCode = typeof rawStatusCode === "number" ? rawStatusCode : Number(rawStatusCode);
  const validStatusCode = !isNaN(statusCode) && statusCode > 0 ? statusCode : undefined;

  let levelVal = getValue(
    record,
    "level",
    "log_level",
    "severity",
    "logType",
    "payload.log_level",
    "payload.level",
    "data.level",
    "data.log_level",
  ).toUpperCase();

  const msgVal = getValue(
    record,
    "message",
    "msg",
    "log",
    "text",
    "description",
    "error",
    "error_message",
    "payload.message",
    "data.message",
  );

  if ((!levelVal || levelVal === "—" || levelVal === "-") && msgVal !== "—") {
    const match = msgVal.match(
      /^(CRITICAL|FATAL|ERROR|WARN|WARNING|INFO|DEBUG|TRACE):?/i,
    );
    if (match) {
      levelVal = match[1].toUpperCase();
      if (levelVal === "WARNING") levelVal = "WARN";
    }
  }

  // Prefer structured error data over generic access-log messages.
  const errorMsgCandidate = getValue(
    record,
    "error_message",
    "payload.error_message",
    "responseError.error_message",
    "response.error_message",
    "data.error_message",
    "payload.error",
    "error",
    "reason",
    "payload.reason",
    "warning",
    "payload.warning",
    "warn",
    "payload.warn",
    "response.message",
    "response.body.message",
    "response.body.error_message",
    "response.body.error",
    "response.body.reason",
    "response.body.detail",
    "response.warn",
    "data.warn",
  );

  const nestedReason =
    findReason(groups.responseBody) ??
    findReason(record.response) ??
    findReason(record.payload) ??
    findReason(record.data) ??
    findReason(record.raw);
  const errorMessage = errorMsgCandidate !== "-" && errorMsgCandidate !== "—" ? errorMsgCandidate : nestedReason;

  let severity: "error" | "warn" | "info" = "info";

  if ((validStatusCode && validStatusCode >= 500) || ["ERROR", "FATAL", "CRITICAL", "ERR"].includes(levelVal)) {
    severity = "error";
  } else if ((validStatusCode && validStatusCode >= 400) || ["WARN", "WARNING"].includes(levelVal)) {
    severity = "warn";
  }

  const statusText = validStatusCode ? getHttpStatusText(validStatusCode) : undefined;
  
  let errorTitle = undefined;
  if (validStatusCode) {
    errorTitle = `HTTP ${validStatusCode} ${statusText}`;
  } else if (severity === "warn") {
    errorTitle = `Warning Log (${levelVal || "WARN"})`;
  } else if (severity === "error") {
    errorTitle = `Error Log (${levelVal || "ERROR"})`;
  }

  return {
    severity,
    statusCode: validStatusCode,
    statusText,
    level: levelVal !== "-" && levelVal !== "—" ? levelVal : undefined,
    errorTitle,
    errorMessage: errorMessage || (severity !== "info" && msgVal !== "-" && msgVal !== "—" ? msgVal : undefined),
  };
}
