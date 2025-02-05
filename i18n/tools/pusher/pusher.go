package pusher

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/common/logger"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/cobra"
)

var (
	IpAddr    string
	Port      uint64
	Scheme    string
	user      string
	password  string
	dir       string
	tw        string
	dist      bool
	namespace string
)

func Init() *cobra.Command {
	pushCmd := cobra.Command{
		Use:   "push",
		Short: "Push the i18n config to nacos.",
		Long:  ``,
	}

	pushCmd.Flags().StringVarP(&IpAddr, "ip", "i", "", "nacos ip")
	pushCmd.Flags().Uint64VarP(&Port, "port", "p", 8848, "nacos port")
	pushCmd.Flags().StringVarP(&Scheme, "scheme", "s", "http", "nacos scheme")
	pushCmd.Flags().StringVarP(&user, "user", "u", "", "nacos user")
	pushCmd.Flags().StringVarP(&password, "password", "a", "", "nacos password")
	pushCmd.Flags().StringVarP(&dir, "dir", "d", "", "i18n config dir")
	pushCmd.Flags().StringVarP(&tw, "tw_config", "t", "", "tw config dir")
	pushCmd.Flags().BoolVarP(&dist, "dist", "b", false, "dist config dir")
	pushCmd.Flags().StringVarP(&namespace, "namespace", "n", "i18n", "nacos namespace")
	return &pushCmd
}

func Push() error {
	logger.Info("Pushing I18N config to nacos...")

	serverConfig := []constant.ServerConfig{
		{
			IpAddr: IpAddr,
			Scheme: "http",
			Port:   Port,
		},
	}
	clientConfig := constant.NewClientConfig(
		constant.WithUsername(user),
		constant.WithPassword(password),
		constant.WithNamespaceId(namespace),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogLevel("debug"),
	)
	clientConfig.AppendToStdout = true

	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  clientConfig,
			ServerConfigs: serverConfig,
		},
	)
	if err != nil {
		logger.Error(err)
		return err
	}
	err = filepath.WalkDir(dir, func(pathName string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return err
		}
		dataID := path.Base(pathName)
		group := path.Base(path.Dir(pathName))
		if strings.Contains(dataID, "_") {
			dataIDSlice := strings.Split(dataID, "_")
			dataID = dataIDSlice[1]
			if dist {
				group = dataIDSlice[0]
			}
		}
		group = "I18N_" + strings.ToUpper(group)

		content, err := os.ReadFile(pathName)
		if err != nil {
			return err
		}

		logger.Infof("Pushing file: %s to group: %s, dataID: %s", pathName, group, dataID)
		fmt.Println("Pushing file: " + pathName + " to group: " + group + ", dataID: " + dataID)

		err = publishToNacos(configClient, group, dataID, pathName, content)
		if err != nil {
			return err
		}
		if tw != "" && strings.HasSuffix(pathName, "_zh.json") {
			newContent, err := genTWContent(tw, content)
			if err != nil {
				return err
			}
			dataID = strings.TrimSuffix(dataID, "zh.json") + "tw.json"
			err = publishToNacos(configClient, group, dataID, pathName, newContent)
			if err != nil {
				return err
			}

			fmt.Println("Pushing file: " + pathName + " to group: " + group + ", dataID: " + dataID)
			logger.Infof("Pushing file: %s to group: %s, dataID: %s", pathName, group, dataID)

		}
		return err
	})
	if err != nil {
		return err
	}
	logger.Info("Push success")

	return err
}

func publishToNacos(configClient config_client.IConfigClient, group, dataID, pathName string, content []byte) error {
	ok, err := configClient.PublishConfig(vo.ConfigParam{
		DataId:  dataID,
		Group:   group,
		Content: string(content),
		Type:    "json",
	})
	if ok {
		logger.Infof("Pushed file: %s", pathName)
	} else {
		return errors.New("failed to push file: " + pathName)
	}
	return err
}

func genTWContent(confPath string, content []byte) ([]byte, error) {
	caseContent, err := os.ReadFile(confPath)
	if err != nil {
		return nil, err
	}
	keyvaluemap := make(map[string]string)

	err = json.Unmarshal(caseContent, &keyvaluemap)
	if err != nil {
		return nil, err
	}
	contentMap := make(map[string]string)
	err = json.Unmarshal(content, &contentMap)
	if err != nil {
		return nil, err
	}
	for k, v := range contentMap {
		for key, value := range keyvaluemap {
			v = strings.ReplaceAll(v, key, value)
		}
		contentMap[k] = v
	}
	// 再次序列化,换行
	res, err := json.MarshalIndent(contentMap, "", "  ")
	if err != nil {
		return nil, err
	}
	return res, nil
}
