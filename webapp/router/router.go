// Package router configures web routes.
package router

import (
	"net/http"

	"github.com/kirsle/go-website-template/webapp/config"
	"github.com/kirsle/go-website-template/webapp/controller/account"
	"github.com/kirsle/go-website-template/webapp/controller/admin"
	"github.com/kirsle/go-website-template/webapp/controller/api"
	"github.com/kirsle/go-website-template/webapp/controller/index"
	"github.com/kirsle/go-website-template/webapp/middleware"
)

func New() http.Handler {
	mux := http.NewServeMux()

	// Register controller endpoints.
	mux.HandleFunc("/", index.Create())
	mux.HandleFunc("/favicon.ico", index.Favicon())
	mux.HandleFunc("/about", index.StaticTemplate("about.html")())
	mux.HandleFunc("/login", account.Login())
	mux.HandleFunc("/logout", account.Logout())
	mux.HandleFunc("/signup", account.Signup())
	mux.HandleFunc("/forgot-password", account.ForgotPassword())
	mux.HandleFunc("/settings/confirm-email", account.ConfirmEmailChange())

	// Login Required. Pages that non-certified users can access.
	mux.Handle("/me", middleware.LoginRequired(account.Dashboard()))
	mux.Handle("/settings", middleware.LoginRequired(account.Settings()))
	mux.Handle("/account/delete", middleware.LoginRequired(account.Delete()))

	// Certification Required. Pages that only full (verified) members can access.
	mux.Handle("/members", middleware.LoginRequired(account.Search()))

	// Admin endpoints.
	mux.Handle("/admin", middleware.AdminRequired(admin.Dashboard()))
	mux.Handle("/admin/user-action", middleware.AdminRequired(admin.UserActions()))

	// JSON API endpoints.
	mux.HandleFunc("/v1/version", api.Version())
	mux.HandleFunc("/v1/users/me", api.LoginOK())
	mux.HandleFunc("/v1/echo", api.Echo())

	// Static files.
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(config.StaticPath))))

	// Global middlewares.
	withCSRF := middleware.CSRF(mux)
	withSession := middleware.Session(withCSRF)
	withRecovery := middleware.Recovery(withSession)
	withLogger := middleware.Logging(withRecovery)
	return withLogger
}
