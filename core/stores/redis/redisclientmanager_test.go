package redis

import (
	"context"
	"testing"

	"github.com/shippomx/zard/core/breaker"
	"github.com/alicebob/miniredis/v2"
	red "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestGetClient(t *testing.T) {
	r := miniredis.RunT(t)
	defer r.Close()
	c, err := getClient(&Redis{
		Addr:  r.Addr(),
		Type:  NodeType,
		brk:   breaker.NewBreaker(),
		hooks: []red.Hook{defaultDurationHook},
	})
	if assert.NoError(t, err) {
		assert.NotNil(t, c)
	}
}

func TestGetClientWithDB(t *testing.T) {
	r := miniredis.RunT(t)
	defer r.Close()
	c, err := getClient(&Redis{
		Addr:  r.Addr(),
		Type:  NodeType,
		brk:   breaker.NewBreaker(),
		hooks: []red.Hook{defaultDurationHook},
	})
	if assert.NoError(t, err) {
		assert.NotNil(t, c)
	}

	assert.Equal(t, "OK", c.Conn().Set(context.Background(), "key", "value", 0).Val())
	// test with db
	r2 := miniredis.RunT(t)
	defer r2.Close()
	c2, err := getClient(&Redis{
		Addr:  r2.Addr(),
		Type:  NodeType,
		DB:    2,
		brk:   breaker.NewBreaker(),
		hooks: []red.Hook{defaultDurationHook},
	})
	assert.NoError(t, err)
	res, err := c2.Conn().Set(context.Background(), "key", "value", 0).Result()
	assert.NoError(t, err)
	assert.Equal(t, "OK", res)
	testClient := red.NewClient(&red.Options{
		Addr: r2.Addr(),
		DB:   2,
	})
	assert.Equal(t, "value", testClient.Get(context.Background(), "key").Val())

	// test with err
	r3 := miniredis.RunT(t)
	r3.SetError("custom error")
	defer r3.Close()
	c3, err := getClient(&Redis{
		Addr:  r3.Addr(),
		Type:  NodeType,
		brk:   breaker.NewBreaker(),
		hooks: []red.Hook{defaultDurationHook},
	})
	assert.NoError(t, err)
	res, err = c3.Conn().Set(context.Background(), "key", "value", 0).Result()
	assert.Error(t, err)
	assert.Equal(t, "", res)
}
