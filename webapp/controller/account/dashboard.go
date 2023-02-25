package account

import (
	"net/http"

	"github.com/kirsle/go-website-template/webapp/session"
	"github.com/kirsle/go-website-template/webapp/templates"
)

// User dashboard or landing page (/me).
func Dashboard() http.HandlerFunc {
	tmpl := templates.Must("account/dashboard.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentUser, err := session.CurrentUser(r)
		if err != nil {
			http.Error(w, "Couldn't get currentUser", http.StatusInternalServerError)
			return
		}

		var vars = map[string]interface{}{
			"User": currentUser,
		}
		if err := tmpl.Execute(w, r, vars); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
