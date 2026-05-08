# Plugin Interface

## Overview

A plugin is a module that wraps a single backend service. It declares its capabilities and implements the methods that correspond to them. Mortar only calls methods that the plugin has declared support for.

The canonical implementation is Go. The TypeScript-style interfaces below remain the source-of-truth contract notation for specs, backend implementations, and mirrored frontend types.

## Capability declaration

```typescript
interface PluginManifest {
  id: string;           // unique instance id from config
  type: string;         // plugin type (e.g. "jellyseerr", "sonarr")
  displayName: string;  // human-readable name for UI
  capabilities: Capability[];
}

type Capability =
  | "requests.video"
  | "requests.audio"
  | "requests.ebook"
  | "library.browse"
  | "library.exists"
  | "library.resume"
  | "downloads.read"
  | "activity.read";
```

## Base interface

All plugins must implement:

```typescript
interface Plugin {
  manifest: PluginManifest;
  health(): Promise<HealthStatus>;
}

interface HealthStatus {
  status: "unknown" | "healthy" | "degraded" | "unreachable";
  reachable: boolean;
  latency_ms: number;
  checked_at: string; // ISO 8601
  detail?: string;    // error message if not reachable
}

// status derivation:
// "unknown"     — plugin has not been checked since server startup (no snapshot row)
// "healthy"     — reachable, latency_ms <= 2000
// "degraded"    — reachable, latency_ms > 2000
// "unreachable" — not reachable (reachable: false)
```

Health is a mandatory base contract, not a capability flag.

## Capability interfaces

Plugins implement only the interfaces matching their declared capabilities.

### `requests.video` / `requests.audio` / `requests.ebook`

```typescript
interface RequestCapable {
  search(query: string): Promise<MediaItem[]>;
  getRequest(id: string): Promise<Request | null>;
  listRequests(options?: { requesterId?: string; status?: RequestStatus }): Promise<Request[]>;
  submitRequest(item: MediaItem, requester: MortarUser): Promise<Request>;
  reviewRequest(id: string, review: RequestReview): Promise<Request>;
}

interface RequestReview {
  decision: "approve" | "decline";
  reviewer: MortarUser;
  reason?: string;
}
```

### `library.browse`

```typescript
interface LibraryBrowsable {
  browse(options: BrowseOptions): Promise<PagedResult<MediaItem>>;
  getItem(id: string): Promise<MediaItem | null>;
  getPlayUrl(item: MediaItem, user: MortarUser): Promise<string>; // deep link URI for external player handoff
}

// getPlayUrl: the plugin resolves the user's identity for the upstream service by finding
// the ExternalAccountLink in user.external_accounts where plugin_id matches its own manifest.id.
// If no matching link exists, the plugin should reject — Mortar gates this call on link presence.

interface BrowseOptions {
  type?: MediaType;
  genre?: string;
  sort?: "added" | "title" | "year";
  page?: number;
  pageSize?: number;
}
```

### `library.exists`

```typescript
interface LibraryExists {
  findMatch(item: MediaItem): Promise<LibraryMatch | null>;
}

interface LibraryMatch {
  plugin_id: string;
  item: MediaItem;
  matched_by: "tmdb_id" | "imdb_id" | "tvdb_id" | "isbn" | "asin" | "title_year";
}
```

### `library.resume`

```typescript
interface LibraryResumeReadable {
  getContinueWatching(
    user: MortarUser,
    options?: { limit?: number }
  ): Promise<ContinueWatchingItem[]>;
}

interface ContinueWatchingItem {
  item: MediaItem;
  progress: number;           // 0.0-1.0
  position_seconds?: number;
  duration_seconds?: number;
  last_watched_at: string;    // ISO 8601
}
```

### `downloads.read`

```typescript
interface DownloadsReadable {
  getQueue(): Promise<DownloadItem[]>;
}

interface DownloadItem {
  id: string;
  name: string;
  progress: number;     // 0.0–1.0
  size_bytes: number;
  speed_bytes_s: number;
  eta_seconds: number | null;
  status: "downloading" | "paused" | "queued" | "processing" | "failed";
  source_plugin?: string; // which *arr triggered this download, if known
}
```

### `activity.read`

```typescript
interface ActivityReadable {
  getActivity(since?: string): Promise<ActivityEvent[]>;
}

interface ActivityEvent {
  id: string;
  source_plugin: string;
  type: ActivityEventType;
  item?: MediaItem;
  message: string;
  timestamp: string; // ISO 8601
  visibility: ActivityVisibility;
  actor_user_id?: string;
  actor_display_name?: string;
}

type ActivityEventType =
  | "downloaded"
  | "added_to_library"
  | "requested"
  | "approved"
  | "declined"
  | "failed"
  | "deleted";

type ActivityVisibility = "all_users" | "admin_only" | "requester_and_admin";

// For events with visibility "requester_and_admin", actor_user_id MUST be set to the
// Mortar user ID of the requester. Mortar uses it to decide per-user event visibility.
```

## Shared types

```typescript
interface MediaItem {
  id: string;             // Mortar internal ID (plugin:externalId)
  external_id: string;    // ID in the source system
  plugin_id: string;      // which plugin this came from
  type: MediaType;
  title: string;
  year?: number;
  overview?: string;
  poster_url?: string;
  genres?: string[];
  tmdb_id?: string;
  imdb_id?: string;
  tvdb_id?: string;
  isbn?: string;          // for ebooks
  asin?: string;          // for audiobooks
}

type MediaType = "movie" | "show" | "audiobook" | "ebook";

interface Request {
  id: string;
  plugin_id: string;
  item: MediaItem;
  requester_id: string;
  status: RequestStatus;
  submitted_at: string;
  updated_at: string;
  fulfilled_at?: string;
}

type RequestStatus = "pending" | "approved" | "available" | "declined" | "failed";

interface MortarUser {
  id: string;
  username: string;
  role: "admin" | "user";
  external_accounts?: ExternalAccountLink[];
}

interface ExternalAccountLink {
  plugin_id: string;
  external_user_id: string;
  external_username?: string;
}

interface PagedResult<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}
```

## Failure semantics

Plugin methods should reject on upstream auth errors, unreachability, or timeouts. Mortar's plugin runtime is responsible for capturing failures per plugin so multi-plugin views can return partial data with notices about unavailable sources.

## Plugin registration

Plugins are registered at server startup based on the config file. The server:

1. Reads `plugins` from config
2. Resolves the plugin type to its implementation
3. Instantiates each plugin with its config
4. Calls `health()` on each plugin and logs the result
5. Makes each plugin available to the router keyed by `id`

Unknown plugin types are a startup error.

## Adding a new plugin

1. Create a package in `internal/plugins/<type>/`
2. Implement `Plugin` and any capability interfaces
3. Declare the type in the plugin registry
4. Add a row to the plugin compatibility table in `docs/plugins.md`
