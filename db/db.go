package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sqlx.DB
}

type Execer interface {
	Query(query string, args ...any) (*sqlx.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error)
	QueryRow(query string, args ...any) *sqlx.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sqlx.Row
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Prepare(query string) (*sqlx.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sqlx.Stmt, error)
}

func New(path string) (*DB, error) {
	db, err := sqlx.Connect("sqlite3", path)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		-- Set the journal mode to Write-Ahead Logging for concurrency
		PRAGMA journal_mode = WAL;

		-- Set synchronous mode to NORMAL for performance and data safety balance
		PRAGMA synchronous = NORMAL;

		-- Set busy timeout to 10 seconds to avoid "database is locked" errors
		PRAGMA busy_timeout = 10000;

		-- Set cache size to 20MB for faster data access
		PRAGMA cache_size = -20000;

		-- Enable foreign key constraint enforcement
		PRAGMA foreign_keys = ON;

		-- Enable auto vacuuming and set it to incremental mode for gradual space reclaiming
		PRAGMA auto_vacuum = INCREMENTAL;

		-- Store temporary tables and data in memory for better performance
		PRAGMA temp_store = MEMORY;

		-- Set the mmap_size to 2GB for faster read/write access using memory-mapped I/O
		PRAGMA mmap_size = 2147483648;

		-- Set the page size to 8KB for balanced memory usage and performance
		PRAGMA page_size = 8192;

		create table if not exists migrations (
			id integer primary key autoincrement,
			name text unique
		);
		`)
	if err != nil {
		return nil, fmt.Errorf("error bringing up db: %w", err)
	}

	return &DB{db}, nil
}
