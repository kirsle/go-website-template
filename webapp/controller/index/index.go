package index

import (
	"net/http"

	"github.com/aichaos/silhouette/webapp/config"
	"github.com/aichaos/silhouette/webapp/log"
	"github.com/aichaos/silhouette/webapp/templates"
)

// Create the controller.
func Create() http.HandlerFunc {
	tmpl := templates.Must("index.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("Beginning of index page")
		if r.URL.Path != "/" || r.Method != http.MethodGet {
			log.Error("404 Not Found: %s", r.URL.Path)
			templates.NotFoundPage(w, r)
			return
		}

		if err := tmpl.Execute(w, r, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

// Favicon
func Favicon() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, config.StaticPath+"/favicon.ico")
	})
}
