package redis

import (
	"errors"
	"time"
)

var (
	// ErrEmptyHost is an error that indicates no redis host is set.
	ErrEmptyHost = errors.New("empty redis host")
	// ErrEmptyType is an error that indicates no redis type is set.
	ErrEmptyType = errors.New("empty redis type")
	// ErrEmptyKey is an error that indicates no redis key is set.
	ErrEmptyKey = errors.New("empty redis key")
	// ErrRoutePolicy is an error that occurs when two policies conflict.
	ErrRoutePolicy = errors.New("route policy conflict")
	// ErrClusterSingleNode is an error that occurs when you use the cluster type as a single node and don't enable readonly.
	ErrClusterSingleNode = errors.New("when you use the cluster type as a single node and don't enable readonly, please use node type")

	// ErrClusterSetDB is an error when you use the cluster mode and set db.
	ErrClusterSetDB = errors.New("when you use the cluster mode, please not set db")
)

type (
	// A RedisConf is a redis config.
	RedisConf struct {
		Host             string
		RouteByLatency   bool   `json:",optional"`
		RouteRandomly    bool   `json:",optional"`
		SingleReplicaSet bool   `json:",optional"`
		Type             string `json:",default=node,options=node|cluster"`
		Pass             string `json:",optional"`
		Tls              bool   `json:",optional"`
		EnableBreaker    bool   `json:",default=true"`
		NonBlock         bool   `json:",default=true"`
		// PingTimeout is the timeout for ping redis.
		PingTimeout    time.Duration `json:",default=1s"`
		DB             uint64        `json:",default=0"`
		PoolSize       int           `json:",optional"`
		MaxActiveConns int           `json:",optional"`
		MaxIdleConns   int           `json:",optional"`
		MinIdleConns   int           `json:",optional"`
	}

	// A RedisKeyConf is a redis config with key.
	RedisKeyConf struct {
		RedisConf
		Key string
	}
)

// NewRedis returns a Redis.
// Deprecated: use MustNewRedis or NewRedis instead.
func (rc RedisConf) NewRedis() *Redis {
	var opts []Option
	if rc.Type == ClusterType {
		opts = append(opts, Cluster())

		if rc.RouteByLatency {
			opts = append(opts, RouteByLatency())
		}

		if rc.RouteRandomly {
			opts = append(opts, RouteRandomly())
		}

		if rc.SingleReplicaSet {
			opts = append(opts, SingleReplicaSet())
		}

	}
	if len(rc.Pass) > 0 {
		opts = append(opts, WithPass(rc.Pass))
	}
	if rc.Tls {
		opts = append(opts, WithTLS())
	}

	return New(rc.Host, opts...)
}

// Validate validates the RedisConf.
func (rc RedisConf) Validate() error {
	if len(rc.Host) == 0 {
		return ErrEmptyHost
	}

	if len(rc.Type) == 0 {
		return ErrEmptyType
	}
	if rc.RouteRandomly && rc.RouteByLatency {
		return ErrRoutePolicy
	}
	if rc.SingleReplicaSet && !(rc.RouteByLatency || rc.RouteRandomly) {
		return ErrClusterSingleNode
	}
	// cluster mode should not set db, except SingleReplicaSet
	if rc.Type == ClusterType && !rc.SingleReplicaSet && rc.DB != 0 {
		return ErrClusterSetDB
	}

	return nil
}

// Validate validates the RedisKeyConf.
func (rkc RedisKeyConf) Validate() error {
	if err := rkc.RedisConf.Validate(); err != nil {
		return err
	}

	if len(rkc.Key) == 0 {
		return ErrEmptyKey
	}

	return nil
}
