package admin

import (
	"net/http"

	"github.com/kirsle/go-website-template/webapp/templates"
)

// Admin dashboard or landing page (/admin).
func Dashboard() http.HandlerFunc {
	tmpl := templates.Must("admin/dashboard.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := tmpl.Execute(w, r, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
