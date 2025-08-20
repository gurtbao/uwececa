package models_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"uwece.ca/app/db"
	"uwece.ca/app/db/dbtest"
	"uwece.ca/app/models"
)

func TestUserInsert(t *testing.T) {
	t.Parallel()
	d := dbtest.GetTestDB(t)

	require.NoError(t, d.RunMigrations(models.Migrations))

	ne := models.NewUser{
		Email:    "me@me.com",
		Password: "1234",
		Name:     "Lorenzo",
	}

	usr, err := models.InsertUser(context.Background(), d, ne)
	require.NoError(t, err)

	require.Equal(t, ne.Email, usr.Email)
	require.Equal(t, ne.Password, usr.Password)
	require.Equal(t, ne.Name, usr.Name)
}

func TestUserInsertGivesConflictError(t *testing.T) {
	t.Parallel()
	d := dbtest.GetTestDB(t)

	require.NoError(t, d.RunMigrations(models.Migrations))

	ne := models.NewUser{
		Email:    "me@me.com",
		Password: "1234",
	}

	_, err := models.InsertUser(context.Background(), d, ne)
	require.NoError(t, err)

	_, err = models.InsertUser(context.Background(), d, ne)
	require.ErrorIs(t, err, db.ErrUnique)
}

func TestGetUserExitsWithNoFilters(t *testing.T) {
	t.Parallel()

	_, err := models.GetUser(context.Background(), nil)
	require.Error(t, err)
}

func TestGetUserGivesNotFoundError(t *testing.T) {
	t.Parallel()
	d := dbtest.GetTestDB(t)

	require.NoError(t, d.RunMigrations(models.Migrations))

	_, err := models.GetUser(context.Background(), d, db.FilterEq("id", 1000))
	require.ErrorIs(t, err, db.ErrNoRows)
}

func TestUserGet(t *testing.T) {
	t.Parallel()
	d := dbtest.GetTestDB(t)

	require.NoError(t, d.RunMigrations(models.Migrations))

	ne := models.NewUser{
		Email:    "me@me.com",
		Password: "1234",
	}

	usr, err := models.InsertUser(context.Background(), d, ne)
	require.NoError(t, err)

	result, err := models.GetUser(context.Background(), d, db.FilterEq("email", "me@me.com"))
	require.NoError(t, err)

	require.Equal(t, usr, result)
}

func TestVerifyUser(t *testing.T) {
	t.Parallel()
	d := dbtest.GetTestDB(t)
	require.NoError(t, d.RunMigrations(models.Migrations))

	id := SeedUser(t, d)

	start := time.Now()

	err := models.VerifyUser(context.Background(), d, db.FilterEq("id", id))
	require.NoError(t, err)

	usr, err := models.GetUser(context.Background(), d, db.FilterEq("id", id))
	require.NoError(t, err)

	require.NotNil(t, usr.VerifiedAt)
	require.True(t, start.Before(usr.UpdatedAt))
}
