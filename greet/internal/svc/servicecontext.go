// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"zero-demo/greet/internal/config"
	"zero-demo/greet/internal/model"
	"zero-demo/user/rpc/userclient"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config      config.Config
	RedisClient *redis.Redis
	UserModel   model.UserModel
	UserRpc     userclient.User
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)

	return &ServiceContext{
		Config:      c,
		RedisClient: redis.MustNewRedis(c.Redis),
		UserModel:   model.NewUserModel(conn, c.Cache),
		UserRpc:     userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
	}
}
