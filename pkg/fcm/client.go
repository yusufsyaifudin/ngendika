package fcm

import (
	"context"
	"encoding/json"
	"fmt"

	"firebase.google.com/go/v4/messaging"
)

var (
	ErrMapDeviceTokenResp = fmt.Errorf("cannot map response to device token")
)

type Client interface {
	SendMulticast(ctx context.Context, key *ServiceAccountKey, in InputSendMulticast) (out OutSendMulticast, err error)
	SendLegacy(ctx context.Context, serverKey string, message *LegacyMessage) (LegacyResponse, error)
}

// MulticastMessage represents a message that can be sent to multiple devices via Firebase Cloud
// Messaging (FCM).
//
// It contains payload information as well as the list of device registration tokens to which the
// message should be sent. A single MulticastMessage may contain up to 500 registration tokens.
//
// We copy this because internal MulticastMessage doesn't come with json tag.
// We need json tag to parse to json string when sending to Kafka or message broker
type MulticastMessage struct {
	Tokens       []string                 `json:"tokens,omitempty"`
	Data         map[string]string        `json:"data,omitempty"`
	Notification *messaging.Notification  `json:"notification,omitempty"`
	Android      *messaging.AndroidConfig `json:"android,omitempty"`
	Webpush      *messaging.WebpushConfig `json:"webpush,omitempty"`
	APNS         *messaging.APNSConfig    `json:"apns,omitempty"`
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
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url" validate:"required"`
	ClientX509CertURL       string `json:"client_x509_cert_url" validate:"required"`
}

// Scan will read the data bytes from database and parse it as ServiceAccountKey
func (m *ServiceAccountKey) Scan(src interface{}) error {
	if m == nil {
		return fmt.Errorf("error scan service account on nil struct")
	}

	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, m)
	case string:
		return json.Unmarshal([]byte(fmt.Sprintf("%s", v)), m)
	}

	return fmt.Errorf("unknown type %T to format as service account key", src)
}

type InputSendMulticast struct {
	Message *MulticastMessage `validate:"required"`
}

type OutSendMulticast struct {
	BatchResponse *MulticastBatchResponse `json:"batch_response,omitempty"`
}

// MulticastSendResponse represents the status of an individual message that was sent as part of a batch request.
type MulticastSendResponse struct {
	DeviceToken string `json:"device_token"`
	Success     bool   `json:"success"`
	MessageID   string `json:"message_id,omitempty"`
	Error       string `json:"error,omitempty"`
}

// MulticastBatchResponse represents the response from the FCM API.
// This copies from FCM lib to make more control about json tag and type data.
// MulticastSendResponse should not slice of pointer
// instead of slice of value.
type MulticastBatchResponse struct {
	SuccessCount int                     `json:"success_count"`
	FailureCount int                     `json:"failure_count"`
	Responses    []MulticastSendResponse `json:"responses,omitempty"` // Pair of device token and response
}
