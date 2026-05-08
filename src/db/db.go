// Package db manages the Mortar SQLite database connection and schema migrations.
package db

import (
	"database/sql"
	_ "embed"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schema string

// DB wraps a sql.DB with Mortar-specific initialisation.
type DB struct {
	*sql.DB
}

// Open opens (or creates) the SQLite database at path, enables WAL mode
// and foreign key enforcement, and applies the embedded schema.
func Open(path string) (*DB, error) {
	sqldb, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("db: open %q: %w", path, err)
	}

	// SQLite allows only one writer at a time; a pool size of 1 prevents
	// "database is locked" errors under concurrent requests.
	sqldb.SetMaxOpenConns(1)

	// Verify the connection is live.
	if err := sqldb.Ping(); err != nil {
		_ = sqldb.Close()
		return nil, fmt.Errorf("db: ping %q: %w", path, err)
	}

	db := &DB{sqldb}
	if err := db.migrate(); err != nil {
		_ = sqldb.Close()
		return nil, err
	}

	return db, nil
}

// migrate runs the embedded schema SQL. The schema uses CREATE TABLE IF NOT
// EXISTS so it is safe to run on every startup against an existing database.
func (db *DB) migrate() error {
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("db: migrate: %w", err)
	}
	return nil
}
