package msgservice

import (
	"net/http"
	"net/url"

	"github.com/yusufsyaifudin/ngendika/internal/logic/fcmservice"

	"github.com/yusufsyaifudin/ylog"

	"github.com/yusufsyaifudin/ngendika/pkg/fcm"
)

// Task is a common payload to be use for worker to handle message based on type.
type Task struct {
	TraceInfo *ylog.Tracer `json:"trace_info" validate:"required"`
	TaskID    string       `json:"task_id" validate:"required"`
	ClientID  string       `json:"client_id" validate:"required"`
	Message   *Message     `json:"message" validate:"required"`
}

type Message struct {
	FCMMulticast *fcm.MulticastMessage `json:"fcm_multicast" validate:"-"`
	FCMLegacy    *fcm.LegacyMessage    `json:"fcm_legacy" validate:"-"`
	Webhook      []TaskPayloadWebhook  `json:"webhook" validate:"-"`
}

// TaskPayloadWebhook is a task to call external service.
// The task is only intended for simple webhook, not support multipart form for upload files.
// PostForm should be ok.
type TaskPayloadWebhook struct {
	ReferenceID string      `json:"reference_id" validate:"required"` // Reference ID from user to know which request is success/fail
	URL         string      `json:"url" validate:"required"`
	Method      string      `json:"method" validate:"required,oneof=GET POST PUT PATCH DELETE"`
	Header      http.Header `json:"header"`
	Body        string      `json:"body"`
	QueryParam  url.Values  `json:"query_param"`
	FormData    url.Values  `json:"form_data"`
}

type TaskResult struct {
	TaskID            string                         `json:"task_id"`
	AppClientID       string                         `json:"app_client_id" `
	FCMMulticast      *fcmservice.FCMMulticastOutput `json:"fcm_multicast,omitempty"`
	FCMMulticastError string                         `json:"fcm_multicast_error,omitempty"`
	FCMLegacy         *fcmservice.FCMLegacyOutput    `json:"fcm_legacy,omitempty"`
	FCMLegacyError    string                         `json:"fcm_legacy_error,omitempty"`
	Webhook           *WebhookOutput                 `json:"webhook,omitempty"`
	WebhookError      string                         `json:"webhook_error,omitempty"`
}
