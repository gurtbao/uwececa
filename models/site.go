package models

import (
	"context"
	"errors"
	"time"

	"uwece.ca/app/db"
)

type Site struct {
	Id int `db:"id"`

	UserId    int    `db:"user_id"`
	Subdomain string `db:"subdomain"`

	HomeContent      string `db:"home_content"`
	Navbar           string `db:"navbar"`
	CustomStylesheet string `db:"custom_stylesheet"`

	VerifiedAt *time.Time `db:"verified_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
	CreatedAt  time.Time  `db:"created_at"`
}

type NewSite struct {
	UserId    int
	Subdomain string

	HomeContent      string
	Navbar           string
	CustomStylesheet string
}

func InsertSite(ctx context.Context, d db.Ex, ns NewSite) (Site, error) {
	query := `
		insert into sites (
			user_id, 
			subdomain, 
			home_content, 
			custom_stylesheet,
			navbar
		) 
		values 
		(?, ?, ?, ?, ?) 
		returning *`

	var site Site
	if err := db.GetContext(ctx, d, &site, query, ns.UserId, ns.Subdomain, ns.HomeContent, ns.CustomStylesheet, ns.Navbar); err != nil {
		return Site{}, db.HandleError(err)
	}

	return site, nil
}

func GetSite(ctx context.Context, d db.Ex, filters ...db.Filter) (Site, error) {
	if len(filters) == 0 {
		return Site{}, errors.New("must define filters")
	}
	where, args := db.BuildWhere(filters)

	var site Site
	if err := db.GetContext(ctx, d, &site, `select * from sites`+where, args...); err != nil {
		return Site{}, db.HandleError(err)
	}

	return site, nil
}

func GetSites(ctx context.Context, d db.Ex, filters ...db.Filter) ([]Site, error) {
	where, args := db.BuildWhere(filters)

	var site []Site
	if err := db.SelectContext(ctx, d, &site, `select * from sites`+where, args...); err != nil {
		return nil, db.HandleError(err)
	}

	return site, nil
}
