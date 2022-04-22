package msgservice

import (
	"context"
)

// Process only contain mux depend on message type, i.e: fcm will go to fcm service, webhook to webhook service.
// This may add complexity since msgservice may depend on another service.
// But, this kind of service layering is what microservice do.
// Imagine that when you create checkout system, you need to validate user and payment info by calling another service.
type Process func(ctx context.Context, task *Task) (out *TaskResult, err error)

type Service interface {
	Process() Process
}
