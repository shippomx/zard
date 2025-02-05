package zrpc

import (
	"os"
	"testing"

	"github.com/shippomx/zard/core/discov"
	"github.com/shippomx/zard/core/service"
	"github.com/shippomx/zard/core/stores/redis"
	"github.com/stretchr/testify/assert"
)

func TestRpcClientConf(t *testing.T) {
	t.Run("direct", func(t *testing.T) {
		conf := NewDirectClientConf([]string{"localhost:1234"}, "foo", "bar")
		assert.True(t, conf.HasCredential())
	})

	t.Run("etcd", func(t *testing.T) {
		conf := NewEtcdClientConf([]string{"localhost:1234", "localhost:5678"},
			"key", "foo", "bar")
		assert.True(t, conf.HasCredential())
	})

	t.Run("etcd with account", func(t *testing.T) {
		conf := NewEtcdClientConf([]string{"localhost:1234", "localhost:5678"},
			"key", "foo", "bar")
		conf.Etcd.User = "user"
		conf.Etcd.Pass = "pass"
		_, err := conf.BuildTarget()
		assert.NoError(t, err)
	})

	t.Run("etcd with tls", func(t *testing.T) {
		conf := NewEtcdClientConf([]string{"localhost:1234", "localhost:5678"},
			"key", "foo", "bar")
		conf.Etcd.CertFile = "cert"
		conf.Etcd.CertKeyFile = "key"
		conf.Etcd.CACertFile = "ca"
		_, err := conf.BuildTarget()
		assert.Error(t, err)
	})

	t.Run("test xds target", func(t *testing.T) {
		conf := RpcClientConf{Target: "localhost:1234"}
		os.Setenv("GRPC_XDS_BOOTSTRAP", "1")
		defer os.Unsetenv("GRPC_XDS_BOOTSTRAP")
		target, err := conf.BuildTarget()
		assert.NoError(t, err)
		assert.Equal(t, "xds:///localhost:1234", target)
		os.Unsetenv("GRPC_XDS_BOOTSTRAP")
		target, err = conf.BuildTarget()
		assert.NoError(t, err)
		assert.Equal(t, "localhost:1234", target)

		conf = RpcClientConf{Target: "xds:///localhost:1234"}
		target, err = conf.BuildTarget()
		assert.NoError(t, err)
		assert.Equal(t, "xds:///localhost:1234", target)
		os.Setenv("GRPC_XDS_BOOTSTRAP", "1")
		defer os.Unsetenv("GRPC_XDS_BOOTSTRAP")
		target, err = conf.BuildTarget()
		assert.NoError(t, err)
		assert.Equal(t, "xds:///localhost:1234", target)
	})
}

func TestRpcServerConf(t *testing.T) {
	conf := RpcServerConf{
		ServiceConf: service.ServiceConf{},
		ListenOn:    "",
		Etcd: discov.EtcdConf{
			Hosts: []string{"localhost:1234"},
			Key:   "key",
		},
		Auth: true,
		Redis: redis.RedisKeyConf{
			RedisConf: redis.RedisConf{
				Type: redis.NodeType,
			},
			Key: "foo",
		},
		StrictControl: false,
		Timeout:       0,
		CpuThreshold:  0,
	}
	assert.True(t, conf.HasEtcd())
	assert.NotNil(t, conf.Validate())
	conf.Redis.Host = "localhost:5678"
	assert.Nil(t, conf.Validate())
}
