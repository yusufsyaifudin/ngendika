package apprepo

import (
	"context"
	"errors"
)

var (
	ErrValidation       = errors.New("validation error")
	ErrAppWrongClientID = errors.New("app client_id is in wrong format")
)

// Repo is App repository service
type Repo interface {
	GetAppByClientID(ctx context.Context, clientID string) (app *App, err error)
	CreateApp(ctx context.Context, app *App) (inserted *App, err error)
}
