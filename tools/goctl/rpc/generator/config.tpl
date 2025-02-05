package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/shippomx/zard/zrpc"{{ if .gateway }}
	"github.com/shippomx/zard/rest"{{end}}
	"github.com/shippomx/zard/core/conf"
	"github.com/shippomx/zard/core/logx"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"

)

type Config struct {
	zrpc.RpcServerConf{{ if .gateway }}
	RestConf rest.RestConf{{end}}
}


func (c *Config) InitNacosEnv() error {
    // 若设置了Nacos相关信息的环境变量，则使用环境变量的配置
    if len(os.Getenv("NACOS_HOST")) > 0 {
        c.Nacos.Ip = os.Getenv("NACOS_HOST")
        // 从环境变量获取NACOS_PORT, 并转换为uint64
        portUint64, err := strconv.ParseUint(os.Getenv("NACOS_PORT"), 10, 64)
        if err != nil {
            logx.Must(err)
            return err
        }
        c.Nacos.Port = portUint64
        c.Nacos.Username = os.Getenv("NACOS_USERNAME")
        c.Nacos.Password = os.Getenv("NACOS_PASSWORD")
        c.Nacos.NamespaceId = os.Getenv("NACOS_NAMESPACE")
        c.Nacos.DataId = os.Getenv("NACOS_DATA_ID")
        c.Nacos.Group = os.Getenv("NACOS_GROUP")
        if c.Nacos.Group == "" {
            c.Nacos.Group = "DEFAULT_GROUP"
        }
        return nil
    }
    return errors.New("nacos env doesn't exist")
}

// MustNacosConf 初始化配置.
func (c *Config) MustNacosConf() error {
    logx.Info("init nacos conf")
    // server conf
    sc := []constant.ServerConfig{
        {
            IpAddr: c.Nacos.Ip,
            Port:   c.Nacos.Port,
        },
    }
    fmt.Println("find nacos config ", c.Nacos)
    // client conf
    cc := constant.ClientConfig{
        NamespaceId:         c.Nacos.NamespaceId, // namespace id
        Username:            c.Nacos.Username,
        Password:            c.Nacos.Password,
        TimeoutMs:           c.Nacos.TimeoutMs,
        NotLoadCacheAtStart: c.Nacos.NotLoadCacheAtStart,
        LogDir:              c.Nacos.LogDir,
        CacheDir:            c.Nacos.CacheDir,
        LogLevel:            c.Nacos.LogLevel,
    }
    // init client
    client, err := clients.NewConfigClient(
        vo.NacosClientParam{
            ClientConfig:  &cc,
            ServerConfigs: sc,
        },
    )
    if err != nil {
        logx.Errorf("init nacos client err:%s", err.Error())
        return err
    }
    // get config
    content, err := client.GetConfig(vo.ConfigParam{
        DataId: c.Nacos.DataId,  // 配置文件名
        Group:  c.Nacos.Group,
    })
    if err != nil {
        logx.Errorf("get nacos config err:%v", err)
        return err
    }
	if err = conf.LoadFromYamlBytes([]byte(content), c); err != nil {
		logx.Errorf("load nacos config err:%s", err.Error())
		return err
	}
    err = client.ListenConfig(vo.ConfigParam{
		DataId: c.Nacos.DataId,
		Group:  "DEFAULT_GROUP",
		OnChange: func(_, _, _, data string) {
			logx.Infof("nacos conf change:%s", data)

			if err = conf.LoadFromYamlBytes([]byte(data), c); err != nil {
				logx.Errorf("update dynamic conf err:%s", err.Error())
				return
			}
		},
	})

    return err
}
