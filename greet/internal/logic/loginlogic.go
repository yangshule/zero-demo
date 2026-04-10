// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"zero-demo/greet/internal/svc"
	"zero-demo/greet/internal/types"
	"zero-demo/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	// todo: add your logic here and delete this line
	rpcResp, err := l.svcCtx.UserRpc.Login(l.ctx, &user.LoginReq{
		Username: req.Username, // 传给 RPC 的账号
		Password: req.Password, // 传给 RPC 的密码
	})

	if err != nil {
		l.Logger.Errorf("RPC 调用失败: %v", err)
		return nil, err
	}

	// 打印 RPC 返回的结果！
	l.Logger.Infof("RPC 调用成功！RPC返回的Token是: %s", rpcResp.Token)
	return &types.LoginResp{
		Token: rpcResp.Token,
	}, nil
}
