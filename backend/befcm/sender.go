package befcm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/yusufsyaifudin/ngendika/backend"
	"github.com/yusufsyaifudin/ngendika/pkg/fcm"
	"github.com/yusufsyaifudin/ngendika/pkg/tracer"
	"github.com/yusufsyaifudin/ngendika/pkg/validator"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

type Backend struct {
	Client fcm.Client
}

var _ backend.Sender = (*Backend)(nil)

func NewBE(httpRoundTripper http.RoundTripper) (*Backend, error) {
	if httpRoundTripper == nil {
		httpRoundTripper = http.DefaultTransport
	}

	fcmClientCfg := fcm.Config{
		RoundTripper: httpRoundTripper,
	}

	fcmClient, err := fcm.NewClient(fcmClientCfg)
	if err != nil {
		err = fmt.Errorf("fcm client failed: %w", err)
		return nil, err
	}

	be := &Backend{
		Client: fcmClient,
	}

	return be, nil
}

func (b *Backend) Send(ctx context.Context, workerID int, serviceProvider backend.PushNotificationProvider, msg *backend.Message) (report *backend.Report, err error) {

	var span trace.Span
	ctx, span = tracer.StartSpan(ctx, "befcm.Send")
	defer span.End()

	message, err := b.ValidateMsg(ctx, msg)
	if err != nil {
		return
	}

	fcmMsg, ok := message.(fcm.MulticastMessage)
	if !ok {
		err = fmt.Errorf("invalid fcm multicast message, got type '%T'", message)
		return
	}

	cred, err := b.ValidateCredJson(ctx, serviceProvider.CredentialJSON)
	if err != nil {
		return
	}

	fcmCred, ok := cred.(fcm.ServiceAccountKey)
	if !ok {
		err = fmt.Errorf("invalid fcm credential json, got type '%T'", message)
		return
	}

	out, err := b.Client.SendMulticast(ctx, &fcmCred, fcm.InputSendMulticast{Message: &fcmMsg})
	if err != nil {
		err = fmt.Errorf("failed send fcm multicast message: %w", err)
		return
	}

	if out.BatchResponse == nil {
		err = fmt.Errorf("cannot process output on nil response: %w", err)
		return
	}

	report = &backend.Report{
		ReferenceID:    msg.ReferenceID,
		WorkerID:       workerID,
		SuccessCount:   out.BatchResponse.SuccessCount,
		FailureCount:   out.BatchResponse.FailureCount,
		NativeResponse: out,
	}

	return
}

func (b *Backend) ValidateCredJson(ctx context.Context, credJson string) (credNative any, err error) {
	var span trace.Span
	ctx, span = tracer.StartSpan(ctx, "befcm.ValidateCredJson")
	defer span.End()

	var cred fcm.ServiceAccountKey
	dec := json.NewDecoder(bytes.NewBufferString(credJson))
	dec.DisallowUnknownFields()
	err = dec.Decode(&cred)
	if err != nil {
		err = fmt.Errorf("fcm config malformed: %w", err)
		return
	}

	err = validator.Validate(cred)
	if err != nil {
		err = fmt.Errorf("fcm config missing fields: %w", err)
		return
	}

	credNative = cred
	return
}

func (b *Backend) ValidateMsg(ctx context.Context, msg *backend.Message) (message any, err error) {
	var span trace.Span
	_, span = tracer.StartSpan(ctx, "befcm.ValidateMsg")
	defer span.End()

	err = validator.Validate(msg)
	if err != nil {
		err = fmt.Errorf("cannot validate the message: %w", err)
		return
	}

	var fcmMsg fcm.MulticastMessage

	// Convert from Go native type to json string
	dataByte, err := json.Marshal(msg.RawPayload)
	if err != nil {
		err = fmt.Errorf("we assume you input payload as json valid object, but it failed to marshal: %w", err)
		return
	}

	// from JSON string, turn back to Go native type but with explicit type: such as fcm.MulticastMessage etc...
	err = json.Unmarshal(dataByte, &fcmMsg)
	if err != nil {
		err = fmt.Errorf("malformed fcm multicast payload: %w", err)
		return
	}

	message = fcmMsg
	return
}

func (b *Backend) Example(_ context.Context) (credNative, message any) {
	credNative = fcm.ServiceAccountKey{
		Type:                    "",
		ProjectID:               "",
		PrivateKeyID:            "",
		PrivateKey:              "",
		ClientEmail:             "",
		ClientID:                "",
		AuthURI:                 "",
		TokenURI:                "",
		AuthProviderX509CertURL: "",
		ClientX509CertURL:       "",
	}

	message = fcm.MulticastMessage{
		Tokens:       nil,
		Data:         nil,
		Notification: nil,
		Android:      nil,
		Webpush:      nil,
		APNS:         nil,
	}

	return
}
