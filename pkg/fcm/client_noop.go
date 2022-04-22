package fcm

import "context"

type Noop struct{}

var _ Client = (*Noop)(nil)

func NewNoop() *Noop {
	return &Noop{}
}

func (n *Noop) SendMulticast(ctx context.Context, key []byte, message *MulticastMessage) (MulticastBatchResult, error) {
	return MulticastBatchResult{}, nil
}

func (n *Noop) SendLegacy(ctx context.Context, serverKey string, message *LegacyMessage) (LegacyResponse, error) {
	return LegacyResponse{}, nil
}
