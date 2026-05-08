// Package api — search and request endpoints.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nbellowe/mortar/src/backend/internal/plugins"
)

// jsonError writes a JSON-encoded error response with the correct Content-Type.
// http.Error always sets Content-Type: text/plain, so we use this helper instead.
func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":%q}`, msg)
}

// submitRequestBody is the JSON body for POST /api/v1/requests.
type submitRequestBody struct {
	MediaID string `json:"media_id"`
	Type    string `json:"type"`
}

// requesterPlugin retrieves the routed Requester plugin for requests.video.
// Returns (nil, false) and writes a 503 if the plugin is unavailable or does
// not implement the Requester interface.
func (h *handler) requesterPlugin(w http.ResponseWriter) (plugins.Requester, bool) {
	pluginID := h.registry.RouteRequest("requests.video")
	if pluginID == "" {
		jsonError(w, "no requests.video plugin configured", http.StatusServiceUnavailable)
		return nil, false
	}

	p := h.registry.Get(pluginID)
	if p == nil {
		jsonError(w, "requests.video plugin not found", http.StatusServiceUnavailable)
		return nil, false
	}

	requester, ok := p.(plugins.Requester)
	if !ok {
		jsonError(w, "requests.video plugin does not implement Requester", http.StatusServiceUnavailable)
		return nil, false
	}

	return requester, true
}

// handleSearch handles GET /api/v1/search?q=<query>.
// Requires the q parameter; returns 400 if missing.
// Calls Search on the routed requests.video plugin and returns a JSON array
// of MediaItem.
func (h *handler) handleSearch(w http.ResponseWriter, r *http.Request) {
	// TODO: fan out to all plugins declaring any requests.* capability (spec §search).
	// Currently routes only to the single requests.video plugin.
	q := r.URL.Query().Get("q")
	if q == "" {
		jsonError(w, "missing required query parameter: q", http.StatusBadRequest)
		return
	}

	requester, ok := h.requesterPlugin(w)
	if !ok {
		return
	}

	items, err := requester.Search(q)
	if err != nil {
		jsonError(w, "search failed", http.StatusInternalServerError)
		return
	}

	if items == nil {
		items = []plugins.MediaItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(items)
}

// handleListRequests handles GET /api/v1/requests.
// Accepts optional query params: status, requester_id.
// Returns a JSON array of Request.
func (h *handler) handleListRequests(w http.ResponseWriter, r *http.Request) {
	requester, ok := h.requesterPlugin(w)
	if !ok {
		return
	}

	var opts plugins.ListRequestsOptions

	if s := r.URL.Query().Get("status"); s != "" {
		status := plugins.RequestStatus(s)
		opts.Status = &status
	}
	if rid := r.URL.Query().Get("requester_id"); rid != "" {
		opts.RequesterID = &rid
	}

	requests, err := requester.ListRequests(opts)
	if err != nil {
		jsonError(w, "list requests failed", http.StatusInternalServerError)
		return
	}

	if requests == nil {
		requests = []plugins.Request{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(requests)
}

// handleSubmitRequest handles POST /api/v1/requests.
// Parses {media_id, type} from the JSON body, constructs a minimal MediaItem,
// and submits it via the routed requests.video plugin.
// Uses an anonymous stub MortarUser (auth comes in a later phase).
// Returns 201 + the created Request on success.
func (h *handler) handleSubmitRequest(w http.ResponseWriter, r *http.Request) {
	requester, ok := h.requesterPlugin(w)
	if !ok {
		return
	}

	var body submitRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.MediaID == "" {
		jsonError(w, "media_id is required", http.StatusBadRequest)
		return
	}
	if body.Type == "" {
		jsonError(w, "type is required", http.StatusBadRequest)
		return
	}

	item := plugins.MediaItem{
		ExternalID: body.MediaID,
		Type:       plugins.MediaType(body.Type),
	}

	// Stub requester — real auth is a later phase.
	user := plugins.MortarUser{
		ID:       "anonymous",
		Username: "anonymous",
		Role:     "user",
	}

	// TODO: check request_snapshots for duplicate pending request (spec AC §8).
	created, err := requester.SubmitRequest(item, user)
	if err != nil {
		jsonError(w, "submit request failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)
}

// handleGetRequest handles GET /api/v1/requests/{id}.
// Returns 200 + the Request if found, 404 if not.
func (h *handler) handleGetRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	requester, ok := h.requesterPlugin(w)
	if !ok {
		return
	}

	req, err := requester.GetRequest(id)
	if err != nil {
		jsonError(w, "get request failed", http.StatusInternalServerError)
		return
	}

	if req == nil {
		jsonError(w, "request not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(req)
}
