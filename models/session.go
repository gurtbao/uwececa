package models

import (
	"context"
	"errors"
	"time"

	"uwece.ca/app/auth"
	"uwece.ca/app/db"
)

type Session struct {
	Id      int        `db:"id"`
	UserId  int        `db:"user_id"`
	Token   auth.Token `db:"token"`
	Expires time.Time  `db:"expires"`
}

type NewSession struct {
	UserId  int
	Token   auth.Token
	Expires time.Time
}

func InsertSession(ctx context.Context, d db.Ex, newSession NewSession) (Session, error) {
	query := `insert into sessions (user_id, token, expires) values (?, ?, ?) returning *`

	var session Session
	err := db.GetContext(ctx, d, &session, query, newSession.UserId, newSession.Token, newSession.Expires)
	if err != nil {
		return Session{}, db.HandleError(err)
	}

	return session, nil
}

func GetSession(ctx context.Context, d db.Ex, filters ...db.Filter) (Session, error) {
	if len(filters) == 0 {
		return Session{}, errors.New("must provide filters to get_session")
	}

	where, args := db.BuildWhere(filters)

	var session Session
	err := db.GetContext(ctx, d, &session, `select * from sessions`+where, args...)
	if err != nil {
		return Session{}, db.HandleError(err)
	}

	return session, nil
}

func GetSessions(ctx context.Context, d db.Ex, filters ...db.Filter) ([]Session, error) {
	where, args := db.BuildWhere(filters)

	var sessions []Session
	err := db.SelectContext(ctx, d, &sessions, `select * from sessions`+where, args...)
	if err != nil {
		return nil, db.HandleError(err)
	}

	return sessions, nil
}
