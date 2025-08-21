package models_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"uwece.ca/app/db"
	"uwece.ca/app/db/dbtest"
	"uwece.ca/app/models"
)

func SeedSite(t *testing.T, d db.Ex) int {
	t.Helper()
	id := SeedUser(t, d)

	s, err := models.InsertSite(context.Background(), d, models.NewSite{
		UserId:    id,
		Subdomain: "zach.30",
	})
	require.NoError(t, err)

	return s.Id
}

func TestSiteInsert(t *testing.T) {
	t.Parallel()
	d := dbtest.GetTestDB(t)

	require.NoError(t, d.RunMigrations(models.Migrations))

	id := SeedUser(t, d)

	ns := models.NewSite{
		UserId:      id,
		Subdomain:   "zach.30",
		HomeContent: "Hi.",
		Navbar:      "[Home](/)",
	}

	site, err := models.InsertSite(context.Background(), d, ns)
	require.NoError(t, err)

	require.Equal(t, ns.UserId, site.UserId)
	require.Equal(t, ns.Subdomain, site.Subdomain)
	require.Equal(t, ns.HomeContent, site.HomeContent)
	require.Equal(t, ns.Navbar, site.Navbar)
	require.Equal(t, ns.CustomStylesheet, site.CustomStylesheet)
}

func TestGetSite(t *testing.T) {
	t.Parallel()
	d := dbtest.GetTestDB(t)

	require.NoError(t, d.RunMigrations(models.Migrations))

	id := SeedUser(t, d)

	ns := models.NewSite{
		UserId:      id,
		Subdomain:   "zach.30",
		HomeContent: "Hi.",
		Navbar:      "[Home](/)",
	}

	site, err := models.InsertSite(context.Background(), d, ns)
	require.NoError(t, err)

	g, err := models.GetSite(context.Background(), d, db.FilterEq("id", site.Id))
	require.NoError(t, err)

	require.Equal(t, site, g)
}

func TestGetSiteErrorOnNoFilters(t *testing.T) {
	t.Parallel()
	_, err := models.GetSite(context.Background(), nil)
	require.Error(t, err)
}

func TestGetSites(t *testing.T) {
	t.Parallel()

	d := dbtest.GetTestDB(t)

	require.NoError(t, d.RunMigrations(models.Migrations))

	id := SeedUser(t, d)

	ns := models.NewSite{
		UserId:      id,
		Subdomain:   "zach.30",
		HomeContent: "Hi.",
		Navbar:      "[Home](/)",
	}

	site, err := models.InsertSite(context.Background(), d, ns)
	require.NoError(t, err)

	sites, err := models.GetSites(context.Background(), d)
	require.NoError(t, err)

	require.Equal(t, site, sites[0])
}
