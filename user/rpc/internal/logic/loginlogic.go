package logic

import (
	"context"

	"zero-demo/user/rpc/internal/svc"
	"zero-demo/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *user.LoginReq) (*user.LoginResp, error) {
	// 1. 打印接收到的请求参数 (方便我们调试)
	l.Logger.Infof("RPC 收到了登录请求！尝试登录的账号: %s", in.Username)

	// 2. 模拟数据库查询校验
	if in.Username == "admin" && in.Password == "123456" {
		return &user.LoginResp{
			Success: true,
			Token:   "super-secret-token-888",
		}, nil
	}

	// 3. 账号密码错误的情况
	return &user.LoginResp{
		Success: false,
		Token:   "",
	}, nil
}
