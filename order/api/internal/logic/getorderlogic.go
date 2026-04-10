// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"zero-demo/order/api/internal/svc"
	"zero-demo/order/api/internal/types"
	"zero-demo/user/rpc/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderLogic {
	return &GetOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrderLogic) GetOrder(req *types.OrderReq) (resp *types.OrderResp, err error) {
	// 1. 模拟从 Order 数据库查到了订单信息（真实开发中这里也要调 Order RPC）
	mockOrderOwner := "admin" // 假设查出来这个订单是属于 admin 用户的

	// 2. 发起跨服 gRPC 调用！去 User 领地查 admin 的详细信息
	l.Logger.Infof("正在跨服呼叫 User RPC，查询用户: %s...", mockOrderOwner)

	// 注意这里用的是 l.svcCtx.UserRpc.GetUserInfo！
	userInfo, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &userclient.UserInfoReq{
		Username: mockOrderOwner,
	})

	if err != nil {
		l.Logger.Errorf("跨服调用失败: %v", err)
		return nil, err
	}

	// 3. 将本地订单数据 + 跨服取回的用户数据，拼装成 BFF 视图返回给前端！
	return &types.OrderResp{
		OrderId:     req.OrderId,
		Status:      "已发货",
		OwnerName:   userInfo.Username, // 从 User RPC 拿到的
		OwnerSecret: userInfo.Password, // 从 User RPC 拿到的！
	}, nil
}
