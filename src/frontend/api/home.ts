/**
 * Home API module.
 * Fetches aggregated home screen data from the Mortar server.
 * The server returns cached data; it does not trigger live probes on this call.
 * Spec: specs/features/home.md
 */

import { api } from "./client";
import type { ContinueWatchingItem, MediaItem } from "../types/plugin";

export interface HealthSummary {
  any_unreachable: boolean;
  total: number;
  unreachable_count: number;
}

export interface HomeResponse {
  recently_added: MediaItem[];
  recently_added_requires_link: boolean;
  continue_watching: ContinueWatchingItem[];
  continue_watching_enabled: boolean;
  continue_watching_requires_link: boolean;
  health_summary: HealthSummary;
}

/**
 * Returns aggregated home screen data: recently added items and a health summary.
 * No auto-polling — the home screen is not a live operational view.
 */
export function getHome(signal?: AbortSignal): Promise<HomeResponse> {
  return api.get<HomeResponse>("/api/v1/home", { signal });
}
