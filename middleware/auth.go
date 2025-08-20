package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"uwece.ca/app/auth"
	"uwece.ca/app/db"
	"uwece.ca/app/models"
)

var userContextKey = struct{ I int }{I: 1}

func loadUser(d *db.DB, w http.ResponseWriter, r *http.Request) (models.User, error) {
	session, ok := auth.GetSession(r)
	if !ok {
		return models.User{}, errors.New("failed to get session (session not found)")
	}

	dbs, err := models.GetSession(r.Context(), d, db.FilterEq("token", session.Token))
	if err != nil {
		if !errors.Is(err, db.ErrNoRows) {
			panic(fmt.Sprintf("db error: %v", err))
		}
		auth.DeleteSession(w)
		return models.User{}, err
	}

	if time.Now().After(dbs.Expires) {
		auth.DeleteSession(w)
		return models.User{}, err
	}

	usr, err := models.GetUser(r.Context(), d, db.FilterEq("id", dbs.UserId))
	if err != nil {
		if !errors.Is(err, db.ErrNoRows) {
			panic(fmt.Sprintf("db error: %v", err))
		}
		auth.DeleteSession(w)
		return models.User{}, err
	}

	return usr, nil
}

func GetUser(r *http.Request) (models.User, bool) {
	u := r.Context().Value(userContextKey)

	usr, ok := u.(models.User)

	return usr, ok
}

func LoadUser(d *db.DB) func(next http.Handler) http.Handler {
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

func RequireLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := GetUser(r)
		if !ok {
			w.Header().Set("Location", "/login")
			w.WriteHeader(http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}
