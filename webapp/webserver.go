package webapp

import (
	"fmt"
	"net/http"

	"github.com/aichaos/silhouette/webapp/log"
	"github.com/aichaos/silhouette/webapp/router"
)

// WebServer is the main entry point for the `webapp web` command.
type WebServer struct {
	// Configuration
	Host string // host interface, default "0.0.0.0"
	Port int    // default 8080
}

// Run the server.
func (ws *WebServer) Run() error {
	// Defaults
	if ws.Host == "" {
		ws.Host = "0.0.0.0"
	}
	if ws.Port == 0 {
		ws.Port = 8080
	}

	s := http.Server{
		Addr:    fmt.Sprintf("%s:%d", ws.Host, ws.Port),
		Handler: router.New(),
	}

	log.Info("Listening at http://%s:%d", ws.Host, ws.Port)
	return s.ListenAndServe()
}
