package appservice

import (
	"context"
	"time"

	"github.com/yusufsyaifudin/ngendika/internal/storage/apprepo"
)

// Service is an interface of final business logic.
// Any input and output from/to this function should be SAFE for external party to consume,
// i.e: request or response from HTTP handler
type Service interface {
	CreateApp(ctx context.Context, input CreateAppIn) (out CreateAppOut, err error)
	GetAppByClientID(ctx context.Context, clientID string) (app *App, err error)
}

// App is like appstore.App but this only use for returning output via external service.
type App struct {
	ID        string
	ClientID  string
	Name      string
	Enabled   bool
	CreatedAt time.Time
}

func AppFromRepo(app *apprepo.App) *App {
	if app == nil {
		return nil
	}

	return &App{
		ID:        app.ID,
		ClientID:  app.ClientID,
		Name:      app.Name,
		Enabled:   app.Enabled,
		CreatedAt: app.CreatedAt.UTC(), // always UTC
	}
}

func AppToRepo(app *App) *apprepo.App {
	if app == nil {
		return nil
	}

	return &apprepo.App{
		ID:        app.ID,
		ClientID:  app.ClientID,
		Name:      app.Name,
		Enabled:   app.Enabled,
		CreatedAt: app.CreatedAt.UTC(), // always UTC
	}
}

// CreateAppIn ...
type CreateAppIn struct {
	ClientID string `validate:"required"`
	Name     string `validate:"required"`
}

type CreateAppOut struct {
	App *App
}
