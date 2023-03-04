package account

import (
	"fmt"
	"net/http"
	nm "net/mail"
	"strings"

	"github.com/google/uuid"
	"github.com/aichaos/silhouette/webapp/config"
	"github.com/aichaos/silhouette/webapp/log"
	"github.com/aichaos/silhouette/webapp/mail"
	"github.com/aichaos/silhouette/webapp/models"
	"github.com/aichaos/silhouette/webapp/redis"
	"github.com/aichaos/silhouette/webapp/session"
	"github.com/aichaos/silhouette/webapp/templates"
)

// SignupToken goes in Redis when the user first gives us their email address. They
// verify their email before signing up, so cache only in Redis until verified.
type SignupToken struct {
	Email string
	Token string
}

// Delete a SignupToken when it's been used up.
func (st SignupToken) Delete() error {
	return redis.Delete(fmt.Sprintf(config.SignupTokenRedisKey, st.Token))
}

// Initial signup controller.
func Signup() http.HandlerFunc {
	tmpl := templates.Must("account/signup.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Template vars.
		var vars = map[string]interface{}{
			"SignupToken":           "",    // non-empty if user has clicked verification link
			"SkipEmailVerification": false, // true if email verification is disabled
			"Email":                 "",    // pre-filled user email
		}

		// Is email verification disabled?
		if config.SkipEmailVerification {
			vars["SkipEmailVerification"] = true
		}

		// Are we called with an email verification token?
		var tokenStr = r.URL.Query().Get("token")
		if r.Method == http.MethodPost {
			tokenStr = r.PostFormValue("token")
		}

		var token SignupToken
		log.Info("SignupToken: %s", tokenStr)
		if tokenStr != "" {
			// Validate it.
			if err := redis.Get(fmt.Sprintf(config.SignupTokenRedisKey, tokenStr), &token); err != nil || token.Token != tokenStr {
				session.FlashError(w, r, "Invalid email verification token. Please try signing up again.")
				templates.Redirect(w, r.URL.Path)
				return
			}

			vars["SignupToken"] = tokenStr
			vars["Email"] = token.Email
		}
		log.Info("Vars: %+v", vars)

		// Posting?
		if r.Method == http.MethodPost {
			var (
				// Collect form fields.
				email   = strings.TrimSpace(strings.ToLower(r.PostFormValue("email")))
				confirm = r.PostFormValue("confirm") == "true"

				// Only on full signup form
				username  = strings.TrimSpace(strings.ToLower(r.PostFormValue("username")))
				password  = strings.TrimSpace(r.PostFormValue("password"))
				password2 = strings.TrimSpace(r.PostFormValue("password2"))
			)

			// Don't let them sneakily change their verified email address on us.
			if vars["SignupToken"] != "" && email != vars["Email"] {
				session.FlashError(w, r, "This email address is not verified. Please start over from the beginning.")
				templates.Redirect(w, r.URL.Path)
				return
			}

			// Reserved username check.
			for _, cmp := range config.ReservedUsernames {
				if username == cmp {
					session.FlashError(w, r, "That username is reserved, please choose a different username.")
					templates.Redirect(w, r.URL.Path+"?token="+tokenStr)
					return
				}
			}

			// Cache username in case of passwd validation errors.
			vars["Email"] = email
			vars["Username"] = username

			// Is the app not configured to send email?
			if !config.Current.Mail.Enabled {
				session.FlashError(w, r, "This app is not configured to send email so you can not sign up at this time. "+
					"Please contact the website administrator about this issue!")
				templates.Redirect(w, r.URL.Path)
				return
			}

			// Validate the email.
			if _, err := nm.ParseAddress(email); err != nil {
				session.FlashError(w, r, "The email address you entered is not valid: %s", err)
				templates.Redirect(w, r.URL.Path)
				return
			}

			// Didn't confirm?
			if !confirm {
				session.FlashError(w, r, "Confirm that you have read the rules.")
				templates.Redirect(w, r.URL.Path)
				return
			}

			// Already an account?
			if _, err := models.FindUser(email); err == nil {
				session.FlashError(w, r, "There is already an account with that e-mail address.")
				templates.Redirect(w, r.URL.Path)
				return
			}

			// Email verification step!
			if !config.SkipEmailVerification && vars["SignupToken"] == "" {
				// Create a SignupToken verification link to send to their inbox.
				token = SignupToken{
					Email: email,
					Token: uuid.New().String(),
				}
				if err := redis.Set(fmt.Sprintf(config.SignupTokenRedisKey, token.Token), token, config.SignupTokenExpires); err != nil {
					session.FlashError(w, r, "Error creating a link to send you: %s", err)
				}

				err := mail.Send(mail.Message{
					To:       email,
					Subject:  "Verify your e-mail address",
					Template: "email/verify_email.html",
					Data: map[string]interface{}{
						"Title": config.Title,
						"URL":   config.Current.BaseURL + "/signup?token=" + token.Token,
					},
				})
				if err != nil {
					session.FlashError(w, r, "Error sending an email: %s", err)
				}

				session.Flash(w, r, "We have sent an e-mail to %s with a link to continue signing up your account. Please go and check your e-mail.", email)
				templates.Redirect(w, r.URL.Path)
				return
			}

			// Full sign-up step (w/ email verification token), validate more things.
			var hasError bool
			if len(password) < 3 {
				session.FlashError(w, r, "Please enter a password longer than 3 characters.")
				hasError = true
			} else if password != password2 {
				session.FlashError(w, r, "Your passwords do not match.")
				hasError = true
			}

			if !config.UsernameRegexp.MatchString(username) {
				session.FlashError(w, r, "Your username must consist of only numbers, letters, - . and be 3-32 characters.")
				hasError = true
			}

			// Looking good?
			if !hasError {
				user, err := models.CreateUser(username, email, password)
				if err != nil {
					session.FlashError(w, r, err.Error())
				} else {
					session.Flash(w, r, "User account created. Now logged in as %s.", user.Username)

					// Burn the signup token.
					if token.Token != "" {
						if err := token.Delete(); err != nil {
							log.Error("SignupToken.Delete(%s): %s", token.Token, err)
						}
					}

					// Log in the user and send them to their dashboard.
					session.LoginUser(w, r, user)
					templates.Redirect(w, "/me")
				}
			}
		}

		if err := tmpl.Execute(w, r, vars); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
