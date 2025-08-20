package site

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"uwece.ca/app/auth"
	"uwece.ca/app/config"
	"uwece.ca/app/db"
	"uwece.ca/app/models"
	"uwece.ca/app/request"
)

// Verification emails will expire in 48 hrs.
const emailExpiryHours = 48

type loginSignupParams struct {
	Variant string
}

func (s *Site) LoginPage(w http.ResponseWriter, r *http.Request) {
	s.RenderTemplate(w, http.StatusOK, "public/login-signup", "layouts/public-base", loginSignupParams{
		Variant: "Login",
	})
}

func (s *Site) SignupPage(w http.ResponseWriter, r *http.Request) {
	s.RenderTemplate(w, http.StatusOK, "public/login-signup", "layouts/public-base", loginSignupParams{
		Variant: "Signup",
	})
}

type emailVerificationPageParams struct {
	Email string
	Name  string
}

func (s *emailVerificationPageParams) From(f request.Form) error {
	s.Email = f.Value("email")
	s.Name = f.Value("name")

	return nil
}

func (s *Site) EmailVerificationPage(w http.ResponseWriter, r *http.Request) {
	var params emailVerificationPageParams
	if err := request.FromMultipart(r, &params); err != nil {
		s.UnhandledError(w, err)
		return
	}
	s.RenderTemplate(w, http.StatusOK, "public/email-verification", "layouts/public-base", params)
}

type signupHandlerParams struct {
	Email    string
	Password string
}

func (s *signupHandlerParams) From(f request.Form) error {
	s.Email = f.Value("email")
	s.Password = f.Value("password")

	if _, err := mail.ParseAddress(s.Email); err != nil {
		return errors.New("Please provide a valid email.")
	}

	if len(s.Password) < 12 {
		return errors.New("Please provide a password of length 12 or greater.")
	}

	if s.Password != f.Value("password-confirm") {
		return errors.New("Password and password confirm must match.")
	}

	return nil
}

func (s *Site) SignupHandler(w http.ResponseWriter, r *http.Request) {
	var params signupHandlerParams
	err := request.FromMultipart(r, &params)
	if err != nil {
		s.AlertError(w, alertErrorParams{
			Variant: "danger",
			Message: err.Error(),
		})
		return
	}

	if s.config.Core.RequiredEmailSuffix != "" {
		if !strings.HasSuffix(params.Email, s.config.Core.RequiredEmailSuffix) {
			s.AlertError(w, alertErrorParams{
				Variant: "danger",
				Message: fmt.Sprintf("Please use your email with the suffix: %s (will be verified).", s.config.Core.RequiredEmailSuffix),
			})
			return
		}
	}

	password := auth.HashPassword(params.Password)

	usr, err := models.InsertUser(r.Context(), s.db, models.NewUser{
		Email:    params.Email,
		Password: password,
	})
	if err != nil {
		s.AlertError(w, alertErrorParams{
			Variant: "danger",
			Message: "Something went wrong, please try again.",
		})
		return
	}

	em, err := models.InsertEmail(r.Context(), s.db, models.NewEmail{
		Token:   auth.NewToken(),
		UserId:  usr.Id,
		Expires: time.Now().Add(emailExpiryHours * time.Hour),
	})
	if err != nil {
		s.AlertError(w, alertErrorParams{
			Variant: "danger",
			Message: "Something went wrong, please try again.",
		})
		return
	}

	if err := s.SendVerificationEmail(usr.Email, "User", em.Token); err != nil {
		slog.Error("failed sending verification email", "error", err)
		s.AlertError(w, alertErrorParams{
			Variant: "danger",
			Message: "Something went wrong, please try again.",
		})
		return
	}

	HxLocation(w, fmt.Sprintf("/signup/email-verification?email=%s&name=name", usr.Email))
}

type loginHandlerParams struct {
	Email    string
	Password string
}

func (l *loginHandlerParams) From(f request.Form) error {
	l.Email = f.Value("email")
	l.Password = f.Value("password")

	if l.Email == "" {
		return errors.New("Please provide a valid email.")
	}

	if l.Password == "" {
		return errors.New("Please provide a valid password.")
	}

	return nil
}

func (s *Site) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var params loginHandlerParams
	if err := request.FromMultipart(r, &params); err != nil {
		s.AlertError(w, alertErrorParams{
			Variant: "danger",
			Message: err.Error(),
		})
	}

	usr, err := models.GetUser(r.Context(), s.db, db.FilterEq("email", params.Email))
	if err != nil {
		if errors.Is(err, db.ErrNoRows) {
			s.AlertError(w, alertErrorParams{
				Variant: "danger",
				Message: "No verified users found with that email and password.",
			})
			return
		}
		s.UnhandledError(w, err)
		return
	}

	if usr.VerifiedAt == nil {
		s.AlertError(w, alertErrorParams{
			Variant: "danger",
			Message: "No verified users found with that email and password.",
		})
		return
	}

	ok := auth.MustVerifyPassword(params.Password, usr.Password)
	if !ok {
		s.AlertError(w, alertErrorParams{
			Variant: "danger",
			Message: "No verified users found with that email and password.",
		})
		return
	}

	session := auth.NewSession()
	_, err = models.InsertSession(r.Context(), s.db, models.NewSession{
		Token:   session.Token,
		Expires: session.Expiry,
		UserId:  usr.Id,
	})
	if err != nil {
		s.UnhandledError(w, err)
		return
	}

	auth.AddSession(w, session)

	HxRedirect(w, "/")
}

func (s *Site) EmailVerificationHandler(w http.ResponseWriter, r *http.Request) {
	token := auth.Token(r.PathValue("token"))

	e, err := models.GetEmail(r.Context(), s.db, db.FilterEq("token", token))
	if err != nil {
		if errors.Is(err, db.ErrNoRows) {
			s.FullpageError(w, fullPageErrorParams{
				Code:    http.StatusNotFound,
				Message: "Token not found.",
			})
			return
		}
		s.UnhandledError(w, err)
		return
	}

	if time.Now().After(e.Expires) {
		if errors.Is(err, db.ErrNoRows) {
			s.FullpageError(w, fullPageErrorParams{
				Code:    http.StatusNotFound,
				Message: "Token not found.",
			})
			return
		}

		s.UnhandledError(w, err)
		return
	}

	if err := models.VerifyUser(r.Context(), s.db, db.FilterEq("id", e.UserId)); err != nil {
		s.UnhandledError(w, err)
		return
	}

	slog.Info("verified user", "id", e.UserId)

	Redirect(w, "/login")
}
