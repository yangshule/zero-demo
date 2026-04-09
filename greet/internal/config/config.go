// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis" // 引入 redis 包)
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf

	Redis      redis.RedisConf
	DataSource string

	Cache cache.CacheConf

	Auth struct {
		AccessSecret string
		AccessExpire int64
	}

	UserRpc zrpc.RpcClientConf
}
