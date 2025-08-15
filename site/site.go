package site

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	chimd "github.com/go-chi/chi/v5/middleware"
	"uwece.ca/app/config"
	"uwece.ca/app/db"
	"uwece.ca/app/middleware"
	"uwece.ca/app/templates"
)

//go:embed templates/* static/*
var embedFS embed.FS

type Site struct {
	config    *config.Config
	db        *db.DB
	templates *templates.Templates
	embedFS   embed.FS
}

func New(c *config.Config, db *db.DB) *Site {
	var tmpl *templates.Templates
	if c.Core.Development {
		tmpl = templates.NewDevTemplates(embedFS, "./site/templates")
	} else {
		tmpl = templates.NewTemplates(embedFS)
	}

	return &Site{
		config:    c,
		db:        db,
		templates: tmpl,
		embedFS:   embedFS,
	}
}

func (s *Site) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.LogRecover)
	r.Use(chimd.Compress(5))

	r.Handle("/static/*", s.Static())
	r.Get("/", s.Index)

	return r
}

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

func (s *Site) RenderPlain(w http.ResponseWriter, statusCode int, name string, params any) {
	s.RenderTemplate(w, statusCode, name, "", params)
}

func (s *Site) RenderTemplate(w http.ResponseWriter, statusCode int, name, base string, params any) {
	err := s.templates.Execute(name, w, base, params)
	if err != nil {
		s.UnhandledError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(statusCode)
}

func (s *Site) Static() http.Handler {
	if s.config.Core.Development {
		return http.StripPrefix("/static", http.FileServer(http.Dir("./site/static")))
	}

	sub, err := fs.Sub(s.embedFS, "static")
	if err != nil {
		slog.Error("no static dir in mainsite", "error", err)
		os.Exit(1)
	}

	return http.StripPrefix("/static", http.FileServer(http.FS(sub)))
}

func HxRefresh(w http.ResponseWriter) {
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func HxRedirect(w http.ResponseWriter, target string) {
	w.Header().Set("HX-Redirect", target)
	w.WriteHeader(http.StatusOK)
}

func HxLocation(w http.ResponseWriter, target string) {
	w.Header().Set("HX-Location", target)
	w.WriteHeader(http.StatusOK)
}

func (s *Site) Index(w http.ResponseWriter, r *http.Request) {
	s.RenderPlain(w, http.StatusOK, "home", nil)
}
