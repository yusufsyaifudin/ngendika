package appservice

import (
	"time"

	"github.com/yusufsyaifudin/ngendika/internal/storage/fcmrepo"
)

// App is like appstore.App but this only use for returning output via external service.
type App struct {
	ClientID  string    `json:"client_id"`
	Name      string    `json:"name"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateAppIn ...
type CreateAppIn struct {
	ClientID string `json:"client_id"`
	Name     string `json:"name"`
}

type CreateAppOut struct {
	App App `json:"app"`
}

type CreateFcmSvcAccKeyIn struct {
	ClientID             string                    `json:"client_id" validate:"required"`
	FCMServiceAccountKey fcmrepo.ServiceAccountKey `json:"fcm_service_account_key" validate:"required"`
}

type CreateFcmSvcAccKeyOut struct {
	ServiceAccountKey fcmrepo.ServiceAccountKey `json:"service_account_key,omitempty"`
	CreatedAt         time.Time                 `json:"created_at"`
}

type GetFcmSvcAccKeyIn struct {
	ClientID string `json:"client_id" validate:"required"`
}

type GetFcmSvcAccKeyOutList struct {
	ID                string                    `json:"id"`
	ServiceAccountKey fcmrepo.ServiceAccountKey `json:"service_account_key"`
}

type GetFcmSvcAccKeyOut struct {
	Lists []GetFcmSvcAccKeyOutList `json:"lists"`
}
