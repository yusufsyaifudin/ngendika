package apprepo

import (
	"context"
)

// Repo is App repository service
type Repo interface {
	GetAppByClientID(ctx context.Context, clientID string) (app App, err error)
	CreateApp(ctx context.Context, app App) (inserted App, err error)
}

type GetFcmSvcAccKeysIn struct {
	ClientID string `validate:"required"`
}
