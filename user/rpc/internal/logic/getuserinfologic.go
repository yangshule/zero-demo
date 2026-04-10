package logic

import (
	"context"

	"zero-demo/user/rpc/internal/svc"
	"zero-demo/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserInfoLogic) GetUserInfo(in *user.UserInfoReq) (*user.UserInfoResp, error) {
	// 调底层 Model 查数据库
	userInfo, err := l.svcCtx.UserModel.FindOneByUsername(l.ctx, in.Username)
	if err != nil {
		return nil, err
	}

	// 把查到的真实数据返回给呼叫方 (Order服务)
	return &user.UserInfoResp{
		Id:       userInfo.Id,
		Username: userInfo.Username,
		Password: userInfo.Password,
	}, nil
}
