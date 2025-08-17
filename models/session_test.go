package models_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"uwece.ca/app/db"
	"uwece.ca/app/db/dbtest"
	"uwece.ca/app/models"
)

func SeedUser(t *testing.T, d db.Ex) int {
	t.Helper()

	usr, err := models.InsertUser(context.Background(), d, models.NewUser{Email: "hi", Password: "hi"})
	require.NoError(t, err)

	return usr.Id
}

func TestInsertSession(t *testing.T) {
	t.Parallel()

	d := dbtest.GetTestDB(t)
	require.NoError(t, d.RunMigrations(models.Migrations))
	id := SeedUser(t, d)

	newS := models.NewSession{
		UserId: id,
		Token:  "1234",
	}

	s, err := models.InsertSession(context.Background(), d, newS)
	require.NoError(t, err)

	require.Equal(t, newS.UserId, s.UserId)
	require.Equal(t, newS.Token, s.Token)
}

func TestGetSession(t *testing.T) {
	t.Parallel()

	d := dbtest.GetTestDB(t)
	require.NoError(t, d.RunMigrations(models.Migrations))
	id := SeedUser(t, d)

	newS := models.NewSession{
		UserId: id,
		Token:  "1234",
	}

	s, err := models.InsertSession(context.Background(), d, newS)
	require.NoError(t, err)

	s2, err := models.GetSession(context.Background(), d, db.FilterEq("token", s.Token))
	require.NoError(t, err)

	require.Equal(t, s.UserId, s2.UserId)
}

func TestGetSessionExistsWithNoFilters(t *testing.T) {
	t.Parallel()

	_, err := models.GetSession(context.Background(), nil)
	require.Error(t, err)
}

func TestGetSessionGivesNotFoundError(t *testing.T) {
	t.Parallel()

	d := dbtest.GetTestDB(t)
	require.NoError(t, d.RunMigrations(models.Migrations))

	_, err := models.GetSession(context.Background(), d, db.FilterEq("user_id", 10000))
	require.ErrorIs(t, err, db.ErrNoRows)
}

func TestGetSessions(t *testing.T) {
	t.Parallel()

	d := dbtest.GetTestDB(t)
	require.NoError(t, d.RunMigrations(models.Migrations))
	id := SeedUser(t, d)

	newSessions := []models.NewSession{
		{
			UserId: id,
			Token:  "1234",
		},
		{
			UserId: id,
			Token:  "123445",
		},
	}

	for _, v := range newSessions {
		_, err := models.InsertSession(context.Background(), d, v)
		require.NoError(t, err)
	}

	s, err := models.GetSessions(context.Background(), d, db.FilterIn("token", []string{"1234", "123445"}))
	require.NoError(t, err)

	require.Len(t, s, 2)
}

func TestGetSessionsFetchNoFilter(t *testing.T) {
	t.Parallel()

	d := dbtest.GetTestDB(t)
	require.NoError(t, d.RunMigrations(models.Migrations))
	id := SeedUser(t, d)

	newSessions := []models.NewSession{
		{
			UserId: id,
			Token:  "1234",
		},
		{
			UserId: id,
			Token:  "123445",
		},
	}

	for _, v := range newSessions {
		_, err := models.InsertSession(context.Background(), d, v)
		require.NoError(t, err)
	}

	s, err := models.GetSessions(context.Background(), d)
	require.NoError(t, err)

	require.Len(t, s, 2)
}

func TestGetSessionsNoErrorNotFound(t *testing.T) {
	t.Parallel()

	d := dbtest.GetTestDB(t)
	require.NoError(t, d.RunMigrations(models.Migrations))

	s, err := models.GetSessions(context.Background(), d)
	require.NoError(t, err)

	require.Empty(t, s)
}
