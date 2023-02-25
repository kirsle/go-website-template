package templates

import "net/http"

// Redirect sends an HTTP header to the browser.
func Redirect(w http.ResponseWriter, url string) {
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusFound)
}
