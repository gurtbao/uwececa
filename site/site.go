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
	"uwece.ca/app/email"
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
	mailer    *email.Mailer
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
		mailer:    email.NewMailer(c),
	}
}

func (s *Site) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.LogRecover)
	r.Use(chimd.Compress(5))

	r.Handle("/static/*", s.Static())

	r.Group(func(r chi.Router) {
		r.Use(middleware.LoadUser(s.db))
		r.Get("/", s.Index)

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireLogin(false))
			r.Get("/login", s.LoginPage)
			r.Post("/login", s.LoginHandler)

			r.Get("/signup", s.SignupPage)
			r.Post("/signup", s.SignupHandler)

			r.Get("/signup/email-verification", s.EmailVerificationPage)
			r.Get("/signup/verify/{token}", s.EmailVerificationHandler)
		})

		r.NotFound(s.NotFound)
	})

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

func (s *Site) Index(w http.ResponseWriter, r *http.Request) {
	s.RenderTemplate(w, http.StatusOK, "public/home", "layouts/public-base", nil)
}

func (s *Site) NotFound(w http.ResponseWriter, r *http.Request) {
	s.FullpageError(w, fullPageErrorParams{
		Code:    http.StatusNotFound,
		Message: "Resource Not Found.",
	})
}

type fullPageErrorParams struct {
	Code    int
	Message string
}

func (s *Site) FullpageError(w http.ResponseWriter, params fullPageErrorParams) {
	s.RenderTemplate(w, params.Code, "public/error", "layouts/public-base", params)
}

type alertErrorParams struct {
	Variant string
	Message string
}

func (s *Site) AlertError(w http.ResponseWriter, params alertErrorParams) {
	s.RenderPlain(w, http.StatusOK, "public/alert-error", params)
}
