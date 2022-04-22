package fcmservice

import (
	"context"
	"time"

	"github.com/yusufsyaifudin/ngendika/pkg/fcm"
)

type Service interface {
	CreateSvcAccKey(ctx context.Context, input CreateSvcAccKeyIn) (out CreateSvcAccKeyOut, err error)
	GetSvcAccKey(ctx context.Context, input GetSvcAccKeyIn) (out GetSvcAccKeyOut, err error)
	GetServerKey(ctx context.Context, input GetServerKeyIn) (out GetServerKeyOut, err error)
	FcmMulticast(ctx context.Context, input *FcmMulticastInput) (out *FCMMulticastOutput, err error)
	FcmLegacy(ctx context.Context, input *FcmLegacyInput) (out *FCMLegacyOutput, err error)
}

type SvcAccKeyMetadata struct {
	Filename string `validate:"required"`
}

type CreateSvcAccKeyIn struct {
	ClientID          string            `validate:"required"` // App Client ID
	ServiceAccountKey []byte            `validate:"required"` // in JSON byte format
	Metadata          SvcAccKeyMetadata `validate:"required"`
}

type CreateSvcAccKeyOut struct {
	ID                string
	AppID             string
	ServiceAccountKey []byte // in JSON stringify format
	CreatedAt         time.Time
}

type GetSvcAccKeyIn struct {
	ClientID string `validate:"required"` // App Client ID
}

type GetSvcAccKeyOutList struct {
	ID                string
	AppID             string
	ServiceAccountKey []byte // in JSON stringify format
	CreatedAt         time.Time
}

type GetSvcAccKeyOut struct {
	Lists []GetSvcAccKeyOutList
}

type GetServerKeyIn struct {
	ClientID string `validate:"required"` // App Client ID
}

type GetServerKeyOutList struct {
	ID        string
	AppID     string
	ServerKey string
	CreatedAt time.Time
}

type GetServerKeyOut struct {
	Lists []GetServerKeyOutList
}

// ------- I/O for FCM Multicast Message

type FcmMulticastInput struct {
	AppClientID string                `validate:"required"`
	Payload     *fcm.MulticastMessage `validate:"omitempty"`
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
	AppClientID string             `validate:"required"`
	Payload     *fcm.LegacyMessage `validate:"omitempty"`
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
