package fcm

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/appleboy/go-fcm"
	"google.golang.org/api/option"
)

type ClientDefault struct{}

// Ensure ClientDefault implements Client
var _ Client = (*ClientDefault)(nil)

func NewClient() (*ClientDefault, error) {
	return &ClientDefault{}, nil
}

func (c *ClientDefault) SendMulticast(ctx context.Context, key []byte, message *MulticastMessage) (MulticastBatchResult, error) {
	if message == nil {
		return MulticastBatchResult{}, nil
	}

	multicastMsg := &messaging.MulticastMessage{
		Tokens:       message.Tokens,
		Data:         message.Data,
		Notification: message.Notification,
		Android:      message.Android,
		Webpush:      message.Webpush,
		APNS:         message.APNS,
	}

	opt := option.WithCredentialsJSON(key)
	firebaseApp, err := firebase.NewApp(ctx, &firebase.Config{}, opt)
	if err != nil {
		return MulticastBatchResult{}, fmt.Errorf("initiate firebase app client error: %w", err)
	}

	msgClient, err := firebaseApp.Messaging(ctx)
	if err != nil {
		return MulticastBatchResult{}, fmt.Errorf("initiate fcm messaging client error: %w", err)
	}

	result, err := msgClient.SendMulticast(ctx, multicastMsg)
	return HandleFCMBatchResponse(result), err
}

func (c *ClientDefault) SendLegacy(ctx context.Context, serverKey string, msg *LegacyMessage) (LegacyResponse, error) {
	if msg == nil {
		return LegacyResponse{}, nil
	}

	client, err := fcm.NewClient(serverKey)
	if err != nil {
		return LegacyResponse{}, fmt.Errorf("fcm client error: %w", err)
	}

	notification := LegacyMessageNotification{}
	if msg.Notification != nil {
		notification = *msg.Notification
	}

	message := &fcm.Message{
		To:                       msg.To,
		RegistrationIDs:          msg.RegistrationIDs,
		Condition:                msg.Condition,
		CollapseKey:              msg.CollapseKey,
		Priority:                 msg.Priority,
		ContentAvailable:         msg.ContentAvailable,
		MutableContent:           msg.MutableContent,
		DelayWhileIdle:           msg.DelayWhileIdle,
		TimeToLive:               msg.TimeToLive,
		DeliveryReceiptRequested: msg.DeliveryReceiptRequested,
		DryRun:                   msg.DryRun,
		RestrictedPackageName:    msg.RestrictedPackageName,
		Notification: &fcm.Notification{
			Title:        notification.Title,
			Body:         notification.Body,
			ChannelID:    notification.ChannelID,
			Icon:         notification.Icon,
			Image:        notification.Image,
			Sound:        notification.Sound,
			Badge:        notification.Badge,
			Tag:          notification.Tag,
			Color:        notification.Color,
			ClickAction:  notification.ClickAction,
			BodyLocKey:   notification.BodyLocKey,
			BodyLocArgs:  notification.BodyLocArgs,
			TitleLocKey:  notification.TitleLocKey,
			TitleLocArgs: notification.TitleLocArgs,
		},
		Data:    msg.Data,
		Apns:    msg.Apns,
		Webpush: msg.Webpush,
	}

	resp, err := client.SendWithContext(ctx, message)
	if err != nil {
		return LegacyResponse{}, fmt.Errorf("response fcm error: %w", err)
	}

	results := make([]LegacyResponseResult, 0)
	for _, res := range resp.Results {
		results = append(results, LegacyResponseResult{
			MessageID:      res.MessageID,
			RegistrationID: res.RegistrationID,
			Error:          res.Error,
		})
	}

	return LegacyResponse{
		MulticastID:           resp.MulticastID,
		Success:               resp.Success,
		Failure:               resp.Failure,
		CanonicalIDs:          resp.CanonicalIDs,
		Results:               results,
		FailedRegistrationIDs: resp.FailedRegistrationIDs,
		MessageID:             resp.MessageID,
		Error:                 resp.Error,
	}, nil
}

// HandleFCMBatchResponse convert from FCM lib messaging.BatchResponse into local struct MulticastBatchResult.
func HandleFCMBatchResponse(result *messaging.BatchResponse) MulticastBatchResult {
	if result == nil {
		return MulticastBatchResult{}
	}

	sendResp := make([]MulticastSendResponse, 0)
	for _, sendRes := range result.Responses {
		if sendRes == nil {
			continue
		}

		sendResp = append(sendResp, MulticastSendResponse{
			Success:   sendRes.Success,
			MessageID: sendRes.MessageID,
			Error:     sendRes.Error,
		})
	}

	// response batch for each fcm service account
	return MulticastBatchResult{
		BatchResponse: &MulticastBatchResponse{
			SuccessCount: result.SuccessCount,
			FailureCount: result.FailureCount,
			Responses:    sendResp,
		},
	}
}
