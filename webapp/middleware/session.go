package middleware

import (
	"context"
	"net/http"

	"github.com/kirsle/go-website-template/webapp/session"
)

// Session middleware.
func Session(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check for the session_id cookie.
		sess := session.LoadOrNew(r)
		ctx := context.WithValue(r.Context(), session.ContextKey, sess)

		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}
