package befcm

import (
	"firebase.google.com/go/v4/messaging"
	"github.com/yusufsyaifudin/ngendika/pkg/fcm"
	"testing"
)

func BenchmarkGolangTypeCoercion(b *testing.B) {
	for i := 0; i < b.N; i++ {

		var message interface{}
		message = fcm.MulticastMessage{
			Tokens: []string{"token"},
			Data: map[string]string{
				"k": "v",
			},
			Notification: &messaging.Notification{
				Title:    "title",
				Body:     "body",
				ImageURL: "http://host.domain",
			},
		}

		_, ok := message.(fcm.MulticastMessage)
		if !ok {
			b.Fatalf("invalid fcm multicast message, got type '%T'", message)
			return
		}

	}
}

func BenchmarkGolangTypeCoercionSwitch(b *testing.B) {
	for i := 0; i < b.N; i++ {

		var message interface{}
		message = fcm.MulticastMessage{
			Tokens: []string{"token"},
			Data: map[string]string{
				"k": "v",
			},
			Notification: &messaging.Notification{
				Title:    "title",
				Body:     "body",
				ImageURL: "http://host.domain",
			},
		}

		switch message.(type) {
		case fcm.MulticastMessage:

		default:
			b.Fatalf("invalid fcm multicast message, got type '%T'", message)
			return
		}

	}
}
