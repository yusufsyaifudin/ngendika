package fcmservice

import (
	"context"
	"time"
)

type Service interface {
	CreateSvcAccKey(ctx context.Context, input CreateSvcAccKeyIn) (out CreateSvcAccKeyOut, err error)
	GetSvcAccKey(ctx context.Context, input GetSvcAccKeyIn) (out GetSvcAccKeyOut, err error)
	GetServerKey(ctx context.Context, input GetServerKeyIn) (out GetServerKeyOut, err error)
}

type SvcAccKeyMetadata struct {
	Filename string `validate:"required"`
}

type CreateSvcAccKeyIn struct {
	ClientID          string            `validate:"required"` // App Client ID
	ServiceAccountKey []byte            `validate:"required"` // in JSON byte format
	Metadata          SvcAccKeyMetadata `validate:"required"`
}

type CreateSvcAccKeyOut struct {
	ID                string
	AppID             string
	ServiceAccountKey []byte // in JSON stringify format
	CreatedAt         time.Time
}

type GetSvcAccKeyIn struct {
	ClientID string `validate:"required"` // App Client ID
}

type GetSvcAccKeyOutList struct {
	ID                string
	AppID             string
	ServiceAccountKey []byte // in JSON stringify format
	CreatedAt         time.Time
}

type GetSvcAccKeyOut struct {
	Lists []GetSvcAccKeyOutList
}

type GetServerKeyIn struct {
	ClientID string `validate:"required"` // App Client ID
}

type GetServerKeyOutList struct {
	ID        string
	AppID     string
	ServerKey string
	CreatedAt time.Time
}

type GetServerKeyOut struct {
	Lists []GetServerKeyOutList
}
