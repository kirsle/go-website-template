package middleware

import (
	"fmt"
	"net/http"
	"time"
)

// Logging middleware.
func Logging(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nw := time.Now()
		handler.ServeHTTP(w, r)
		fmt.Printf("%s %s %s %s\n", r.RemoteAddr, r.Method, r.URL, time.Since(nw))
	})
}
