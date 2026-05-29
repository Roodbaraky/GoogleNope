export interface ApiErrorBody {
  error?: string;
}

export function apiErrorMessage(error: unknown, fallback = 'Request failed'): string {
  if (typeof error === 'object' && error !== null && 'error' in error) {
    const httpError = error as { error?: ApiErrorBody | string; message?: string };
    if (typeof httpError.error === 'string') {
      return httpError.error;
    }

    if (httpError.error?.error) {
      return httpError.error.error;
    }

    if (httpError.message) {
      return httpError.message;
    }
  }

  return fallback;
}
