package main

import (
	"database/sql"
	"strings"
)

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func migrate(db *sql.DB) error {
	// users table
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	email TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	supabase_user_id TEXT UNIQUE,
	created_at TIMESTAMP NOT NULL
);
`)
	if err != nil {
		return err
	}

	// Add supabase_user_id column if it doesn't exist (for existing databases)
	// Check if column exists by querying table info
	var columnExists bool
	rows, err := db.Query(`PRAGMA table_info(users)`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var cid int
			var name, dataType string
			var notNull, pk int
			var defaultValue interface{}
			if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err == nil {
				if name == "supabase_user_id" {
					columnExists = true
					break
				}
			}
		}
	}
	
	// Add column if it doesn't exist
	// Note: SQLite doesn't allow adding UNIQUE constraint directly to existing table
	// So we add the column first, then create a unique index
	if !columnExists {
		_, err = db.Exec(`ALTER TABLE users ADD COLUMN supabase_user_id TEXT`)
		if err != nil {
			// Log but don't fail - column might have been added concurrently
			_ = err
		} else {
			// Create unique index on supabase_user_id
			// Ignore error if index already exists
			_, _ = db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_supabase_user_id ON users(supabase_user_id)`)
		}
	}

	// plushies table with TEXT user_id referencing users.supabase_user_id
	// SQLite doesn't support ALTER COLUMN, so we need to recreate the table
	// First, check if plushies table exists and has INTEGER user_id
	var tableInfo string
	err = db.QueryRow(`
		SELECT sql FROM sqlite_master 
		WHERE type='table' AND name='plushies'
	`).Scan(&tableInfo)

	if err == nil && tableInfo != "" {
		// Table exists, check if user_id is INTEGER
		if contains(tableInfo, "user_id INTEGER") {
			// Create new table with TEXT user_id
			_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS plushies_new (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id TEXT NOT NULL,
	name TEXT NOT NULL,
	kind TEXT NOT NULL,
	adopted_at TEXT,
	image_path TEXT,
	conversation_history TEXT,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	FOREIGN KEY (user_id) REFERENCES users(supabase_user_id) ON DELETE CASCADE
);
`)
			if err != nil {
				return err
			}

			// Copy data from old table to new table
			_, _ = db.Exec(`
INSERT INTO plushies_new (id, user_id, name, kind, adopted_at, image_path, conversation_history, created_at, updated_at)
SELECT id, CAST(user_id AS TEXT), name, kind, adopted_at, image_path, conversation_history, created_at, updated_at
FROM plushies;
`)

			// Drop old table and rename new table
			_, err = db.Exec(`DROP TABLE plushies;`)
			if err != nil {
				return err
			}
			_, err = db.Exec(`ALTER TABLE plushies_new RENAME TO plushies;`)
			if err != nil {
				return err
			}
		}
	} else {
		// Table doesn't exist or is already migrated, create with TEXT user_id
		_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS plushies (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id TEXT NOT NULL,
	name TEXT NOT NULL,
	kind TEXT NOT NULL,
	adopted_at TEXT,
	image_path TEXT,
	conversation_history TEXT,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	FOREIGN KEY (user_id) REFERENCES users(supabase_user_id) ON DELETE CASCADE
);
`)
		if err != nil {
			return err
		}
	}

	return nil
}
