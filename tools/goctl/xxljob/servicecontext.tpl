package svc

import (
	"{{.projectPackage}}/internal/config"
)

type ServiceContext struct {
	Config config.Config
	// EmailClient IEmailClient.
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
