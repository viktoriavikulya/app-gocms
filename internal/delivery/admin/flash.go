package admin

import (
	"net/http"
	"net/url"
	"strings"
	"time"
)

const flashCookieName = "appcms_flash"

func setFlashError(w http.ResponseWriter, message string) {
	if strings.TrimSpace(message) == "" {
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     flashCookieName,
		Value:    url.QueryEscape(message),
		Path:     "/",
		MaxAge:   60,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func ReadAndClearFlash(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie(flashCookieName)
	if err != nil || cookie.Value == "" {
		return ""
	}
	clearFlashCookie(w)
	value, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return cookie.Value
	}
	return value
}

func clearFlashCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     flashCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
