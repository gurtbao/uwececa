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
	"uwece.ca/app/models"
	"uwece.ca/app/site/middleware"
	"uwece.ca/app/templates"
	"uwece.ca/app/utils"
	"uwece.ca/app/web"
)

//go:embed templates/* static/*
var embedFS embed.FS

type Site struct {
	config    *config.Config
	db        *db.DB
	templates *templates.Templates
	embedFS   embed.FS
	mailer    *utils.Mailer
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
		mailer:    utils.NewMailer(c),
	}
}

func (s *Site) Routes() http.Handler {
	w := web.NewHandlerWrapper(s)
	r := chi.NewRouter()
	r.Use(web.MidLogRecover)
	r.Use(chimd.Compress(5))

	r.Handle("/static/*", s.Static())

	r.Group(func(r chi.Router) {
		r.Use(middleware.LoadUser(s.db))
		r.Get("/", w.Wrap(s.Index))

		// No login group.
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireLogin(false))
			r.Get("/login", w.Wrap(s.LoginPage))
			r.Post("/login", w.Wrap(s.LoginHandler))

			r.Get("/signup", w.Wrap(s.SignupPage))
			r.Post("/signup", w.Wrap(s.SignupHandler))

			r.Get("/signup/verify/{token}", w.Wrap(s.EmailVerificationHandler))
		})

		// Required Login Group.
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireLogin(true))
			r.Get("/logout", w.Wrap(s.Logout))

			r.Group(func(r chi.Router) {
				r.Use(middleware.LoadBlog(s.db))
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireBlog(false))
					r.Get("/new-blog", w.Wrap(s.NewBlogPage))
					r.Post("/new-blog", w.Wrap(s.NewBlogHandler))
				})
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireBlog(true), middleware.RequireBlogVerified(false))
					r.Get("/site/blog-unverified", w.Wrap(s.BlogUnverified))
				})
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireBlog(true))
					r.Use(middleware.RequireBlogVerified(true))
					r.Get("/site", w.Wrap(s.Index))
				})
			})
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

func (s *Site) RenderPlain(w http.ResponseWriter, r *http.Request, statusCode int, name string, params any) error {
	return s.RenderTemplate(w, r, statusCode, name, "", params)
}

func (s *Site) RenderTemplate(w http.ResponseWriter, r *http.Request, statusCode int, name, base string, params any) error {
	p := PageParams{
		Data:         params,
		LoggedInUser: middleware.GetUserRef(r),
	}
	err := s.templates.Execute(name, w, base, p)
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
	Data         any
}

func (s *Site) Index(w http.ResponseWriter, r *http.Request) error {
	return s.RenderTemplate(w, r, http.StatusOK, "public/home", "layouts/public-base", nil)
}

func (s *Site) NotFound(w http.ResponseWriter, r *http.Request) error {
	return s.FullpageError(w, r, fullPageErrorParams{
		Code:    http.StatusNotFound,
		Message: "Resource not found.",
	})
}

type fullPageErrorParams struct {
	Code    int
	Message string
}

func (s *Site) FullpageError(w http.ResponseWriter, r *http.Request, params fullPageErrorParams) error {
	return s.RenderTemplate(w, r, params.Code, "public/error", "layouts/public-base", params)
}

type alertErrorParams struct {
	Variant string
	Message string
}

func (s *Site) AlertError(w http.ResponseWriter, r *http.Request, params alertErrorParams) error {
	return s.RenderPlain(w, r, http.StatusOK, "public/alert-error", params)
}
