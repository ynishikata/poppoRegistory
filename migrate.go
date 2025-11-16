package main

import (
	"database/sql"
)

func migrate(db *sql.DB) error {
	// users table
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	email TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL
);
`)
	if err != nil {
		return err
	}

	// plushies table
	_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS plushies (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	name TEXT NOT NULL,
	kind TEXT NOT NULL,
	adopted_at TEXT,
	image_path TEXT,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
`)
	return err
}


