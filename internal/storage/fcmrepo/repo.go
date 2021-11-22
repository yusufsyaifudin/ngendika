package fcmrepo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	ErrValidation = errors.New("validation error")
)

// RepoFCMServerKey .
type RepoFCMServerKey interface {
	Create(ctx context.Context, param FCMServerKey) (inserted FCMServerKey, err error)
	FetchAll(ctx context.Context, appID string) (serverKeys []FCMServerKey, err error)
}

// RepoFCMServiceAccountKey .
type RepoFCMServiceAccountKey interface {
	Create(ctx context.Context, cert FCMServiceAccountKey) (inserted FCMServiceAccountKey, err error)
	FetchAll(ctx context.Context, appID string) (fcmServiceAccountKeys []FCMServiceAccountKey, err error)
}

// Repo .
type Repo interface {
	FCMServerKey() RepoFCMServerKey
	FCMServiceAccountKey() RepoFCMServiceAccountKey
}

// FCMServerKey is FCM server key structure
type FCMServerKey struct {
	ID        string    `json:"id" db:"id" validate:"required,uuid"`
	AppID     string    `json:"app_id" db:"app_id" validate:"required"`
	ServerKey string    `json:"server_key" db:"server_key" validate:"required"`
	CreatedAt time.Time `json:"created_at" db:"created_at" validate:"required"`
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
	AppID             string            `json:"app_id" db:"app_id" validate:"required"`
	ServiceAccountKey ServiceAccountKey `json:"service_account_key" db:"service_account_key" validate:"required"`
	CreatedAt         time.Time         `json:"created_at" db:"created_at" validate:"required"`
}
