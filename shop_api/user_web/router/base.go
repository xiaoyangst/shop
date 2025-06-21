package router

import (
	"github.com/gin-gonic/gin"
	"shop_api/user_web/api"
)

func InitBaseRouter(Router *gin.RouterGroup) {
	BaseRouter := Router.Group("base")
	{
		BaseRouter.GET("captcha", api.GetCaptcha) // 生成验证码
		BaseRouter.POST("send_sms", api.SendSMS)  // 发送验证码
	}
}
