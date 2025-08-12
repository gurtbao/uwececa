package db_test

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"uwece.ca/app/db"
	"uwece.ca/app/db/dbtest"
)

// Ensures two migrations with the same name will not be ran.
func TestMigrations(t *testing.T) {
	t.Parallel()

	d := dbtest.GetTestDB(t)

	err := d.RunMigrations([]db.Migration{
		db.FuncMigration("hallo", func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`
				create table if not exists migration_testing (id integer primary key);
		
				insert into migration_testing (id) values (1);
				`)
			if err != nil {
				return err
			}

			return nil
		}),

		db.FuncMigration("hallo", func(tx *sqlx.Tx) error {
			_, err := tx.Exec("delete from migration_testing")
			if err != nil {
				return err
			}

			return nil
		}),
	})
	require.NoError(t, err)

	var exists bool
	err = d.QueryRow("select exists (select 1 from migration_testing where id = 1)").Scan(&exists)
	require.NoError(t, err)

	require.True(t, exists)
}
