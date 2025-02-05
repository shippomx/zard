package xxljob

import "github.com/shippomx/zard/core/service"

type Conf struct {
	service.ServiceConf
	ServerAddr   string
	AccessToken  string
	ExecutorIP   string `json:",optional"`
	ExecutorPort string `json:",default=8082"`
}
