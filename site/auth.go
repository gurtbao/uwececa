package site

import (
	"errors"
	"net/http"
	"net/mail"

	"uwece.ca/app/auth"
	"uwece.ca/app/db"
	"uwece.ca/app/models"
	"uwece.ca/app/request"
	"uwece.ca/app/form"
	"uwece.ca/app/models"
)

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
<<<<<<< HEAD
	err := request.FromMultipart(r, &params)
=======
	err := form.FromRequest(r, &params)
>>>>>>> 4b4e84b (site: Add login / signup flow.)
	if err != nil {
		s.AlertError(w, alertErrorParams{
			Variant: "danger",
			Message: err.Error(),
		})
		return
	}

	password := auth.HashPassword(params.Password)

	_, err = models.InsertUser(r.Context(), s.db, models.NewUser{
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

	HxRedirect(w, "/login")
}

type loginHandlerParams struct {
	Email    string
	Password string
}

<<<<<<< HEAD
func (l loginHandlerParams) From(f request.Form) error {
=======
func (l loginHandlerParams) From(f form.Form) error {
>>>>>>> 4b4e84b (site: Add login / signup flow.)
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
				Message: "No users found with that email and password.",
			})
			return
		}
		s.UnhandledError(w, err)
		return
	}

	ok := auth.MustVerifyPassword(params.Password, usr.Password)
	if !ok {
		s.AlertError(w, alertErrorParams{
			Variant: "danger",
			Message: "No users found with that email and password.",
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
