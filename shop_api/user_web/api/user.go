package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"shop_api/user_web/forms"
	"shop_api/user_web/global"
	"shop_api/user_web/global/response"
	"shop_api/user_web/middlewares"
	"shop_api/user_web/models"
	pb "shop_api/user_web/proto"
	"strconv"
	"strings"
	"time"
)

func removeTopStruct(fileds map[string]string) map[string]string {
	rsp := map[string]string{}
	for field, err := range fileds {
		rsp[field[strings.Index(field, ".")+1:]] = err
	}
	return rsp
}

func HandleValidatorError(c *gin.Context, err error) {
	var errs validator.ValidationErrors
	ok := errors.As(err, &errs)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"msg": err.Error(),
		})
	}
	c.JSON(http.StatusBadRequest, gin.H{
		"error": removeTopStruct(errs.Translate(global.Trans)),
	})
	return
}

// HandleGrpcErrorToHttp 将grpc的code转换成http的状态码
func HandleGrpcErrorToHttp(err error, c *gin.Context) {
	if err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{
					"msg": e.Message(),
				})
			case codes.Internal:
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": "内部错误",
				})
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, gin.H{
					"msg": "参数错误",
				})
			case codes.Unavailable:
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": "用户服务不可用",
				})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": "其它错误",
				})
			}
		}

	}
}

func GetUserList(ctx *gin.Context) {
	zap.S().Debugf("获取用户列表页")

	// 连接用户服务
	userConn, err := grpc.Dial(fmt.Sprintf("%s:%d", global.ServerConfig.UserSrvInfo.Host, global.ServerConfig.UserSrvInfo.Port), grpc.WithInsecure())
	if err != nil {
		zap.S().Errorw("[GetUserList] 连接 【用户服务失败】", "msg", err.Error())
	}

	defer userConn.Close()

	// 创建 grpc 客户端
	userSrvClient := pb.NewUserServiceClient(userConn)

	// 由用户自定义 pn 和 psize
	// web --> gin --> grpc --> 在具体的服务中处理
	pn := ctx.DefaultQuery("pn", "0")
	pnInt, _ := strconv.Atoi(pn)
	pSize := ctx.DefaultQuery("psize", "1")
	pSizeInt, _ := strconv.Atoi(pSize)
	zap.S().Infof("获取用户列表页: 页码: %s, 页大小: %s", pn, pSize)

	// 调用 grpc 服务，获取用户列表
	rsp, err := userSrvClient.GetUserList(context.Background(), &pb.PageInfo{
		PageIndex: uint32(pnInt),
		PageSize:  uint32(pSizeInt),
	})

	if err != nil {
		zap.S().Errorw("[GetUserList] 调用 【用户服务失败】", "msg", err.Error())
		HandleGrpcErrorToHttp(err, ctx)
		return
	}

	// 把从 grpc 服务中获取的数据，转换为 web 端需要的数据
	result := make([]interface{}, 0)
	for _, user := range rsp.Data {
		user := response.UserResp{
			Id:       user.Id,
			NickName: user.Nickname,
			Birthday: response.JsonTime(time.Unix(int64(user.Birthday), 0)),
			Gender:   user.Gender,
			Mobile:   user.Mobile,
		}
		result = append(result, user)
	}

	ctx.JSON(http.StatusOK, result)
}

func PassWordLogin(c *gin.Context) {
	passwordLoginForm := forms.PasswordLoginForm{}
	// 绑定表单数据到结构体
	if err := c.ShouldBindJSON(&passwordLoginForm); err != nil {
		HandleValidatorError(c, err)
		return
	}

	// 验证码验证
	if !store.Verify(passwordLoginForm.CaptchaId, passwordLoginForm.Captcha, true) {
		c.JSON(http.StatusBadRequest, map[string]string{
			"captcha": "验证码错误",
		})
		return
	}

	// 连接用户服务
	userConn, err := grpc.Dial(fmt.Sprintf("%s:%d", global.ServerConfig.UserSrvInfo.Host, global.ServerConfig.UserSrvInfo.Port), grpc.WithInsecure())
	if err != nil {
		zap.S().Errorw("[PassWordLogin] 连接 【用户服务失败】", "msg", err.Error())
	}

	defer userConn.Close()

	// 创建 grpc 客户端
	userSrvClient := pb.NewUserServiceClient(userConn)

	// 获取手机号，判断用户是否存在
	rsp, err := userSrvClient.GetUserByMobile(context.Background(), &pb.MobileRequest{
		Mobile: passwordLoginForm.Mobile,
	})

	if err != nil { // 用户不存在
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.NotFound:
				c.JSON(http.StatusBadRequest, map[string]string{
					"mobile": "用户不存在",
				})
			default:
				c.JSON(http.StatusInternalServerError, map[string]string{
					"mobile": "登录失败",
				})
			}
			return
		}
	} else {
		// 验证密码
		pwdRsp, pwdErr := userSrvClient.CheckPassword(context.Background(), &pb.CheckPasswordInfo{
			Password:          passwordLoginForm.Password,
			EncryptedPassword: rsp.Password,
		})
		if pwdErr != nil { // 密码验证失败
			c.JSON(http.StatusInternalServerError, map[string]string{
				"password": "登录失败",
			})
		} else {
			if !pwdRsp.Success {
				c.JSON(http.StatusBadRequest, map[string]string{
					"password": "密码错误",
				})
			} else {
				// 登录成功, 生成 token
				j := middlewares.NewJWT()
				claims := models.CustomClaims{
					ID:          uint(rsp.Id),
					NickName:    rsp.Nickname,
					AuthorityId: rsp.Role,
					StandardClaims: jwt.StandardClaims{
						NotBefore: time.Now().Unix(),
						ExpiresAt: time.Now().Unix() + 60*60*24*30, // token 有效期为 30 天
						Issuer:    "xy",
					},
				}
				token, err := j.CreateToken(claims)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"msg": "生成 token 失败",
					})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"token":      token,
					"id":         rsp.Id,
					"nickname":   rsp.Nickname,
					"expires_at": (time.Now().Unix() + 60*60*24*30) * 1000,
				})
			}
		}
	}

}
