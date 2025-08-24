package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"uwece.ca/app/db"
	"uwece.ca/app/models"
	"uwece.ca/app/web"
)

var userContextKey = struct{ I int }{I: 1}

func loadUser(d *db.DB, w http.ResponseWriter, r *http.Request) (models.User, error) {
	session, ok := web.GetSession(r)
	if !ok {
		return models.User{}, errors.New("failed to get session (session not found)")
	}

	dbs, err := models.GetSession(r.Context(), d, db.FilterEq("token", session.Token))
	if err != nil {
		if !errors.Is(err, db.ErrNoRows) {
			panic(fmt.Sprintf("db error: %v", err))
		}
		web.DeleteSession(w)
		return models.User{}, err
	}

	if time.Now().After(dbs.Expires) {
		web.DeleteSession(w)
		return models.User{}, err
	}

	usr, err := models.GetUser(r.Context(), d, db.FilterEq("id", dbs.UserId))
	if err != nil {
		if !errors.Is(err, db.ErrNoRows) {
			panic(fmt.Sprintf("db error: %v", err))
		}
		web.DeleteSession(w)
		return models.User{}, err
	}

	return usr, nil
}

func GetUser(r *http.Request) (models.User, bool) {
	u := r.Context().Value(userContextKey)

	usr, ok := u.(models.User)

	return usr, ok
}

func GetUserRef(r *http.Request) *models.User {
	usr, ok := GetUser(r)
	if !ok {
		return nil
	}

	return &usr
}

func LoadUser(d *db.DB) web.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			usr, err := loadUser(d, w, r)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			ctx := context.WithValue(r.Context(), userContextKey, usr)

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

func RequireLogin(t bool) web.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, ok := GetUser(r)
			if !ok == t {
				if t {
					web.Redirect(w, "/login")
				} else {
					web.Redirect(w, "/site")
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
