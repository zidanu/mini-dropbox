package metadata

import (
	"database/sql"
	"github.com/google/uuid"
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

func (d *Database) SaveFile(file *File) error {
	query := `
	INSERT INTO files (
		id, path, hash, size, mod_time, is_dir, version, remote_hash, last_sync_time, created_at, deleted
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := d.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) GetFile(path string) (*File, error) {
	query := `
	SELECT id, path, hash, size, mod_time, is_dir, version, remote_hash, last_sync_time, created_at, deleted
	FROM files
	WHERE path = ?
	`

	var file File
	var idStr string

	err := d.db.QueryRow(query, path).Scan(
		&idStr,
		&file.Path,
		&file.Hash,
		&file.Size,
		&file.ModTime,
		&file.IsDir,
		&file.Version,
		&file.RemoteHash,
		&file.LastSyncTime,
		&file.CreatedAt,
		&file.Deleted,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	file.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}

	return &file, nil
}
