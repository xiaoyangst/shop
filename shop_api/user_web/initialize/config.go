package initialize

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"shop_api/user_web/global"
)

func InitConfig() {
	// 设置文件名（不需要扩展名）
	viper.SetConfigName("config_debug")
	// 设置文件类型
	viper.SetConfigType("yaml")
	// 设置查找路径
	viper.AddConfigPath("user_web/")

	// 读取配置
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	// 将配置映射到结构体
	if err := viper.Unmarshal(&global.ServerConfig); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	zap.S().Infof("配置信息: %v", global.ServerConfig)

	// 动态监视配置文件变化
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		zap.S().Infof("Config file changed: %s", e.Name)
		if err := viper.Unmarshal(&global.ServerConfig); err != nil {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
		zap.S().Infof("配置信息: %v", global.ServerConfig)
	})

}
