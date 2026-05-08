/**
 * Base API client for the Mortar server.
 *
 * All frontend code must call the Go server through this client.
 * No frontend code should call upstream services (Jellyfin, Sonarr, etc.) directly.
 *
 * Feature-specific methods are added in later phases as sub-modules of this client.
 */

/** The base URL of the Mortar Go server. Override via environment in development. */
const BASE_URL: string =
  process.env['EXPO_PUBLIC_MORTAR_API_URL'] ?? 'http://localhost:3000';

/** HTTP methods supported by the client. */
type Method = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';

/** Options for a client request, excluding method and body (handled per-function). */
interface RequestOptions {
  /** Additional headers to merge into the request. */
  headers?: Record<string, string>;
  /** AbortSignal for cancellation. */
  signal?: AbortSignal;
}

/**
 * MortarAPIError is thrown when the server returns a non-2xx response.
 */
export class MortarAPIError extends Error {
  constructor(
    public readonly status: number,
    public readonly statusText: string,
    message: string,
  ) {
    super(message);
    this.name = 'MortarAPIError';
  }
}

/**
 * Performs an HTTP request against the Mortar server and returns the parsed
 * JSON body. Throws MortarAPIError on non-2xx responses.
 */
async function request<T>(
  method: Method,
  path: string,
  body?: unknown,
  options: RequestOptions = {},
): Promise<T> {
  const url = `${BASE_URL}${path}`;
  const headers: Record<string, string> = {
    ...options.headers,
  };

  // Only include Content-Type when there is a body to send.
  if (body !== undefined) {
    headers['Content-Type'] = 'application/json';
  }

  const init: RequestInit = {
    method,
    headers,
    credentials: 'include', // include HttpOnly session cookie
    signal: options.signal,
  };

  if (body !== undefined) {
    init.body = JSON.stringify(body);
  }

  const response = await fetch(url, init);

  if (!response.ok) {
    let message = response.statusText;
    try {
      const err = (await response.json()) as { message?: string };
      if (err.message) message = err.message;
    } catch {
      // ignore parse errors; keep statusText as message
    }
    throw new MortarAPIError(response.status, response.statusText, message);
  }

  // 204 No Content: return undefined cast as T
  if (response.status === 204) {
    return undefined as unknown as T;
  }

  return response.json() as Promise<T>;
}

/**
 * The Mortar API client.
 * Provides low-level GET / POST / PATCH / DELETE helpers.
 * Feature modules extend this with typed methods.
 */
export const api = {
  get<T>(path: string, options?: RequestOptions): Promise<T> {
    return request<T>('GET', path, undefined, options);
  },

  post<T>(path: string, body?: unknown, options?: RequestOptions): Promise<T> {
    return request<T>('POST', path, body, options);
  },

  patch<T>(path: string, body?: unknown, options?: RequestOptions): Promise<T> {
    return request<T>('PATCH', path, body, options);
  },

  put<T>(path: string, body?: unknown, options?: RequestOptions): Promise<T> {
    return request<T>('PUT', path, body, options);
  },

  delete<T>(path: string, options?: RequestOptions): Promise<T> {
    return request<T>('DELETE', path, undefined, options);
  },
};
