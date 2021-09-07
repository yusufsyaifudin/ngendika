package apprepo

import (
	"context"
)

// Repo is App repository service
type Repo interface {
	GetAppByClientID(ctx context.Context, clientID string) (app App, err error)
	CreateApp(ctx context.Context, app App) (inserted App, err error)

	CreateFCMServiceAccountKey(ctx context.Context, cert FCMServiceAccountKey) (inserted FCMServiceAccountKey, err error)
	GetFCMServiceAccountKeys(ctx context.Context, clientID string) (fcmServiceAccountKeys []FCMServiceAccountKey, err error)

	CreateFCMServerKey(ctx context.Context, param FCMServerKey) (inserted FCMServerKey, err error)
	GetFCMServerKeys(ctx context.Context, clientID string) (serverKeys []FCMServerKey, err error)
}
