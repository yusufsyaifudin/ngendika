package backend

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

var once sync.Once

func BenchmarkSenderMultiplexer_Send(b *testing.B) {
	beNoop, err := NewNoopSender()
	assert.NotNil(b, beNoop)
	assert.NoError(b, err)

	once.Do(func() {
		MustRegister("noop", beNoop)
	})

	be := MuxBackend()

	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		provider := PushNotificationProvider{
			ID:             int64(i),
			AppID:          1234,
			Provider:       "noop",
			Label:          "noop-test",
			CredentialJSON: `{"config": {"username": "user", "password": "password"}}`,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		msg := &Message{
			ReferenceID: fmt.Sprintf("ref-%d-%d", i, time.Now().UnixNano()),
			RawPayload:  nil,
		}

		report, err := be.Send(ctx, i, provider, msg)
		if report == nil {
			b.Logf("nil report on ref id %s\n", msg.ReferenceID)
		}

		if err != nil {
			b.Logf("error send: %s\n", err)
		}
	}
}

func BenchmarkSyncMap(b *testing.B) {
	type Map struct {
		m sync.Map
	}

	m := Map{}
	m.m.Store("foo", "bar")

	for i := 0; i < b.N; i++ {
		val, ok := m.m.Load("foo")
		if !ok {
			b.Fatalf("not found key foo")
			return
		}

		if val != "bar" {
			b.Fatalf("not expected value")
			return
		}
	}
}

func BenchmarkSyncRwMutex(b *testing.B) {
	type Map struct {
		m    map[string]string
		lock sync.RWMutex
	}

	m := Map{
		m: make(map[string]string),
	}
	m.lock.Lock()
	m.m["foo"] = "bar"
	m.lock.Unlock()

	for i := 0; i < b.N; i++ {
		m.lock.RLock()
		val, ok := m.m["foo"]
		m.lock.RUnlock()
		if !ok {
			b.Fatalf("not found key foo")
			return
		}

		if val != "bar" {
			b.Fatalf("not expected value")
			return
		}

	}
}
