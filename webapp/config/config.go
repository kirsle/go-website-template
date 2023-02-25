// Package config holds some (mostly static) configuration for the app.
package config

import (
	"regexp"
	"time"
)

// Branding
const (
	Title    = "Untitled Website"
	Subtitle = "A good base for a simple Go web application."
)

// Paths and layouts
const (
	TemplatePath = "./web/templates"
	StaticPath   = "./web/static"
	SettingsPath = "./settings.toml"
)

// Security
const (
	BcryptCost            = 14
	SessionCookieName     = "session_id"
	CSRFCookieName        = "xsrf_token"
	CSRFInputName         = "_csrf" // html input name
	SessionCookieMaxAge   = 60 * 60 * 24 * 30
	SessionRedisKeyFormat = "session/%s"
	MultipartMaxMemory    = 1024 * 1024 * 1024 * 20 // 20 MB
)

// Authentication
const (
	// Skip the email verification step. The signup page will directly ask for
	// email+username+password rather than only email and needing verification.
	SkipEmailVerification = false

	SignupTokenRedisKey   = "signup-token/%s"
	ResetPasswordRedisKey = "reset-password/%s"
	ChangeEmailRedisKey   = "change-email/%s"
	SignupTokenExpires    = 24 * time.Hour // used for all tokens so far

	// Rate limits
	RateLimitRedisKey        = "rate-limit/%s/%s" // namespace, id
	LoginRateLimitWindow     = 1 * time.Hour
	LoginRateLimit           = 10 // 10 failed login attempts = locked for full hour
	LoginRateLimitCooldownAt = 3  // 3 failed attempts = start throttling
	LoginRateLimitCooldown   = 30 * time.Second

	// Contact form rate limits for logged-out users to curb spam robots:
	// - One message can be submitted every 2 minutes
	// - If they post 10 minutes in an hour they are paused for one hour.
	ContactRateLimitWindow     = 1 * time.Hour
	ContactRateLimit           = 10
	ContactRateLimitCooldownAt = 1
	ContactRateLimitCooldown   = 2 * time.Minute

	// How frequently to refresh LastLoginAt since sessions are long-lived.
	LastLoginAtCooldown = 8 * time.Hour
)

var (
	UsernameRegexp    = regexp.MustCompile(`^[a-z0-9_-]{3,32}$`)
	ReservedUsernames = []string{
		"admin",
		"admins",
		"administrator",
		"moderator",
		"support",
		"staff",
	}
)

// Photo Galleries
const (
	MaxPhotoWidth     = 1280
	ProfilePhotoWidth = 512

	// Quotas for uploaded photos.
	PhotoQuotaUncertified = 6
	PhotoQuotaCertified   = 100
)

// Forum settings
const (
	// Only ++ the Views count per user per thread within a small
	// window of time - if a user keeps reloading the same thread
	// rapidly it does not increment the view counter more.
	ThreadViewDebounceRedisKey = "debounce-view/user=%d/thr=%d"
	ThreadViewDebounceCooldown = 1 * time.Hour
)

// Poll settings
var (
	// Max number of responses to accept for a poll (how many form
	// values the app will read in). NOTE: also enforced in frontend
	// UX in new_post.html, update there if you change this.
	PollMaxAnswers = 100

	// Poll color CSS classes (Bulma). Plugged in to templates like:
	// <progress class="$CLASS">
	// Values will wrap around for long polls.
	PollProgressBarClasses = []string{
		"progress is-success",
		"progress is-link",
		"progress is-warning",
		"progress is-danger",
		"progress is-primary",
		"progress is-info",
	}
)

// Variables set by main.go to make them readily available.
var (
	RuntimeVersion   string
	RuntimeBuild     string
	RuntimeBuildDate string
	Debug            bool // app is in debug mode
)
