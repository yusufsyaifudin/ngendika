package msgservice

import (
	"net/http"
	"net/url"

	"github.com/yusufsyaifudin/ngendika/pkg/fcm"
	"github.com/yusufsyaifudin/ngendika/pkg/logger"
)

// Task is a common payload to be use for worker to handle message based on type.
type Task struct {
	TraceInfo logger.Tracer `json:"trace_info" validate:"required"`
	TaskID    string        `json:"task_id" validate:"required"`

	AppClientID             string                   `json:"app_client_id" validate:"required"`
	TaskPayloadFCMMulticast *TaskPayloadFCMMulticast `json:"task_payload_fcm_multicast,omitempty" validate:"required_if=TaskType fcm_multicast"`
	TaskPayloadFCMLegacy    *TaskPayloadFCMLegacy    `json:"task_payload_fcm_legacy,omitempty" validate:"required_if=TaskType fcm_legacy"`
	TaskPayloadWebhook      []TaskPayloadWebhook     `json:"task_payload_webhook,omitempty" validate:"dive,required_if=TaskType webhook"`
}

// TaskPayloadFCMMulticast is FCM message to send with mode multicast
type TaskPayloadFCMMulticast struct {
	Msg *fcm.MulticastMessage `json:"msg" validate:"required"`
}

// TaskPayloadFCMLegacy is FCM message to send with mode legacy API
type TaskPayloadFCMLegacy struct {
	Msg *fcm.LegacyMessage `json:"msg" validate:"required"`
}

// TaskPayloadWebhook is a task to call external service.
// The task is only intended for simple webhook, not support multipart form for upload files.
// PostForm should be ok.
type TaskPayloadWebhook struct {
	URL        string      `json:"url" validate:"required"`
	Method     string      `json:"method" validate:"required,oneof=GET POST PUT PATCH DELETE"`
	Header     http.Header `json:"header"`
	Body       string      `json:"body"`
	QueryParam url.Values  `json:"query_param"`
	FormData   url.Values  `json:"form_data"`
}

type TaskResult struct {
	TaskID            string              `json:"task_id"`
	AppClientID       string              `json:"app_client_id" `
	FCMMulticast      *FCMMulticastOutput `json:"fcm_multicast,omitempty"`
	FCMMulticastError string              `json:"fcm_multicast_error,omitempty"`
	FCMLegacy         *FCMLegacyOutput    `json:"fcm_legacy,omitempty"`
	FCMLegacyError    string              `json:"fcm_legacy_error,omitempty"`
	Webhook           *WebhookOutput      `json:"webhook,omitempty"`
	WebhookError      string              `json:"webhook_error,omitempty"`
}
