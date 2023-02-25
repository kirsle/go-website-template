// Package session handles user login and other cookies.
package session

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kirsle/go-website-template/webapp/config"
	"github.com/kirsle/go-website-template/webapp/log"
	"github.com/kirsle/go-website-template/webapp/mail"
	"github.com/kirsle/go-website-template/webapp/models"
	"github.com/kirsle/go-website-template/webapp/redis"
)

// Session cookie object that is kept server side in Redis.
type Session struct {
	UUID         string    `json:"-"` // not stored
	LoggedIn     bool      `json:"loggedIn"`
	UserID       uint64    `json:"userId,omitempty"`
	Flashes      []string  `json:"flashes,omitempty"`
	Errors       []string  `json:"errors,omitempty"`
	Impersonator uint64    `json:"impersonator,omitempty"`
	LastSeen     time.Time `json:"lastSeen"`
}

const (
	ContextKey     = "session"
	CurrentUserKey = "current_user"
	CSRFKey        = "csrf"
)

// New creates a blank session object.
func New() *Session {
	return &Session{
		UUID:    uuid.New().String(),
		Flashes: []string{},
		Errors:  []string{},
	}
}

// Load the session from the browser session_id token and Redis or creates a new session.
func LoadOrNew(r *http.Request) *Session {
	var sess = New()

	// Read the session cookie value.
	cookie, err := r.Cookie(config.SessionCookieName)
	if err != nil {
		log.Debug("session.LoadOrNew: cookie error, new sess: %s", err)
		return sess
	}

	// Look up this UUID in Redis.
	sess.UUID = cookie.Value
	key := fmt.Sprintf(config.SessionRedisKeyFormat, sess.UUID)

	err = redis.Get(key, sess)
	// log.Error("LoadOrNew: raw from Redis: %+v", sess)
	if err != nil {
		log.Error("session.LoadOrNew: didn't find %s in Redis: %s", key, err)
	}

	return sess
}

// Save the session and send a cookie header.
func (s *Session) Save(w http.ResponseWriter) {
	// Roll a UUID session_id value.
	if s.UUID == "" {
		s.UUID = uuid.New().String()
	}

	// Ensure it is a valid UUID.
	if _, err := uuid.Parse(s.UUID); err != nil {
		log.Error("Session.Save: got an invalid UUID session_id: %s", err)
		s.UUID = uuid.New().String()
	}

	// Ping last seen.
	s.LastSeen = time.Now()

	// Save their session object in Redis.
	key := fmt.Sprintf(config.SessionRedisKeyFormat, s.UUID)
	if err := redis.Set(key, s, config.SessionCookieMaxAge*time.Second); err != nil {
		log.Error("Session.Save: couldn't write to Redis: %s", err)
	}

	cookie := &http.Cookie{
		Name:     config.SessionCookieName,
		Value:    s.UUID,
		MaxAge:   config.SessionCookieMaxAge,
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}

// Get the session from the current HTTP request context.
func Get(r *http.Request) *Session {
	if r == nil {
		panic("session.Get: http.Request is required")
	}

	ctx := r.Context()
	if sess, ok := ctx.Value(ContextKey).(*Session); ok {
		return sess
	}

	// If the session isn't on the request, it means I broke something.
	log.Error("session.Get(): didn't find session in request context!")
	return nil
}

var portSuffixRegexp = regexp.MustCompile(`:(\d+)$`)

// RemoteAddr returns the user's remote IP address. If UseXForwardedFor is enabled in settings.json,
// the HTTP header X-Forwarded-For may be returned here or otherwise the request RemoteAddr is returned.
func RemoteAddr(r *http.Request) string {
	var remoteAddr = r.RemoteAddr // Usually "ip:port" format
	if config.Current.UseXForwardedFor {
		xff := r.Header.Get("X-Forwarded-For")
		if len(xff) > 0 {
			remoteAddr = strings.SplitN(xff, ",", 2)[0]
		}
	}

	// Return just the IP and not the port suffix.
	return portSuffixRegexp.ReplaceAllString(remoteAddr, "")
}

// ReadFlashes returns and clears the Flashes and Errors for this session.
func (s *Session) ReadFlashes(w http.ResponseWriter) (flashes, errors []string) {
	flashes = s.Flashes
	errors = s.Errors
	s.Flashes = []string{}
	s.Errors = []string{}
	if len(flashes)+len(errors) > 0 {
		s.Save(w)
	}
	return flashes, errors
}

// Flash adds a transient message to the user's session to show on next page load.
func Flash(w http.ResponseWriter, r *http.Request, msg string, args ...interface{}) {
	sess := Get(r)
	sess.Flashes = append(sess.Flashes, fmt.Sprintf(msg, args...))
	sess.Save(w)
}

// FlashError adds a transient error message to the session.
func FlashError(w http.ResponseWriter, r *http.Request, msg string, args ...interface{}) {
	sess := Get(r)
	sess.Errors = append(sess.Errors, fmt.Sprintf(msg, args...))
	sess.Save(w)
}

// LoginUser marks a session as logged in to an account.
func LoginUser(w http.ResponseWriter, r *http.Request, u *models.User) error {
	if u == nil || u.ID == 0 {
		return errors.New("not a valid user account")
	}

	sess := Get(r)
	sess.LoggedIn = true
	sess.UserID = u.ID
	sess.Impersonator = 0
	sess.Save(w)

	// Ping the user's last login time.
	u.LastLoginAt = time.Now()
	return u.Save()
}

// ImpersonateUser assumes the role of the user impersonated by an admin uid.
func ImpersonateUser(w http.ResponseWriter, r *http.Request, u *models.User, impersonator *models.User, reason string) error {
	if u == nil || u.ID == 0 {
		return errors.New("not a valid user account")
	}
	if impersonator == nil || impersonator.ID == 0 || !impersonator.IsAdmin {
		return errors.New("impersonator not a valid admin account")
	}

	sess := Get(r)
	sess.LoggedIn = true
	sess.UserID = u.ID
	sess.Impersonator = impersonator.ID
	sess.Save(w)

	// Email the admins.
	if err := mail.Send(mail.Message{
		To:       config.Current.AdminEmail,
		Subject:  "Admin 'user impersonate' has been used",
		Template: "email/admin_impersonate.html",
		Data: map[string]interface{}{
			"Impersonator": impersonator,
			"User":         u,
			"Reason":       reason,
			"AdminURL":     config.Current.BaseURL + "/admin/feedback",
		},
	}); err != nil {
		log.Error("/contact page: couldn't send email: %s", err)
	}

	return u.Save()
}

// Impersonated returns if the current session has an impersonator.
func Impersonated(r *http.Request) bool {
	sess := Get(r)
	if sess == nil {
		return false
	}

	return sess.Impersonator > 0
}

// LogoutUser signs a user out.
func LogoutUser(w http.ResponseWriter, r *http.Request) {
	sess := Get(r)
	sess.LoggedIn = false
	sess.UserID = 0
	sess.Save(w)
}
