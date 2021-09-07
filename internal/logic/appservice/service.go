package appservice

import (
	"context"
)

// Service is an interface of final business logic.
// Any input and output from/to this function should be SAFE for external party to consume,
// i.e: request or response from HTTP handler
type Service interface {
	CreateApp(ctx context.Context, input InputCreateApp) (out OutputCreateApp, err error)
	CreateFcmServiceAccountKey(ctx context.Context, input InputCreateFcmServiceAccountKey) (out OutputCreateFcmServiceAccountKey, err error)
}
