package api

import (
	"fmt"
	"net/http"
)

// Echo API tests the validity of a user's session cookie.
func Echo() http.HandlerFunc {
	// Response JSON schema.
	type Response struct {
		OK    bool        `json:"OK"`
		Error string      `json:"error,omitempty"`
		Echo  interface{} `json:"echo"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse request payload.
		var req interface{}
		if err := ParseJSON(r, &req); err != nil {
			SendJSON(w, http.StatusBadRequest, Response{
				Error: fmt.Sprintf("Error with request payload: %s", err),
			})
			return
		}

		// Send success response.
		SendJSON(w, http.StatusOK, Response{
			OK:   true,
			Echo: req,
		})
	})
}
