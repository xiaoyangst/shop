package main

import (
	"fmt"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"shop_api/user_web/global"
	"shop_api/user_web/initialize"
	myvalidator "shop_api/user_web/validator"
)

func main() {

	// 初始化 日志
	initialize.InitLogger()

	// 初始化 配置
	initialize.InitConfig()

	// 初始化 路由信息
	router := initialize.Routers()

	// 初始化 翻译器
	if err := initialize.InitTrans("zh"); err != nil {
		zap.S().Panicf("初始化翻译器错误：%s", err.Error())
		return
	}

	// 注册验证器
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("mobile", myvalidator.ValidateMobile); err != nil {
			zap.S().Panicf("注册手机号码验证器失败：%s", err.Error())
			return
		}
	}

	// 启动服务
	port := global.ServerConfig.Port
	zap.S().Debugf("服务器启动，端口号：%d", port)
	if err := router.Run(fmt.Sprintf(":%d", port)); err != nil {
		zap.S().Panicf("启动失败：%s", err.Error())
		return
	}
}
