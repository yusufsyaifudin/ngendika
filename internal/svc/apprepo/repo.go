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
	Create(ctx context.Context, in InputCreate) (out OutCreate, err error)
	Upsert(ctx context.Context, in InputUpsert) (out OutUpsert, err error)
	GetByClientID(ctx context.Context, in InputGetByClientID) (out OutGetByClientID, err error)
	List(ctx context.Context, in InputList) (out OutList, err error)
	DelByClientID(ctx context.Context, in InputDelByClientID) (out OutDelByClientID, err error)
}

type InputCreate struct {
	App App `validate:"required"`
}

type OutCreate struct {
	App App
}

type InputUpsert struct {
	App App `validate:"required"`
}

type OutUpsert struct {
	App App
}

type InputGetByClientID struct {
	ClientID string `validate:"required,lowercase"`
	Enabled  *bool  `validate:"-"`
}

type OutGetByClientID struct {
	App App
}

type InputList struct {
	Limit    int64 `validate:"required"`
	BeforeID int64 `validate:"min=0"`
	AfterID  int64 `validate:"min=0"`
}

type OutList struct {
	Total int64
	Apps  []App
}

type InputDelByClientID struct {
	ClientID  string `validate:"required,lowercase"`
	DeletedAt int64  `validate:"required"`
}

type OutDelByClientID struct {
	Success bool
}
