/**
 * Plugin interface types — TypeScript mirror of internal/plugins/plugin.go.
 *
 * These types MUST stay in sync with the Go definitions. The plugin-interface
 * spec (specs/plugins/plugin-interface.md) is the source of truth; do not
 * add, remove, or rename fields without updating both files.
 */

// ---------------------------------------------------------------------------
// Shared types
// ---------------------------------------------------------------------------

export type MediaType = 'movie' | 'show' | 'audiobook' | 'ebook';

export interface MediaItem {
  /** Mortar internal ID in the format "plugin_id:external_id" */
  id: string;
  /** ID in the source system */
  external_id: string;
  /** Which plugin this item came from */
  plugin_id: string;
  type: MediaType;
  title: string;
  year?: number;
  overview?: string;
  poster_url?: string;
  genres?: string[];
  tmdb_id?: string;
  imdb_id?: string;
  tvdb_id?: string;
  /** For ebooks */
  isbn?: string;
  /** For audiobooks */
  asin?: string;
}

export type RequestStatus =
  | 'pending'
  | 'approved'
  | 'available'
  | 'declined'
  | 'failed';

export interface Request {
  id: string;
  plugin_id: string;
  item: MediaItem;
  requester_id: string;
  status: RequestStatus;
  /** ISO 8601 */
  submitted_at: string;
  /** ISO 8601 */
  updated_at: string;
  /** ISO 8601, present when fulfilled */
  fulfilled_at?: string;
}

export type ActivityEventType =
  | 'downloaded'
  | 'added_to_library'
  | 'requested'
  | 'approved'
  | 'declined'
  | 'failed'
  | 'deleted';

export type ActivityVisibility =
  | 'all_users'
  | 'admin_only'
  | 'requester_and_admin';

export interface ActivityEvent {
  id: string;
  source_plugin: string;
  type: ActivityEventType;
  item?: MediaItem;
  message: string;
  /** ISO 8601 */
  timestamp: string;
  visibility: ActivityVisibility;
  /**
   * Must be set for events with visibility "requester_and_admin".
   * Mortar uses it to determine per-user event visibility.
   */
  actor_user_id?: string;
  actor_display_name?: string;
}

export interface DownloadItem {
  id: string;
  name: string;
  /** 0.0–1.0 */
  progress: number;
  size_bytes: number;
  speed_bytes_s: number;
  /** null when unknown */
  eta_seconds: number | null;
  status: 'downloading' | 'paused' | 'queued' | 'processing' | 'failed';
  /** Which *arr triggered this download, if known */
  source_plugin?: string;
}

export interface HealthStatus {
  status: 'unknown' | 'healthy' | 'degraded' | 'unreachable';
  reachable: boolean;
  latency_ms: number;
  /** ISO 8601 */
  checked_at: string;
  /** Error message when not reachable. Null when healthy (Go *string → JSON null). */
  detail?: string | null;
}

export interface MortarUser {
  id: string;
  username: string;
  role: 'admin' | 'user';
  external_accounts?: ExternalAccountLink[];
}

export interface ExternalAccountLink {
  plugin_id: string;
  external_user_id: string;
  external_username?: string;
}

// ---------------------------------------------------------------------------
// Capability types
// ---------------------------------------------------------------------------

export type Capability =
  | 'requests.video'
  | 'requests.audio'
  | 'requests.ebook'
  | 'library.browse'
  | 'library.exists'
  | 'library.resume'
  | 'downloads.read'
  | 'activity.read';

// Note: spec uses displayName; wire format is display_name for consistency.
export interface PluginManifest {
  id: string;
  type: string;
  display_name: string;
  capabilities: Capability[];
}

// ---------------------------------------------------------------------------
// Pagination
// ---------------------------------------------------------------------------

export interface PagedResult<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}
