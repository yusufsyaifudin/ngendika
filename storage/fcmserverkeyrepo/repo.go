package fcmserverkeyrepo

import (
	"context"
)

// Repo .
type Repo interface {
	CreateFCMServerKey(ctx context.Context, param FCMServerKey) (inserted FCMServerKey, err error)
	GetFCMServerKeys(ctx context.Context, appID string) (serverKeys []FCMServerKey, err error)
}
