package backend

import (
	"context"
	"fmt"
	"github.com/yusufsyaifudin/ngendika/pkg/tracer"
	"go.opentelemetry.io/otel/trace"
	"strings"
	"sync"
	"unicode/utf8"
)

type SenderMultiplexer struct {
	lock   sync.RWMutex
	sender map[string]Sender
}

var beMux = &SenderMultiplexer{
	sender: map[string]Sender{},
}

func MuxBackend() SenderMux {
	return beMux
}

func MustRegister(provider string, sender Sender) {
	err := Register(provider, sender)
	if err != nil {
		panic(err)
	}
}

// Register new provider with the implemented Sender
func Register(provider string, sender Sender) (err error) {
	provider = strings.TrimSpace(provider)
	if provider == "" {
		err = fmt.Errorf("cannot assign empty provider name")
		return
	}

	if provider != strings.ToLower(provider) {
		err = fmt.Errorf("provider name must only contain lower case")
		return
	}

	if !utf8.ValidString(provider) {
		err = fmt.Errorf("provider name must only use utf8 characters")
		return
	}

	if sender == nil {
		err = fmt.Errorf("cannot assign nil sender")
		return
	}

	beMux.lock.Lock()
	defer beMux.lock.Unlock()

	if _, exist := beMux.sender[provider]; exist {
		err = fmt.Errorf("%w '%s'", ErrProviderAlreadyRegistered, provider)
		return
	}

	beMux.sender[provider] = sender
	return
}

func (s *SenderMultiplexer) Send(ctx context.Context, workerID int, serviceProvider PushNotificationProvider, msg *Message) (report *Report, err error) {
	var span trace.Span
	ctx, span = tracer.StartSpan(ctx, "backendmux.Send")
	defer span.End()

	if msg == nil {
		err = fmt.Errorf("passed message is nil, we cannot process that")
		return
	}

	beMux.lock.RLock()
	defer beMux.lock.RUnlock()

	// select the appropriate registered client based on provider key
	client, exist := s.sender[serviceProvider.Provider]
	if !exist {
		err = fmt.Errorf("sender for provider '%s' is not registered", serviceProvider.Provider)
		return
	}

	report, err = client.Send(ctx, workerID, serviceProvider, msg)
	if err != nil {
		return
	}

	return
}

func (s *SenderMultiplexer) ValidateCredJson(ctx context.Context, provider string, credJson string) (credNative interface{}, err error) {
	if credJson == "" {
		err = fmt.Errorf("passed empty credential json")
		return
	}

	beMux.lock.RLock()
	defer beMux.lock.RUnlock()

	// select the appropriate registered client based on provider key
	client, exist := s.sender[provider]
	if !exist {
		err = fmt.Errorf("dry run for provider '%s' is not registered", provider)
		return
	}

	credNative, err = client.ValidateCredJson(ctx, credJson)
	return
}

func (s *SenderMultiplexer) ValidateMsg(ctx context.Context, provider string, msg *Message) (message interface{}, err error) {
	if msg == nil {
		err = fmt.Errorf("passed message is nil, we cannot process that")
		return
	}

	beMux.lock.RLock()
	defer beMux.lock.RUnlock()

	// select the appropriate registered client based on provider key
	client, exist := s.sender[provider]
	if !exist {
		err = fmt.Errorf("dry run for provider '%s' is not registered", provider)
		return
	}

	message, err = client.ValidateMsg(ctx, msg)
	return
}

func (s *SenderMultiplexer) Examples(ctx context.Context) (examples []Example) {
	beMux.lock.RLock()
	defer beMux.lock.RUnlock()

	examples = make([]Example, 0)
	for provider := range s.sender {
		client, exist := s.sender[provider]
		if !exist {
			continue
		}

		cred, msg := client.Example(ctx)

		examples = append(examples, Example{
			Provider:      provider,
			BackendConfig: cred,
			Message:       msg,
		})
	}

	return
}

func (s *SenderMultiplexer) ListProviders(_ context.Context) (providers []string) {
	beMux.lock.RLock()
	defer beMux.lock.RUnlock()

	providers = make([]string, 0)
	for provider := range s.sender {
		providers = append(providers, provider)
	}

	return
}
