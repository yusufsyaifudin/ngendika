package backend

import (
	"bytes"
	"context"
	"fmt"
	"github.com/segmentio/encoding/json"
)

type NoopBackend struct{}

var _ Sender = (*NoopBackend)(nil)

func NewNoopSender() *NoopBackend {
	be := &NoopBackend{}

	return be
}

func (b *NoopBackend) Send(ctx context.Context, workerID int, serviceProvider PushNotificationProvider, msg *Message) (report *Report, err error) {
	_, err = b.ValidateMsg(ctx, msg)
	if err != nil {
		return
	}

	_, err = b.ValidateCredJson(ctx, serviceProvider.CredentialJSON)
	if err != nil {
		return
	}

	report = &Report{
		ReferenceID:    msg.ReferenceID,
		WorkerID:       workerID,
		SuccessCount:   1,
		FailureCount:   0,
		NativeResponse: nil,
	}

	return
}

func (b *NoopBackend) ValidateCredJson(_ context.Context, credJson string) (credNative any, err error) {
	dec := json.NewDecoder(bytes.NewBufferString(credJson))
	dec.DisallowUnknownFields()
	err = dec.Decode(&credNative)
	if err != nil {
		err = fmt.Errorf("noop config malformed: %w", err)
		return
	}
	return
}

func (b *NoopBackend) ValidateMsg(ctx context.Context, msg *Message) (message any, err error) {
	dataByte, err := json.Marshal(msg.RawPayload)
	if err != nil {
		err = fmt.Errorf("we assume you input payload as json valid object, but it failed to marshal: %w", err)
		return
	}

	// from JSON string, turn back to Go native type but with explicit type: such as fcm.MulticastMessage etc...
	err = json.Unmarshal(dataByte, &message)
	if err != nil {
		err = fmt.Errorf("malformed noop payload: %w", err)
		return
	}

	return
}

func (b *NoopBackend) Example(ctx context.Context) (credNative, message any) {
	credNative = struct {
		UpperCase bool
	}{
		UpperCase: true,
	}

	message = struct {
		Message string
	}{
		Message: "message to upper case",
	}

	return
}
