package svc

import (
	"eiot/pkg/config"
	"eiot/pkg/mqtt"
)

// ServiceContext 业务上下文
type ServiceContext struct {
	Config *config.Config
	EMQX   *mqtt.EMQXClient
}

// NewServiceContext 构造 ServiceContext
func NewServiceContext(c *config.Config) *ServiceContext {
	return &ServiceContext{Config: c}
}
