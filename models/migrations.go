package models

import (
	"uwece.ca/app/db"
)

var Migrations = []db.Migration{
	db.FuncMigration("0001_add_users", func(tx db.Ex) error {
		_, err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS users (
    			id INTEGER PRIMARY KEY AUTOINCREMENT,
    			net_id VARCHAR(255) NOT NULL UNIQUE,
    			password VARCHAR(255) NOT NULL,
				name VARCHAR(255) NOT NULL,
    			verified_at timestamp,
    			created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    			updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
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
	db.FuncMigration("0003_add_email_links", func(tx db.Ex) error {
		_, err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS emails (
				id integer primary key AUTOINCREMENT,
				user_id integer not null references users (id),
				token varchar(32) not null,
				expires timestamp not null 
			);

			create index emails_user_id_idx on sessions (user_id);
			create index emails_token_idx on sessions (token);
		`)
		return err
	}),
	db.FuncMigration("0004_add_sites", func(tx db.Ex) error {
		_, err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS sites (
				id integer primary key AUTOINCREMENT,

				user_id integer not null unique references users (id),
				subdomain varchar(255) not null unique,
				home_content varchar,
				navbar varchar,
				custom_stylesheet varchar,

				verified_at timestamp,
    			created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    			updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
			);

			create index sites_user_id_idx on sites (user_id);
			create index sites_subdomain_idx on sites (subdomain);
		`)
		return err
	}),
}
