package i18ngo

import (
	"errors"
	"os"
	"strconv"
	"sync"

	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	t           Translator
	mu          sync.RWMutex
	initialized bool
)

// metric
var (
	requestsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "i18n_requests_total",
			Help: "Total Request number of I18N middeleware",
		},
	)
	ErrorsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "i18n_errors_total",
			Help: "Total Err number of I18N middeleware",
		},
	)
	UpdateTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "i18n_update_total",
			Help: "Total Update number of I18N middeleware",
		},
	)
)

type Translator interface {
	Init(clientConf constant.ClientConfig, servers []constant.ServerConfig) error
	InitByTranslatorConfig(conf *TranslatorConf) error
	LocalizeText(text string, locale LocaleT) (string, error)
	LocalizeSprintf(text string, locale LocaleT, args ...string) (string, error)
	Close()
}

func SetTranslator() error {
	mu.Lock()
	defer mu.Unlock()
	if initialized {
		return errors.New("translator already initialized")
	}
	t = &nacosTranslator
	// get conf by env
	conf := getEnvConf()
	err := t.Init(conf.Client, conf.Servers)
	if err != nil {
		return err
	}
	prometheus.MustRegister(requestsTotal, ErrorsTotal, UpdateTotal)
	initialized = true
	return nil
}

func SetTranslatorByConf(conf *NacosConf) error {
	mu.Lock()
	defer mu.Unlock()

	if initialized {
		return errors.New("translator already initialized")
	}
	t = &nacosTranslator
	err := t.Init(conf.Client, conf.Servers)
	if err != nil {
		return err
	}
	prometheus.MustRegister(requestsTotal, ErrorsTotal, UpdateTotal)
	initialized = true
	return nil
}

func SetTranslatorByConfig(conf *TranslatorConf) error {
	mu.Lock()
	defer mu.Unlock()
	if initialized {
		return errors.New("translator already initialized")
	}
	t = &nacosTranslator
	err := t.InitByTranslatorConfig(conf)
	if err != nil {
		return err
	}
	prometheus.MustRegister(requestsTotal, ErrorsTotal, UpdateTotal)
	initialized = true
	return nil
}

func GetTranslator() Translator {
	mu.RLock()
	defer mu.RUnlock()

	if !initialized {
		return nil
	}
	return t
}

func getEnvConf() *NacosConf {
	host := os.Getenv("NACOS_HOST")
	port := os.Getenv("NACOS_PORT")
	username := os.Getenv("NACOS_USERNAME")
	password := os.Getenv("NACOS_PASSWORD")
	namespaceId := os.Getenv("NACOS_NAMESPACE_ID")
	scheme := os.Getenv("NACOS_SCHEME")
	// dataId := viper.GetString("DATA_ID")
	// fmt.Printf("nacos conf[host:%s,port:%d,username:%s,password:%s,namespaceId:%s,scheme:%s,group:%s]\n", host, port, username, password, namespaceId, scheme, group)
	client := constant.ClientConfig{
		NamespaceId: namespaceId,
		LogLevel:    "info",
		Username:    username,
		Password:    password,
	}
	iport, err := strconv.Atoi(port)
	if err != nil {
		panic(err)
	}
	servers := []constant.ServerConfig{
		{
			IpAddr:      host,
			ContextPath: "/nacos",
			Port:        uint64(iport),
			Scheme:      scheme,
		},
	}
	return &NacosConf{
		Client:  client,
		Servers: servers,
	}
}
