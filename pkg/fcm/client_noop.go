package fcm

import "context"

type Noop struct{}

func (n Noop) SendMulticast(ctx context.Context, key *ServiceAccountKey, in InputSendMulticast) (out OutSendMulticast, err error) {
	return
}

func (n Noop) SendLegacy(ctx context.Context, serverKey string, message *LegacyMessage) (LegacyResponse, error) {
	return LegacyResponse{}, nil
}

var _ Client = (*Noop)(nil)

func NewNoop() *Noop {
	return &Noop{}
}
