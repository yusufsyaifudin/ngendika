package fcm

import (
	"context"

	"firebase.google.com/go/v4/messaging"
)

type Client interface {
	SendMulticast(ctx context.Context, key []byte, message *MulticastMessage) (MulticastBatchResult, error)
	SendLegacy(ctx context.Context, serverKey string, message *LegacyMessage) (LegacyResponse, error)
}

// ServiceAccountKey represent service account key json
type ServiceAccountKey struct {
	Type                    string `json:"type" validate:"required"`
	ProjectID               string `json:"project_id" validate:"required"`
	PrivateKeyID            string `json:"private_key_id" validate:"required"`
	PrivateKey              string `json:"private_key" validate:"required"`
	ClientEmail             string `json:"client_email" validate:"required"`
	ClientID                string `json:"client_id" validate:"required"`
	AuthURI                 string `json:"auth_uri" validate:"required"`
	TokenURI                string `json:"token_uri" validate:"required"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url validate:"required""`
	ClientX509CertURL       string `json:"client_x509_cert_url" validate:"required"`
}

// MulticastMessage is like https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages
// but instead single token, it use array of tokens.
// Since target to send a message to is must be one of: token, topic or condition,
// hence you cannot set topic or condition in multicast message.
type MulticastMessage struct {
	Tokens       []string                 `json:"tokens,omitempty"`
	Data         map[string]string        `json:"data,omitempty"`
	Notification *messaging.Notification  `json:"notification,omitempty"`
	Android      *messaging.AndroidConfig `json:"android,omitempty"`
	Webpush      *messaging.WebpushConfig `json:"webpush,omitempty"`
	APNS         *messaging.APNSConfig    `json:"apns,omitempty"`
}

// MulticastSendResponse represents the status of an individual message that was sent as part of a batch request.
type MulticastSendResponse struct {
	Success   bool   `json:"success,omitempty"`
	MessageID string `json:"message_id,omitempty"`
	Error     error  `json:"error,omitempty"`
}

// MulticastBatchResponse represents the response from the FCM API.
// This copies from FCM lib to make more control about json tag and type data: MulticastSendResponse should not slice of pointer
// instead of slice of value.
type MulticastBatchResponse struct {
	SuccessCount int                     `json:"success_count"`
	FailureCount int                     `json:"failure_count"`
	Responses    []MulticastSendResponse `json:"responses,omitempty"`
}

type MulticastBatchResult struct {
	BatchResponse *MulticastBatchResponse `json:"batch_response,omitempty"`
}
