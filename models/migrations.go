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
}
