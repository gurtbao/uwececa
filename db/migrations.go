package db

import (
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
)

type Migration struct {
	Name string
	Func func(tx *sqlx.Tx) error
}

func FuncMigration(name string, fn func(tx *sqlx.Tx) error) Migration {
	return Migration{
		Name: name,
		Func: fn,
	}
}

func (d *DB) RunMigrations(migrations []Migration) error {
	for _, v := range migrations {
		if err := runMigration(d, v); err != nil {
			return err
		}
	}

	return nil
}

func runMigration(d *DB, m Migration) error {
	tx, err := d.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRow("select exists (select 1 from migrations where name = ?)", m.Name).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		err = m.Func(tx)
		if err != nil {
			return fmt.Errorf("error applying migration %s: %w", m.Name, err)
		}

		_, err := tx.Exec("insert into migrations (name) values (?)", m.Name)
		if err != nil {
			return fmt.Errorf("error marking migration as complete: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		slog.Info("migration applied sucessfully", "name", m.Name)
	} else {
		slog.Debug("skipped migration", "name", m.Name)
	}

	return nil
}
