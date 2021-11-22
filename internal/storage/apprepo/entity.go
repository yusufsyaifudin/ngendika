package apprepo

import (
	"strings"
	"time"
)

// App save all apps and it's connection info relation
type App struct {
	ID        string    `json:"id" db:"id" validate:"-"`                      // primary key
	ClientID  string    `json:"client_id" db:"client_id" validate:"required"` // unique
	Name      string    `json:"name" db:"name" validate:"required"`
	Enabled   bool      `json:"enabled" db:"enabled" validate:"required"`
	CreatedAt time.Time `json:"created_at" db:"created_at" validate:"required"`
}

func NewApp(clientID, name string) App {
	return App{
		ClientID:  strings.ToLower(clientID),
		Name:      name,
		Enabled:   true,
		CreatedAt: time.Now().UTC(),
	}
}