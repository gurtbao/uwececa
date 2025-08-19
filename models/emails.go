package models

import (
	"context"
	"errors"
	"time"

	"uwece.ca/app/auth"
	"uwece.ca/app/db"
)

type Email struct {
	Id      int        `db:"id"`
	UserId  int        `db:"user_id"`
	Token   auth.Token `db:"token"`
	Expires time.Time  `db:"expires"`
}

type NewEmail struct {
	UserId  int
	Token   auth.Token
	Expires time.Time
}

func InsertEmail(ctx context.Context, d db.Ex, newEmail NewEmail) (Email, error) {
	query := `insert into emails (user_id, token, expires) values (?, ?, ?) returning *`

	var email Email
	err := db.GetContext(ctx, d, &email, query, newEmail.UserId, newEmail.Token, newEmail.Expires)
	if err != nil {
		return Email{}, db.HandleError(err)
	}

	return email, nil
}

func GetEmail(ctx context.Context, d db.Ex, filters ...db.Filter) (Email, error) {
	if len(filters) == 0 {
		return Email{}, errors.New("must provide filters to get_email")
	}

	where, args := db.BuildWhere(filters)

	var email Email
	err := db.GetContext(ctx, d, &email, `select * from emails`+where, args...)
	if err != nil {
		return Email{}, db.HandleError(err)
	}

	return email, nil
}

func GetEmails(ctx context.Context, d db.Ex, filters ...db.Filter) ([]Email, error) {
	where, args := db.BuildWhere(filters)

	var emails []Email
	err := db.SelectContext(ctx, d, &emails, `select * from emails`+where, args...)
	if err != nil {
		return nil, db.HandleError(err)
	}

	return emails, nil
}
