package pnprepo

import (
	"context"
	"errors"
)

var (
	ErrValidation = errors.New("validation error")
)

type Repo interface {
	Insert(ctx context.Context, in InputInsert) (out OutInsert, err error)
	GetByLabels(ctx context.Context, in InGetByLabels) (out OutGetByLabels, err error)
}

// PushNotificationProvider is resembles the table structure.
// We separate the table model into service model (entity that be use for entire app)
// to scoping the database entity to HTTP (request/response) entity.
type PushNotificationProvider struct {
	ID             int64  `db:"id" validate:"required"`
	AppID          int64  `db:"app_id" validate:"required"`
	Provider       string `db:"provider" validate:"required"`
	Label          string `db:"label" validate:"required"`
	CredentialJSON string `db:"credential_json" validate:"required"`

	// Timestamp using integer as unix microsecond in UTC
	CreatedAt int64 `db:"created_at" validate:"required"`
	UpdatedAt int64 `db:"updated_at" validate:"required"`
}

type InputInsert struct {
	PnProvider PushNotificationProvider `validate:"required"`
}

type OutInsert struct {
	PnProvider PushNotificationProvider
}

type InGetByLabels struct {
	AppID    int64    `validate:"required"`
	Provider string   `validate:"required"`
	Labels   []string `validate:"required"`
}

type OutGetByLabels struct {
	PnProvider []PushNotificationProvider
}
