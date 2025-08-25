package site

import (
	"fmt"
	"log/slog"
	"net/http"

	"uwece.ca/app/templates"
)

func (s *Site) UnhandledError(w http.ResponseWriter, err error) {
	slog.Error("mainsite internal server error", "error", err)
	var response string
	if s.config.Core.Development {
		response = fmt.Sprintf("Internal Server Error: %v", err)
	} else {
		response = "Internal Server Error! Please Contact The Maintainers."
	}
	w.WriteHeader(http.StatusInternalServerError)
	_, err = w.Write([]byte(response))
	if err != nil {
		panic(fmt.Sprintf("failed to send internal server error status to client: %v", err))
	}
}

func (s *Site) FullpageError(w http.ResponseWriter, r *http.Request, code int, message string) error {
	ctx := s.BaseContext(r)
	ctx.Add("Code", code)
	ctx.Add("Message", message)

	return s.Render(w, code, "layouts/public-base", "public/error", ctx)
}

func (s *Site) NotFound(w http.ResponseWriter, r *http.Request) error {
	return s.FullpageError(w, r, http.StatusNotFound, "No resources found.")
}

func (s *Site) DangerAlert(w http.ResponseWriter, message string) error {
	return s.RenderPlain(w, http.StatusOK, "public/alert", templates.Context{
		"message": message,
		"variant": "danger",
	})
}

func (s *Site) WarnAlert(w http.ResponseWriter, message string) error {
	return s.RenderPlain(w, http.StatusOK, "public/alert", templates.Context{
		"message": message,
		"variant": "warning",
	})
}
