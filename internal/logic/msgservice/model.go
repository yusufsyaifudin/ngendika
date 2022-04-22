package msgservice

import (
	"net/http"
)

// ------- I/O for Webhook

// WebhookInput input webhook
type WebhookInput struct {
	AppClientID string               `validate:"required"`
	Webhook     []TaskPayloadWebhook `validate:"omitempty,max=10"`
}

// WebhookOutput .
type WebhookOutput struct {
	Result []WebhookResult `json:"result,omitempty"`
}

// WebhookResult result of webhook
type WebhookResult struct {
	ReferenceID string        `json:"reference_id"`
	Header      http.Header   `json:"header,omitempty"`
	Body        string        `json:"body,omitempty"`
	Error       *WebhookError `json:"error,omitempty"`
}
