package apprepo

// App save all apps and it's connection info relation
// Json tag is used for caching.
type App struct {
	ID       int64  `json:"id" db:"id" validate:"required"`               // primary key
	ClientID string `json:"client_id" db:"client_id" validate:"required"` // unique
	Name     string `json:"name" db:"name" validate:"required"`

	// Timestamp using integer as unix microsecond in UTC
	CreatedAt int64 `json:"created_at" db:"created_at" validate:"required"`
	UpdatedAt int64 `json:"updated_at" db:"updated_at" validate:"required"`
	DeletedAt int64 `json:"deleted_at" db:"deleted_at" validate:"-"`
}
