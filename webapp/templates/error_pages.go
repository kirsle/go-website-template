package templates

import (
	"net/http"
)

// NotFoundPage is an HTTP handler for 404 pages.
var NotFoundPage = func() http.HandlerFunc {
	return MakeErrorPage("Not Found", "The page you requested was not here.", http.StatusNotFound)
}()

// ForbiddenPage is an HTTP handler for 404 pages.
var ForbiddenPage = func() http.HandlerFunc {
	return MakeErrorPage("Forbidden", "You do not have permission for this page.", http.StatusForbidden)
}()

func MakeErrorPage(header string, message string, statusCode int) http.HandlerFunc {
	tmpl := Must("errors/error.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		if err := tmpl.Execute(w, r, map[string]interface{}{
			"Header":  header,
			"Message": message,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
