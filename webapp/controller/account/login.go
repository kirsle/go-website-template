package account

import (
	"net/http"
	"strings"

	"github.com/aichaos/silhouette/webapp/config"
	"github.com/aichaos/silhouette/webapp/log"
	"github.com/aichaos/silhouette/webapp/models"
	"github.com/aichaos/silhouette/webapp/ratelimit"
	"github.com/aichaos/silhouette/webapp/session"
	"github.com/aichaos/silhouette/webapp/templates"
)

// Login controller.
func Login() http.HandlerFunc {
	tmpl := templates.Must("account/login.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var next = r.FormValue("next")

		// Posting?
		if r.Method == http.MethodPost {
			var (
				// Collect form fields.
				username = strings.ToLower(r.PostFormValue("username"))
				password = r.PostFormValue("password")
			)

			// Look up their account.
			user, err := models.FindUser(username)
			if err != nil {
				session.FlashError(w, r, "Incorrect username or password.")
				templates.Redirect(w, r.URL.Path)
				return
			}

			// Rate limit failed login attempts.
			limiter := &ratelimit.Limiter{
				Namespace:  "login",
				ID:         user.ID,
				Limit:      config.LoginRateLimit,
				Window:     config.LoginRateLimitWindow,
				CooldownAt: config.LoginRateLimitCooldownAt,
				Cooldown:   config.LoginRateLimitCooldown,
			}

			// Verify password.
			if err := user.CheckPassword(password); err != nil {
				if err := limiter.Ping(); err != nil {
					session.FlashError(w, r, err.Error())
					templates.Redirect(w, r.URL.Path)
					return
				}

				session.FlashError(w, r, "Incorrect username or password.")
				templates.Redirect(w, r.URL.Path)
				return
			}

			// Is their account banned or disabled?
			if user.Status != models.UserStatusActive {
				session.FlashError(w, r, "Your account has been %s. If you believe this was done in error, please contact support.", user.Status)
				templates.Redirect(w, r.URL.Path)
				return
			}

			// OK. Log in the user's session.
			session.LoginUser(w, r, user)

			// Clear their rate limiter.
			if err := limiter.Clear(); err != nil {
				log.Error("Failed to clear login rate limiter: %s", err)
			}

			// Redirect to their dashboard.
			session.Flash(w, r, "Login successful.")
			if strings.HasPrefix(next, "/") {
				templates.Redirect(w, next)
			} else {
				templates.Redirect(w, "/me")
			}
			return
		}

		var vars = map[string]interface{}{
			"Next": next,
		}
		if err := tmpl.Execute(w, r, vars); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

// Logout controller.
func Logout() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session.Flash(w, r, "You have been successfully logged out.")
		session.LogoutUser(w, r)
		templates.Redirect(w, "/")
	})
}
