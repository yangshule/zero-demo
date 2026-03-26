// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"fmt"

	"zero-demo/greet/internal/model"
	"zero-demo/greet/internal/svc"
	"zero-demo/greet/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.RegisterResp, err error) {
	// todo: add your logic here and delete this line
	// 1. 组装数据
	// 我们把 API 层收到的 req 数据，转换成 Model 层需要的 User 结构体
	newUser := &model.User{
		Name:     req.Username,
		Password: req.Password, // 实际项目中这里要加密，现在先存明文
		Number:   req.Number,
	}

	// 2. 写入数据库 (调用工具箱里的 UserModel)
	// Insert 会自动生成 SQL：INSERT INTO user ...
	res, err := l.svcCtx.UserModel.Insert(l.ctx, newUser)

	// 3. 错误处理
	if err != nil {
		// 如果是唯一键冲突（比如学号重复），需要特殊处理
		// 这里简单返回错误信息
		return nil, fmt.Errorf("注册失败: %v", err)
	}

	// 4. 获取新插入的 ID (可选)
	id, _ := res.LastInsertId()

	// 5. 返回成功响应
	return &types.RegisterResp{
		Message: fmt.Sprintf("注册成功! 用户ID: %d", id),
	}, nil

}
