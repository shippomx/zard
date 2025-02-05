package redis

import (
	"context"
	"crypto/tls"
	"io"
	"runtime"

	"github.com/shippomx/zard/core/syncx"
	red "github.com/redis/go-redis/v9"
)

const (
	defaultDatabase = 0
	maxRetries      = 3
	idleConns       = 8
)

var (
	clientManager = syncx.NewResourceManager()
	// nodePoolSize is default pool size for node type of redis.
	nodePoolSize = 10 * runtime.GOMAXPROCS(0)
)

func getClient(r *Redis) (*red.Client, error) {
	val, err := clientManager.GetResource(r.Addr, func() (io.Closer, error) {
		var tlsConfig *tls.Config
		if r.tls {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		db := defaultDatabase
		if r.DB != 0 {
			db = int(r.DB)
		}
		poolSize := nodePoolSize
		if r.poolSize != 0 {
			poolSize = r.poolSize
		}
		idleC := idleConns
		if r.minIdleConns != 0 {
			idleC = r.minIdleConns
		}
		store := red.NewClient(&red.Options{
			Addr:           r.Addr,
			Password:       r.Pass,
			DB:             db,
			MaxRetries:     maxRetries,
			MinIdleConns:   idleC,
			TLSConfig:      tlsConfig,
			PoolSize:       poolSize,
			MaxActiveConns: r.maxActiveConns,
			MaxIdleConns:   r.maxIdleConns,
		})
		defaultRedisHook := []red.Hook{
			defaultDurationHook,
		}
		if r.EnableBrk {
			defaultRedisHook = append(defaultRedisHook, breakerHook{
				brk: r.brk,
			})
		}
		hooks := append(defaultRedisHook, r.hooks...)
		for _, hook := range hooks {
			store.AddHook(hook)
		}
		ctx, cancel := context.WithTimeout(context.Background(), defaultPingTimeout)
		defer cancel()
		info := store.InfoMap(ctx, "Server")
		version := info.Item("Server", "redis_version")
		connCollector.registerClient(&statGetter{
			clientType: NodeType,
			key:        r.Addr,
			version:    version,
			poolSize:   poolSize,
			poolStats: func() *red.PoolStats {
				return store.PoolStats()
			},
		})
		return store, nil
	})
	if err != nil {
		return nil, err
	}
	return val.(*red.Client), nil
}
