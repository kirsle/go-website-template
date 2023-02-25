package main

import (
	"fmt"
	"os"

	webapp "github.com/kirsle/go-website-template/webapp"
	"github.com/kirsle/go-website-template/webapp/config"
	"github.com/kirsle/go-website-template/webapp/log"
	"github.com/kirsle/go-website-template/webapp/models"
	"github.com/kirsle/go-website-template/webapp/redis"
	"github.com/urfave/cli/v2"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Build-time values.
var (
	Build     = "n/a"
	BuildDate = "n/a"
)

func init() {
	config.RuntimeVersion = webapp.Version
	config.RuntimeBuild = Build
	config.RuntimeBuildDate = BuildDate
}

func main() {
	app := &cli.App{
		Name:  "webapp",
		Usage: "a niche social networking webapp",
		Commands: []*cli.Command{
			{
				Name:  "web",
				Usage: "start the web server",
				Flags: []cli.Flag{
					// Debug mode.
					&cli.BoolFlag{
						Name:    "debug",
						Aliases: []string{"d"},
						Usage:   "debug mode (logging and reloading templates)",
					},

					// HTTP settings.
					&cli.StringFlag{
						Name:    "host",
						Aliases: []string{"H"},
						Value:   "0.0.0.0",
						Usage:   "host address to listen on",
					},
					&cli.IntFlag{
						Name:    "port",
						Aliases: []string{"P"},
						Value:   8080,
						Usage:   "port number to listen on",
					},
				},
				Action: func(c *cli.Context) error {
					if c.Bool("debug") {
						config.Debug = true
						log.SetDebug(true)
					}

					initdb(c)
					initcache(c)

					log.Debug("Debug logging enabled.")

					app := &webapp.WebServer{
						Host: c.String("host"),
						Port: c.Int("port"),
					}

					return app.Run()
				},
			},
			{
				Name:  "user",
				Usage: "manage user accounts such as to create admins",
				Subcommands: []*cli.Command{
					{
						Name:  "add",
						Usage: "add a new user account",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "username",
								Aliases:  []string{"u"},
								Required: true,
								Usage:    "username, case insensitive",
							},
							&cli.StringFlag{
								Name:     "email",
								Aliases:  []string{"e"},
								Required: true,
								Usage:    "email address",
							},
							&cli.StringFlag{
								Name:     "password",
								Aliases:  []string{"p"},
								Required: true,
								Usage:    "set user password",
							},
							&cli.BoolFlag{
								Name:  "admin",
								Usage: "set admin status",
							},
						},
						Action: func(c *cli.Context) error {
							initdb(c)

							log.Info("Creating user account: %s", c.String("username"))
							user, err := models.CreateUser(
								c.String("username"),
								c.String("email"),
								c.String("password"),
							)

							if err != nil {
								return err
							}

							// Making an admin?
							if c.Bool("admin") {
								log.Warn("Promoting user to admin status")
								user.IsAdmin = true
								user.Save()
							}
							return nil
						},
					},
				},
			},
			{
				Name:  "backfill",
				Usage: "One-off maintenance tasks and data backfills for database migrations",
				Subcommands: []*cli.Command{
					{
						Name:  "example",
						Usage: "repopulate Filesizes on all photos and comment_photos which have a zero stored in the DB",
						Action: func(c *cli.Context) error {
							initdb(c)

							log.Info("Example backfill function.")

							return nil
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func initdb(c *cli.Context) {
	// Load the settings.json
	config.LoadSettings()

	var gormcfg = &gorm.Config{}
	if c.Bool("debug") {
		gormcfg = &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		}
	}

	// Initialize the database.
	log.Info("Initializing DB")
	if config.Current.Database.IsSQLite {
		db, err := gorm.Open(sqlite.Open(config.Current.Database.SQLite), gormcfg)
		if err != nil {
			panic("failed to open SQLite DB")
		}
		models.DB = db
	} else if config.Current.Database.IsPostgres {
		db, err := gorm.Open(postgres.Open(config.Current.Database.Postgres), gormcfg)
		if err != nil {
			panic(fmt.Sprintf("failed to open Postgres DB: %s", err))
		}
		models.DB = db
	} else {
		log.Fatal("A choice of SQL database is required.")
	}

	// Auto-migrate the DB.
	models.AutoMigrate()
}

func initcache(c *cli.Context) {
	// Initialize Redis.
	log.Info("Initializing Redis")
	redis.Setup(c.String("redis"))
}
