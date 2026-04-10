package logic

import (
	"context"
	"time"

	"zero-demo/user/rpc/internal/svc"
	"zero-demo/user/rpc/model"
	"zero-demo/user/rpc/user"

	"github.com/golang-jwt/jwt/v4"
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
	// 1. 调用 Model 层：根据用户名去 MySQL(或 Redis) 里查数据！
	// (因为你建表时加了 UNIQUE KEY，goctl 自动帮你生成了这个牛逼的方法)
	userInfo, err := l.svcCtx.UserModel.FindOneByUsername(l.ctx, in.Username)

	if err != nil {
		// 如果报 ErrNotFound，说明没这个号
		if err == model.ErrNotFound {
			l.Logger.Errorf("账号不存在: %s", in.Username)
			return &user.LoginResp{Success: false, Token: ""}, nil
		}
		// 数据库炸了等其他严重错误
		l.Logger.Errorf("数据库查询错误: %v", err)
		return nil, err
	}

	// 2. 校验密码 (实际企业开发中，这里必须是加密比对，比如 bcrypt，今天先用明文感受流程)
	if userInfo.Password == in.Password {
		l.Logger.Infof("密码正确！生成 Token...")
		now := time.Now().Unix()
		accessExpire := l.svcCtx.Config.JwtAuth.AccessExpire
		accessSecret := l.svcCtx.Config.JwtAuth.AccessSecret

		token, err := l.getJwtToken(accessSecret, now, accessExpire, userInfo.Id)
		if err != nil {
			return nil, err
		}
		return &user.LoginResp{
			Success: true,
			Token:   token,
		}, nil
	}

	// 3. 密码错误
	l.Logger.Infof("密码错误！")
	return &user.LoginResp{Success: false, Token: ""}, nil
}

func (l *LoginLogic) getJwtToken(secretKey string, iat, seconds, userId int64) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds // 过期时间
	claims["iat"] = iat           // 签发时间
	claims["userId"] = userId     // 【关键】把用户ID塞进 Token 里

	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims

	return token.SignedString([]byte(secretKey))
}
