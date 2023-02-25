package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/kirsle/go-website-template/webapp/config"
	"github.com/kirsle/go-website-template/webapp/log"
	"github.com/kirsle/go-website-template/webapp/session"
	"github.com/kirsle/go-website-template/webapp/templates"
)

// CSRF middleware. Other places to look: pkg/session/session.go, pkg/templates/template_funcs.go
func CSRF(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get or create the cookie CSRF value.
		token := MakeCSRFCookie(r, w)
		ctx := context.WithValue(r.Context(), session.CSRFKey, token)

		// If it's a JSON post, allow it thru.
		if r.Header.Get("Content-Type") == "application/json" {
			handler.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// If we are running a POST request, validate the CSRF form value.
		if r.Method != http.MethodGet {
			r.ParseMultipartForm(config.MultipartMaxMemory)
			check := r.FormValue(config.CSRFInputName)
			if check != token {
				log.Error("CSRF mismatch! %s <> %s", check, token)
				templates.MakeErrorPage(
					"CSRF Error",
					"An error occurred while processing your request. Please go back and try again.",
					http.StatusForbidden,
				)(w, r.WithContext(ctx))
				return
			}
		}

		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}

// MakeCSRFCookie gets or creates the CSRF cookie and returns its value.
func MakeCSRFCookie(r *http.Request, w http.ResponseWriter) string {
	// Has a token already?
	cookie, err := r.Cookie(config.CSRFCookieName)
	if err == nil {
		// log.Debug("MakeCSRFCookie: user has token %s", cookie.Value)
		return cookie.Value
	}

	// Generate a new CSRF token.
	token := uuid.New().String()
	cookie = &http.Cookie{
		Name:     config.CSRFCookieName,
		Value:    token,
		HttpOnly: true,
		Path:     "/",
	}
	// log.Debug("MakeCSRFCookie: giving cookie value %s to user", token)
	http.SetCookie(w, cookie)

	return token
}
