package appstate

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nbellowe/mortar/src/backend/internal/config"
	"github.com/nbellowe/mortar/src/backend/internal/plugins"
	"github.com/nbellowe/mortar/src/db"
	"golang.org/x/crypto/bcrypt"
)

const sessionLifetime = 30 * 24 * time.Hour

type Store struct {
	db *db.DB
}

func New(database *db.DB) *Store {
	if database == nil {
		return nil
	}
	return &Store{db: database}
}

func randomID() (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("appstate: random id: %w", err)
	}
	return hex.EncodeToString(buf[:]), nil
}

func (s *Store) SyncUsersFromConfig(users []config.UserConfig) error {
	if len(users) == 0 || s == nil || s.db == nil {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("appstate: begin sync users: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().UTC().Format(time.RFC3339)
	for _, u := range users {
		userID, err := upsertUser(tx, u, now)
		if err != nil {
			return err
		}
		if err := syncExternalAccounts(tx, userID, u.ExternalAccounts, now); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("appstate: commit sync users: %w", err)
	}
	return nil
}

func upsertUser(tx *sql.Tx, u config.UserConfig, now string) (string, error) {
	var userID string
	queryErr := tx.QueryRow(`SELECT id FROM users WHERE username = ?`, u.Username).Scan(&userID)
	if queryErr != nil && !errors.Is(queryErr, sql.ErrNoRows) {
		return "", fmt.Errorf("appstate: query user %q: %w", u.Username, queryErr)
	}
	isNewUser := errors.Is(queryErr, sql.ErrNoRows)
	if isNewUser {
		var err error
		userID, err = randomID()
		if err != nil {
			return "", err
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("appstate: hash password for %q: %w", u.Username, err)
	}

	if isNewUser {
		_, execErr := tx.Exec(
			`INSERT INTO users (id, username, role, password_hash, created_at) VALUES (?, ?, ?, ?, ?)`,
			userID, u.Username, u.Role, string(hash), now,
		)
		if execErr != nil {
			return "", fmt.Errorf("appstate: insert user %q: %w", u.Username, execErr)
		}
		return userID, nil
	}

	if _, err := tx.Exec(`UPDATE users SET role = ?, password_hash = ? WHERE id = ?`, u.Role, string(hash), userID); err != nil {
		return "", fmt.Errorf("appstate: update user %q: %w", u.Username, err)
	}
	return userID, nil
}

func syncExternalAccounts(tx *sql.Tx, userID string, accounts []config.ExternalAccountConfig, now string) error {
	if _, err := tx.Exec(`DELETE FROM external_account_links WHERE user_id = ?`, userID); err != nil {
		return fmt.Errorf("appstate: clear external accounts for %q: %w", userID, err)
	}
	for _, ext := range accounts {
		id, err := randomID()
		if err != nil {
			return err
		}
		var username interface{}
		if strings.TrimSpace(ext.ExternalUsername) != "" {
			username = ext.ExternalUsername
		}
		if _, err := tx.Exec(
			`INSERT INTO external_account_links (id, user_id, plugin_id, upstream_user_id, upstream_username, created_at)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			id, userID, ext.PluginID, ext.ExternalUserID, username, now,
		); err != nil {
			return fmt.Errorf("appstate: insert external account for user %q plugin %q: %w", userID, ext.PluginID, err)
		}
	}
	return nil
}

func (s *Store) AuthenticateUser(username, password string) (*plugins.MortarUser, error) {
	var (
		userID       string
		role         string
		passwordHash string
	)
	err := s.db.QueryRow(`SELECT id, role, password_hash FROM users WHERE username = ?`, username).
		Scan(&userID, &role, &passwordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("appstate: authenticate %q: %w", username, err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, nil
	}
	return s.UserByID(userID)
}

func (s *Store) UserByID(userID string) (*plugins.MortarUser, error) {
	var user plugins.MortarUser
	err := s.db.QueryRow(`SELECT id, username, role FROM users WHERE id = ?`, userID).
		Scan(&user.ID, &user.Username, &user.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("appstate: user by id %q: %w", userID, err)
	}

	accounts, err := s.externalAccountsForUser(userID)
	if err != nil {
		return nil, err
	}
	user.ExternalAccounts = accounts
	return &user, nil
}

func (s *Store) externalAccountsForUser(userID string) ([]plugins.ExternalAccountLink, error) {
	rows, err := s.db.Query(
		`SELECT plugin_id, upstream_user_id, upstream_username
		   FROM external_account_links
		  WHERE user_id = ?
		  ORDER BY plugin_id`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("appstate: external accounts for %q: %w", userID, err)
	}
	defer rows.Close()

	var accounts []plugins.ExternalAccountLink
	for rows.Next() {
		var (
			account  plugins.ExternalAccountLink
			username sql.NullString
		)
		if err := rows.Scan(&account.PluginID, &account.ExternalUserID, &username); err != nil {
			return nil, fmt.Errorf("appstate: scan external account for %q: %w", userID, err)
		}
		if username.Valid {
			account.ExternalUsername = &username.String
		}
		accounts = append(accounts, account)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("appstate: iterate external accounts for %q: %w", userID, err)
	}
	if accounts == nil {
		return []plugins.ExternalAccountLink{}, nil
	}
	return accounts, nil
}

func (s *Store) CreateSession(userID string) (string, error) {
	token, err := randomID()
	if err != nil {
		return "", err
	}
	expiresAt := time.Now().UTC().Add(sessionLifetime).Format(time.RFC3339)
	id, err := randomID()
	if err != nil {
		return "", err
	}
	if _, err := s.db.Exec(`INSERT INTO sessions (id, user_id, token, expires_at) VALUES (?, ?, ?, ?)`, id, userID, token, expiresAt); err != nil {
		return "", fmt.Errorf("appstate: create session for %q: %w", userID, err)
	}
	return token, nil
}

func (s *Store) DeleteSession(token string) error {
	if token == "" {
		return nil
	}
	if _, err := s.db.Exec(`DELETE FROM sessions WHERE token = ?`, token); err != nil {
		return fmt.Errorf("appstate: delete session: %w", err)
	}
	return nil
}

func (s *Store) UserBySessionToken(token string) (*plugins.MortarUser, error) {
	var (
		userID    string
		expiresAt string
	)
	err := s.db.QueryRow(`SELECT user_id, expires_at FROM sessions WHERE token = ?`, token).Scan(&userID, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("appstate: session lookup: %w", err)
	}

	expiry, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("appstate: parse session expiry: %w", err)
	}
	if time.Now().UTC().After(expiry) {
		_ = s.DeleteSession(token)
		return nil, nil
	}
	return s.UserByID(userID)
}

func (s *Store) LookupMortarUserIDByExternalAccount(pluginID, externalUserID string) (string, error) {
	var userID string
	err := s.db.QueryRow(
		`SELECT user_id FROM external_account_links WHERE plugin_id = ? AND upstream_user_id = ?`,
		pluginID, externalUserID,
	).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("appstate: lookup external account %q/%q: %w", pluginID, externalUserID, err)
	}
	return userID, nil
}

func (s *Store) RecordHealthSnapshot(pluginID string, status plugins.HealthStatus) error {
	var detail interface{}
	if status.Detail != nil {
		detail = *status.Detail
	}
	if _, err := s.db.Exec(
		`INSERT INTO health_snapshots (plugin_id, status, checked_at, detail)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(plugin_id) DO UPDATE SET
		   status = excluded.status,
		   checked_at = excluded.checked_at,
		   detail = excluded.detail`,
		pluginID, status.Status, status.CheckedAt, detail,
	); err != nil {
		return fmt.Errorf("appstate: record health snapshot for %q: %w", pluginID, err)
	}
	return nil
}

func (s *Store) LoadHealthSnapshots() (map[string]plugins.HealthStatus, error) {
	rows, err := s.db.Query(`SELECT plugin_id, status, checked_at, detail FROM health_snapshots`)
	if err != nil {
		return nil, fmt.Errorf("appstate: load health snapshots: %w", err)
	}
	defer rows.Close()

	out := make(map[string]plugins.HealthStatus)
	for rows.Next() {
		var (
			pluginID  string
			snapshot  plugins.HealthStatus
			detailVal sql.NullString
		)
		if err := rows.Scan(&pluginID, &snapshot.Status, &snapshot.CheckedAt, &detailVal); err != nil {
			return nil, fmt.Errorf("appstate: scan health snapshot: %w", err)
		}
		snapshot.Reachable = snapshot.Status == "healthy" || snapshot.Status == "degraded"
		if detailVal.Valid {
			snapshot.Detail = &detailVal.String
		}
		out[pluginID] = snapshot
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("appstate: iterate health snapshots: %w", err)
	}
	return out, nil
}

func snapshotID(pluginID, requestID string) string {
	return "snapshot:" + pluginID + ":" + requestID
}

func (s *Store) UpsertRequestSnapshot(req plugins.Request, mortarUserID string) error {
	if mortarUserID == "" || req.ID == "" || req.PluginID == "" {
		return nil
	}
	var fulfilled interface{}
	if req.FulfilledAt != nil {
		fulfilled = *req.FulfilledAt
	}
	if _, err := s.db.Exec(
		`INSERT INTO request_snapshots (id, user_id, plugin_id, upstream_request_id, media_item_id, status, created_at, updated_at, fulfilled_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   status = excluded.status,
		   updated_at = excluded.updated_at,
		   fulfilled_at = excluded.fulfilled_at`,
		snapshotID(req.PluginID, req.ID),
		mortarUserID,
		req.PluginID,
		req.ID,
		req.Item.ID,
		req.Status,
		req.SubmittedAt,
		req.UpdatedAt,
		fulfilled,
	); err != nil {
		return fmt.Errorf("appstate: upsert request snapshot %q: %w", req.ID, err)
	}
	return nil
}

func (s *Store) PendingRequestExists(mediaItemID string) (bool, error) {
	if mediaItemID == "" {
		return false, nil
	}
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(1) FROM request_snapshots WHERE media_item_id = ? AND status = 'pending'`, mediaItemID).Scan(&count); err != nil {
		return false, fmt.Errorf("appstate: pending request exists for %q: %w", mediaItemID, err)
	}
	return count > 0, nil
}
