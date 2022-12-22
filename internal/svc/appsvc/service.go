package appsvc

import (
	"context"
	"time"
)

// Service is an interface of final business logic.
// Any input and output from/to this function should be SAFE for external party to consume,
// i.e: request or response from HTTP handler
type Service interface {
	CreateApp(ctx context.Context, input InputCreateApp) (out OutCreateApp, err error)
	PutApp(ctx context.Context, input InputPutApp) (out OutPutApp, err error)
	GetApp(ctx context.Context, input InputGetApp) (out OutGetApp, err error)
	ListApp(ctx context.Context, input InputListApp) (out OutListApp, err error)
	DelApp(ctx context.Context, input InputDelApp) (out OutDelApp, err error)
}

// App is like appstore.AppRepo but this only use for returning output via external service.
// This must not have any json or yaml tag, any output method (HTTP, gRPC, etc) must define its own entity standard.
// Service level just only act as input -> process -> output, not taking care of request/response traffic.
type App struct {
	ID        int64  `validate:"required"`
	ClientID  string `validate:"required"`
	Name      string `validate:"required"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

// InputCreateApp ...
type InputCreateApp struct {
	ClientID string `validate:"required,alphanum,lowercase"`
	Name     string `validate:"required"`
}

type OutCreateApp struct {
	App App
}

type InputPutApp struct {
	ClientID string `validate:"required,alphanum,lowercase"`
	Name     string `validate:"required"`
}

type OutPutApp struct {
	App App
}

type InputGetApp struct {
	ClientID string `validate:"required,lowercase"`
	Enabled  *bool  `validate:"-"`
}

type OutGetApp struct {
	App App
}

type InputListApp struct {
	Limit    int64 `validate:"min=0"`
	BeforeID int64 `validate:"min=0"`
	AfterID  int64 `validate:"min=0"`
}

type OutListApp struct {
	Total int64
	Limit int64
	Apps  []App
}

type InputDelApp struct {
	ClientID string `validate:"required,lowercase"`
}

type OutDelApp struct {
	Success bool
}
