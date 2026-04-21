package config

import "github.com/zeromicro/go-zero/rest"

type ProxyConfig struct {
	Enable             bool   `json:",default=false"`
	Http               string `json:",optional"`
	Https              string `json:",optional"`
	NoProxy            string `json:",optional"`
	InsecureSkipVerify bool   `json:",default=false"`
}

type Config struct {
	rest.RestConf
	Auth  struct { // JWT 认证需要的密钥和过期时间配置
		AccessSecret string
		AccessExpire int64
	}
	Proxy ProxyConfig `json:",optional"`
}

var (
	Version   string
	BuildDate string
)
