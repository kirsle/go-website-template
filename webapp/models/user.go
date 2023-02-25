package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/kirsle/go-website-template/webapp/config"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User account table.
type User struct {
	ID             uint64 `gorm:"primaryKey"`
	Username       string `gorm:"uniqueIndex"`
	Email          string `gorm:"uniqueIndex"`
	HashedPassword string
	IsAdmin        bool       `gorm:"index"`
	Status         UserStatus `gorm:"index"` // active, disabled

	CreatedAt   time.Time `gorm:"index"`
	UpdatedAt   time.Time `gorm:"index"`
	LastLoginAt time.Time `gorm:"index"`
}

// Preload related tables for the user (classmethod).
func (u *User) Preload() *gorm.DB {
	// You can eager-load related tables like: (see gorm docs)
	//return DB.Preload("ProfileField").Preload("ProfilePhoto")
	return DB
}

// UserStatus options.
type UserStatus string

const (
	UserStatusActive   = "active"
	UserStatusDisabled = "disabled"
	UserStatusBanned   = "banned"
)

// CreateUser. It is assumed username and email are correctly formatted.
func CreateUser(username, email, password string) (*User, error) {
	// Verify username and email are unique.
	if _, err := FindUser(username); err == nil {
		return nil, errors.New("That username already exists. Please try a different username.")
	} else if _, err := FindUser(email); err == nil {
		return nil, errors.New("That email address is already registered.")
	}

	u := &User{
		Username: username,
		Email:    email,
		Status:   UserStatusActive,
	}

	if err := u.HashPassword(password); err != nil {
		return nil, err
	}

	result := DB.Create(u)
	return u, result.Error
}

// GetUser by ID.
func GetUser(userId uint64) (*User, error) {
	user := &User{}
	result := user.Preload().First(&user, userId)
	return user, result.Error
}

// GetUsers queries for multiple user IDs and returns users in the same order.
func GetUsers(currentUser *User, userIDs []uint64) ([]*User, error) {
	userMap, err := MapUsers(currentUser, userIDs)
	if err != nil {
		return nil, err
	}

	// Re-order them per the original sequence.
	var users = []*User{}
	for _, uid := range userIDs {
		if user, ok := userMap[uid]; ok {
			users = append(users, user)
		}
	}

	return users, nil
}

// FindUser by username or email.
func FindUser(username string) (*User, error) {
	if username == "" {
		return nil, errors.New("username is required")
	}

	u := &User{}
	if strings.ContainsRune(username, '@') {
		result := u.Preload().Where("email = ?", username).Limit(1).First(u)
		return u, result.Error
	}
	result := u.Preload().Where("username = ?", username).Limit(1).First(u)
	return u, result.Error
}

// UserSearch config.
type UserSearch struct {
	EmailOrUsername string
}

// SearchUsers from the perspective of a given user.
func SearchUsers(user *User, search *UserSearch, pager *Pagination) ([]*User, error) {
	if search == nil {
		search = &UserSearch{}
	}

	var (
		users        = []*User{}
		query        *gorm.DB
		wheres       = []string{}
		placeholders = []interface{}{}
	)

	if search.EmailOrUsername != "" {
		ilike := "%" + strings.TrimSpace(strings.ToLower(search.EmailOrUsername)) + "%"
		wheres = append(wheres, "(email LIKE ? OR username LIKE ?)")
		placeholders = append(placeholders, ilike, ilike)
	}

	query = (&User{}).Preload().Where(
		strings.Join(wheres, " AND "),
		placeholders...,
	).Order(pager.Sort)
	query.Model(&User{}).Count(&pager.Total)
	result := query.Offset(pager.GetOffset()).Limit(pager.PerPage).Find(&users)

	return users, result.Error
}

// UserMap helps map a set of users to look up by ID.
type UserMap map[uint64]*User

// MapUsers looks up a set of user IDs in bulk and returns a UserMap suitable for templates.
// Useful to avoid circular reference issues with Photos especially; the Site Gallery queries
// photos of ALL users and MapUsers helps stitch them together for the frontend.
func MapUsers(user *User, userIDs []uint64) (UserMap, error) {
	var (
		usermap  = UserMap{}
		set      = map[uint64]interface{}{}
		distinct = []uint64{}
	)

	// Uniqueify users.
	for _, uid := range userIDs {
		if _, ok := set[uid]; ok {
			continue
		}
		set[uid] = nil
		distinct = append(distinct, uid)
	}

	var (
		users  = []*User{}
		result = (&User{}).Preload().Where("id IN ?", distinct).Find(&users)
	)

	if result.Error == nil {
		for _, row := range users {
			usermap[row.ID] = row
		}
	}

	return usermap, result.Error
}

// Has a user ID in the map?
func (um UserMap) Has(id uint64) bool {
	_, ok := um[id]
	return ok
}

// Get a user from the UserMap.
func (um UserMap) Get(id uint64) *User {
	if user, ok := um[id]; ok {
		return user
	}
	return nil
}

// HashPassword sets the user's hashed (bcrypt) password.
func (u *User) HashPassword(password string) error {
	passwd, err := bcrypt.GenerateFromPassword([]byte(password), config.BcryptCost)
	if err != nil {
		return err
	}
	u.HashedPassword = string(passwd)
	return nil
}

// CheckPassword verifies the password is correct. Returns nil on success.
func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(password))
}

// Save user.
func (u *User) Save() error {
	result := DB.Save(u)
	return result.Error
}

// Delete a user. NOTE: use the models/deletion/DeleteUser() function
// instead of this to do a deep scrub of all related data!
func (u *User) Delete() error {
	return DB.Delete(u).Error
}

// Print user object as pretty JSON.
func (u *User) Print() string {
	var (
		buf bytes.Buffer
		enc = json.NewEncoder(&buf)
	)
	enc.SetIndent("", "    ")
	enc.Encode(u)
	return buf.String()
}
