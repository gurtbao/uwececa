package site

import "net/http"

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
