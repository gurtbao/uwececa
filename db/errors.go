package db

import (
	"database/sql"
	"errors"

	sqlite "github.com/mattn/go-sqlite3"
)

var (
	ErrNoRows     = errors.New("no rows found for query")
	ErrForeignKey = errors.New("foreign key violation")
	ErrNotNull    = errors.New("not null violation")
	ErrUnique     = errors.New("unique violation")
)

func HandleError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNoRows
	}

	var sqliteErr sqlite.Error
	if errors.As(err, &sqliteErr) {
		err := sqliteErr.ExtendedCode
		if errors.Is(err, sqlite.ErrConstraintUnique) {
			return ErrUnique
		}

		if errors.Is(err, sqlite.ErrConstraintForeignKey) {
			return ErrForeignKey
		}

		if errors.Is(err, sqlite.ErrConstraintNotNull) {
			return ErrNotNull
		}

		if errors.Is(err, sqlite.ErrConstraintPrimaryKey) {
			return ErrUnique
		}

		if errors.Is(err, sqlite.ErrConstraintUnique) {
			return ErrUnique
		}
	}

	return err
}
