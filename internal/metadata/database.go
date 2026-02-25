package metadata

import (
	"database/sql"
	"github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(filepath string) (*Database, error) {
	sqlDB, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}
	database := &Database{db: sqlDB}

	if err := database.initTables(); err != nil {
		return nil, err
	}

	return database, nil
}

func (d *Database) initTables() error {
	query := `
	CREATE TABLES IF NOT EXISTS files (
		id TEXT PRIMARY KEY,
		path TEXT NOT NULL UNIQUE,
		hash TEXT NOT NULL,
		size INTEGER NOT NULL,
		mod_time DATETIME NOT NULL,
		is_dir BOOLEAN NOT NULL,
		version INTEGER DEFAULT 0,
		remote_hash TEXT,
		last_sync_time DATETIME,
		created_at DATETIME,
		deleted BOOLEAN DEFAULT 0
	);

	CREATE INDEX IF NOT EXISTS idx_path ON files(path);
	`

	_, err := d.db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}
