package backend

import (
	"context"
	"fmt"
)

var (
	ErrProviderAlreadyRegistered = fmt.Errorf("provider already registered")
)

// Sender is an interface
type Sender interface {

	// Send is used when we want really send the message using the selected NoopBackend.
	Send(ctx context.Context, workerID int, serviceProvider PushNotificationProvider, msg *Message) (report *Report, err error)

	// ValidateCredJson to convert credential JSON into native Go type.
	ValidateCredJson(_ context.Context, credJson string) (credNative any, err error)

	// ValidateMsg is used when we want to only validate the message, it must not require the credential to operate.
	// It only return the message in Go native type or error.
	ValidateMsg(ctx context.Context, msg *Message) (message any, err error)

	Example(ctx context.Context) (credNative, message any)
}

// SenderMux used by internal application to route to the specific Sender based on provider passed in the params.
type SenderMux interface {

	// Send is used when we want really send the message using the selected NoopBackend.
	Send(ctx context.Context, workerID int, serviceProvider PushNotificationProvider, msg *Message) (report *Report, err error)

	ValidateCredJson(ctx context.Context, provider string, credJson string) (credNative interface{}, err error)

	// ValidateMsg is used when we want to only validate the message, it must not require the credential to operate.
	// It only return the message in Go native type or error.
	ValidateMsg(ctx context.Context, provider string, msg *Message) (message interface{}, err error)

	// Examples will return example of credential JSON and message payload for all providers
	Examples(ctx context.Context) (examples []Example)

	// ListProviders will return all available providers registered in global Backend SenderMux
	ListProviders(ctx context.Context) (providers []string)
}

type Message struct {
	ReferenceID string      `validate:"required"`
	RawPayload  interface{} `validate:"required"`
}

// Report is a struct that hold the report
type Report struct {
	ReferenceID    string `json:"reference_id"`
	WorkerID       int    `json:"worker_id"`
	SuccessCount   int    `json:"success_count"`
	FailureCount   int    `json:"failure_count"`
	NativeResponse any    `json:"native_response"`
}

type Example struct {
	Provider      string `json:"provider"`
	BackendConfig any    `json:"backend_config,omitempty"`
	Message       any    `json:"message,omitempty"`
}
