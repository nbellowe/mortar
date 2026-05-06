# Plugin Interface

## Overview

A plugin is a module that wraps a single backend service. It declares its capabilities and implements the methods that correspond to them. Mortar only calls methods that the plugin has declared support for.

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
  | "downloads.read"
  | "activity.read"
  | "health.ping";
```

## Base interface

All plugins must implement:

```typescript
interface Plugin {
  manifest: PluginManifest;
  health(): Promise<HealthStatus>;
}
```

## Capability interfaces

Plugins implement only the interfaces matching their declared capabilities.

### `health.ping`

```typescript
interface HealthStatus {
  reachable: boolean;
  latency_ms: number;
  checked_at: string; // ISO 8601
  detail?: string;    // error message if not reachable
}
```

### `requests.video` / `requests.audio` / `requests.ebook`

```typescript
interface RequestCapable {
  search(query: string, type: MediaType): Promise<MediaItem[]>;
  getRequest(id: string): Promise<Request | null>;
  listRequests(options?: { userId?: string; status?: RequestStatus }): Promise<Request[]>;
  submitRequest(item: MediaItem, requester: MortarUser): Promise<Request>;
}
```

### `library.browse`

```typescript
interface LibraryBrowsable {
  browse(options: BrowseOptions): Promise<PagedResult<MediaItem>>;
  getItem(id: string): Promise<MediaItem | null>;
  getPlayUrl(item: MediaItem, user: MortarUser): Promise<string>; // deep link to player
}

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
  exists(item: MediaItem): Promise<boolean>;
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
}

type ActivityEventType =
  | "downloaded"
  | "added_to_library"
  | "requested"
  | "approved"
  | "rejected"
  | "failed"
  | "deleted";
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

type MediaType = "movie" | "show" | "audiobook" | "ebook" | "music";

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
}

interface PagedResult<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}
```

## Plugin registration

Plugins are registered at server startup based on the config file. The server:

1. Reads `plugins` from config
2. Resolves the plugin type to its implementation
3. Instantiates each plugin with its config
4. Calls `health()` on each plugin and logs the result
5. Makes each plugin available to the router keyed by `id`

Unknown plugin types are a startup error.

## Adding a new plugin

1. Create a file in `src/plugins/<type>.ts`
2. Implement `Plugin` and any capability interfaces
3. Declare the type in the plugin registry
4. Add a row to the plugin compatibility table in `docs/plugins.md`
