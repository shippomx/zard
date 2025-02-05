package nacos

var DEFAULT_NACOS_PATH = "/etc/nacosconfig"

type Config struct {
	Ip   string
	Port uint64

	// 此配置仅动态配置会用到，服务注册发现固定public，不用此字段
	NamespaceId string `json:",default=public"`
	// 此配置仅动态配置会用到
	Group               string `json:",default=DEFAULT_GROUP"`
	DataId              string `json:",default=gate.yaml"`
	Username            string `json:",default=nacos"`
	Password            string `json:",default=nacos"`
	TimeoutMs           uint64 `json:",default=5000"`
	NotLoadCacheAtStart bool   `json:",default=true"`
	LogDir              string `json:",default=/tmp/nacos/log"`
	CacheDir            string `json:",default=/tmp/nacos/cache"`
	LogLevel            string `json:",default=debug"`
}
