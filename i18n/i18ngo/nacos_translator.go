package i18ngo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/common/logger"
	"github.com/nacos-group/nacos-sdk-go/v2/util"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

type NacosTranslator struct {
	mu          sync.Mutex
	defaultType LocaleT
	cache       []sync.Map
	NacosConf   vo.NacosClientParam
	clients     []config_client.IConfigClient
}

var nacosTranslator = NacosTranslator{
	cache: make([]sync.Map, len(Locales)),
}

var RE = regexp.MustCompile(`\{[a-zA-Z0-9_\-.\s]+?\}|\%s`)

func (t *NacosTranslator) StoreLocale(locale LocaleT, m *sync.Map) error {
	index := slices.Index(Locales, locale)
	if index == -1 {
		return errors.New("locale not support " + string(locale))
	}
	if len(t.cache) <= index {
		return errors.New("overflow t.cache")
	}
	m.Range(func(k, v interface{}) bool {
		t.cache[index].Store(k, v)
		return true
	})
	return nil
}

func (t *NacosTranslator) UpdateLocale(locale LocaleT, m map[string]string) error {
	index := slices.Index(Locales, locale)
	if index == -1 {
		return errors.New("locale  not found " + string(locale))
	}
	for k, v := range m {
		old, exists := t.cache[index].Load(k)
		if !exists {
			UpdateTotal.Inc()
			t.cache[index].Store(k, v)
			continue
		}
		if old.(string) == v {
			continue
		}
		UpdateTotal.Inc()
		t.cache[index].Store(k, v)
	}
	return nil
}

func UnmarshalToSyncMap(config string) (*sync.Map, error) {
	var configMap map[string]string
	if err := json.Unmarshal([]byte(config), &configMap); err != nil {
		return nil, err
	}
	var m sync.Map
	for k, v := range configMap {
		m.Store(k, v)
	}
	return &m, nil
}

func (t *NacosTranslator) InitByTranslatorConfig(conf *TranslatorConf) error {
	util.LocalIP() // fuck sdk data race bug
	if len(conf.DefaultLocale) != 0 {
		t.defaultType = conf.DefaultLocale
	}
	if len(conf.DataIDs) == 0 {
		conf.DataIDs = append(conf.DataIDs, Locales...)
	}

	if len(conf.Groups) == 0 {
		conf.Groups = []string{"DEFAULT_GROUP"}
	}
	t.NacosConf = vo.NacosClientParam{
		ClientConfig:  &conf.ClientConfig,
		ServerConfigs: conf.Servers,
	}

	wg := sync.WaitGroup{}
	var errMap sync.Map
	for _, group := range conf.Groups {
		group = "I18N_" + strings.ToUpper(group)
		for _, dataID := range conf.DataIDs {
			dataIDStr := string(dataID) + ".json"
			wg.Add(1)

			go func(group, dataID string) {
				defer wg.Done()
				if !t.ExistsTranslator(group, dataID) {
					return
				}
				if err := t.SetupTranslator(group, dataID); err != nil {
					errMap.Store(dataID, err)
					return
				}
			}(group, dataIDStr)
		}
	}
	wg.Wait()
	agerros := []error{}
	errMap.Range(func(key, value interface{}) bool {
		agerros = append(agerros, fmt.Errorf("dataid: %s ", key.(string)), value.(error))
		return true
	})
	if len(agerros) > 0 {
		logger.Debug("init translator error", agerros)
	}
	return nil
}

func (t *NacosTranslator) Client() (config_client.IConfigClient, error) {
	cc := *t.NacosConf.ClientConfig
	sc := []constant.ServerConfig{}
	sc = append(sc, t.NacosConf.ServerConfigs...)
	vv := vo.NacosClientParam{
		ClientConfig:  &cc,
		ServerConfigs: sc,
	}
	return clients.NewConfigClient(vv)
}

func (t *NacosTranslator) Init(clientConf constant.ClientConfig, servers []constant.ServerConfig) error {
	groups := os.Getenv("TRANSLATOR_GROUPS")
	if groups == "" {
		groups = "DEFAULT_GROUP"
	}
	dataIDs := os.Getenv("TRANSLATOR_DATA_IDS")
	dataIDSlice := []string{}
	if dataIDs == "" {
		for _, locale := range Locales {
			dataIDSlice = append(dataIDSlice, string(locale)+".json")
		}
	} else {
		dataIDSlice = strings.Split(dataIDs, ",")
	}
	t.NacosConf = vo.NacosClientParam{
		ClientConfig:  &clientConf,
		ServerConfigs: servers,
	}

	wg := sync.WaitGroup{}
	var errMap sync.Map
	for _, group := range strings.Split(groups, ",") {
		for _, dataID := range dataIDSlice {
			wg.Add(1)
			go func(group, dataID string) {
				defer wg.Done()
				if !t.ExistsTranslator(group, dataID) {
					return
				}
				err := t.SetupTranslator(group, dataID)
				if err != nil {
					errMap.Store(dataID, err)
					return
				}
			}(group, dataID)

		}
	}
	wg.Wait()
	agerros := []error{}
	errMap.Range(func(key, value interface{}) bool {
		agerros = append(agerros, fmt.Errorf("dataid: %s ", key.(string)), value.(error))
		return true
	})
	if len(agerros) > 0 {
		return errors.Join(agerros...)
	}
	return nil
}

func (t *NacosTranslator) Close() {
	for _, m := range t.clients {
		m.CloseClient()
	}
}

func (t *NacosTranslator) LocalizeText(text string, locale LocaleT) (string, error) {
	requestsTotal.Inc()
	index := slices.Index(Locales, locale)
	if index == -1 {
		ErrorsTotal.Inc()
		return "", fmt.Errorf("cache not found for locale %s", locale)
	}
	v, ok := t.cache[index].Load(text)
	if !ok {
		if t.defaultType != "" && t.defaultType != locale {
			return t.LocalizeText(text, t.defaultType)
		}
		ErrorsTotal.Inc()
		return "", fmt.Errorf("cache not found for text %s", text)
	}

	return v.(string), nil
}

// LocalizeSprintf Localizes the given text and formats it with the given arguments.
//
// The function takes a text string, a locale string, and a variable number of string arguments.
// It returns the localized and formatted text and an error if the operation fails.
func (t *NacosTranslator) LocalizeSprintf(text string, locale LocaleT, args ...string) (string, error) {
	content, err := t.LocalizeText(text, locale)
	if err != nil {
		return "", err
	}
	return Sprintf(content, nil, args...), nil
}

// LocalizeInfof Localizes the given text and replaces placeholders with the given key-value arguments.
//
// The function takes a text string, a locale string, and a variable number of key-value arguments.
// It returns the localized text with replaced placeholders and an error if the operation fails.
func (t *NacosTranslator) LocalizeInfof(text string, locale LocaleT, kvargs ...string) (string, error) {
	if len(kvargs)%2 != 0 {
		return "", errors.New("infof requires an even number of arguments")
	}

	content, err := t.LocalizeText(text, locale)
	if err != nil {
		return "", err
	}

	mp := make(map[string]string, len(kvargs)/2)
	for i := 0; i < len(kvargs); i += 2 {
		mp[kvargs[i]] = kvargs[i+1]
	}

	return Sprintf(content, mp), nil
}

// Sprintf formats a string with the given key-value arguments and positional arguments.
//
// The function takes a format string, a key-value map, and a variable number of positional string arguments.
// It returns the formatted string.
func Sprintf(format string, kv map[string]string, args ...string) string {
	argIndex := 0
	return RE.ReplaceAllStringFunc(format, func(match string) string {
		if argIndex < len(args) && match == "%s" {
			replacement := fmt.Sprintf("%v", args[argIndex])
			argIndex++
			return replacement
		}
		if strings.HasPrefix(match, "{") && strings.HasSuffix(match, "}") {
			inStr := match[1 : len(match)-1]
			index, err := strconv.Atoi(inStr)
			if err == nil && index < len(args) && index >= 0 {
				return fmt.Sprintf("%v", args[index])
			}
			if v, ok := kv[inStr]; ok {
				return v
			}
		}
		return match
	})
}

func (t *NacosTranslator) UpdateCache(dataID string, config string) error {
	if strings.HasSuffix(dataID, ".json") {
		locale := strings.TrimSuffix(dataID, ".json")
		m, err := UnmarshalToSyncMap(config)
		if err != nil {
			return err
		}
		err = t.UpdateLocale(LocaleT(locale), SyncMapToMap(m))
		if err != nil {
			return err
		}
	} else {
		return errors.New("unknown file type")
	}
	return nil
}

func SyncMapToMap(m *sync.Map) map[string]string {
	result := map[string]string{}
	m.Range(func(k, v interface{}) bool {
		result[k.(string)] = v.(string)
		return true
	})
	return result
}

func (t *NacosTranslator) InitCache(dataID string, config string) error {
	if strings.HasSuffix(dataID, ".json") {
		locale := strings.TrimSuffix(dataID, ".json")
		m, err := UnmarshalToSyncMap(config)
		if err != nil {
			return err
		}
		err = t.StoreLocale(LocaleT(locale), m)
		if err != nil {
			return err
		}
	} else {
		return errors.New("unknown file type")
	}
	return nil
}

func (t *NacosTranslator) ExistsTranslator(group, dataID string) bool {
	client, err := t.Client()
	if err != nil {
		return false
	}

	page, err := client.SearchConfig(vo.SearchConfigParam{
		Group:  group, // 默认group
		DataId: dataID,
		Search: "accurate",
	})
	if err != nil {
		fmt.Println(err)
		return false
	}
	for _, config := range page.PageItems {
		if config.DataId == dataID && config.Group == group {
			return true
		}
	}
	return false
}

func (t *NacosTranslator) SetupTranslator(group, dataID string) error {
	client, err := t.Client()
	if err != nil {
		return err
	}
	t.mu.Lock()
	t.clients = append(t.clients, client)
	t.mu.Unlock()
	config, err := client.GetConfig(vo.ConfigParam{
		DataId: dataID, // 配置文件名
		Group:  group,  // 默认group
	})
	if err != nil {
		return err
	}
	err = t.InitCache(dataID, config)
	if err != nil {
		return err
	}
	err = client.ListenConfig(vo.ConfigParam{
		DataId: dataID, // 配置文件名
		Group:  group,  // 默认group
		OnChange: func(_, _, dataId, data string) {
			err := t.UpdateCache(dataId, data)
			if err != nil {
				logger.Errorf("update cache error: %v", err)
			}
		},
	})
	if err != nil {
		return err
	}

	return nil
}
