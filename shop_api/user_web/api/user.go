package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
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

// 创建 gRPC 用户服务客户端
func getUserSrvClient() (pb.UserServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", global.ServerConfig.UserSrvInfo.Host, global.ServerConfig.UserSrvInfo.Port), grpc.WithInsecure())
	if err != nil {
		zap.S().Errorw("[gRPC连接失败]", "msg", err.Error())
		return nil, nil, err
	}
	return pb.NewUserServiceClient(conn), conn, nil
}

// 创建 JWT Token
func createToken(user *pb.UserInfoResponse) (string, int64, error) {
	j := middlewares.NewJWT()
	expiresAt := time.Now().Unix() + 60*60*24*30
	claims := models.CustomClaims{
		ID:          uint(user.Id),
		NickName:    user.Nickname,
		AuthorityId: user.Role,
		StandardClaims: jwt.StandardClaims{
			NotBefore: time.Now().Unix(),
			ExpiresAt: expiresAt,
			Issuer:    "xy",
		},
	}
	token, err := j.CreateToken(claims)
	return token, expiresAt * 1000, err
}

func GetUserList(c *gin.Context) {
	zap.S().Debug("获取用户列表页")
	userSrvClient, conn, err := getUserSrvClient()
	if err != nil {
		HandleGrpcErrorToHttp(err, c)
		return
	}
	defer conn.Close()

	pn, _ := strconv.Atoi(c.DefaultQuery("pn", "0"))
	psize, _ := strconv.Atoi(c.DefaultQuery("psize", "1"))

	rsp, err := userSrvClient.GetUserList(context.Background(), &pb.PageInfo{
		PageIndex: uint32(pn),
		PageSize:  uint32(psize),
	})
	if err != nil {
		HandleGrpcErrorToHttp(err, c)
		return
	}

	users := make([]interface{}, 0)
	for _, u := range rsp.Data {
		users = append(users, response.UserResp{
			Id:       u.Id,
			NickName: u.Nickname,
			Birthday: response.JsonTime(time.Unix(int64(u.Birthday), 0)),
			Gender:   u.Gender,
			Mobile:   u.Mobile,
		})
	}
	c.JSON(http.StatusOK, users)
}

func PassWordLogin(c *gin.Context) {
	var form forms.PasswordLoginForm
	if err := c.ShouldBindJSON(&form); err != nil {
		HandleValidatorError(c, err)
		return
	}

	// 图形验证码验证
	if !store.Verify(form.CaptchaId, form.Captcha, true) {
		c.JSON(http.StatusBadRequest, gin.H{"captcha": "验证码错误"})
		return
	}

	userSrvClient, conn, err := getUserSrvClient()
	if err != nil {
		HandleGrpcErrorToHttp(err, c)
		return
	}
	defer conn.Close()

	// 获取用户信息
	user, err := userSrvClient.GetUserByMobile(context.Background(), &pb.MobileRequest{Mobile: form.Mobile})
	if err != nil {
		HandleGrpcErrorToHttp(err, c)
		return
	}

	// 验证密码
	pwdCheck, err := userSrvClient.CheckPassword(context.Background(), &pb.CheckPasswordInfo{
		Password:          form.Password,
		EncryptedPassword: user.Password,
	})
	if err != nil || !pwdCheck.Success {
		c.JSON(http.StatusBadRequest, gin.H{"password": "密码错误"})
		return
	}

	token, exp, err := createToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "生成 token 失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      token,
		"id":         user.Id,
		"nickname":   user.Nickname,
		"expires_at": exp,
	})
}

func Register(c *gin.Context) {
	var form forms.RegisterForm
	if err := c.ShouldBindJSON(&form); err != nil {
		HandleValidatorError(c, err)
		return
	}
	// 手机号验证码验证
	if !validateSmsCode(form.Mobile, form.Code) {
		c.JSON(http.StatusBadRequest, gin.H{"code": "验证码错误"})
		return
	}

	userSrvClient, conn, err := getUserSrvClient()
	if err != nil {
		HandleGrpcErrorToHttp(err, c)
		return
	}
	defer conn.Close()

	_, err = userSrvClient.CreateUser(context.Background(), &pb.CreateUserInfo{
		Nickname: form.Mobile,
		Password: form.Password,
		Mobile:   form.Mobile,
	})
	if err != nil {
		HandleGrpcErrorToHttp(err, c)
		return
	}

	user, err := userSrvClient.GetUserByMobile(context.Background(), &pb.MobileRequest{Mobile: form.Mobile})
	if err != nil {
		HandleGrpcErrorToHttp(err, c)
		return
	}

	token, exp, err := createToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "生成 token 失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      token,
		"id":         user.Id,
		"nickname":   user.Nickname,
		"expires_at": exp,
	})

	clearSmsCode(form.Mobile)
}

func validateSmsCode(mobile, code string) bool {
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", global.ServerConfig.RedisInfo.Host, global.ServerConfig.RedisInfo.Port),
	})
	defer rdb.Close()

	val, err := rdb.Get(context.Background(), mobile+"_sms_code").Result()
	return err == nil && val == code
}

func clearSmsCode(mobile string) {
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", global.ServerConfig.RedisInfo.Host, global.ServerConfig.RedisInfo.Port),
	})
	defer rdb.Close()

	if err := rdb.Del(context.Background(), mobile+"_sms_code").Err(); err != nil {
		zap.S().Errorw("[Register] 删除验证码失败", "msg", err.Error())
	}
}
