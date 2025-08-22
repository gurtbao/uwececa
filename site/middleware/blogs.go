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

func loadBlog(d *db.DB, r *http.Request) (BlogContext, error) {
	var ctx BlogContext
	usr, ok := GetUser(r)
	if !ok {
		return ctx, errors.New("Load blog requires a user in the request ctx.")
	}

	site, err := models.GetSite(r.Context(), d, db.FilterEq("user_id", usr.Id))
	if err != nil {
		if !errors.Is(err, db.ErrNoRows) {
			panic(fmt.Sprintf("db error: %v", err))
		}

		return ctx, err
	}

	ctx = BlogContext{
		Id:       site.Id,
		Verified: site.VerifiedAt != nil,
	}

	return ctx, nil
}

func GetBlog(r *http.Request) (BlogContext, bool) {
	u := r.Context().Value(blogContextKey)
	s, ok := u.(BlogContext)
	return s, ok
}

type BlogContext struct {
	Id       int
	Verified bool
}

func LoadBlog(db *db.DB) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := loadBlog(db, r)
			if err != nil {
				h.ServeHTTP(w, r)
			}

			ctx := context.WithValue(r.Context(), blogContextKey, c)

			r = r.WithContext(ctx)

			h.ServeHTTP(w, r)
		})
	}
}

func RequireBlog(t bool) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, ok := GetBlog(r)
			if t != ok {
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

func RequireBlogVerified(t bool) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, ok := GetBlog(r)
			if !ok {
				panic("RequireBlogVerified requires a blog in ctx.")
			}

			if b.Verified != t {
				if t {
					redirect(w, "/site/blog-unverified")
				} else {
					redirect(w, "/site")
				}
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
