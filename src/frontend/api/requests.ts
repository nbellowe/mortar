/**
 * Requests API module.
 * Covers search, request listing, request submission, and request detail.
 * All calls go to the Mortar server — never directly to upstream services.
 */

import { api } from "./client";
import { LibraryMatch, MediaItem, Request } from "../types/plugin";

export interface SearchResponse {
  items: MediaItem[];
  failed_plugins: string[];
  existing_requests: Request[];
  available_matches: LibraryMatch[];
}

export interface RequestsResponse {
  items: Request[];
  review_urls: Record<string, string>;
}

/**
 * Search for media across all plugins with a requests.* capability.
 * Results are deduplicated by external ID on the server.
 */
export async function searchMedia(
  query: string,
  signal?: AbortSignal,
): Promise<SearchResponse> {
  const encoded = encodeURIComponent(query);
  return api.get<SearchResponse>(`/api/v1/search?q=${encoded}`, { signal });
}

/**
 * List requests, optionally filtered by requester and/or status.
 */
export async function fetchRequests(options?: {
  requesterId?: string;
  status?: string;
  signal?: AbortSignal;
}): Promise<RequestsResponse> {
  const params = new URLSearchParams();
  if (options?.requesterId) params.set("requester_id", options.requesterId);
  if (options?.status) params.set("status", options.status);
  const qs = params.toString();
  return api.get<RequestsResponse>(`/api/v1/requests${qs ? `?${qs}` : ""}`, {
    signal: options?.signal,
  });
}

/**
 * Submit a new request for the given media item.
 * Uses hardcoded requester "anonymous" until auth is implemented.
 */
export async function submitRequest(
  item: MediaItem,
  signal?: AbortSignal,
): Promise<Request> {
  return api.post<Request>(
    "/api/v1/requests",
    {
      item_id: item.id,
      media_id: item.external_id,
      plugin_id: item.plugin_id,
      type: item.type,
      title: item.title,
      tmdb_id: item.tmdb_id,
      imdb_id: item.imdb_id,
      tvdb_id: item.tvdb_id,
      isbn: item.isbn,
      asin: item.asin,
    },
    { signal },
  );
}

/**
 * Fetch a single request by its Mortar-internal ID.
 */
export async function fetchRequest(
  id: string,
  signal?: AbortSignal,
): Promise<Request> {
  return api.get<Request>(`/api/v1/requests/${encodeURIComponent(id)}`, {
    signal,
  });
}
