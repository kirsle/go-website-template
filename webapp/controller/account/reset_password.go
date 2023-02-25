package account

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/kirsle/go-website-template/webapp/config"
	"github.com/kirsle/go-website-template/webapp/log"
	"github.com/kirsle/go-website-template/webapp/mail"
	"github.com/kirsle/go-website-template/webapp/models"
	"github.com/kirsle/go-website-template/webapp/redis"
	"github.com/kirsle/go-website-template/webapp/session"
	"github.com/kirsle/go-website-template/webapp/templates"
)

// ResetToken goes in Redis.
type ResetToken struct {
	UserID uint64
	Token  string
}

// Delete the token.
func (t ResetToken) Delete() error {
	return redis.Delete(fmt.Sprintf(config.ResetPasswordRedisKey, t.Token))
}

// ForgotPassword controller.
func ForgotPassword() http.HandlerFunc {
	tmpl := templates.Must("account/forgot_password.html")

	vagueSuccessMessage := "If that username or email existed, we have sent " +
		"an email to the address on file with a link to reset your password. " +
		"Please check your email inbox for the link."

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			tokenStr = r.FormValue("token") // GET or POST
			token    ResetToken
			user     *models.User
		)

		// If given a token, validate it first.
		if tokenStr != "" {
			if err := redis.Get(fmt.Sprintf(config.ResetPasswordRedisKey, tokenStr), &token); err != nil || token.Token != tokenStr {
				session.FlashError(w, r, "Invalid password reset token. Please try again from the beginning.")
				templates.Redirect(w, r.URL.Path)
				return
			}

			// Get the target user by ID.
			if target, err := models.GetUser(token.UserID); err != nil {
				session.FlashError(w, r, "Couldn't look up the user for this token. Please try again.")
				templates.Redirect(w, r.URL.Path)
				return
			} else {
				user = target
			}
		}

		// POSTing:
		// - To begin the reset flow (username only)
		// - To finalize (username + passwords + validated token)
		if r.Method == http.MethodPost {
			var (
				username  = strings.TrimSpace(strings.ToLower(r.PostFormValue("username")))
				password1 = strings.TrimSpace(r.PostFormValue("password"))
				password2 = strings.TrimSpace(r.PostFormValue("confirm"))
			)

			// Find the user. If we came here by token, we already have it,
			// otherwise the username post param is required.
			if user == nil {
				if username == "" {
					session.FlashError(w, r, "Username or email address is required.")
					templates.Redirect(w, r.URL.Path)
					return
				}

				target, err := models.FindUser(username)
				if err != nil {
					session.Flash(w, r, vagueSuccessMessage)
					templates.Redirect(w, r.URL.Path)
					return
				}

				user = target
			}

			// With a validated token?
			if token.Token != "" {
				if password1 == "" {
					session.FlashError(w, r, "A password is required.")
					templates.Redirect(w, r.URL.Path+"?token="+token.Token)
					return
				} else if password1 != password2 {
					session.FlashError(w, r, "Your passwords do not match.")
					templates.Redirect(w, r.URL.Path+"?token="+token.Token)
					return
				}

				// Set the new password.
				user.HashPassword(password1)
				if err := user.Save(); err != nil {
					session.FlashError(w, r, "Error saving your user: %s", err)
					templates.Redirect(w, r.URL.Path+"?token="+token.Token)
					return
				} else {
					// All done! Burn the reset token.
					if err := token.Delete(); err != nil {
						log.Error("ResetToken.Delete(%s): %s", token.Token, err)
					}

					if err := session.LoginUser(w, r, user); err != nil {
						session.FlashError(w, r, "Your password was reset and you can now log in.")
						templates.Redirect(w, "/login")
						return
					} else {
						session.Flash(w, r, "Your password has been reset and you are now logged in to your account.")
						templates.Redirect(w, "/me")
						return
					}
				}
			}

			// Create a reset token.
			token := ResetToken{
				UserID: user.ID,
				Token:  uuid.New().String(),
			}
			if err := redis.Set(fmt.Sprintf(config.ResetPasswordRedisKey, token.Token), token, config.SignupTokenExpires); err != nil {
				session.FlashError(w, r, "Couldn't create a reset token: %s", err)
				templates.Redirect(w, r.URL.Path)
				return
			}

			// Email them their reset link.
			if err := mail.Send(mail.Message{
				To:       user.Email,
				Subject:  "Reset your forgotten password",
				Template: "email/reset_password.html",
				Data: map[string]interface{}{
					"Username": user.Username,
					"URL":      config.Current.BaseURL + "/forgot-password?token=" + token.Token,
				},
			}); err != nil {
				session.FlashError(w, r, "Error sending an email: %s", err)
			}

			// Success message and redirect away.
			session.Flash(w, r, vagueSuccessMessage)
			templates.Redirect(w, r.URL.Path)
			return
		}

		var vars = map[string]interface{}{
			"Token": token,
			"User":  user,
		}
		if err := tmpl.Execute(w, r, vars); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
