// nacos discover service client

package nds

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/shippomx/zard/core/logx"
	"github.com/shippomx/zard/core/proc"
	"github.com/shippomx/zard/rest/httpc/v2"
	"github.com/shippomx/zard/rest/httpc/v2/admin"
)

var (
	nacosTransport = &NacosTransport{}
	DefaultTimeout = 10 * time.Second
)

func init() {
	if admin.SchemeFunc == nil {
		admin.SchemeFunc = make(map[string]http.RoundTripper)
	}
	// 从环境初始化客户端
	nacosTransport.initByEnv()
	admin.SchemeFunc["nacos"] = nacosTransport
	admin.SchemeFunc["nacoss"] = nacosTransport
	if nacosTransport.inited {
		admin.AddService(func(client interface{}, fn func()) {
			switch c := client.(type) {
			case *httpc.HTTPClient:
				logx.Debug("http client transport re register")
				nacosTransport.RegisterTransport(c.GetHTTPTransport())
				fn()
			case *http.Client:
				logx.Debug("nacos transport register")
			default:
				logx.Error("http client no init")
			}
		})
	}
}

type NacosTransport struct {
	inited bool
	config *NacosDiscoveryConfig
	// transport       *http.Transport
	nacostransport *http.Transport
	nacosClient    naming_client.INamingClient
	mu             sync.Mutex
}

func MustNewNacosTransport(c *NacosDiscoveryConfig) *NacosTransport {
	n, err := NewNacosTransport(c)
	logx.Must(err)
	return n
}

func NewNacosTransport(c *NacosDiscoveryConfig) (*NacosTransport, error) {
	n := &NacosTransport{}
	err := n.Register(c)
	return n, err
}

// 只对Register写时加锁，其它不加锁 提高效率.
func (n *NacosTransport) Register(c *NacosDiscoveryConfig) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.config = c
	sc := []constant.ServerConfig{
		{
			IpAddr: c.IPAddr,
			Port:   c.Port,
		},
	}
	timeoutMS, err := I64ToU64(c.Timeout.Milliseconds())
	if err != nil {
		return err
	}
	cc := constant.ClientConfig{
		NamespaceId:         c.NamespaceID,
		Username:            c.Username,
		Password:            c.Password,
		TimeoutMs:           timeoutMS,
		NotLoadCacheAtStart: c.NotLoadCacheAtStart,
		LogLevel:            c.LogLevel,
	}

	// init client
	client, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  &cc,
		ServerConfigs: sc,
	})
	if err != nil {
		return err
	}
	// close old client
	if n.nacosClient != nil {
		n.nacosClient.CloseClient()
	}

	n.nacosClient = client
	n.inited = true
	return nil
}

func (n *NacosTransport) RegisterTransport(t *http.Transport) {
	n.mu.Lock()
	defer n.mu.Unlock()
	nt := t.Clone()
	nt.DialContext = n.dialContext
	nt.DialTLSContext = n.dialContext
	n.nacostransport = nt
}

func (n *NacosTransport) Close() {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.nacostransport != nil {
		n.nacostransport.CloseIdleConnections()
	}
	if n.nacosClient != nil {
		n.nacosClient.CloseClient()
	}
	n.inited = false
}

func RegisterNacosTransport(t *http.Transport) {
	nacosTransport.RegisterTransport(t)
}

func Register(c *NacosDiscoveryConfig) {
	err := nacosTransport.Register(c)
	logx.Must(err)
}

func Close() {
	if nacosTransport != nil && nacosTransport.inited {
		if nacosTransport.nacosClient != nil {
			nacosTransport.nacosClient.CloseClient()
		}
	}
	nacosTransport.inited = false
}

func (n *NacosTransport) initByEnv() {
	c := &NacosDiscoveryConfig{}
	if os.Getenv("NACOS_SERVICE_NAME") != "" && !n.inited {
		logx.Info("[http] [client] initing nacos discover  from env")
		if os.Getenv("NACOS_PORT") == "" {
			c.Port = 8848
		} else {
			portUint64, err := strconv.ParseUint(os.Getenv("NACOS_PORT"), 10, 64)
			logx.Must(err)
			c.Port = portUint64
		}
		c.IPAddr = os.Getenv("NACOS_HOST")
		c.Username = os.Getenv("NACOS_USERNAME")

		c.Password = os.Getenv("NACOS_PASSWORD")
		c.GroupName = os.Getenv("NACOS_GROUP")
		if c.GroupName == "" {
			c.GroupName = "DEFAULT_GROUP"
		}
		cluster := os.Getenv("NACOS_CLUSTERS")
		if cluster == "" {
			c.Clusters = []string{"DEFAULT"}
		} else {
			c.Clusters = strings.Split(cluster, ",")
		}

		c.NamespaceID = os.Getenv("NACOS_NAMESPACE")
		if c.NamespaceID == "" {
			c.NamespaceID = "public"
		}
		c.LogLevel = os.Getenv("NACOS_LOG_LEVEL")
		if c.LogLevel == "" {
			c.LogLevel = "info"
		}
		c.Timeout = DefaultTimeout
		n.config = c
		// server conf
		sc := []constant.ServerConfig{
			{
				IpAddr: c.IPAddr,
				Port:   c.Port,
			},
		}
		timeoutMS, err := I64ToU64(c.Timeout.Milliseconds())
		if err != nil {
			logx.Must(err)
		}
		cc := constant.ClientConfig{
			NamespaceId:         c.NamespaceID, // namespace id
			Username:            c.Username,
			Password:            c.Password,
			TimeoutMs:           timeoutMS,
			NotLoadCacheAtStart: true,
			LogLevel:            c.LogLevel,
		}

		// init client
		client, err := clients.NewNamingClient(vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		})
		if err != nil {
			logx.Must(err)
		}
		serviceName := os.Getenv("NACOS_SERVICE_NAME")
		_, err = client.GetService(vo.GetServiceParam{
			ServiceName: serviceName,
			GroupName:   c.GroupName,
			Clusters:    c.Clusters,
		})
		if err != nil {
			logx.Warn("[http] [client] nacos discover try get service ", serviceName, err.Error())
		}
		n.nacosClient = client
		n.nacostransport = &http.Transport{}
		n.inited = true
		proc.AddShutdownListener(func() {
			if n.nacosClient != nil {
				n.nacosClient.CloseClient()
			}
		})

	}
}

func (n *NacosTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	newr := r.Clone(r.Context())
	if newr.URL == nil {
		return nil, errors.New("url  is nil")
	}
	if newr.URL.Scheme == "nacos" {
		newr.URL.Scheme = "http"
	}
	if newr.URL.Scheme == "nacoss" {
		newr.URL.Scheme = "https"
	}

	if n.nacostransport != nil {
		return n.nacostransport.RoundTrip(newr)
	}
	return nil, errors.New("nacos transport not init,pls RegisterTransport")
}

func (n *NacosTransport) dialContext(_ context.Context, network, addr string) (net.Conn, error) {
	if n.inited {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		if n.config == nil {
			return nil, errors.New("nacos config is nil")
		}
		instance, err := n.nacosClient.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
			ServiceName: host,
			GroupName:   n.config.GroupName,
			Clusters:    n.config.Clusters,
		})
		if err != nil {
			return nil, err
		}
		if instance.Port != 0 {
			port = strconv.FormatUint(instance.Port, 10)
		}
		addr := net.JoinHostPort(instance.Ip, port)
		return net.Dial(network, addr)
	}
	return net.Dial(network, addr)
}

func I64ToU64(i int64) (uint64, error) {
	if i < 0 {
		return 0, fmt.Errorf("u64 cannot be negative")
	}
	return uint64(i), nil
}
