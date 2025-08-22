package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"uwece.ca/app/db"
	"uwece.ca/app/models"
)

func redirect(w http.ResponseWriter, target string) {
	w.Header().Set("Location", target)
	w.WriteHeader(http.StatusFound)
}

func RequireBlog(d *db.DB, t bool) fn {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			usr, ok := GetUser(r)
			if !ok {
				panic("RequireBlog requires a user.")
			}

			_, err := models.GetSite(r.Context(), d, db.FilterEq("user_id", usr.Id))
			if errors.Is(err, db.ErrNoRows) && t {
				redirect(w, "/new-blog")
				return
			}
			if errors.Is(err, db.ErrNoRows) && !t {
				h.ServeHTTP(w, r)
				return
			}
			if err != nil && !errors.Is(err, db.ErrNoRows) {
				slog.Error("failed loading site from db", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if !t {
				redirect(w, "/site")
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
