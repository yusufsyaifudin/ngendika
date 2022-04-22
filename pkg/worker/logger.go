package worker

import (
	"context"
	"log"
)

type Logger interface {
	Info(ctx context.Context, msg string)
}

type stdOut struct{}

func (*stdOut) Info(ctx context.Context, msg string) {
	log.Println(msg)
}
