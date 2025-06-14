package main

import (
	"fmt"
	"go.uber.org/zap"
	"shop_api/user_web/global"
	"shop_api/user_web/initialize"
)

func main() {

	// 初始化 日志
	initialize.InitLogger()

	// 初始化 配置
	initialize.InitConfig()

	// 初始化 路由信息
	router := initialize.Routers()

	// 启动服务
	port := global.ServerConfig.Port
	zap.S().Debugf("服务器启动，端口号：%d", port)
	if err := router.Run(fmt.Sprintf(":%d", port)); err != nil {
		zap.S().Panicf("启动失败：%s", err.Error())
		return
	}
}
