package api

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

// Envelope is the standard JSON response envelope.
type Envelope struct {
	Data       interface{} `json:"data"`
	StatusCode int
}

// ParseJSON request body.
func ParseJSON(r *http.Request, v interface{}) error {
	if r.Header.Get("Content-Type") != "application/json" {
		return errors.New("request Content-Type must be application/json")
	}

	// Parse request body.
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	// Parse params from JSON.
	if err := json.Unmarshal(body, v); err != nil {
		return err
	}

	return nil
}

// SendJSON response.
func SendJSON(w http.ResponseWriter, statusCode int, v interface{}) {
	buf, err := json.Marshal(Envelope{
		Data:       v,
		StatusCode: statusCode,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(buf)
}
