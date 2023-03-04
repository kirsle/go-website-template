package templates

import (
	"net/http"
	"time"

	"github.com/aichaos/silhouette/webapp/config"
	"github.com/aichaos/silhouette/webapp/session"
)

// MergeVars mixes in globally available template variables. The http.Request is optional.
func MergeVars(r *http.Request, m map[string]interface{}) {
	m["Title"] = config.Title
	m["Subtitle"] = config.Title
	m["BuildHash"] = config.RuntimeBuild
	m["BuildDate"] = config.RuntimeBuildDate
	m["Subtitle"] = config.Subtitle
	m["YYYY"] = time.Now().Year()

	if r == nil {
		return
	}

	m["Request"] = r
}

// MergeUserVars mixes in global template variables: LoggedIn and CurrentUser. The http.Request is optional.
func MergeUserVars(r *http.Request, m map[string]interface{}) {
	// Defaults
	m["LoggedIn"] = false
	m["CurrentUser"] = nil
	m["SessionImpersonated"] = false

	if r == nil {
		return
	}

	m["SessionImpersonated"] = session.Impersonated(r)

	if user, err := session.CurrentUser(r); err == nil {
		m["LoggedIn"] = true
		m["CurrentUser"] = user
	}
}
