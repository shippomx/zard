package rest

import (
	"fmt"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/shippomx/zard/core/nacos"

	"github.com/shippomx/zard/core/service"
)

type (
	// MiddlewaresConf is the config of middlewares.
	MiddlewaresConf struct {
		Trace      bool `json:",default=true"`
		Log        bool `json:",default=true"`
		Prometheus bool `json:",default=true"`
		MaxConns   bool `json:",default=true"`
		Breaker    bool `json:",default=true"`
		Shedding   bool `json:",default=false"`
		Timeout    bool `json:",default=true"`
		Recover    bool `json:",default=true"`
		Metrics    bool `json:",default=true"`
		MaxBytes   bool `json:",default=true"`
		Gunzip     bool `json:",default=true"`
		UserID     bool `json:",default=true"`
	}

	// A PrivateKeyConf is a private key config.
	PrivateKeyConf struct {
		Fingerprint string
		KeyFile     string
	}

	// A SignatureConf is a signature config.
	SignatureConf struct {
		Strict      bool          `json:",default=false"`
		Expiry      time.Duration `json:",default=1h"`
		PrivateKeys []PrivateKeyConf
	}

	// A RestConf is a http service config.
	// Why not name it as Conf, because we need to consider usage like:
	//  type Config struct {
	//     zrpc.RpcConf
	//     rest.RestConf
	//  }
	// if with the name Conf, there will be two Conf inside Config.
	RestConf struct {
		service.ServiceConf
		Host            string `json:",default=0.0.0.0"`
		Port            int
		CertFile        string `json:",optional"`
		KeyFile         string `json:",optional"`
		Verbose         bool   `json:",optional"`
		EnableAccessLog bool   `json:",optional,default=true"` // enable/disable access log.
		MaxConns        int    `json:",default=10000"`
		MaxBytes        int64  `json:",default=1048576"`
		// milliseconds
		// nolint:all
		Timeout int64 `json:",default=3000"`
		// nolint:all
		CpuThreshold int64         `json:",default=0,range=[0:1000)"`
		Signature    SignatureConf `json:",optional"`
		// There are default values for all the items in Middlewares.
		Middlewares MiddlewaresConf
		// TraceIgnorePaths is paths blacklist for trace middleware.
		TraceIgnorePaths []string `json:",optional"`
		NacosDiscovery   bool     `json:",default=false"`
	}
)

// HasNacos checks if there is nacos settings in config.
func (rc RestConf) HasNacos() bool {
	return len(rc.Nacos.Ip) > 0
}

func (rc RestConf) RegistNacos() error {
	// register service to nacos
	sc := []constant.ServerConfig{
		*constant.NewServerConfig(rc.Nacos.Ip, rc.Nacos.Port),
	}

	cc := &constant.ClientConfig{
		NamespaceId:         "public",
		Username:            rc.Nacos.Username,
		Password:            rc.Nacos.Password,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "debug",
	}

	opts := nacos.NewNacosConfig(rc.Name, fmt.Sprintf("%s:%d", rc.Host, rc.Port), sc, cc)
	err := nacos.RegisterService(opts)
	if err != nil {
		return err
	}

	return nil
}
