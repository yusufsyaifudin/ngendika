package pnpsvc

import (
	"context"
	"errors"
	"github.com/yusufsyaifudin/ngendika/backend"
	"github.com/yusufsyaifudin/ngendika/internal/svc/pnprepo"
	"time"
)

var (
	ErrValidation = errors.New("validation error")
)

type Service interface {
	Create(ctx context.Context, in InCreate) (out OutCreate, err error)
	GetByLabels(ctx context.Context, in InGetByLabels) (out OutGetByLabels, err error)
	Examples(ctx context.Context) (out OutExamples)
}

type InCreatePnProvider struct {
	Provider       string      `validate:"required"`
	Label          string      `validate:"required"`
	CredentialJSON interface{} `validate:"required"`
}

type InCreate struct {
	AppID      int64              `validate:"required"`
	PnProvider InCreatePnProvider `validate:"required"`
}

type OutCreate struct {
	ServiceProvider backend.PushNotificationProvider
}

type InGetByLabels struct {
	AppID    int64  `validate:"required"`
	Provider string `validate:"required"`

	// Label comma separated values
	Label string `validate:"required"`
}

type OutGetByLabels struct {
	PnProviders []backend.PushNotificationProvider
}

type OutExamples struct {
	Items []backend.Example
}

// -- func helper

func FromRepo(e pnprepo.PushNotificationProvider) (o backend.PushNotificationProvider) {
	o = backend.PushNotificationProvider{
		ID:             e.ID,
		AppID:          e.AppID,
		Provider:       e.Provider,
		Label:          e.Label,
		CredentialJSON: e.CredentialJSON,
		CreatedAt:      time.UnixMicro(e.CreatedAt).UTC(),
		UpdatedAt:      time.UnixMicro(e.UpdatedAt).UTC(),
	}

	return o
}
