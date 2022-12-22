package msgsvc

import (
	"context"
	"github.com/yusufsyaifudin/ngendika/backend"
	"github.com/yusufsyaifudin/ngendika/internal/svc/appsvc"
)

// Service .
type Service interface {
	// Process only contain mux depend on message type, i.e: fcm will go to fcm service, webhook to webhook service.
	// This may add complexity since msgservice may depend on another service.
	// But, this kind of service layering is what microservice do.
	// Imagine that when you create checkout system, you need to validate user and payment info by calling another service.
	Process(ctx context.Context, input *InputProcess) (out *OutProcess, err error)
}

// InputProcess never be as request response payload!
type InputProcess struct {
	TaskID   string `validate:"required"`
	ClientID string `validate:"required"`
	Label    string `validate:"required"`

	// We can send multiple payload at a time in one providers.
	// For example: {"email" [{"subject": "1", "recipients": ["a"]}, {"subject": "2" "recipients": ["b"]}]}
	Payloads map[string][]interface{} `validate:"required"`
}

type ReportGroup struct {
	PNP            backend.PushNotificationProvider `json:"pnp"`
	BackendErrors  []string                         `json:"backend_errors,omitempty"`
	BackendReports []*backend.Report                `json:"backend_reports,omitempty"`
}

type OutProcess struct {
	TaskID      string
	App         appsvc.App
	Errors      []string
	ReportGroup []ReportGroup
}
