package metadata

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
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

	if err := database.InitTable(); err != nil {
		return nil, err
	}

	return database, nil
}

func (d *Database) InitTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS files (
		db_id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT NOT NULL UNIQUE,
		hash TEXT NOT NULL,
		size INTEGER NOT NULL,
		mod_time DATETIME NOT NULL,
		is_dir BOOLEAN NOT NULL,
		version INTEGER DEFAULT 0,
		remote_hash TEXT,
		last_sync_time DATETIME,
		created_at DATETIME,
		deleted BOOLEAN DEFAULT 0,
		inode INTEGER NOT NULL
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
		path, hash, size, mod_time, is_dir, version, remote_hash, last_sync_time, created_at, deleted, inode
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := d.db.Exec(query, file.Path, file.Hash, file.Size, file.ModTime, file.IsDir, file.Version, file.RemoteHash, file.LastSyncTime, file.CreatedAt, file.Deleted, file.Inode)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) GetFileByPath(path string) (*File, error) {
	query := `
	SELECT path, hash, size, mod_time, is_dir, version, remote_hash, last_sync_time, created_at, deleted, inode
	FROM files
	WHERE path = ?
	`

	var file File

	err := d.db.QueryRow(query, path).Scan(
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
		&file.Inode,
	)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	return &file, nil
}

/*
	 this function should be called only when:
		file event has been processed and server has been updated
		file deletion occured, set file.Deleted to true
*/
func (d *Database) UpdateFile(file *File) error {
	query := `
	UPDATE files
	SET hash = ?, size = ?, mod_time = ?, version = version + 1, remote_hash = ?, last_sync_time = ?, deleted = ?
	WHERE path = ?
	`

	_, err := d.db.Exec(query, file.Hash, file.Size, file.ModTime, file.Version, file.RemoteHash, file.LastSyncTime, file.Deleted, file.Path)
	if err != nil {
		return err
	}

	return nil
}

// only use this to clear up the metadata db when needed
func (d *Database) DeleteFile(path string) error {
	query := `
	DELETE FROM files WHERE path = ?
	`

	_, err := d.db.Exec(query, path)
	if err != nil {
		return err
	}

	return nil
}
