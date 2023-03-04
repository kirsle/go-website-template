package api

import (
	"encoding/json"
	"net/http"

	"github.com/aichaos/silhouette/webapp/config"
)

// Version details of the running app.
func Version() http.HandlerFunc {
	// Response JSON schema.
	type Response struct {
		Version   string `json:"version"`
		Build     string `json:"build"`
		BuildDate string `json:"buildDate"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf, err := json.Marshal(Response{
			Version:   config.RuntimeVersion,
			Build:     config.RuntimeBuild,
			BuildDate: config.RuntimeBuildDate,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(buf)
	})
}
