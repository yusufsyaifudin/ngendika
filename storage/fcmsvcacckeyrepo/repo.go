package fcmsvcacckeyrepo

import (
	"context"
)

// Repo .
type Repo interface {
	CreateFCMServiceAccountKey(ctx context.Context, cert FCMServiceAccountKey) (inserted FCMServiceAccountKey, err error)
	GetFCMSvcAccKeys(ctx context.Context, appID string) (fcmServiceAccountKeys []FCMServiceAccountKey, err error)
}
