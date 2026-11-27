/**
 * Typed client for the vecai Go backend (docs/api.md contract).
 *
 * The backend is a separate Gin server (default http://localhost:8080) that
 * speaks a JSON envelope: {"success": boolean, "message"?, "data"?, "error"?}.
 * This module unwraps that envelope so callers work with plain data or a
 * typed ApiError.
 *
 * The base URL comes from NEXT_PUBLIC_API_URL so production builds can point
 * at a deployed API without code changes; local dev falls back to :8080.
 */

export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export type Envelope<T> = {
  success: boolean;
  message?: string;
  data?: T;
  error?: string;
};

/** Error thrown by apiFetch when the backend reports a failure. */
export class ApiError extends Error {
  readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

/**
 * POST JSON to the backend and unwrap the response envelope.
 * Throws ApiError with the backend's error text on non-2xx responses
 * (e.g. 400 "invalid email address format", 401 "authentication required").
 */
export async function apiPost<T>(
  path: string,
  body: unknown,
  token?: string,
): Promise<T> {
  return apiRequest<T>("POST", path, {
    body: JSON.stringify(body),
    token,
  });
}

/** GET a resource from the backend, optionally authenticated. */
export async function apiGet<T>(path: string, token?: string): Promise<T> {
  return apiRequest<T>("GET", path, { token });
}

/** PUT a resource to the backend, optionally authenticated. */
export async function apiPut<T>(
  path: string,
  body: unknown,
  token?: string,
): Promise<T> {
  return apiRequest<T>("PUT", path, { body: JSON.stringify(body), token });
}

/** DELETE a resource from the backend, optionally authenticated. */
export async function apiDelete<T>(
  path: string,
  token?: string,
): Promise<T> {
  return apiRequest<T>("DELETE", path, { token });
}

async function apiRequest<T>(
  method: "GET" | "POST" | "PUT" | "DELETE",
  path: string,
  options: { body?: string; token?: string },
): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${path}`, {
    method,
    headers: {
      "Content-Type": "application/json",
      ...(options.token ? { Authorization: `Bearer ${options.token}` } : {}),
    },
    body: options.body,
  });

  const json = (await res.json().catch(() => null)) as Envelope<T> | null;

  if (!res.ok || !json?.success) {
    throw new ApiError(
      res.status,
      json?.error ?? "Unable to reach the server. Please try again.",
    );
  }
  // Some DELETE/PUT may return null data
  return (json.data ?? (null as T));
}
