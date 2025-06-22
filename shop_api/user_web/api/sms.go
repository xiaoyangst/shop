package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"math/rand"
	"net/http"
	"shop_api/user_web/forms"
	"shop_api/user_web/global"
	"time"
)

// 自己模拟的短信发送接口，因为短信服务需要付费，所以这里就不实现了
// 更为主要的是，现在各大短信服务都需要多个的证件，比以前要麻烦很多

func SendSMS(c *gin.Context) {
	var req forms.SMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "参数错误"})
		return
	}

	code := generateCode()

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", global.ServerConfig.RedisInfo.Host,
			global.ServerConfig.RedisInfo.Port),
	})

	err := rdb.Set(c, req.Mobile+"_sms_code", code, time.Duration(global.ServerConfig.RedisInfo.Expire)*time.Second).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": code,
			"msg":  "验证码发送失败",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code": code,
			"msg":  "验证码发送成功",
		})
	}
}

func generateCode() string {
	rand.Seed(time.Now().UnixNano())
	return string([]byte{
		byte(rand.Intn(9) + 49),
		byte(rand.Intn(9) + 49),
		byte(rand.Intn(9) + 49),
		byte(rand.Intn(9) + 49),
		byte(rand.Intn(9) + 49),
		byte(rand.Intn(9) + 49),
	})
}
