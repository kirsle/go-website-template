package middleware

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/aichaos/silhouette/webapp/config"
	"github.com/aichaos/silhouette/webapp/log"
	"github.com/aichaos/silhouette/webapp/models"
	"github.com/aichaos/silhouette/webapp/session"
	"github.com/aichaos/silhouette/webapp/templates"
)

// LoginRequired middleware.
func LoginRequired(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// User must be logged in.
		user, err := session.CurrentUser(r)
		if err != nil {
			log.Error("LoginRequired: %s", err)
			session.FlashError(w, r, "You must be signed in to view this page.")
			templates.Redirect(w, "/login?next="+url.QueryEscape(r.URL.String()))
			return
		}

		// Are they banned or disabled?
		if user.Status == models.UserStatusDisabled {
			session.LogoutUser(w, r)
			session.FlashError(w, r, "Your account has been disabled and you are now logged out.")
			templates.Redirect(w, "/")
			return
		} else if user.Status == models.UserStatusBanned {
			session.LogoutUser(w, r)
			session.FlashError(w, r, "Your account has been banned and you are now logged out.")
			templates.Redirect(w, "/")
			return
		}

		// Ping LastLoginAt for long lived sessions, but not if impersonated.
		if time.Since(user.LastLoginAt) > config.LastLoginAtCooldown && !session.Impersonated(r) {
			user.LastLoginAt = time.Now()
			if err := user.Save(); err != nil {
				log.Error("LoginRequired: couldn't refresh LastLoginAt for user %s: %s", user.Username, err)
			}
		}

		// Stick the CurrentUser in the request context so future calls to session.CurrentUser can read it.
		ctx := context.WithValue(r.Context(), session.CurrentUserKey, user)
		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminRequired middleware.
func AdminRequired(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// User must be logged in.
		currentUser, err := session.CurrentUser(r)
		if err != nil {
			log.Error("AdminRequired: %s", err)
			session.FlashError(w, r, "You must be signed in to view this page.")
			templates.Redirect(w, "/login?next="+url.QueryEscape(r.URL.String()))
			return
		}

		// Stick the CurrentUser in the request context so future calls to session.CurrentUser can read it.
		ctx := context.WithValue(r.Context(), session.CurrentUserKey, currentUser)

		// Admin required.
		if !currentUser.IsAdmin {
			log.Error("AdminRequired: %s", err)
			errhandler := templates.MakeErrorPage("Admin Required", "You do not have permission for this page.", http.StatusForbidden)
			errhandler.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}
