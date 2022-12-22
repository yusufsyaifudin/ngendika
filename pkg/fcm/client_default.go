package fcm

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/yusufsyaifudin/ngendika/pkg/tracer"
	"go.opentelemetry.io/otel/trace"
	"net/http"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	goFCM "github.com/appleboy/go-fcm"
	"github.com/go-playground/validator/v10"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

type Config struct {
	RoundTripper http.RoundTripper // use shared Round Tripper
}

type ClientDefault struct {
	HTTPClient   *http.Client
	RoundTripper http.RoundTripper
}

// Ensure ClientDefault implements Client
var _ Client = (*ClientDefault)(nil)

func NewClient(cfg Config) (*ClientDefault, error) {
	err := validator.New().Struct(cfg)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	client := &ClientDefault{
		HTTPClient: &http.Client{
			Transport: cfg.RoundTripper,
		},
		RoundTripper: cfg.RoundTripper,
	}

	return client, nil
}

// SendMulticast .
// > The return value is a BatchResponse whose responses list corresponds to the order of the input tokens.
// > This is useful when you want to check which tokens resulted in errors.
// > https://stackoverflow.com/a/72117397/5489910
func (c *ClientDefault) SendMulticast(ctx context.Context, key *ServiceAccountKey, in InputSendMulticast) (out OutSendMulticast, err error) {
	var span trace.Span
	_, span = tracer.StartSpan(ctx, "fcm.SendMulticast")
	defer span.End()

	if in.Message == nil {
		return
	}

	if key == nil {
		err = fmt.Errorf("cannot send multicast on nil service account key")
		return
	}

	message := in.Message

	multicastMsg := &messaging.MulticastMessage{
		Tokens:       message.Tokens,
		Data:         message.Data,
		Notification: message.Notification,
		Android:      message.Android,
		Webpush:      message.Webpush,
		APNS:         message.APNS,
	}

	scopes := []string{
		"https://www.googleapis.com/auth/cloud-platform",
	}

	keyBytes, err := json.Marshal(key)
	if err != nil {
		err = fmt.Errorf("service account key cannot be converted to json bytes: %w", err)
		return
	}

	cred, err := google.CredentialsFromJSON(ctx, keyBytes, scopes...)
	if err != nil {
		err = fmt.Errorf("find default cred error: %w", err)
		return
	}

	config := &firebase.Config{
		ProjectID: cred.ProjectID,
	}

	httpTransport := &oauth2.Transport{
		Base:   c.RoundTripper,
		Source: cred.TokenSource,
	}

	// don't reuse client because each client have different token source
	httpClient := &http.Client{
		Transport: httpTransport,
	}

	opt := []option.ClientOption{
		option.WithHTTPClient(httpClient),
	}

	firebaseApp, err := firebase.NewApp(ctx, config, opt...)
	if err != nil {
		err = fmt.Errorf("initiate firebase app client error: %w", err)
		return
	}

	msgClient, err := firebaseApp.Messaging(ctx)
	if err != nil {
		err = fmt.Errorf("initiate fcm messaging client error: %w", err)
		return
	}

	_, err = msgClient.SendMulticastDryRun(ctx, multicastMsg)
	if err != nil {
		err = fmt.Errorf("fcm multicast may contain invalid payload: %w", err)
		return
	}

	result, err := msgClient.SendMulticast(ctx, multicastMsg)
	if err != nil {
		err = fmt.Errorf("fcm client got error: %w", err)
		return
	}

	out, err = HandleFCMBatchResponse(multicastMsg.Tokens, result)
	return
}

func (c *ClientDefault) SendLegacy(ctx context.Context, serverKey string, msg *LegacyMessage) (LegacyResponse, error) {
	if msg == nil {
		return LegacyResponse{}, nil
	}

	client, err := goFCM.NewClient(serverKey,
		goFCM.WithHTTPClient(c.HTTPClient),
	)
	if err != nil {
		return LegacyResponse{}, fmt.Errorf("fcm client error: %w", err)
	}

	notification := LegacyMessageNotification{}
	if msg.Notification != nil {
		notification = *msg.Notification
	}

	message := &goFCM.Message{
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
		Notification: &goFCM.Notification{
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

// HandleFCMBatchResponse convert from FCM lib messaging.BatchResponse into local struct MulticastBatchResponse.
// This map the array result from FCM to the Device Token.
// As per SO Answer https://stackoverflow.com/a/72117397/5489910 refer to
// https://firebase.google.com/docs/cloud-messaging/send-message#send-messages-to-multiple-devices
// > For Firebase Admin SDKs, this operation uses the sendAll() API under the hood, as shown in the examples.
// > The return value is a BatchResponse whose responses list corresponds to the order of the input tokens.
// > This is useful when you want to check which tokens resulted in errors.
//
// So, index 0 in responses is the result for device token index 0 in the request.
func HandleFCMBatchResponse(tokens []string, result *messaging.BatchResponse) (out OutSendMulticast, err error) {
	if result == nil {
		return
	}

	reqTokenLen := len(tokens)
	respLen := len(result.Responses)
	if reqTokenLen != respLen {
		out = OutSendMulticast{
			BatchResponse: &MulticastBatchResponse{
				SuccessCount: result.SuccessCount,
				FailureCount: result.FailureCount,
			},
		}

		err = fmt.Errorf(
			"%w: total token (%d) vs response length (%d)",
			ErrMapDeviceTokenResp, reqTokenLen, respLen,
		)
		return
	}

	totalHandledMsg := result.SuccessCount + result.FailureCount
	if reqTokenLen != totalHandledMsg {
		out = OutSendMulticast{
			BatchResponse: &MulticastBatchResponse{
				SuccessCount: result.SuccessCount,
				FailureCount: result.FailureCount,
			},
		}

		err = fmt.Errorf(
			"%w: total token (%d) vs success + failure count (%d)",
			ErrMapDeviceTokenResp, reqTokenLen, totalHandledMsg,
		)
		return
	}

	sendResp := make([]MulticastSendResponse, 0)
	for i, sendRes := range result.Responses {
		if sendRes == nil {
			continue
		}

		errStr := ""
		if sendRes.Error != nil {
			errStr = sendRes.Error.Error()
		}

		sendResp = append(sendResp, MulticastSendResponse{
			DeviceToken: tokens[i],
			Success:     sendRes.Success,
			MessageID:   sendRes.MessageID,
			Error:       errStr,
		})
	}

	// response batch for each fcm service account
	out = OutSendMulticast{
		BatchResponse: &MulticastBatchResponse{
			SuccessCount: result.SuccessCount,
			FailureCount: result.FailureCount,
			Responses:    sendResp,
		},
	}

	return
}
