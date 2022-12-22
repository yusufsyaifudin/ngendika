package backend

import (
	"encoding/json"
	"time"
)

type PushNotificationProvider struct {
	ID             int64  `json:"id" validate:"required"`
	AppID          int64  `json:"app_id" validate:"required"`
	Provider       string `json:"provider" validate:"required"`
	Label          string `json:"label" validate:"required"`
	CredentialJSON string `json:"credential_json" validate:"required"`

	// Timestamp using integer as unix microsecond in UTC
	CreatedAt time.Time `json:"created_at" validate:"required"`
	UpdatedAt time.Time `json:"updated_at" validate:"required"`
}

func (p PushNotificationProvider) MarshalJSON() ([]byte, error) {

	var credJson interface{}
	if _err := json.Unmarshal([]byte(p.CredentialJSON), &credJson); _err != nil {
		credJson = p.CredentialJSON
	}

	alias := struct {
		ID             int64       `json:"id"`
		AppID          int64       `json:"app_id"`
		Provider       string      `json:"provider"`
		Label          string      `json:"label"`
		CredentialJSON interface{} `json:"credential_json"`
		CreatedAt      time.Time   `json:"created_at"`
		UpdatedAt      time.Time   `json:"updated_at"`
	}{
		ID:             p.ID,
		AppID:          p.AppID,
		Provider:       p.Provider,
		Label:          p.Label,
		CredentialJSON: credJson,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}

	return json.Marshal(alias)
}
