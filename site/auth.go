package site

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"uwece.ca/app/db"
	"uwece.ca/app/models"
	"uwece.ca/app/utils"
	"uwece.ca/app/web"
)

// Verification emails will expire in 48 hrs.
const emailExpiryHours = 48

type loginSignupParams struct {
	Variant string
}

func (s *Site) LoginPage(w http.ResponseWriter, r *http.Request) error {
	return s.RenderTemplate(w, r, http.StatusOK, "public/login-signup", "layouts/public-base", loginSignupParams{
		Variant: "Login",
	})
}

func (s *Site) SignupPage(w http.ResponseWriter, r *http.Request) error {
	return s.RenderTemplate(w, r, http.StatusOK, "public/login-signup", "layouts/public-base", loginSignupParams{
		Variant: "Signup",
	})
}

type signupHandlerParams struct {
	NetID    string
	Password string
	Name     string
}

func (s *signupHandlerParams) From(f web.Form) error {
	s.NetID = strings.ToLower(f.Value("netid"))
	s.Password = f.Value("password")
	s.Name = f.Value("name")

	if 1 > len(s.NetID) && len(s.NetID) > 35 {
		return errors.New("Please provide a valid netID (0 < len < 35).")
	}

	nID := strings.Map(func(r rune) rune {
		if r < '0' || 'z' < r {
			return -1
		}

		return r
	}, s.NetID)

	if nID != s.NetID {
		return errors.New("Please provide us a valid netID (1-9,a-z).")
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

type signupResponseParams struct {
	Email string
	Name  string
}

func (s *Site) SignupHandler(w http.ResponseWriter, r *http.Request) error {
	var params signupHandlerParams
	err := web.FromMultipart(r, &params)
	if err != nil {
		return s.AlertError(w, r, alertErrorParams{
			Variant: "danger",
			Message: err.Error(),
		})
	}

	password := utils.HashPassword(params.Password)

	usr, err := models.InsertUser(r.Context(), s.db, models.NewUser{
		Email:    fmt.Sprintf("%s@%s", params.NetID, s.config.Core.EmailDomain),
		Password: password,
		Name:     params.Name,
	})
	if err != nil {
		if !errors.Is(err, db.ErrUnique) {
			slog.Error("failed inserting user into database during signup", "error", err, "net_id", params.NetID)
		}
		return s.AlertError(w, r, alertErrorParams{
			Variant: "danger",
			Message: "Something went wrong, please try again.",
		})
	}

	em, err := models.InsertEmail(r.Context(), s.db, models.NewEmail{
		Token:   utils.NewToken(),
		UserId:  usr.Id,
		Expires: time.Now().Add(emailExpiryHours * time.Hour),
	})
	if err != nil {
		slog.Error("failed inserting email into database during signup", "error", err, "net_id", params.NetID)
		return s.AlertError(w, r, alertErrorParams{
			Variant: "danger",
			Message: "Something went wrong, please try again.",
		})
	}

	if err := s.SendVerificationEmail(usr.Email, usr.Name, em.Token); err != nil {
		slog.Error("failed sending verification email", "error", err, "email", usr.Email)
		return s.AlertError(w, r, alertErrorParams{
			Variant: "danger",
			Message: "Something went wrong, please try again.",
		})
	}

	return s.RenderPlain(w, r, http.StatusOK, "public/signup-response", signupResponseParams{
		Email: usr.Email,
		Name:  usr.Name,
	})
}

type loginHandlerParams struct {
	NetID    string
	Password string
}

func (l *loginHandlerParams) From(f web.Form) error {
	l.NetID = f.Value("netid")
	l.Password = f.Value("password")

	if l.NetID == "" {
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
		return s.AlertError(w, r, alertErrorParams{
			Variant: "danger",
			Message: err.Error(),
		})
	}

	usr, err := models.GetUser(r.Context(), s.db, db.FilterEq("email", fmt.Sprintf("%s@%s", params.NetID, s.config.Core.EmailDomain)))
	if err != nil {
		if errors.Is(err, db.ErrNoRows) {
			return s.AlertError(w, r, alertErrorParams{
				Variant: "danger",
				Message: "No verified users found with that email and password.",
			})
		}
		return err
	}

	if usr.VerifiedAt == nil {
		return s.AlertError(w, r, alertErrorParams{
			Variant: "danger",
			Message: "No verified users found with that email and password.",
		})
	}

	ok := utils.MustVerifyPassword(params.Password, usr.Password)
	if !ok {
		return s.AlertError(w, r, alertErrorParams{
			Variant: "danger",
			Message: "No verified users found with that email and password.",
		})
	}

	session := web.NewSession()
	_, err = models.InsertSession(r.Context(), s.db, models.NewSession{
		Token:   session.Token,
		Expires: session.Expiry,
		UserId:  usr.Id,
	})
	if err != nil {
		return err
	}

	web.AddSession(w, session)
	return web.HxRedirect(w, "/site")
}

func (s *Site) EmailVerificationHandler(w http.ResponseWriter, r *http.Request) error {
	token := utils.Token(r.PathValue("token"))

	e, err := models.GetEmail(r.Context(), s.db, db.FilterEq("token", token))
	if err != nil {
		if errors.Is(err, db.ErrNoRows) {
			return s.FullpageError(w, r, fullPageErrorParams{
				Code:    http.StatusNotFound,
				Message: "Token not found.",
			})
		}
		return err
	}

	if time.Now().After(e.Expires) {
		if errors.Is(err, db.ErrNoRows) {
			return s.FullpageError(w, r, fullPageErrorParams{
				Code:    http.StatusNotFound,
				Message: "Token not found.",
			})
		}

		return err
	}

	if err := models.UpdateUser(r.Context(), s.db, db.Updates(db.Update("updated_at", time.Now()), db.Update("verified_at", time.Now())), db.FilterEq("id", e.UserId)); err != nil {
		return err
	}

	slog.Info("verified user", "id", e.UserId)

	return web.Redirect(w, "/login")
}

func (s *Site) Logout(w http.ResponseWriter, r *http.Request) error {
	web.DeleteSession(w)
	return web.Redirect(w, "/")
}
