// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"zero-demo/greet/internal/svc"
	"zero-demo/greet/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserInfoLogic {
	return &UserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserInfoLogic) UserInfo() (resp *types.UserInfoResp, err error) {
	// todo: add your logic here and delete this line
	// 1. 从 Context 中获取 userId
	// 注意：key "userId" 必须和你生成 Token 时写的 key 一模一样
	// 调试大法：先打印看看 ctx 里到底是啥
	val := l.ctx.Value("userId")
	l.Logger.Infof("从Token解析出的userId类型: %T, 值: %v", val, val)

	var userId int64

	// 2.使用 switch 来处理各种可能的数字类型
	switch v := val.(type) {
	case int64:
		userId = v
	case float64: // JWT 默认经常解析成 float64
		userId = int64(v)
	case json.Number: // go-zero 有时配置为 json.Number
		userId, err = v.Int64()
		if err != nil {
			return nil, errors.New("Token userId 转换失败")
		}
	case nil:
		return nil, errors.New("Token 里没有 userId 字段")
	default:
		return nil, fmt.Errorf("未知的 userId 类型: %T", v)
	}

	// 3. 根据 ID 查数据库
	userInfo, err := l.svcCtx.UserModel.FindOne(l.ctx, userId)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 4. 返回信息
	return &types.UserInfoResp{
		Id:     userInfo.Id,
		Name:   userInfo.Name,
		Number: userInfo.Number,
	}, nil
}
