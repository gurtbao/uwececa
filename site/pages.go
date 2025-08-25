package site

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"uwece.ca/app/services"
	"uwece.ca/app/utils"
	"uwece.ca/app/web"
)

func (s *Site) Index(w http.ResponseWriter, r *http.Request) error {
	ctx := s.BaseContext(r)

	return s.Render(w, http.StatusOK, "layouts/public-base", "public/home", ctx)
}

func (s *Site) LoginPage(w http.ResponseWriter, r *http.Request) error {
	ctx := s.BaseContext(r)
	ctx.Add("variant", "Login")

	return s.Render(w, http.StatusOK, "layouts/public-base", "public/login-signup", ctx)
}

func (s *Site) LoginHandler(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	var req services.UserLoginRequest
	if err := s.decoder.Decode(&req, r.Form); err != nil {
		slog.Warn("form decode error", "error", err)
		return s.DangerAlert(w, "Error decoding form, please try again.")
	}

	res, err := s.users.Login(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrValidationFailed):
			return s.DangerAlert(w, err.Error())
		case errors.Is(err, services.ErrUserDoesNotExist):
			fallthrough
		case errors.Is(err, services.ErrUserWrongPassword):
			return s.DangerAlert(w, "No user account found with that netid and password.")
		case errors.Is(err, services.ErrUserNotVerified):
			return s.WarnAlert(w, "User account not verified, please check your email for a verification link.")
		}

		return err
	}

	web.AddSession(w, res.Session)
	return web.HxRedirect(w, "/site")
}

func (s *Site) SignupPage(w http.ResponseWriter, r *http.Request) error {
	ctx := s.BaseContext(r)
	ctx.Add("variant", "Signup")

	return s.Render(w, http.StatusOK, "layouts/public-base", "public/login-signup", ctx)
}

func (s *Site) SignupHandler(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	var req services.UserSignupRequest
	if err := s.decoder.Decode(&req, r.Form); err != nil {
		slog.Warn("form decode error", "error", err)
		return s.DangerAlert(w, "Error decoding form, please try again.")
	}

	res, err := s.users.Signup(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrValidationFailed):
			return s.DangerAlert(w, err.Error())
		case errors.Is(err, services.ErrUserExists):
			return s.DangerAlert(w, "An account with that netid already exists.")
		}

		return err
	}

	ctx := s.BaseContext(r)
	ctx.Add("email", res.Email)

	return s.RenderPlain(w, http.StatusOK, "public/signup-response", ctx)
}

func (s *Site) VerificationHandler(w http.ResponseWriter, r *http.Request) error {
	token := utils.Token(r.PathValue("token"))

	if err := s.users.Verify(r.Context(), token); err != nil {
		switch {
		case errors.Is(err, services.ErrTokenExpired):
			return s.FullpageError(w, r, http.StatusBadRequest, "Token Expired.")
		case errors.Is(err, services.ErrTokenNotFound):
			return s.FullpageError(w, r, http.StatusNotFound, "Token Does Not Exist.")
		}
		return err
	}

	return web.Redirect(w, "/login")
}

func (s *Site) NewBlogPage(w http.ResponseWriter, r *http.Request) error {
	usr := ExtractUser(r)
	ctx := s.BaseContext(r)
	ctx.Add("name", strings.ToLower(usr.Name))

	return s.Render(w, http.StatusOK, "layouts/public-base", "public/new-blog", ctx)
}

func (s *Site) NewBlogHandler(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	var req services.BlogNewRequest
	if err := s.decoder.Decode(&req, r.Form); err != nil {
		slog.Warn("form decode error", "error", err)
		return s.DangerAlert(w, "Error decoding form, please try again.")
	}

	usr := ExtractUser(r)

	if err := s.blogs.New(r.Context(), req, usr.Id); err != nil {
		switch {
		case errors.Is(err, services.ErrValidationFailed):
			return s.DangerAlert(w, err.Error())
		case errors.Is(err, services.ErrSubdomainNotAvailable):
			return s.DangerAlert(w, "The subdomain you selected was not available, maybe try using your last name.")
		}

		return err
	}

	ctx := s.BaseContext(r)
	return s.RenderPlain(w, http.StatusOK, "public/new-blog-response", ctx)
}
