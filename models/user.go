package models

import (
	"context"
	"errors"

	"uwece.ca/app/db"
)

type User struct {
	Id    int    `db:"id"`
	Email string `db:"email"`

	Password string `db:"password"`

	VerifiedAt *string `db:"verified_at"`
	CreatedAt  string  `db:"created_at"`
	UpdatedAt  string  `db:"updated_at"`
}

type NewUser struct {
	Email    string
	Password string
}

func InsertUser(ctx context.Context, d db.Ex, user NewUser) (User, error) {
	query := `insert into users (email, password) values (?, ?) returning *`

	var usr User
	err := db.GetContext(ctx, d, &usr, query, user.Email, user.Password)
	if err != nil {
		return User{}, db.HandleError(err)
	}

	return usr, nil
}

func GetUser(ctx context.Context, d db.Ex, filters ...db.Filter) (User, error) {
	if len(filters) == 0 {
		return User{}, errors.New("get user called without filters")
	}

	where, args := db.BuildWhere(filters)

	var usr User
	err := db.GetContext(ctx, d, &usr, `select * from users`+where, args...)
	if err != nil {
		return User{}, db.HandleError(err)
	}

	return usr, nil
}
