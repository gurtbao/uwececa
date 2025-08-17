package models

import (
	"uwece.ca/app/db"
)

var Migrations = []db.Migration{
	db.FuncMigration("0001_add_users", func(tx db.Ex) error {
		_, err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS users (
    			id INTEGER PRIMARY KEY AUTOINCREMENT,
    			email VARCHAR(255) NOT NULL UNIQUE,
    			password VARCHAR(255) NOT NULL,
    			verified_at integer,
    			created_at integer NOT NULL DEFAULT CURRENT_TIMESTAMP,
    			updated_at integer NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
		`)
		return err
	}),
	db.FuncMigration("0002_add_sessions", func(tx db.Ex) error {
		_, err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS sessions (
				id integer primary key AUTOINCREMENT,
				user_id integer not null references users (id),
				token varchar(32) not null,
				expires timestamp not null 
			);

			create index sessions_user_id_idx on sessions (user_id);
			create index sessions_token_idx on sessions (token);
		`)
		return err
	}),
}
