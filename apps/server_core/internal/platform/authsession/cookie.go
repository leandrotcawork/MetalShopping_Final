package authsession

import (
	"net/http"
	"time"
)

func SetCSRFCookie(w http.ResponseWriter, config Config, csrfToken string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     config.CSRFCookieName,
		Value:    csrfToken,
		Path:     config.CookiePath,
		Domain:   config.CookieDomain,
		HttpOnly: false,
		Secure:   config.CookieSecure,
		SameSite: config.CookieSameSite,
		Expires:  expiresAt.UTC(),
	})
}

func ClearCSRFCookie(w http.ResponseWriter, config Config) {
	http.SetCookie(w, &http.Cookie{
		Name:     config.CSRFCookieName,
		Value:    "",
		Path:     config.CookiePath,
		Domain:   config.CookieDomain,
		HttpOnly: false,
		Secure:   config.CookieSecure,
		SameSite: config.CookieSameSite,
		MaxAge:   -1,
	})
}

func SetSessionCookie(w http.ResponseWriter, config Config, sessionID string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     config.SessionCookieName,
		Value:    sessionID,
		Path:     config.CookiePath,
		Domain:   config.CookieDomain,
		HttpOnly: true,
		Secure:   config.CookieSecure,
		SameSite: config.CookieSameSite,
		Expires:  expiresAt.UTC(),
	})
}

func ClearSessionCookie(w http.ResponseWriter, config Config) {
	http.SetCookie(w, &http.Cookie{
		Name:     config.SessionCookieName,
		Value:    "",
		Path:     config.CookiePath,
		Domain:   config.CookieDomain,
		HttpOnly: true,
		Secure:   config.CookieSecure,
		SameSite: config.CookieSameSite,
		MaxAge:   -1,
	})
}

func SetLoginStateCookie(w http.ResponseWriter, config Config, state string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     config.StateCookieName,
		Value:    state,
		Path:     config.CookiePath,
		Domain:   config.CookieDomain,
		HttpOnly: true,
		Secure:   config.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt.UTC(),
	})
}

func ClearLoginStateCookie(w http.ResponseWriter, config Config) {
	http.SetCookie(w, &http.Cookie{
		Name:     config.StateCookieName,
		Value:    "",
		Path:     config.CookiePath,
		Domain:   config.CookieDomain,
		HttpOnly: true,
		Secure:   config.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
