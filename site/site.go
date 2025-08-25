package site

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimd "github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/schema"
	"uwece.ca/app/config"
	"uwece.ca/app/db"
	"uwece.ca/app/mailer"
	"uwece.ca/app/services"
	"uwece.ca/app/templates"
	"uwece.ca/app/web"
)

//go:embed templates/* static
var embedFS embed.FS

type Site struct {
	blogs     *services.BlogService
	users     *services.UserService
	templates *templates.Templates
	config    *config.Config
	decoder   *schema.Decoder
}

func New(cfg *config.Config, db *db.DB, mailer mailer.Mailer) *Site {
	var tmpl *templates.Templates
	if cfg.Core.Development {
		tmpl = templates.NewDevTemplates(embedFS, "./site/templates")
	} else {
		tmpl = templates.NewTemplates(embedFS)
	}

	return &Site{
		users:     services.NewUserService(db, mailer, cfg),
		blogs:     services.NewBlogService(db),
		config:    cfg,
		templates: tmpl,
		decoder:   schema.NewDecoder(),
	}
}

func (s *Site) Routes() http.Handler {
	w := web.NewHandlerWrapper(s)
	r := chi.NewMux()
	r.Use(web.MidLogRecover)
	r.Use(chimd.Compress(5))

	r.Handle("/static/*", s.Static())

	r.Group(func(r chi.Router) {
		r.Use(s.LoadUser)
		r.Handle("/", w.Wrap(s.Index))

		r.Group(func(r chi.Router) {
			r.Use(RequireLogin(false))
			r.Get("/login", w.Wrap(s.LoginPage))
			r.Post("/login", w.Wrap(s.LoginHandler))
			r.Get("/signup", w.Wrap(s.SignupPage))
			r.Post("/signup", w.Wrap(s.SignupHandler))
			r.Get("/signup/verify/{token}", w.Wrap(s.VerificationHandler))
		})

		r.Group(func(r chi.Router) {
			r.Use(RequireLogin(true))
			r.Use(RequireBlog(false, false))
			r.Get("/new-blog", w.Wrap(s.NewBlogPage))
			r.Post("/new-blog", w.Wrap(s.NewBlogHandler))
		})

		r.NotFound(w.Wrap(s.NotFound))
	})

	return r
}

func (s *Site) BaseContext(r *http.Request) templates.Context {
	return templates.Context{
		"current_user": nil,
	}
}

func (s *Site) Render(w http.ResponseWriter, statusCode int, base, name string, params templates.Context) error {
	err := s.templates.Execute(name, w, base, params)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(statusCode)

	return nil
}

func (s *Site) RenderPlain(w http.ResponseWriter, statusCode int, name string, params templates.Context) error {
	return s.Render(w, statusCode, "", name, params)
}

func (s *Site) Static() http.Handler {
	if s.config.Core.Development {
		return http.StripPrefix("/static", http.FileServer(http.Dir("./site/static")))
	}

	sub, err := fs.Sub(embedFS, "static")
	if err != nil {
		slog.Error("no static dir in mainsite", "error", err)
		os.Exit(1)
	}

	return http.StripPrefix("/static", http.FileServer(http.FS(sub)))
}
