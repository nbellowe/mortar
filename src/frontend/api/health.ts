/**
 * Health API module.
 * Fetches plugin health snapshots from the Mortar server.
 * The server returns last-known state; it does not trigger a live probe on this call.
 */

import { api } from './client';
import { PluginHealth } from '../types/health';

/**
 * Returns the last-known health snapshot for all configured plugins.
 */
export async function fetchHealth(signal?: AbortSignal): Promise<PluginHealth[]> {
  return api.get<PluginHealth[]>('/api/health', { signal });
}
