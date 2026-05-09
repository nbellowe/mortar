package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/nbellowe/mortar/src/backend/internal/plugins"
)

type contextKey string

const (
	sessionCookieName contextKey = "mortar_session"
	currentUserKey    contextKey = "current_user"
)

type authSessionResponse struct {
	User plugins.MortarUser `json:"user"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func currentUser(r *http.Request) *plugins.MortarUser {
	user, _ := r.Context().Value(currentUserKey).(*plugins.MortarUser)
	return user
}

func (h *handler) enforceExternalLinks() bool {
	return h.store != nil
}

func setCurrentUser(r *http.Request, user *plugins.MortarUser) *http.Request {
	ctx := context.WithValue(r.Context(), currentUserKey, user)
	return r.WithContext(ctx)
}

func (h *handler) sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.store == nil {
			next.ServeHTTP(w, setCurrentUser(r, &plugins.MortarUser{
				ID:               "test-admin",
				Username:         "test-admin",
				Role:             "admin",
				ExternalAccounts: []plugins.ExternalAccountLink{},
			}))
			return
		}

		cookie, err := r.Cookie(string(sessionCookieName))
		if err != nil || cookie.Value == "" {
			next.ServeHTTP(w, r)
			return
		}

		user, err := h.store.UserBySessionToken(cookie.Value)
		if err != nil || user == nil {
			next.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(w, setCurrentUser(r, user))
	})
}

func (h *handler) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if currentUser(r) == nil {
			jsonError(w, "authentication required", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *handler) requireAdmin(next http.Handler) http.Handler {
	return h.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := currentUser(r)
		if user == nil || user.Role != "admin" {
			jsonError(w, "admin access required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}))
}

func writeSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     string(sessionCookieName),
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   30 * 24 * 60 * 60,
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     string(sessionCookieName),
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func (h *handler) handleAuthSession(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if user == nil {
		jsonError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(authSessionResponse{User: *user})
}

func (h *handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		jsonError(w, "login unavailable", http.StatusServiceUnavailable)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var body loginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	body.Username = strings.TrimSpace(body.Username)
	if body.Username == "" || body.Password == "" {
		jsonError(w, "username and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.store.AuthenticateUser(body.Username, body.Password)
	if err != nil {
		jsonError(w, "login failed", http.StatusInternalServerError)
		return
	}
	if user == nil {
		jsonError(w, "invalid username or password", http.StatusUnauthorized)
		return
	}

	token, err := h.store.CreateSession(user.ID)
	if err != nil {
		jsonError(w, "login failed", http.StatusInternalServerError)
		return
	}
	writeSessionCookie(w, token)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(authSessionResponse{User: *user})
}

func (h *handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	if h.store != nil {
		if cookie, err := r.Cookie(string(sessionCookieName)); err == nil {
			_ = h.store.DeleteSession(cookie.Value)
		}
	}
	clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}
