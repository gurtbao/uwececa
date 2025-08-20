package web

import (
	"log/slog"
	"net/http"
)

type HandleFunc func(w http.ResponseWriter, r *http.Request) error

func (fn HandleFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := fn(w, r)
	if err != nil {
		slog.Error("handler error", "handled_by", "web", "error", err, "path", r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("Internal Server Error."))
		if err != nil {
			slog.Error("error sending internal server error to client", "error", err)
		}
	}
}

type ErrorWrapper interface {
	UnhandledError(w http.ResponseWriter, err error)
}

type HandlerWrapper struct {
	errWrapper ErrorWrapper
}

func NewHandlerWrapper(wr ErrorWrapper) *HandlerWrapper {
	return &HandlerWrapper{
		errWrapper: wr,
	}
}

func (h *HandlerWrapper) Wrap(fn HandleFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r)
		if err != nil {
			h.errWrapper.UnhandledError(w, err)
		}
	}
}
