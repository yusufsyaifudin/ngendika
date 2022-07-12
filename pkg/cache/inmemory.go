package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/VictoriaMetrics/fastcache"
)

type InMemory struct {
	DB *fastcache.Cache
}

var _ Cache = (*InMemory)(nil)

func NewInMemory() (*InMemory, error) {
	db := fastcache.New(32 * 1048576) // 32MB
	return &InMemory{
		DB: db,
	}, nil
}

func (i *InMemory) GetAs(_ context.Context, key string, out interface{}) error {
	result := i.DB.Get(nil, []byte(key))
	if result == nil {
		return ErrKeyNotExist
	}

	return json.Unmarshal(result, out)
}

// SetExp using InMemory does not support expired.
func (i *InMemory) SetExp(_ context.Context, key string, inValue interface{}, _ time.Duration) error {
	val, err := json.Marshal(inValue)
	if err != nil {
		err = fmt.Errorf("cannot marshal json value: %w", err)
		return err
	}

	i.DB.Set([]byte(key), val)
	return nil
}

func (i *InMemory) Delete(ctx context.Context, key string) error {
	i.DB.Del([]byte(key))
	return nil
}
