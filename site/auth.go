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
	"uwece.ca/app/db"
	"uwece.ca/app/models"
	"uwece.ca/app/web"
)

// Verification emails will expire in 48 hrs.
const emailExpiryHours = 48

type loginSignupParams struct {
	PageParams
	Variant string
}

func (s *Site) LoginPage(w http.ResponseWriter, r *http.Request) error {
	return s.RenderTemplate(w, http.StatusOK, "public/login-signup", "layouts/public-base", loginSignupParams{
		PageParams: PageParamsFromReq(r),
		Variant:    "Login",
	})
}

func (s *Site) SignupPage(w http.ResponseWriter, r *http.Request) error {
	return s.RenderTemplate(w, http.StatusOK, "public/login-signup", "layouts/public-base", loginSignupParams{
		PageParams: PageParamsFromReq(r),
		Variant:    "Signup",
	})
}

type emailVerificationPageParams struct {
	PageParams
	Email string
	Name  string
}

func (s *emailVerificationPageParams) From(f web.Form) error {
	s.Email = f.Value("email")
	s.Name = f.Value("name")

	return nil
}

func (s *Site) EmailVerificationPage(w http.ResponseWriter, r *http.Request) error {
	var params emailVerificationPageParams
	if err := web.FromMultipart(r, &params); err != nil {
		return err
	}
	params.PageParams = PageParamsFromReq(r)
	return s.RenderTemplate(w, http.StatusOK, "public/email-verification", "layouts/public-base", params)
}

type signupHandlerParams struct {
	Email    string
	Password string
	Name     string
}

func (s *signupHandlerParams) From(f web.Form) error {
	s.Email = f.Value("email")
	s.Password = f.Value("password")
	s.Name = f.Value("name")

	if _, err := mail.ParseAddress(s.Email); err != nil {
		return errors.New("Please provide a valid email.")
	}

	if s.Name == "" {
		return errors.New("Please provide us with a name :).")
	}

	if len(s.Password) < 12 {
		return errors.New("Please provide a password of length 12 or greater.")
	}

	if s.Password != f.Value("password-confirm") {
		return errors.New("Password and password confirm must match.")
	}

	return nil
}

func (s *Site) SignupHandler(w http.ResponseWriter, r *http.Request) error {
	var params signupHandlerParams
	err := web.FromMultipart(r, &params)
	if err != nil {
		return s.AlertError(w, alertErrorParams{
			Variant: "danger",
			Message: err.Error(),
		})
	}

	if s.config.Core.RequiredEmailSuffix != "" {
		if !strings.HasSuffix(params.Email, s.config.Core.RequiredEmailSuffix) {
			return s.AlertError(w, alertErrorParams{
				Variant: "danger",
				Message: fmt.Sprintf("Please use your email with the suffix: %s (will be verified).", s.config.Core.RequiredEmailSuffix),
			})
		}
	}

	password := auth.HashPassword(params.Password)

	usr, err := models.InsertUser(r.Context(), s.db, models.NewUser{
		Email:    params.Email,
		Password: password,
		Name:     params.Name,
	})
	if err != nil {
		return s.AlertError(w, alertErrorParams{
			Variant: "danger",
			Message: "Something went wrong, please try again.",
		})
	}

	em, err := models.InsertEmail(r.Context(), s.db, models.NewEmail{
		Token:   auth.NewToken(),
		UserId:  usr.Id,
		Expires: time.Now().Add(emailExpiryHours * time.Hour),
	})
	if err != nil {
		return s.AlertError(w, alertErrorParams{
			Variant: "danger",
			Message: "Something went wrong, please try again.",
		})
	}

	if err := s.SendVerificationEmail(usr.Email, usr.Name, em.Token); err != nil {
		slog.Error("failed sending verification email", "error", err, "email", usr.Email)
		return s.AlertError(w, alertErrorParams{
			Variant: "danger",
			Message: "Something went wrong, please try again.",
		})
	}

	return web.HxLocation(w, fmt.Sprintf("/signup/email-verification?email=%s&name=%s", usr.Email, usr.Name))
}

type loginHandlerParams struct {
	Email    string
	Password string
}

func (l *loginHandlerParams) From(f web.Form) error {
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

func (s *Site) LoginHandler(w http.ResponseWriter, r *http.Request) error {
	var params loginHandlerParams
	if err := web.FromMultipart(r, &params); err != nil {
		return s.AlertError(w, alertErrorParams{
			Variant: "danger",
			Message: err.Error(),
		})
	}

	usr, err := models.GetUser(r.Context(), s.db, db.FilterEq("email", params.Email))
	if err != nil {
		if errors.Is(err, db.ErrNoRows) {
			return s.AlertError(w, alertErrorParams{
				Variant: "danger",
				Message: "No verified users found with that email and password.",
			})
		}
		return err
	}

	if usr.VerifiedAt == nil {
		return s.AlertError(w, alertErrorParams{
			Variant: "danger",
			Message: "No verified users found with that email and password.",
		})
	}

	ok := auth.MustVerifyPassword(params.Password, usr.Password)
	if !ok {
		return s.AlertError(w, alertErrorParams{
			Variant: "danger",
			Message: "No verified users found with that email and password.",
		})
	}

	session := auth.NewSession()
	_, err = models.InsertSession(r.Context(), s.db, models.NewSession{
		Token:   session.Token,
		Expires: session.Expiry,
		UserId:  usr.Id,
	})
	if err != nil {
		return err
	}

	auth.AddSession(w, session)
	return web.HxRedirect(w, "/")
}

func (s *Site) EmailVerificationHandler(w http.ResponseWriter, r *http.Request) error {
	token := auth.Token(r.PathValue("token"))

	e, err := models.GetEmail(r.Context(), s.db, db.FilterEq("token", token))
	if err != nil {
		if errors.Is(err, db.ErrNoRows) {
			return s.FullpageError(w, fullPageErrorParams{
				Code:    http.StatusNotFound,
				Message: "Token not found.",
			})
		}
		return err
	}

	if time.Now().After(e.Expires) {
		if errors.Is(err, db.ErrNoRows) {
			return s.FullpageError(w, fullPageErrorParams{
				Code:    http.StatusNotFound,
				Message: "Token not found.",
			})
		}

		return err
	}

	if err := models.VerifyUser(r.Context(), s.db, db.FilterEq("id", e.UserId)); err != nil {
		return err
	}

	slog.Info("verified user", "id", e.UserId)

	return web.Redirect(w, "/login")
}
