package msgservice

import (
	"net/http"

	"github.com/yusufsyaifudin/ngendika/pkg/fcm"
)

// ------- I/O for FCM Multicast Message

type FcmMulticastInput struct {
	AppID                   string                   `validate:"required"`
	TaskPayloadFCMMulticast *TaskPayloadFCMMulticast `validate:"required"`
}

// FCMMulticastResult output data for FCMMulticastOutput
type FCMMulticastResult struct {
	FCMKeyID    string                    `json:"fcm_key_id"`
	Error       string                    `json:"error,omitempty"`
	BatchResult *fcm.MulticastBatchResult `json:"batch_result,omitempty"`
}

// FCMMulticastOutput output param for SendFCMMessageMulticast
type FCMMulticastOutput struct {
	Result []FCMMulticastResult `json:"result,omitempty"`
}

// ------- I/O for Legacy FCM

// FcmLegacyInput input param for sending FCM Legacy message
type FcmLegacyInput struct {
	AppID                string                `validate:"required"`
	TaskPayloadFCMLegacy *TaskPayloadFCMLegacy `validate:"required"`
}

// FCMLegacyResult output data for FCMMulticastOutput
type FCMLegacyResult struct {
	FCMKeyID    string              `json:"fcm_key_id"`
	Error       string              `json:"error,omitempty"`
	BatchResult *fcm.LegacyResponse `json:"batch_result,omitempty"`
}

// FCMLegacyOutput output param for SendFCMMessageLegacy
type FCMLegacyOutput struct {
	Result []FCMLegacyResult `json:"result,omitempty"`
}

// ------- I/O for Webhook

// WebhookInput input webhook
type WebhookInput struct {
	AppClientID string               `validate:"required"`
	Webhook     []TaskPayloadWebhook `validate:"required,max=10"`
}

// WebhookOutput .
type WebhookOutput struct {
	Result []WebhookResult `json:"result,omitempty"`
}

// WebhookResult result of webhook
type WebhookResult struct {
	Header http.Header   `json:"header,omitempty"`
	Body   string        `json:"body,omitempty"`
	Error  *WebhookError `json:"error,omitempty"`
}
