package cache

import (
	"context"
	"fmt"
	"time"
)

var (
	ErrKeyNotExist = fmt.Errorf("cache key not exists")
)

type Cache interface {
	GetAs(ctx context.Context, key string, out interface{}) error
	SetExp(ctx context.Context, key string, inValue interface{}, expireDur time.Duration) error
	Delete(ctx context.Context, key string) error
}
