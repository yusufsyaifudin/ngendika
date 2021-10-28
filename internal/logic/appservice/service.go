package appservice

import (
	"context"
)

// Service is an interface of final business logic.
// Any input and output from/to this function should be SAFE for external party to consume,
// i.e: request or response from HTTP handler
type Service interface {
	CreateApp(ctx context.Context, input CreateAppIn) (out CreateAppOut, err error)

	CreateFcmSvcAccKey(ctx context.Context, input CreateFcmSvcAccKeyIn) (out CreateFcmSvcAccKeyOut, err error)
	GetFcmSvcAccKey(ctx context.Context, input GetFcmSvcAccKeyIn) (out GetFcmSvcAccKeyOut, err error)
}
