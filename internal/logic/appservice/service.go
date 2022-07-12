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
	CreateApp(ctx context.Context, input InputCreateApp) (out OutCreateApp, err error)
	GetApp(ctx context.Context, input InputGetApp) (out OutGetApp, err error)
}

// App is like appstore.App but this only use for returning output via external service.
type App struct {
	ID        string    `json:"id"`
	ClientID  string    `json:"clientID"`
	Name      string    `json:"name"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
}

func AppFromRepo(app apprepo.App) App {
	a := App(app)
	return a
}

// InputCreateApp ...
type InputCreateApp struct {
	ClientID string `json:"clientID" validate:"required,alphanum,lowercase"`
	Name     string `json:"name" validate:"required"`
}

type OutCreateApp struct {
	App App `json:"app"`
}

type InputGetApp struct {
	ClientID string `validate:"required,lowercase"`
	Enabled  *bool  `validate:"-"`
}

type OutGetApp struct {
	App App `json:"app"`
}
