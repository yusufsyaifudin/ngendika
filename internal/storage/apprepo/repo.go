package apprepo

import (
	"context"
	"errors"
)

var (
	ErrValidation = errors.New("validation error")
)

// Repo is App repository service
type Repo interface {
	CreateApp(ctx context.Context, in InputCreateApp) (out OutCreateApp, err error)
	GetApp(ctx context.Context, in InputGetApp) (out OutGetApp, err error)
}

type InputCreateApp struct {
	App App `validate:"required"`
}

type OutCreateApp struct {
	App App
}

type InputGetApp struct {
	ClientID string `validate:"required,lowercase"`
	Enabled  *bool  `validate:"-"`
}

type OutGetApp struct {
	App App
}
