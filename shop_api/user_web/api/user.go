package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"shop_api/user_web/global"
	"shop_api/user_web/global/response"
	pb "shop_api/user_web/proto"
	"time"
)

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

	// 调用用户服务
	userSrvClient := pb.NewUserServiceClient(userConn)
	rsp, err := userSrvClient.GetUserList(context.Background(), &pb.PageInfo{
		PageIndex: 1,
		PageSize:  5,
	})

	if err != nil {
		zap.S().Errorw("[GetUserList] 调用 【用户服务失败】", "msg", err.Error())
		HandleGrpcErrorToHttp(err, ctx)
		return
	}

	// 返回结果
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
