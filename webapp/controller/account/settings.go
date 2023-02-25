package account

import (
	"fmt"
	"net/http"
	nm "net/mail"
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

// ChangeEmailToken for Redis.
type ChangeEmailToken struct {
	Token    string
	UserID   uint64
	NewEmail string
}

// Delete the change email token.
func (t ChangeEmailToken) Delete() error {
	return redis.Delete(fmt.Sprintf(config.ChangeEmailRedisKey, t.Token))
}

// User settings page. (/settings).
func Settings() http.HandlerFunc {
	tmpl := templates.Must("account/settings.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := map[string]interface{}{}

		// Load the current user in case of updates.
		user, err := session.CurrentUser(r)
		if err != nil {
			session.FlashError(w, r, "Couldn't get CurrentUser: %s", err)
			templates.Redirect(w, r.URL.Path)
			return
		}

		// Are we POSTing?
		if r.Method == http.MethodPost {
			intent := r.PostFormValue("intent")
			switch intent {
			case "settings":
				var (
					oldPassword = r.PostFormValue("old_password")
					changeEmail = strings.TrimSpace(strings.ToLower(r.PostFormValue("change_email")))
					password1   = strings.TrimSpace(r.PostFormValue("new_password"))
					password2   = strings.TrimSpace(r.PostFormValue("new_password2"))
				)

				// Their old password is needed to make any changes to their account.
				if err := user.CheckPassword(oldPassword); err != nil {
					session.FlashError(w, r, "Could not make changes to your account settings as the 'current password' you entered was incorrect.")
					templates.Redirect(w, r.URL.Path)
					return
				}

				// Changing their email?
				if changeEmail != user.Email {
					// Validate the email.
					if _, err := nm.ParseAddress(changeEmail); err != nil {
						session.FlashError(w, r, "The email address you entered is not valid: %s", err)
						templates.Redirect(w, r.URL.Path)
						return
					}

					// Email must not already exist.
					if _, err := models.FindUser(changeEmail); err == nil {
						session.FlashError(w, r, "That email address is already in use.")
						templates.Redirect(w, r.URL.Path)
						return
					}

					// Create a tokenized link.
					token := ChangeEmailToken{
						Token:    uuid.New().String(),
						UserID:   user.ID,
						NewEmail: changeEmail,
					}
					if err := redis.Set(fmt.Sprintf(config.ChangeEmailRedisKey, token.Token), token, config.SignupTokenExpires); err != nil {
						session.FlashError(w, r, "Failed to create change email token: %s", err)
						templates.Redirect(w, r.URL.Path)
						return
					}

					err := mail.Send(mail.Message{
						To:       changeEmail,
						Subject:  "Verify your e-mail address",
						Template: "email/verify_email.html",
						Data: map[string]interface{}{
							"Title":       config.Title,
							"URL":         config.Current.BaseURL + "/settings/confirm-email?token=" + token.Token,
							"ChangeEmail": true,
						},
					})
					if err != nil {
						session.FlashError(w, r, "Error sending a confirmation email to %s: %s", changeEmail, err)
					} else {
						session.Flash(w, r, "Please verify your new email address. A link has been sent to %s to confirm.", changeEmail)
					}
				}

				// Changing their password?
				if password1 != "" {
					if password2 != password1 {
						log.Error("pw1=%s  pw2=%s", password1, password2)
						session.FlashError(w, r, "Couldn't change your password: your new passwords do not match.")
					} else {
						// Hash the new password.
						if err := user.HashPassword(password1); err != nil {
							session.FlashError(w, r, "Failed to hash your new password: %s", err)
						} else {
							// Save the user row.
							if err := user.Save(); err != nil {
								session.FlashError(w, r, "Failed to update your password in the database: %s", err)
							} else {
								session.Flash(w, r, "Your password has been updated.")
							}
						}
					}
				}
			default:
				session.FlashError(w, r, "Unknown POST intent value. Please try again.")
			}

			templates.Redirect(w, r.URL.Path)
			return
		}

		if err := tmpl.Execute(w, r, vars); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

// ConfirmEmailChange after a user tries to change their email.
func ConfirmEmailChange() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenStr = r.FormValue("token")

		if tokenStr != "" {
			var token ChangeEmailToken
			if err := redis.Get(fmt.Sprintf(config.ChangeEmailRedisKey, tokenStr), &token); err != nil {
				session.FlashError(w, r, "Invalid token. Please try again to change your email address.")
				templates.Redirect(w, "/")
				return
			}

			// Verify new email still doesn't already exist.
			if _, err := models.FindUser(token.NewEmail); err == nil {
				session.FlashError(w, r, "Couldn't update your email address: it is already in use by another member.")
				templates.Redirect(w, "/")
				return
			}

			// Look up the user.
			user, err := models.GetUser(token.UserID)
			if err != nil {
				session.FlashError(w, r, "Didn't find the user that this email change was for. Please try again.")
				templates.Redirect(w, "/")
				return
			}

			// Burn the token.
			if err := token.Delete(); err != nil {
				log.Error("ChangeEmail: couldn't delete Redis token: %s", err)
			}

			// Make the change.
			user.Email = token.NewEmail
			if err := user.Save(); err != nil {
				session.FlashError(w, r, "Couldn't save the change to your user: %s", err)
			} else {
				session.Flash(w, r, "Your email address has been confirmed and updated.")
				templates.Redirect(w, "/")
			}
		} else {
			session.FlashError(w, r, "Invalid change email token. Please try again.")
		}

		templates.Redirect(w, "/")
	})
}
