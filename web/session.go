package web

import (
	"net/http"
	"time"

	"uwece.ca/app/utils"
)

const (
	cookieName  = "uwececa_session_token_v1"
	expiryHours = 48
)

type Session struct {
	Token  utils.Token
	Expiry time.Time
}

// Create a new session token (a hex encoded string of len tokenBytes).
func NewSession() Session {
	return Session{
		Token:  utils.NewToken(),
		Expiry: time.Now().Add(expiryHours * time.Hour),
	}
}

// Send session token to user's browser.
func AddSession(w http.ResponseWriter, s Session) {
	cookie := &http.Cookie{
		Name:   cookieName,
		Value:  string(s.Token),
		Quoted: false,

		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		Secure:   true,
		Expires:  s.Expiry,
	}

	http.SetCookie(w, cookie)
}

func GetSession(r *http.Request) (Session, bool) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return Session{}, false
	}

	s := Session{
		Token:  utils.Token(cookie.Value),
		Expiry: cookie.Expires,
	}

	return s, true
}

func DeleteSession(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   cookieName,
		Value:  "deleted",
		Quoted: false,

		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		Expires:  time.Now(),
	}

	http.SetCookie(w, cookie)
}
