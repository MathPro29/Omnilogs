import axios from "axios";

interface ApiErrorBody {
  message?: string;
  error?: {
    message?: string;
  };
}

export function getErrorMessage(error: unknown, fallback: string): string {
  if (axios.isAxiosError<ApiErrorBody>(error)) {
    return (
      error.response?.data?.error?.message ??
      error.response?.data?.message ??
      error.message ??
      fallback
    );
  }
  return error instanceof Error ? error.message : fallback;
}
