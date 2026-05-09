/**
 * Downloads API module.
 * Fetches the unified download queue from the Mortar server.
 * Results are pre-sorted by status priority on the server.
 */

import { api } from './client';
import type { DownloadItem } from '../types/plugin';

export interface DownloadsResponse {
  items: DownloadItem[];
  /** Display names of plugins that errored and could not contribute items. */
  failed_plugins: string[];
}

export function getDownloads(signal?: AbortSignal): Promise<DownloadsResponse> {
  return api.get<DownloadsResponse>('/api/v1/downloads', { signal });
}
