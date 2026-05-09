import { api } from "./client";
import type { MediaItem } from "../types/plugin";

export interface LibraryBrowseResponse {
  items: MediaItem[];
  total: number;
  page: number;
  page_size: number;
  available_genres: string[];
  requires_link: boolean;
  plugin_display_name?: string;
}

export interface LibraryPlayResponse {
  url: string;
}

export function browseLibrary(options?: {
  type?: string;
  genre?: string;
  sort?: string;
  page?: number;
  pageSize?: number;
  signal?: AbortSignal;
}): Promise<LibraryBrowseResponse> {
  const params = new URLSearchParams();
  if (options?.type) params.set("type", options.type);
  if (options?.genre) params.set("genre", options.genre);
  if (options?.sort) params.set("sort", options.sort);
  if (options?.page) params.set("page", String(options.page));
  if (options?.pageSize) params.set("page_size", String(options.pageSize));
  const qs = params.toString();
  return api.get<LibraryBrowseResponse>(
    `/api/v1/library${qs ? `?${qs}` : ""}`,
    { signal: options?.signal },
  );
}

export function playLibraryItem(
  itemId: string,
  signal?: AbortSignal,
): Promise<LibraryPlayResponse> {
  return api.post<LibraryPlayResponse>(
    "/api/v1/library/play",
    { item_id: itemId },
    { signal },
  );
}
