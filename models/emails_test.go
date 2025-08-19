package models_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"uwece.ca/app/db"
	"uwece.ca/app/db/dbtest"
	"uwece.ca/app/models"
)

func TestInsertEmail(t *testing.T) {
	t.Parallel()

	d := dbtest.GetTestDB(t)
	require.NoError(t, d.RunMigrations(models.Migrations))
	id := SeedUser(t, d)

	newS := models.NewEmail{
		UserId: id,
		Token:  "1234",
	}

	s, err := models.InsertEmail(context.Background(), d, newS)
	require.NoError(t, err)

	require.Equal(t, newS.UserId, s.UserId)
	require.Equal(t, newS.Token, s.Token)
}

func TestGetEmail(t *testing.T) {
	t.Parallel()

	d := dbtest.GetTestDB(t)
	require.NoError(t, d.RunMigrations(models.Migrations))
	id := SeedUser(t, d)

	newS := models.NewEmail{
		UserId: id,
		Token:  "1234",
	}

	s, err := models.InsertEmail(context.Background(), d, newS)
	require.NoError(t, err)

	s2, err := models.GetEmail(context.Background(), d, db.FilterEq("token", s.Token))
	require.NoError(t, err)

	require.Equal(t, s.UserId, s2.UserId)
}

func TestGetEmailExistsWithNoFilters(t *testing.T) {
	t.Parallel()

	_, err := models.GetEmail(context.Background(), nil)
	require.Error(t, err)
}

func TestGetEmailGivesNotFoundError(t *testing.T) {
	t.Parallel()

	d := dbtest.GetTestDB(t)
	require.NoError(t, d.RunMigrations(models.Migrations))

	_, err := models.GetEmail(context.Background(), d, db.FilterEq("user_id", 10000))
	require.ErrorIs(t, err, db.ErrNoRows)
}

func TestGetEmails(t *testing.T) {
	t.Parallel()

	d := dbtest.GetTestDB(t)
	require.NoError(t, d.RunMigrations(models.Migrations))
	id := SeedUser(t, d)

	newEmails := []models.NewEmail{
		{
			UserId: id,
			Token:  "1234",
		},
		{
			UserId: id,
			Token:  "123445",
		},
	}

	for _, v := range newEmails {
		_, err := models.InsertEmail(context.Background(), d, v)
		require.NoError(t, err)
	}

	s, err := models.GetEmails(context.Background(), d, db.FilterIn("token", []string{"1234", "123445"}))
	require.NoError(t, err)

	require.Len(t, s, 2)
}

func TestGetEmailsFetchNoFilter(t *testing.T) {
	t.Parallel()

	d := dbtest.GetTestDB(t)
	require.NoError(t, d.RunMigrations(models.Migrations))
	id := SeedUser(t, d)

	newEmails := []models.NewEmail{
		{
			UserId: id,
			Token:  "1234",
		},
		{
			UserId: id,
			Token:  "123445",
		},
	}

	for _, v := range newEmails {
		_, err := models.InsertEmail(context.Background(), d, v)
		require.NoError(t, err)
	}

	s, err := models.GetEmails(context.Background(), d)
	require.NoError(t, err)

	require.Len(t, s, 2)
}

func TestGetEmailsNoErrorNotFound(t *testing.T) {
	t.Parallel()

	d := dbtest.GetTestDB(t)
	require.NoError(t, d.RunMigrations(models.Migrations))

	s, err := models.GetEmails(context.Background(), d)
	require.NoError(t, err)

	require.Empty(t, s)
}
