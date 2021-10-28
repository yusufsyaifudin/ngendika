package fcmserverkeyrepo

import (
	"time"
)

// FCMServerKey is FCM server key structure
type FCMServerKey struct {
	ID        string    `json:"id" db:"id" validate:"required,uuid"`
	AppID     string    `json:"app_id" db:"app_id" validate:"required"`
	ServerKey string    `json:"server_key" db:"server_key" validate:"required"`
	CreatedAt time.Time `json:"created_at" db:"created_at" validate:"required"`
}
