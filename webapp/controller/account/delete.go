package account

import (
	"net/http"
	"strings"

	"github.com/kirsle/go-website-template/webapp/models/deletion"
	"github.com/kirsle/go-website-template/webapp/session"
	"github.com/kirsle/go-website-template/webapp/templates"
)

// Delete account page (self service).
func Delete() http.HandlerFunc {
	tmpl := templates.Must("account/delete.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentUser, err := session.CurrentUser(r)
		if err != nil {
			session.FlashError(w, r, "Couldn't get your current user: %s", err)
			templates.Redirect(w, "/")
			return
		}

		// Confirm deletion.
		if r.Method == http.MethodPost {
			var password = strings.TrimSpace(r.PostFormValue("password"))
			if err := currentUser.CheckPassword(password); err != nil {
				session.FlashError(w, r, "You must enter your correct account password to delete your account.")
				templates.Redirect(w, r.URL.Path)
				return
			}

			// Delete their account!
			if err := deletion.DeleteUser(currentUser); err != nil {
				session.FlashError(w, r, "Error while deleting your account: %s", err)
				templates.Redirect(w, r.URL.Path)
				return
			}

			// Sign them out.
			session.LogoutUser(w, r)
			session.Flash(w, r, "Your account has been deleted.")
			templates.Redirect(w, "/")
			return
		}

		var vars = map[string]interface{}{}
		if err := tmpl.Execute(w, r, vars); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
