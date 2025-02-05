package redis

import (
	"testing"

	"github.com/shippomx/zard/core/stringx"
	"github.com/stretchr/testify/assert"
)

func TestRedisConf(t *testing.T) {
	tests := []struct {
		name string
		RedisConf
		ok bool
	}{
		{
			name: "missing host",
			RedisConf: RedisConf{
				Host: "",
				Type: NodeType,
				Pass: "",
			},
			ok: false,
		},
		{
			name: "missing type",
			RedisConf: RedisConf{
				Host: "localhost:6379",
				Type: "",
				Pass: "",
			},
			ok: false,
		},
		{
			name: "ok",
			RedisConf: RedisConf{
				Host: "localhost:6379",
				Type: NodeType,
				Pass: "",
			},
			ok: true,
		},
		{
			name: "ok",
			RedisConf: RedisConf{
				Host: "localhost:6379",
				Type: ClusterType,
				Pass: "pwd",
				Tls:  true,
			},
			ok: true,
		},
		{
			name: "missing route policy",
			RedisConf: RedisConf{
				Host:           "localhost:6379",
				Type:           ClusterType,
				Pass:           "pwd",
				RouteByLatency: true,
				RouteRandomly:  true,
				Tls:            true,
			},
			ok: false,
		},
		{
			name: "set db when cluster type",
			RedisConf: RedisConf{
				Host:           "localhost:6379",
				Type:           ClusterType,
				Pass:           "pwd",
				RouteByLatency: true,
				RouteRandomly:  true,
				Tls:            true,
				DB:             1,
			},
			ok: false,
		},
		{
			name: "set db when cluster type with SingleReplicaSet ",
			RedisConf: RedisConf{
				Host:             "localhost:6379",
				Type:             ClusterType,
				Pass:             "pwd",
				RouteByLatency:   true,
				Tls:              true,
				DB:               1,
				SingleReplicaSet: true,
			},
			ok: true,
		},
	}

	for _, test := range tests {
		t.Run(stringx.RandId(), func(t *testing.T) {
			if test.ok {
				assert.Nil(t, test.RedisConf.Validate())
				assert.NotNil(t, test.RedisConf.NewRedis())
			} else {
				assert.NotNil(t, test.RedisConf.Validate())
			}
		})
	}
}

func TestRedisKeyConf(t *testing.T) {
	tests := []struct {
		name string
		RedisKeyConf
		ok bool
	}{
		{
			name: "missing host",
			RedisKeyConf: RedisKeyConf{
				RedisConf: RedisConf{
					Host: "",
					Type: NodeType,
					Pass: "",
				},
				Key: "foo",
			},
			ok: false,
		},
		{
			name: "missing key",
			RedisKeyConf: RedisKeyConf{
				RedisConf: RedisConf{
					Host: "localhost:6379",
					Type: NodeType,
					Pass: "",
				},
				Key: "",
			},
			ok: false,
		},
		{
			name: "ok",
			RedisKeyConf: RedisKeyConf{
				RedisConf: RedisConf{
					Host: "localhost:6379",
					Type: NodeType,
					Pass: "",
				},
				Key: "foo",
			},
			ok: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.ok {
				assert.Nil(t, test.RedisKeyConf.Validate())
			} else {
				assert.NotNil(t, test.RedisKeyConf.Validate())
			}
		})
	}
}
