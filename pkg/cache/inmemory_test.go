package cache_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yusufsyaifudin/ngendika/pkg/cache"
)

func TestNewInMemory(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		c, err := cache.NewInMemory()
		assert.NotNil(t, c)
		assert.NoError(t, err)
	})
}

func TestInMemory_GetAs(t *testing.T) {
	type S struct {
		Value string
	}

	t.Run("no key found", func(t *testing.T) {
		c, err := cache.NewInMemory()
		assert.NotNil(t, c)
		assert.NoError(t, err)

		var out S
		err = c.GetAs(context.Background(), "key", &out)
		assert.Error(t, err)
		assert.ErrorIs(t, err, cache.ErrKeyNotExist)
	})

	t.Run("success", func(t *testing.T) {
		c, err := cache.NewInMemory()
		assert.NotNil(t, c)
		assert.NoError(t, err)

		in := S{
			Value: "this is value",
		}

		err = c.SetExp(context.Background(), "key", in, -1)
		assert.NoError(t, err)

		var out S
		err = c.GetAs(context.Background(), "key", &out)
		assert.NoError(t, err)
		assert.Equal(t, in, out)
	})
}

func TestInMemory_SetExp(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		c, err := cache.NewInMemory()
		assert.NotNil(t, c)
		assert.NoError(t, err)

		in := map[string]interface{}{
			"key": make(chan int, 1),
		}

		err = c.SetExp(context.Background(), "key", in, -1)
		assert.Error(t, err)

		var eType *json.UnsupportedTypeError
		assert.ErrorAs(t, err, &eType)
		assert.NotNil(t, eType)
	})

	t.Run("success", func(t *testing.T) {
		c, err := cache.NewInMemory()
		assert.NotNil(t, c)
		assert.NoError(t, err)

		type S struct {
			Value string
		}

		in := S{
			Value: "this is value",
		}

		err = c.SetExp(context.Background(), "key", in, -1)
		assert.NoError(t, err)
	})
}

func TestInMemory_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c, err := cache.NewInMemory()
		assert.NotNil(t, c)
		assert.NoError(t, err)

		type S struct {
			Value string
		}

		in := S{
			Value: "this is value",
		}

		err = c.SetExp(context.Background(), "key", in, -1)
		assert.NoError(t, err)

		err = c.Delete(context.Background(), "key")
		assert.NoError(t, err)
	})
}
