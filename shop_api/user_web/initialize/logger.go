package initialize

import (
	"go.uber.org/zap"
)

func InitLogger() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
}

// 用户如何使用日志
// zap.L().Info("success") 或者  zap.S().Info("success")
