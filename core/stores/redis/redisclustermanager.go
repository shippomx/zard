package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shippomx/zard/core/syncx"

	red "github.com/redis/go-redis/v9"
)

const addrSep = ","

var (
	clusterManager = syncx.NewResourceManager()
	// clusterPoolSize is default pool size for cluster type of redis.
	clusterPoolSize = 5 * runtime.GOMAXPROCS(0)
)

func getCluster(r *Redis) (*red.ClusterClient, error) {
	val, err := clusterManager.GetResource(r.Addr, func() (io.Closer, error) {
		var tlsConfig *tls.Config
		if r.tls {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		poolSize := clusterPoolSize
		if r.poolSize != 0 {
			poolSize = r.poolSize
		}
		idleC := idleConns
		if r.minIdleConns != 0 {
			idleC = r.minIdleConns
		}
		clusterOptions := &red.ClusterOptions{
			Addrs:          splitClusterAddrs(r.Addr),
			Password:       r.Pass,
			MaxRetries:     maxRetries,
			MinIdleConns:   idleC,
			TLSConfig:      tlsConfig,
			RouteRandomly:  r.RouteRandomly,
			RouteByLatency: r.RouteByLatency,
			PoolSize:       poolSize,
			MaxActiveConns: r.maxActiveConns,
			MaxIdleConns:   r.maxIdleConns,
		}
		var dcClient *DiscoverResolverClient
		var err error
		if r.SingleReplicaSet {
			var options []*red.Options
			for _, address := range splitClusterAddrs(r.Addr) {
				options = append(options, &red.Options{
					Addr:       address,
					Password:   r.Pass,
					TLSConfig:  tlsConfig,
					MaxRetries: maxRetries,
				})
			}
			dcClient, err = NewDiscoverResolverClient(options)
			if err != nil {
				return nil, err
			}
			if len(dcClient.GetClusterNodes()) == 0 {
				return nil, fmt.Errorf("no nodes in cluster")
			}
			// clusterSlots will be called concurrently.
			clusterSlots := func(ctx context.Context) ([]red.ClusterSlot, error) {
				slots := []red.ClusterSlot{
					{
						Start: 0,
						End:   16383,
						Nodes: dcClient.GetClusterNodes(), // GetClusterNodes needs to ensure concurrency safety.
					},
				}
				return slots, nil
			}
			clusterOptions.ClusterSlots = clusterSlots
			if r.DB != 0 {
				clusterOptions.OnConnect = func(ctx context.Context, conn *red.Conn) error {
					res, err := conn.Select(ctx, int(r.DB)).Result()
					if err != nil {
						return err
					}
					if res != "OK" {
						return fmt.Errorf("failed to select db: %s", res)
					}
					return nil
				}
			}
		}

		store := red.NewClusterClient(clusterOptions)

		if r.SingleReplicaSet {
			dcClient.mu.Lock()
			dcClient.onUpdate = func(ctx context.Context) {
				store.ReloadState(ctx)
			}
			dcClient.mu.Unlock()

		}
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
			version:    version,
			clientType: ClusterType,
			key:        r.Addr,
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

	return val.(*red.ClusterClient), nil
}

func splitClusterAddrs(addr string) []string {
	addrs := strings.Split(addr, addrSep)
	unique := make(map[string]struct{})
	for _, each := range addrs {
		unique[strings.TrimSpace(each)] = struct{}{}
	}

	addrs = addrs[:0]
	for k := range unique {
		addrs = append(addrs, k)
	}

	return addrs
}

func GetHost(addr string) (string, string) {
	host, port, err := net.SplitHostPort(addr)
	if err == nil {
		return host, port
	}
	return "", ""
}

type DiscoverResolverClient struct {
	onUpdate func(ctx context.Context)
	mu       sync.RWMutex
	stop     chan struct{}
	NodeMap  map[string]NodeInfo
}

type NodeInfo struct {
	Address string
	Role    string
	Port    int
}

func (dc *DiscoverResolverClient) GetClusterNodes() []red.ClusterNode {
	var nodes []red.ClusterNode
	var slaves []red.ClusterNode
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	for _, v := range dc.NodeMap {
		if strings.Contains(v.Role, "master") {
			nodes = append(nodes, red.ClusterNode{
				Addr: net.JoinHostPort(v.Address, strconv.Itoa(v.Port)),
			})
		}
		if strings.Contains(v.Role, "slave") {
			slaves = append(slaves, red.ClusterNode{
				Addr: net.JoinHostPort(v.Address, strconv.Itoa(v.Port)),
			})
		}
	}
	nodes = append(nodes, slaves...)
	return nodes
}

func GetNodeMap(options []*red.Options) map[string]NodeInfo {
	nodeMap := make(map[string]NodeInfo)
	for _, option := range options {
		host, port := GetHost(option.Addr)
		addrs, _ := net.LookupHost(host)
		for _, addr := range addrs {
			newOption := *option
			newOption.Addr = net.JoinHostPort(addr, port)
			_client := red.NewClient(&newOption)
			ctx, cancel := context.WithTimeout(context.Background(), defaultPingTimeout)
			defer cancel()
			info, err := _client.Info(ctx).Result()
			runId := ""
			role := ""
			if err == nil {
				for _, line := range strings.Split(info, "\n") {
					if strings.HasPrefix(line, "run_id") {
						runId = strings.Split(line, ":")[1]
					}
					if strings.HasPrefix(line, "role") {
						role = strings.Split(line, ":")[1]
					}
				}
				iport, err := strconv.Atoi(port)
				if err != nil {
					iport = 6379
				}
				if runId != "" && role != "" {
					nodeMap[runId] = NodeInfo{
						Address: addr,
						Role:    role,
						Port:    iport,
					}
				}
			}
			_client.Close()
		}
	}
	return nodeMap
}

func (dc *DiscoverResolverClient) SetOnUpdate(fn func(ctx context.Context)) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.onUpdate = fn
}

func NewDiscoverResolverClient(options []*red.Options) (*DiscoverResolverClient, error) {
	client := &DiscoverResolverClient{}
	ticker := time.NewTicker(defaultPingTimeout * 5)
	client.mu.Lock()
	client.stop = make(chan struct{})
	client.NodeMap = GetNodeMap(options)
	client.mu.Unlock()
	if client.onUpdate != nil {
		client.onUpdate(context.Background())
	}
	go func() {
	loop:
		for {
			select {
			case <-client.stop:
				ticker.Stop()
				break loop
			case <-ticker.C:
				updated := false
				newNodeMap := GetNodeMap(options)
				oldNodeMap := make(map[string]NodeInfo)
				client.mu.RLock()
				for k, v := range client.NodeMap {
					oldNodeMap[k] = v
				}
				client.mu.RUnlock()
				for runId, nodeInfo := range newNodeMap {
					if oldNodeMap[runId] != nodeInfo {
						client.mu.Lock()
						client.NodeMap[runId] = nodeInfo
						client.mu.Unlock()
						updated = true
					}
				}

				for runId, nodeInfo := range oldNodeMap {
					if _, ok := newNodeMap[runId]; !ok {
						newOption := *options[0]
						newOption.Addr = net.JoinHostPort(nodeInfo.Address, strconv.Itoa(nodeInfo.Port))
						redClient := red.NewClient(&newOption)
						_, err := redClient.Ping(context.Background()).Result()
						redClient.Close()
						if err != nil {
							client.mu.Lock()
							delete(client.NodeMap, runId)
							client.mu.Unlock()
							updated = true
						}

					}
				}
				if updated && client.onUpdate != nil {
					client.onUpdate(context.Background())
				}
			}
		}
	}()

	return client, nil
}
