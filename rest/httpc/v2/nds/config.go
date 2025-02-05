package nds

import "time"

type NacosDiscoveryConfig struct {
	// 连接地址
	IPAddr string `json:",default=127.0.0.1"`
	// 端口
	Port uint64 `json:",default=8848"`
	// 命名空间
	NamespaceID string `json:",default=public"`

	Timeout             time.Duration `json:",default=10000ms"`
	Username            string        `json:",optional"`
	Password            string        `json:",optional"`
	LogLevel            string        `json:",default=info"`
	NotLoadCacheAtStart bool          `json:",default=true"`

	Clusters  []string `json:",optional"`
	GroupName string   `json:",optional"`
}
