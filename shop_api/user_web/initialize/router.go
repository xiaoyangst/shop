package initialize

import (
	"github.com/gin-gonic/gin"
	"shop_api/user_web/middlewares"
	"shop_api/user_web/router"
)

func Routers() *gin.Engine {
	r := gin.Default()
	// 配置跨域
	r.Use(middlewares.Cors())
	// 配置路由组
	ApiRouter := r.Group("/u/v1")
	router.InitUserRouter(ApiRouter)
	router.InitBaseRouter(ApiRouter)

	return r
}
