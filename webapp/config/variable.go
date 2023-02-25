package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/google/uuid"
	"github.com/kirsle/go-website-template/webapp/log"
)

// Current loaded settings.toml
var Current = DefaultVariable()

// Variable configuration attributes (loaded from settings.toml).
type Variable struct {
	BaseURL          string
	AdminEmail       string
	CronAPIKey       string
	Mail             Mail
	Redis            Redis
	Database         Database
	BareRTC          BareRTC
	UseXForwardedFor bool
}

// DefaultVariable returns the default settings.toml data.
func DefaultVariable() Variable {
	return Variable{
		BaseURL: "http://localhost:8080",
		Mail: Mail{
			Enabled: false,
			Host:    "localhost",
			Port:    25,
			From:    "no-reply@localhost",
		},
		Redis: Redis{
			Host: "localhost",
			Port: 6379,
		},
		Database: Database{
			SQLite:   "database.sqlite",
			Postgres: "host=localhost user=webapp password=webapp dbname=webapp port=5679 sslmode=disable TimeZone=America/Los_Angeles",
		},
		CronAPIKey: uuid.New().String(),
	}
}

// LoadSettings loads the settings.toml file or, if not existing, creates it with the default settings.
func LoadSettings() {
	if _, err := os.Stat(SettingsPath); !os.IsNotExist(err) {
		log.Info("Loading settings from %s", SettingsPath)
		content, err := ioutil.ReadFile(SettingsPath)
		if err != nil {
			panic(fmt.Sprintf("LoadSettings: couldn't read settings.toml: %s", err))
		}

		var v Variable
		err = toml.Unmarshal(content, &v)
		if err != nil {
			panic(fmt.Sprintf("LoadSettings: couldn't parse settings.toml: %s", err))
		}

		Current = v
	} else {
		var buf bytes.Buffer
		enc := toml.NewEncoder(&buf)
		err := enc.Encode(DefaultVariable())
		if err != nil {
			panic(fmt.Sprintf("LoadSettings: couldn't marshal default settings: %s", err))
		}

		ioutil.WriteFile(SettingsPath, buf.Bytes(), 0600)
		log.Warn("NOTICE: Created default settings.toml file - review it and configure mail servers and database!")
	}

	// If there is no DB configured, exit now.
	if !Current.Database.IsSQLite && !Current.Database.IsPostgres {
		log.Error("No database configured in settings.toml. Choose SQLite or Postgres and update the DB connector string!")
		os.Exit(1)
	}
}

// Mail settings.
type Mail struct {
	Enabled  bool
	Host     string // localhost
	Port     int    // 25
	From     string // noreply@localhost
	Username string // SMTP credentials
	Password string
}

// Redis settings.
type Redis struct {
	Host string
	Port int
	DB   int
}

// Database settings.
type Database struct {
	IsSQLite   bool
	IsPostgres bool
	SQLite     string
	Postgres   string
}

// BareRTC chat room settings.
type BareRTC struct {
	JWTSecret string
	URL       string
}
