package initialize

import (
	"github.com/gin-gonic/gin"
	"shop-api/user_web/router"
)

func Routers() *gin.Engine {
	r := gin.Default()
	ApiRouter := r.Group("/u/v1")
	router.InitUserRouter(ApiRouter)

	return r
}
