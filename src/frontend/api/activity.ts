/**
 * Activity API module.
 * Fetches the cross-service activity feed from the Mortar server.
 */

import { api } from './client';
import type { ActivityEvent } from '../types/plugin';

export interface ActivityResponse {
  events: ActivityEvent[];
  failed_plugins: string[];
}

/**
 * Returns the activity feed, optionally filtered to events after `since`.
 * The server guarantees events are sorted newest-first.
 */
export function getActivity(since?: string): Promise<ActivityResponse> {
  const params = since ? `?since=${encodeURIComponent(since)}` : '';
  return api.get<ActivityResponse>(`/api/v1/activity${params}`);
}
