package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
)

type RedisConfig struct {
	DB redis.UniversalClient `validate:"required"`
}

type Redis struct {
	Conf RedisConfig
}

var _ Cache = (*Redis)(nil)

func NewRedis(conf RedisConfig) (*Redis, error) {
	err := validator.New().Struct(conf)
	if err != nil {
		err = fmt.Errorf("error validate cache redis: %w", err)
		return nil, err
	}

	return &Redis{Conf: conf}, nil
}

func (r *Redis) GetAs(ctx context.Context, key string, out interface{}) error {
	val, err := r.Conf.DB.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		err = fmt.Errorf("%w: %s", ErrKeyNotExist, err)
		return err
	}

	if err != nil {
		err = fmt.Errorf("error occured on redis: %w", err)
		return err
	}

	return json.Unmarshal([]byte(val), out)
}

func (r *Redis) SetExp(ctx context.Context, key string, inValue interface{}, expireDur time.Duration) error {
	val, err := json.Marshal(inValue)
	if err != nil {
		err = fmt.Errorf("cannot marshal json value: %w", err)
		return err
	}

	return r.Conf.DB.Set(ctx, key, val, expireDur).Err()
}

func (r *Redis) Delete(ctx context.Context, key string) error {
	err := r.Conf.DB.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		err = fmt.Errorf("%w: %s", ErrKeyNotExist, err)
		return err
	}

	if err != nil {
		err = fmt.Errorf("error occured on redis: %w", err)
		return err
	}

	return nil
}
