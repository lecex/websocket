package main

import (
	// 公共引入

	"github.com/micro/go-micro/v2/util/log"
	"github.com/micro/go-micro/v2/web"

	"github.com/lecex/core/env"
	_ "github.com/lecex/core/plugins"
	"github.com/lecex/websocket/config"
	"github.com/lecex/websocket/handler"
)

func main() {
	var Conf = config.Conf
	service := web.NewService(
		web.Name(Conf.Name),
		web.Version(Conf.Version),
		web.Address(env.Getenv("SERVER_ADDRESS", ":8082")),
	)
	service.Init()
	// 注册服务
	handler.Register(service)
	// Run the server
	log.Fatal("serviser run ... Version:" + Conf.Version)
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
