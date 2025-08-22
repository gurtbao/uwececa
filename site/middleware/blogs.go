package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"uwece.ca/app/db"
	"uwece.ca/app/models"
)

func redirect(w http.ResponseWriter, target string) {
	w.Header().Set("Location", target)
	w.WriteHeader(http.StatusFound)
}

var blogContextKey = struct{ K int }{1}

func loadBlog(d *db.DB, r *http.Request) (int, error) {
	usr, ok := GetUser(r)
	if !ok {
		return 0, errors.New("Load blog requires a user in the request ctx.")
	}

	site, err := models.GetSite(r.Context(), d, db.FilterEq("user_id", usr.Id))
	if err != nil {
		if !errors.Is(err, db.ErrNoRows) {
			panic(fmt.Sprintf("db error: %v", err))
		}

		return 0, err
	}

	return site.Id, nil
}

func GetBlog(r *http.Request) (int, bool) {
	u := r.Context().Value(blogContextKey)
	s, ok := u.(int)
	return s, ok
}

func LoadBlog(db *db.DB) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			blogID, err := loadBlog(db, r)
			if err != nil {
				h.ServeHTTP(w, r)
			}

			ctx := context.WithValue(r.Context(), blogContextKey, blogID)

			r = r.WithContext(ctx)

			h.ServeHTTP(w, r)
		})
	}
}

func RequireBlog(t bool) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, ok := GetBlog(r)
			if !ok == t {
				if t {
					redirect(w, "/new-blog")
				} else {
					redirect(w, "/site")
				}

				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
