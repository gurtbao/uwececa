package middleware

import (
	"log/slog"
	"net/http"
	"time"

	chimid "github.com/go-chi/chi/v5/middleware"
)

type fn func(http.Handler) http.Handler

func LogRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now().UTC()
		ww := chimid.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			if err := recover(); err != nil {
				slog.Error(
					"request handler panic",
					"status", ww.Status(),
					"duration", time.Since(start),
					"path", r.URL.Path,
					"panic", err,
					"ua", r.Header.Get("User-Agent"),
				)
			} else {
				logFn := slog.Debug
				mills := time.Since(start) / time.Millisecond

				if mills >= 500 {
					logFn = slog.Info
				}
				if ww.Status() >= 500 {
					logFn = slog.Error
				}

				logFn(
					"http request",
					"status", ww.Status(),
					"duration", time.Since(start),
					"path", r.URL.Path,
					"ua", r.Header.Get("User-Agent"),
				)
			}
		}()

		next.ServeHTTP(ww, r)
	})
}
