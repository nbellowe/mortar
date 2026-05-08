/**
 * Health feature types.
 * Extends HealthStatus from the plugin interface with plugin identity fields
 * returned by the /api/health endpoint.
 */

import { HealthStatus } from './plugin';

export interface PluginHealth extends HealthStatus {
  plugin_id: string;
  plugin_type: string;
  display_name: string;
}
