package models

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"uwece.ca/app/db"
)

type User struct {
	Id    int    `db:"id"`
	NetID string `db:"net_id"`
	Name  string `db:"name"`

	Password string `db:"password"`

	VerifiedAt *time.Time `db:"verified_at"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
}

type NewUser struct {
	NetID    string
	Name     string
	Password string
}

func InsertUser(ctx context.Context, d db.Ex, user NewUser) (User, error) {
	query := `insert into users (net_id, name, password) values (?, ?, ?) returning *`

	var usr User
	err := db.GetContext(ctx, d, &usr, query, user.NetID, user.Name, user.Password)
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

func UpdateUser(ctx context.Context, d db.Ex, updates []db.UpdateData, filters ...db.Filter) error {
	if len(filters) == 0 {
		slog.Debug("calling user update without filters")
	}
	where, args := db.BuildWhere(filters)
	keys, values := db.BuildUpdate(updates)

	values = append(values, args...)

	if _, err := d.ExecContext(ctx, `update users`+keys+where, values...); err != nil {
		return db.HandleError(err)
	}

	return nil
}
