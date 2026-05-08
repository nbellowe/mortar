/**
 * Requests API module.
 * Covers search, request listing, request submission, and request detail.
 * All calls go to the Mortar server — never directly to upstream services.
 */

import { api } from './client';
import { MediaItem, Request } from '../types/requests';

/**
 * Search for media across all plugins with a requests.* capability.
 * Results are deduplicated by external ID on the server.
 */
export async function searchMedia(query: string, signal?: AbortSignal): Promise<MediaItem[]> {
  const encoded = encodeURIComponent(query);
  return api.get<MediaItem[]>(`/api/search?q=${encoded}`, { signal });
}

/**
 * List requests, optionally filtered by requester and/or status.
 */
export async function fetchRequests(options?: {
  requesterId?: string;
  status?: string;
  signal?: AbortSignal;
}): Promise<Request[]> {
  const params = new URLSearchParams();
  if (options?.requesterId) params.set('requester_id', options.requesterId);
  if (options?.status) params.set('status', options.status);
  const qs = params.toString();
  return api.get<Request[]>(`/api/requests${qs ? `?${qs}` : ''}`, { signal: options?.signal });
}

/**
 * Submit a new request for the given media item.
 * Uses hardcoded requester "anonymous" until auth is implemented.
 */
export async function submitRequest(item: MediaItem, signal?: AbortSignal): Promise<Request> {
  return api.post<Request>('/api/requests', { item_id: item.id, requester_id: 'anonymous' }, { signal });
}

/**
 * Fetch a single request by its Mortar-internal ID.
 */
export async function fetchRequest(id: string, signal?: AbortSignal): Promise<Request> {
  return api.get<Request>(`/api/requests/${encodeURIComponent(id)}`, { signal });
}
