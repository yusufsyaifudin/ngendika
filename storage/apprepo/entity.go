package apprepo

import (
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/encoding/json"
)

// App save all apps and it's connection info relation
type App struct {
	ClientID  string    `json:"client_id" db:"client_id" validate:"required"` // unique, primary key
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

// ServiceAccountKey is represented FCM Service Account Key format
type ServiceAccountKey struct {
	Type                    string `json:"type" validate:"required"`
	ProjectID               string `json:"project_id" validate:"required"`
	PrivateKeyID            string `json:"private_key_id" validate:"required"`
	PrivateKey              string `json:"private_key" validate:"required"`
	ClientEmail             string `json:"client_email" validate:"required"`
	ClientID                string `json:"client_id" validate:"required"`
	AuthURI                 string `json:"auth_uri" validate:"required"`
	TokenURI                string `json:"token_uri" validate:"required"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url" validate:"required"`
	ClientX509CertURL       string `json:"client_x509_cert_url" validate:"required"`
}

func (m *ServiceAccountKey) Scan(src interface{}) error {
	return json.Unmarshal([]byte(fmt.Sprintf("%s", src)), m)
}

// FCMServiceAccountKey is FCM service account key table structure
type FCMServiceAccountKey struct {
	ID                string            `json:"id" db:"id" validate:"required,uuid"`
	AppClientID       string            `json:"app_client_id" db:"app_client_id" validate:"required"`
	ServiceAccountKey ServiceAccountKey `json:"service_account_key" db:"service_account_key" validate:"required"`
	CreatedAt         time.Time         `json:"created_at" db:"created_at" validate:"required"`
}

// FCMServerKey is FCM server key structure
type FCMServerKey struct {
	ID          string    `json:"id" db:"id" validate:"required,uuid"`
	AppClientID string    `json:"app_client_id" db:"app_client_id" validate:"required"`
	ServerKey   string    `json:"server_key" db:"server_key" validate:"required"`
	CreatedAt   time.Time `json:"created_at" db:"created_at" validate:"required"`
}
