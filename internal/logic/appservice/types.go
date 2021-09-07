package appservice

import (
	"time"

	"github.com/yusufsyaifudin/ngendika/storage/apprepo"
)

// App is like appstore.App but this only use for returning output via external service.
type App struct {
	ClientID  string    `json:"client_id"`
	Name      string    `json:"name"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

// InputCreateApp ...
type InputCreateApp struct {
	ClientID string `json:"client_id"`
	Name     string `json:"name"`
}

type OutputCreateApp struct {
	App App `json:"app"`
}

type InputCreateFcmServiceAccountKey struct {
	ClientID             string                    `json:"client_id" validate:"required"`
	FCMServiceAccountKey apprepo.ServiceAccountKey `json:"fcm_service_account_key" validate:"required"`
}

type OutputCreateFcmServiceAccountKey struct {
	ServiceAccountKey apprepo.ServiceAccountKey `json:"service_account_key,omitempty"`
	CreatedAt         time.Time                 `json:"created_at"`
}
