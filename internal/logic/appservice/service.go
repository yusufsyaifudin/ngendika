package appservice

import (
	"context"
	"time"
)

// Service is an interface of final business logic.
// Any input and output from/to this function should be SAFE for external party to consume,
// i.e: request or response from HTTP handler
type Service interface {
	CreateApp(ctx context.Context, input CreateAppIn) (out CreateAppOut, err error)
	GetAppByClientID(ctx context.Context, clientID string) (app App, err error)
}

// App is like appstore.App but this only use for returning output via external service.
type App struct {
	ID        string
	ClientID  string
	Name      string
	Enabled   bool
	CreatedAt time.Time
}

// CreateAppIn ...
type CreateAppIn struct {
	ClientID string `validate:"required"`
	Name     string `validate:"required"`
}

type CreateAppOut struct {
	App App
}
