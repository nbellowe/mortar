// Package plugins defines the plugin interface and all shared types used
// across Mortar's plugin system. The types here mirror the TypeScript
// interfaces in src/types/plugin.ts — do not diverge.
package plugins

// ---------------------------------------------------------------------------
// Shared types
// ---------------------------------------------------------------------------

// MediaType classifies the kind of media an item represents.
type MediaType string

const (
	MediaTypeMovie     MediaType = "movie"
	MediaTypeShow      MediaType = "show"
	MediaTypeAudiobook MediaType = "audiobook"
	MediaTypeEbook     MediaType = "ebook"
)

// MediaItem is the normalized representation of a piece of media.
// The ID field uses the Mortar-internal format "plugin_id:external_id".
type MediaItem struct {
	ID         string    `json:"id"`          // Mortar internal ID (plugin:externalId)
	ExternalID string    `json:"external_id"` // ID in the source system
	PluginID   string    `json:"plugin_id"`   // which plugin this came from
	Type       MediaType `json:"type"`
	Title      string    `json:"title"`
	Year       *int      `json:"year,omitempty"`
	Overview   *string   `json:"overview,omitempty"`
	PosterURL  *string   `json:"poster_url,omitempty"`
	Genres     []string  `json:"genres,omitempty"`
	TmdbID     *string   `json:"tmdb_id,omitempty"`
	ImdbID     *string   `json:"imdb_id,omitempty"`
	TvdbID     *string   `json:"tvdb_id,omitempty"`
	ISBN       *string   `json:"isbn,omitempty"` // for ebooks
	ASIN       *string   `json:"asin,omitempty"` // for audiobooks
}

// RequestStatus represents the lifecycle state of a request.
type RequestStatus string

const (
	RequestStatusPending   RequestStatus = "pending"
	RequestStatusApproved  RequestStatus = "approved"
	RequestStatusAvailable RequestStatus = "available"
	RequestStatusDeclined  RequestStatus = "declined"
	RequestStatusFailed    RequestStatus = "failed"
)

// Request represents a user's request for a media item.
type Request struct {
	ID          string        `json:"id"`
	PluginID    string        `json:"plugin_id"`
	Item        MediaItem     `json:"item"`
	RequesterID string        `json:"requester_id"`
	Status      RequestStatus `json:"status"`
	SubmittedAt string        `json:"submitted_at"` // ISO 8601
	UpdatedAt   string        `json:"updated_at"`   // ISO 8601
	FulfilledAt *string       `json:"fulfilled_at,omitempty"`
}

// ActivityEventType classifies an activity event.
type ActivityEventType string

const (
	ActivityEventDownloaded     ActivityEventType = "downloaded"
	ActivityEventAddedToLibrary ActivityEventType = "added_to_library"
	ActivityEventRequested      ActivityEventType = "requested"
	ActivityEventApproved       ActivityEventType = "approved"
	ActivityEventDeclined       ActivityEventType = "declined"
	ActivityEventFailed         ActivityEventType = "failed"
	ActivityEventDeleted        ActivityEventType = "deleted"
)

// ActivityVisibility controls which users can see an event.
type ActivityVisibility string

const (
	ActivityVisibilityAllUsers          ActivityVisibility = "all_users"
	ActivityVisibilityAdminOnly         ActivityVisibility = "admin_only"
	ActivityVisibilityRequesterAndAdmin ActivityVisibility = "requester_and_admin"
)

// ActivityEvent is a single entry in the cross-service activity timeline.
// For events with visibility "requester_and_admin", ActorUserID must be set
// to the Mortar user ID of the requester.
type ActivityEvent struct {
	ID               string             `json:"id"`
	SourcePlugin     string             `json:"source_plugin"`
	Type             ActivityEventType  `json:"type"`
	Item             *MediaItem         `json:"item,omitempty"`
	Message          string             `json:"message"`
	Timestamp        string             `json:"timestamp"` // ISO 8601
	Visibility       ActivityVisibility `json:"visibility"`
	ActorUserID      *string            `json:"actor_user_id,omitempty"`
	ActorDisplayName *string            `json:"actor_display_name,omitempty"`
}

// DownloadItem represents a single item in a download queue.
type DownloadItem struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Progress     float64 `json:"progress"` // 0.0–1.0
	SizeBytes    int64   `json:"size_bytes"`
	SpeedBytesS  int64   `json:"speed_bytes_s"`
	EtaSeconds   *int64  `json:"eta_seconds"`             // null when unknown
	Status       string  `json:"status"`                  // "downloading"|"paused"|"queued"|"processing"|"failed"
	SourcePlugin *string `json:"source_plugin,omitempty"` // which *arr triggered this, if known
}

// HealthStatus is the result of a plugin health check.
// Status is derived by Mortar: "unknown" if never checked, "healthy" if
// reachable and latency_ms <= 2000, "degraded" if reachable and
// latency_ms > 2000, "unreachable" if not reachable.
type HealthStatus struct {
	Status    string  `json:"status"` // "unknown"|"healthy"|"degraded"|"unreachable"
	Reachable bool    `json:"reachable"`
	LatencyMs int64   `json:"latency_ms"`
	CheckedAt string  `json:"checked_at"`       // ISO 8601
	Detail    *string `json:"detail,omitempty"` // error message if not reachable
}

// MortarUser is a Mortar-authenticated user with optional upstream links.
type MortarUser struct {
	ID               string                `json:"id"`
	Username         string                `json:"username"`
	Role             string                `json:"role"` // "admin"|"user"
	ExternalAccounts []ExternalAccountLink `json:"external_accounts,omitempty"`
}

// ExternalAccountLink links a Mortar user to a user identity in an upstream service.
type ExternalAccountLink struct {
	PluginID         string  `json:"plugin_id"`
	ExternalUserID   string  `json:"external_user_id"`
	ExternalUsername *string `json:"external_username,omitempty"`
}

// ---------------------------------------------------------------------------
// Capability types
// ---------------------------------------------------------------------------

// Capability is a named plugin capability flag.
type Capability string

const (
	CapabilityRequestsVideo Capability = "requests.video"
	CapabilityRequestsAudio Capability = "requests.audio"
	CapabilityRequestsEbook Capability = "requests.ebook"
	CapabilityLibraryBrowse Capability = "library.browse"
	CapabilityLibraryExists Capability = "library.exists"
	CapabilityLibraryResume Capability = "library.resume"
	CapabilityDownloadsRead Capability = "downloads.read"
	CapabilityActivityRead  Capability = "activity.read"
)

// Note: the spec uses displayName but the wire format is display_name,
// consistent with all other Mortar API fields.
// PluginManifest describes a plugin's identity and declared capabilities.
type PluginManifest struct {
	ID           string       `json:"id"`
	Type         string       `json:"type"`
	DisplayName  string       `json:"display_name"`
	Capabilities []Capability `json:"capabilities"`
}

// ---------------------------------------------------------------------------
// Plugin interfaces
// ---------------------------------------------------------------------------

// Plugin is the base interface every plugin must implement.
// Health is mandatory regardless of declared capabilities.
type Plugin interface {
	Manifest() PluginManifest
	Health() (HealthStatus, error)
}

// RequestReview is the payload for approve/decline actions.
type RequestReview struct {
	Decision string     `json:"decision"` // "approve"|"decline"
	Reviewer MortarUser `json:"reviewer"`
	Reason   *string    `json:"reason,omitempty"`
}

// ListRequestsOptions filters the ListRequests call. All fields are optional.
type ListRequestsOptions struct {
	RequesterID *string
	Status      *RequestStatus
}

// Requester is implemented by plugins with any requests.* capability.
type Requester interface {
	Search(query string) ([]MediaItem, error)
	GetRequest(id string) (*Request, error)
	ListRequests(opts ListRequestsOptions) ([]Request, error)
	SubmitRequest(item MediaItem, requester MortarUser) (Request, error)
	ReviewRequest(id string, review RequestReview) (Request, error)
}

// RequestReviewURLProvider is an optional interface for plugins that can link
// admins out to the upstream request-management surface.
type RequestReviewURLProvider interface {
	ReviewURL(id string) string
}

// BrowseOptions parameterises a library browse call.
type BrowseOptions struct {
	Type     *MediaType `json:"type,omitempty"`
	Genre    *string    `json:"genre,omitempty"`
	Sort     *string    `json:"sort,omitempty"` // "added"|"title"|"year"
	Page     *int       `json:"page,omitempty"`
	PageSize *int       `json:"page_size,omitempty"`
}

// PagedResult is a paginated response wrapper.
type PagedResult[T any] struct {
	Items    []T `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// LibraryBrowser is implemented by plugins with the library.browse capability.
type LibraryBrowser interface {
	Browse(options BrowseOptions) (PagedResult[MediaItem], error)
	GetItem(id string) (*MediaItem, error)
	GetPlayURL(item MediaItem, user MortarUser) (string, error)
}

// LibraryMatch is the result of a library existence check.
type LibraryMatch struct {
	PluginID  string    `json:"plugin_id"`
	Item      MediaItem `json:"item"`
	MatchedBy string    `json:"matched_by"` // "tmdb_id"|"imdb_id"|"tvdb_id"|"isbn"|"asin"|"title_year"
}

// LibraryExists is implemented by plugins with the library.exists capability.
type LibraryExists interface {
	FindMatch(item MediaItem) (*LibraryMatch, error)
}

// ContinueWatchingItem is a resume point for a partially-watched item.
type ContinueWatchingItem struct {
	Item            MediaItem `json:"item"`
	Progress        float64   `json:"progress"` // 0.0–1.0
	PositionSeconds *int64    `json:"position_seconds,omitempty"`
	DurationSeconds *int64    `json:"duration_seconds,omitempty"`
	LastWatchedAt   string    `json:"last_watched_at"` // ISO 8601
}

// ContinueWatchingOptions parameterises a GetContinueWatching call.
type ContinueWatchingOptions struct {
	Limit *int
}

// LibraryResumeReadable is implemented by plugins with the library.resume capability.
type LibraryResumeReadable interface {
	GetContinueWatching(user MortarUser, opts ContinueWatchingOptions) ([]ContinueWatchingItem, error)
}

// DownloadsReadable is implemented by plugins with the downloads.read capability.
type DownloadsReadable interface {
	GetQueue() ([]DownloadItem, error)
}

// ActivityReadable is implemented by plugins with the activity.read capability.
type ActivityReadable interface {
	GetActivity(since *string) ([]ActivityEvent, error)
}
