# Go Website Template

This repository starts at a good base for a Go web app featuring authentication, log in/out
and support for Postgres or SQLite databases. It is a pared-down copy of a site I built that
uses largely just the standard Go net/http library (and so implements session cookies, CSRF
protection, and so on from scratch).

## Features

This code serves as a good base for a Go-backed web app that serves server-side rendered
templates and has basic account management functions (sign up, in/out, admin).

* Fairly vanilla standard library
    * Uses the standard `net/http` and `html/template` modules and wires it all up from scratch.
    * Middlewares including session cookie (Redis backed), Login Required, Admin Required,
      CSRF protection, logging and panic recovery all written from scratch.
    * Session cookie features include "flashed" success/error messages that display on next
      page load.
* **Database**:
    * Uses the [gorm](https://gorm.io) ORM so that I can easily run SQLite locally but Postgres
      on my production server.
    * A basic setup with a single `users` table and some useful query examples including
      pagination.
    * PostgreSQL and SQLite officially supported.
* **Redis** cache:
    * For signup workflows needing email verification
    * For rate limiting failed password attempts
    * Simple API built out that (de)serializes w/e you give it as JSON into the Redis cache.
* **JSON APIs:**
    * A couple example JSON (ajax) endpoints and the start of a standardized wrapper format
      (in webapp/controller/api/json_layer.go).
* User accounts:
    * Create account (with email verification required, or not - it's hardcoded in config.go)
    * Log in or out
    * Admin flag (endpoints to e.g. ban users may be missing)
    * CLI interface to create users locally (skipping email verification) or create the first
      admin account.
* A few basic pages: about, dashboards, etc.
* A simple front-end website using the [Bulma](https://bulma.io) CSS library.

## Renaming the module

The first thing you'll want to do is rename the Go module which can be done like so:

```bash
# what you're renaming it into:
MODULE_NAME="github.com/my-username/my-project"

# From the working directory of this git repository checkout, e.g.
find . -type f -name "*.go" -print0 | xargs -0 sed -i '' -e "s,github.com/kirsle/go-website-template,${MODULE_NAME},g"
sed -i'' -e "s,github.com/kirsle/go-website-template,${MODULE_NAME},g" go.mod
```

TODO: rest of this document.

## Dependencies

You may need to run the following services along with this app:

* A **Redis cache** server: [redis.io](https://redis.io)
* (Optional) a **PostgreSQL database:** [postgresql.org](https://www.postgresql.org/)

The website can also run out of a local SQLite database which is convenient
for local development. The production server runs on PostgreSQL and the
web app is primarily designed for that.

## Building the App

This app is written in Go: [go.dev](https://go.dev). You can probably
get it from your package manager, e.g.

* macOS: `brew install golang` with homebrew: [brew.sh](https://brew.sh)
* Linux: it's in your package manager, e.g. `apt install golang`

Use the Makefile (with GNU `make` or similar):

* `make setup`: install Go dependencies
* `make build`: builds the program to ./webapp
* `make run`: run the app from Go sources in debug mode

Or read the Makefile to see what the underlying `go` commands are,
e.g. `go run cmd/webapp/main.go web`

## Configuring

On first run it will generate a `settings.toml` file in the current
working directory (which is intended to be the root of the git clone,
with the ./web folder). Edit it to configure mail settings or choose
a database.

For simple local development, just set `"UseSQLite": true` and the
app will run with a SQLite database.

## Usage

The `webapp` binary has sub-commands to either run the web server
or perform maintenance tasks such as creating admin user accounts.

Run `webapp --help` for its documentation.

Run `webapp web` to start the web server.

```bash
webapp web --host 0.0.0.0 --port 8080 --debug
```

## Create Admin User Accounts

Use the `webapp user add` command like so:

```bash
$ webapp user add --admin \
  --email name@domain.com \
  --password secret \
  --username admin
```

Shorthand options `-e`, `-p` and `-u` can work in place of the longer
options `--email`, `--password` and `--username` respectively.

After the first admin user is created, you may promote other users thru
the web app by using the admin controls on their profile page.

## A Brief Tour of the Code

* `cmd/webapp/main.go`: the entry point for the Go program.
* `pkg/webserver.go`: the entry point for the web server.
* `pkg/config`: mostly hard-coded configuration values - all of the page
  sizes and business logic controls are in here, set at compile time. For
  ease of local development you may want to toggle SkipEmailValidation in
  here - the signup form will then directly allow full signup with a user
  and password.
* `pkg/controller`: the various web endpoint controllers are here,
  categorized into subpackages (account, forum, inbox, photo, etc.)
* `pkg/log`: the logging to terminal functions.
* `pkg/mail`: functions for delivering HTML email messages.
* `pkg/markdown`: functions to render GitHub Flavored Markdown.
* `pkg/middleware`: HTTP middleware functions, for things such as:
    * Session cookies
    * Authentication (LoginRequired, AdminRequired)
    * CSRF protection
    * Logging HTTP requests
    * Panic recovery for unhandled server errors
* `pkg/models`: the SQL database models and query functions are here.
    * `pkg/models/deletion`: the code to fully scrub wipe data for
      user deletion (GDPR/CCPA compliance).
* `pkg/ratelimit`: rate limiter for login attempts etc.
* `pkg/redis`: Redis cache functions - get/set JSON values for things like
  session cookie storage and temporary rate limits.
* `pkg/router`: the HTTP route URLs for the controllers are here.
* `pkg/session`: functions to read/write the user's session cookie
  (log in/out, get current user, flash messages)
* `pkg/templates`: functions to handle HTTP responses - render HTML
  templates, issue redirects, error pages, ...
* `pkg/utility`: miscellaneous useful functions for the app.

## License

MIT, or whatever - I don't care.