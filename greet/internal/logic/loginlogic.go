// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"errors"
	"time"

	"zero-demo/greet/internal/model"
	"zero-demo/greet/internal/svc"
	"zero-demo/greet/internal/types"
	"zero-demo/user/rpc/user"

	"github.com/golang-jwt/jwt/v4"
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
		Username: "admin",  // 传给 RPC 的账号
		Password: "123456", // 传给 RPC 的密码
	})

	if err != nil {
		l.Logger.Errorf("RPC 调用失败: %v", err)
		return nil, err
	}

	// 打印 RPC 返回的结果！
	l.Logger.Infof("RPC 调用成功！RPC返回的Token是: %s", rpcResp.Token)

	// 1. 去数据库查用户 (根据学号)
	userInfo, err := l.svcCtx.UserModel.FindOneByNumber(l.ctx, req.Number)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 2. 比对密码
	// 注意：实际项目中密码是加密存的，这里演示先直接比对字符串
	if userInfo.Password != req.Password {
		return nil, errors.New("密码错误")
	}

	// 3. 生成 JWT Token
	now := time.Now().Unix()
	accessExpire := l.svcCtx.Config.Auth.AccessExpire
	accessSecret := l.svcCtx.Config.Auth.AccessSecret

	token, err := l.getJwtToken(accessSecret, now, accessExpire, userInfo.Id)
	if err != nil {
		return nil, err
	}

	return &types.LoginResp{
		AccessToken:  token,
		AccessExpire: now + accessExpire,
	}, nil
}

// 辅助方法：生成 JWT
func (l *LoginLogic) getJwtToken(secretKey string, iat, seconds, userId int64) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds // 过期时间
	claims["iat"] = iat           // 签发时间
	claims["userId"] = userId     // 【关键】把用户ID塞进 Token 里

	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims

	return token.SignedString([]byte(secretKey))
}
