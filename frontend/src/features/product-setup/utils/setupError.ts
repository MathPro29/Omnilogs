import axios from "axios";

interface ApiErrorEnvelope {
  message?: string;
}

export function getSetupErrorMessage(
  error: unknown,
  fallback: string,
): string {
  if (axios.isAxiosError<ApiErrorEnvelope>(error)) {
    return error.response?.data?.message ?? error.message ?? fallback;
  }
  return error instanceof Error ? error.message : fallback;
}
