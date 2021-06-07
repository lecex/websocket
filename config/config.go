package config

import (
	"github.com/lecex/core/config"
	"github.com/lecex/core/env"
)

// 	Conf 配置
// 	Name // 服务名称
//	Method // 方法
//	Auth // 是否认证授权
//	Policy // 是否认证权限
//	Name // 权限名称
//	Description // 权限解释
var Conf config.Config = config.Config{
	Name:    env.Getenv("MICRO_WEB_NAMESPACE", "go.micro.web.") + "websocket",
	Version: "v1.3.3",
}
