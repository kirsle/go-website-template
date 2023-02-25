// Package models handles the database.
package models

import "gorm.io/gorm"

// DB to be set by calling app (SQLite or Postgres connection).
var DB *gorm.DB

// AutoMigrate the schema. List all your new models here for DB creation.
func AutoMigrate() {
	DB.AutoMigrate(&User{})
}
