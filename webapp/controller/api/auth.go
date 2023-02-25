package api

import (
	"encoding/json"
	"net/http"

	"github.com/kirsle/go-website-template/webapp/session"
)

// LoginOK API tests the validity of a user's session cookie.
func LoginOK() http.HandlerFunc {
	type Response struct {
		Success  bool   `json:"success"`
		UserID   uint64 `json:"userId"`
		Username string `json:"username"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if we're logged in.
		var res Response
		if user, err := session.CurrentUser(r); err == nil {
			res = Response{
				Success:  true,
				UserID:   user.ID,
				Username: user.Username,
			}
		}

		buf, err := json.Marshal(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(buf)
	})
}
