package site

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"uwece.ca/app/models"
	"uwece.ca/app/services"
	"uwece.ca/app/web"
)

var (
	userContextKey = struct{ I int }{I: 1}
	blogContextKey = struct{ K int }{1}
)

// Load a user into a request.
func (s *Site) LoadUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		se, ok := web.GetSession(r)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		usr, err := s.users.LoadSession(r.Context(), se.Token)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrSessionDoesNotExist):
				slog.Info("user attempted to log in with non-existent session", "token", se.Token)
			case errors.Is(err, services.ErrUserDoesNotExist):
				slog.Info("user attempted to log in with session for non-existend user", "token", se.Token)
			case errors.Is(err, services.ErrSessionExpired):
			default:
				slog.Error("error loading session from database", "error", err)
			}

			web.DeleteSession(w)
		} else {
			ctx := context.WithValue(r.Context(), userContextKey, usr)

			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

func ExtractUser(r *http.Request) *models.User {
	u := r.Context().Value(userContextKey)

	usr, ok := u.(models.User)
	if !ok {
		return nil
	}

	return &usr
}

func RequireLogin(t bool) web.Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pass := false
			usr := ExtractUser(r)
			if usr == nil {
				if !t {
					pass = true
				}
			} else {
				if t {
					pass = true
				}
			}

			if pass {
				h.ServeHTTP(w, r)
				return
			}

			if t {
				web.Redirect(w, "/login")
			} else {
				web.Redirect(w, "/site")
			}
		})
	}
}

func (s *Site) LoadBlog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := ExtractUser(r)
		if user == nil {
			panic("User required in handler for LoadBlog.")
		}

		blog, err := s.blogs.LoadBlogFromUser(r.Context(), user.Id)
		if err != nil {
			if !errors.Is(err, services.ErrBlogDoesNotExist) {
				slog.Error("error loading blog from database", "user_id", user.Id, "error", err)
			}
		} else {
			ctx := context.WithValue(r.Context(), blogContextKey, blog)

			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

func ExtractBlog(r *http.Request) *models.Site {
	b := r.Context().Value(blogContextKey)

	blog, ok := b.(models.Site)
	if !ok {
		return nil
	}

	return &blog
}

func RequireBlog(t, verified bool) web.Middleware {
	if !t && verified {
		panic("You can only require a blog to be verified if you are requiring it to exist.")
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pass := false
			verified_failed := false
			blog := ExtractBlog(r)
			if blog == nil {
				if !t {
					pass = true
				}
			} else {
				if t {
					pass = true
				}
			}

			if verified && blog.VerifiedAt == nil {
				pass = false
				verified_failed = true
			}

			if pass {
				h.ServeHTTP(w, r)
				return
			}

			if t {
				if verified_failed {
					web.Redirect(w, "/blog-unverified")
				} else {
					web.Redirect(w, "/new-blog")
				}
			} else {
				web.Redirect(w, "/site")
			}
		})
	}
}
