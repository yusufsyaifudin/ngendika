package cache_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/yusufsyaifudin/ngendika/pkg/cache"
)

func prepareMiniRedis(t *testing.T) *redis.Client {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
		DB:   1,
	})

	return client
}

func TestNewRedis(t *testing.T) {
	t.Run("bad dep", func(t *testing.T) {
		c, err := cache.NewRedis(cache.RedisConfig{})
		assert.Nil(t, c)
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		redisConn := prepareMiniRedis(t)
		conf := cache.RedisConfig{
			DB: redisConn,
		}

		c, err := cache.NewRedis(conf)
		assert.NotNil(t, c)
		assert.NoError(t, err)
	})
}

func TestRedis_GetAs(t *testing.T) {
	type S struct {
		Value string
	}

	t.Run("no key found", func(t *testing.T) {
		redisConn := prepareMiniRedis(t)
		conf := cache.RedisConfig{
			DB: redisConn,
		}

		c, err := cache.NewRedis(conf)
		assert.NotNil(t, c)
		assert.NoError(t, err)

		var out S
		err = c.GetAs(context.Background(), "key", &out)
		assert.Error(t, err)
		assert.ErrorIs(t, err, cache.ErrKeyNotExist)
	})

	t.Run("redis error: closed connection", func(t *testing.T) {
		redisConn := prepareMiniRedis(t)
		conf := cache.RedisConfig{
			DB: redisConn,
		}

		c, err := cache.NewRedis(conf)
		assert.NotNil(t, c)
		assert.NoError(t, err)

		// close redis to state error
		err = redisConn.Close()

		var out S
		err = c.GetAs(context.Background(), "key", &out)
		assert.Error(t, err)
		assert.ErrorIs(t, err, redis.ErrClosed)
	})

	t.Run("success", func(t *testing.T) {
		redisConn := prepareMiniRedis(t)
		conf := cache.RedisConfig{
			DB: redisConn,
		}

		c, err := cache.NewRedis(conf)
		assert.NotNil(t, c)
		assert.NoError(t, err)

		in := S{
			Value: "this is value",
		}

		err = c.SetExp(context.Background(), "key", in, time.Second)
		assert.NoError(t, err)

		var out S
		err = c.GetAs(context.Background(), "key", &out)
		assert.NoError(t, err)
		assert.Equal(t, in, out)
	})
}

func TestRedis_SetExp(t *testing.T) {
	t.Run("error marshal data", func(t *testing.T) {
		redisConn := prepareMiniRedis(t)
		conf := cache.RedisConfig{
			DB: redisConn,
		}

		c, err := cache.NewRedis(conf)
		assert.NotNil(t, c)
		assert.NoError(t, err)

		in := map[string]interface{}{
			"key": make(chan int, 1),
		}

		err = c.SetExp(context.Background(), "key", in, time.Second)
		assert.Error(t, err)

		var eType *json.UnsupportedTypeError
		assert.ErrorAs(t, err, &eType)
		assert.NotNil(t, eType)
	})

	t.Run("success with expiration", func(t *testing.T) {
		redisConn := prepareMiniRedis(t)
		conf := cache.RedisConfig{
			DB: redisConn,
		}

		c, err := cache.NewRedis(conf)
		assert.NotNil(t, c)
		assert.NoError(t, err)

		type S struct {
			Value string
		}

		in := S{
			Value: "this is value",
		}

		err = c.SetExp(context.Background(), "key", in, time.Second)
		assert.NoError(t, err)
	})
}

func TestRedis_Delete(t *testing.T) {
	t.Run("success: not exist key", func(t *testing.T) {
		redisConn := prepareMiniRedis(t)
		conf := cache.RedisConfig{
			DB: redisConn,
		}

		c, err := cache.NewRedis(conf)
		assert.NotNil(t, c)
		assert.NoError(t, err)

		err = c.Delete(context.Background(), "key")
		assert.NoError(t, err)
	})

	t.Run("redis error: closed connection", func(t *testing.T) {
		redisConn := prepareMiniRedis(t)
		conf := cache.RedisConfig{
			DB: redisConn,
		}

		c, err := cache.NewRedis(conf)
		assert.NotNil(t, c)
		assert.NoError(t, err)

		type S struct {
			Value string
		}

		in := S{
			Value: "this is value",
		}

		err = c.SetExp(context.Background(), "key", in, time.Second)
		assert.NoError(t, err)

		// close connection
		err = redisConn.Close()
		assert.NoError(t, err)

		err = c.Delete(context.Background(), "key")
		assert.Error(t, err)
		assert.ErrorIs(t, err, redis.ErrClosed)
	})

	t.Run("success", func(t *testing.T) {
		redisConn := prepareMiniRedis(t)
		conf := cache.RedisConfig{
			DB: redisConn,
		}

		c, err := cache.NewRedis(conf)
		assert.NotNil(t, c)
		assert.NoError(t, err)

		type S struct {
			Value string
		}

		in := S{
			Value: "this is value",
		}

		err = c.SetExp(context.Background(), "key", in, time.Second)
		assert.NoError(t, err)

		err = c.Delete(context.Background(), "key")
		assert.NoError(t, err)
	})
}
