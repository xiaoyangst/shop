package config

type UserSrvConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type JWTConfig struct {
	SigningKey  string `mapstructure:"key"`
	ExpiresTime int64  `mapstructure:"expires"`
	Issuer      string `mapstructure:"issuer"`
}

type RedisConfig struct {
	Host   string `mapstructure:"host"`
	Port   int    `mapstructure:"port"`
	Expire int    `mapstructure:"expire"`
}

type ServerConfig struct {
	Host        string        `mapstructure:"name"`
	Port        int           `mapstructure:"port"`
	RedisInfo   RedisConfig   `mapstructure:"redis"`
	JWTInfo     JWTConfig     `mapstructure:"jwt"`
	UserSrvInfo UserSrvConfig `mapstructure:"user_srv"`
}
