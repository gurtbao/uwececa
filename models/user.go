package models

import (
	"context"
	"errors"
	"time"

	"uwece.ca/app/db"
)

type User struct {
	Id    int    `db:"id"`
	Email string `db:"email"`
	Name  string `db:"name"`

	Password string `db:"password"`

	VerifiedAt *time.Time `db:"verified_at"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
}

type NewUser struct {
	Email    string
	Name     string
	Password string
}

func InsertUser(ctx context.Context, d db.Ex, user NewUser) (User, error) {
	query := `insert into users (email, name, password) values (?, ?, ?) returning *`

	var usr User
	err := db.GetContext(ctx, d, &usr, query, user.Email, user.Name, user.Password)
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

func VerifyUser(ctx context.Context, d db.Ex, filters ...db.Filter) error {
	where, args := db.BuildWhere(filters)
	now := time.Now()

	newArgs := []any{now, now}
	newArgs = append(newArgs, args...)

	if _, err := d.ExecContext(ctx, `update users set verified_at = ?, updated_at = ?`+where, newArgs...); err != nil {
		return db.HandleError(err)
	}

	return nil
}
