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
	"uwece.ca/app/models"
	"uwece.ca/app/templates"
	"uwece.ca/app/web"
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
	w := web.NewHandlerWrapper(s)
	r := chi.NewRouter()
	r.Use(middleware.LogRecover)
	r.Use(chimd.Compress(5))

	r.Handle("/static/*", s.Static())

	r.Group(func(r chi.Router) {
		r.Use(middleware.LoadUser(s.db))
		r.Get("/", w.Wrap(s.Index))

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireLogin(false))
			r.Get("/login", w.Wrap(s.LoginPage))
			r.Post("/login", w.Wrap(s.LoginHandler))

			r.Get("/signup", w.Wrap(s.SignupPage))
			r.Post("/signup", w.Wrap(s.SignupHandler))

			r.Get("/signup/email-verification", w.Wrap(s.EmailVerificationPage))
			r.Get("/signup/verify/{token}", w.Wrap(s.EmailVerificationHandler))
		})

		r.NotFound(w.Wrap(s.NotFound))
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

func (s *Site) RenderPlain(w http.ResponseWriter, statusCode int, name string, params any) error {
	return s.RenderTemplate(w, statusCode, name, "", params)
}

func (s *Site) RenderTemplate(w http.ResponseWriter, statusCode int, name, base string, params any) error {
	err := s.templates.Execute(name, w, base, params)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(statusCode)

	return nil
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

// Params for a base page.
type PageParams struct {
	LoggedInUser *models.User
}

func PageParamsFromReq(r *http.Request) PageParams {
	return PageParams{
		LoggedInUser: middleware.GetUserRef(r),
	}
}

func (s *Site) Index(w http.ResponseWriter, r *http.Request) error {
	return s.RenderTemplate(w, http.StatusOK, "public/home", "layouts/public-base", PageParamsFromReq(r))
}

func (s *Site) NotFound(w http.ResponseWriter, r *http.Request) error {
	return s.FullpageError(w, newFullpageErrorParams(r, http.StatusNotFound, "Resouce Not Found."))
}

type fullPageErrorParams struct {
	PageParams
	Code    int
	Message string
}

func newFullpageErrorParams(r *http.Request, code int, message string) fullPageErrorParams {
	return fullPageErrorParams{
		PageParams: PageParamsFromReq(r),
		Code:       code,
		Message:    message,
	}
}

func (s *Site) FullpageError(w http.ResponseWriter, params fullPageErrorParams) error {
	return s.RenderTemplate(w, params.Code, "public/error", "layouts/public-base", params)
}

type alertErrorParams struct {
	Variant string
	Message string
}

func (s *Site) AlertError(w http.ResponseWriter, params alertErrorParams) error {
	return s.RenderPlain(w, http.StatusOK, "public/alert-error", params)
}
