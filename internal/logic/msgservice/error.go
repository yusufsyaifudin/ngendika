package msgservice

import (
	"fmt"
)

type WebhookErrorCode string

const (
	ErrWebhookRequest WebhookErrorCode = "ErrWebhookRequest"
)

// WebhookError structure of webhook error
type WebhookError struct {
	Code          WebhookErrorCode `json:"code"`
	MessageDetail string           `json:"message_detail"`
}

func (we *WebhookError) Error() string {
	return fmt.Sprintf("%s: %s", we.Code, we.MessageDetail)
}
