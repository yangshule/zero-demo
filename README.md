# 0. 说明

此文档是基于linux的go-zero学习项目，使用docker打包，并使用docker-compose集成管理。

下方所有浏览器访问`localhost`改为虚拟机ip，`port`改为`greet-api.yaml`中的端口`Port`值

# 1. goctl

## 1.1 goctl是什么

简单来说，`goctl`（通常读作 *go-control*）是 `go-zero` 微服务框架的**灵魂引擎**。它是一个极其强大的**命令行代码生成工具**.

它主要负责：

1.**解析图纸，生成HTTP API代码**

它会根据`.api`文件生成路由模块、Handler层、Logic层等规范的go文档，我们只需要在Logic中写核心逻辑即可。

2.**生成微服务RPC代码**

它会根据`.proto`文件，生成基于gRPC的高效通信代码，让微服务之间的调用向调用本地函数一样。

3.**一键生成数据库Model（ORM）**

我们创建了数据库，我们只需将表传给goctl，其会生成一套包含增删查改（CRUD）的go代码

4.**生成DevOps基础设施**

其能生成`Dockerfile`、`k8s`部署文件等

## 1.2 安装goctl

```shell
# 安装 goctl
go install github.com/zeromicro/go-zero/tools/goctl@latest

# 验证安装
goctl --version
```

配置环境变量：

```shell
#添加环境变量
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
#让环境变量立即生效
source ~/.bashrc
```

检测安装位置：

```shell
#会打印goctl的路径，否则失败
ls $(go env GOPATH)/bin/goctl
```

## 1.3 使用goctl

```shell
# 1. 创建项目目录
mkdir zero-demo && cd zero-demo

# 2. 初始化 Go module （这是 Go 项目的身份证）
go mod init zero-demo

# 3. 使用 goctl 生成一个名为 greet（问候）的服务（文件夹）
goctl api new greet
```

**greet文件夹**的目录结构：

`greet.api`:菜单(提示你的服务能干什么)

`internal/logic/`:厨房(核心业务逻辑)

`internal/handler/`:服务员(将用户的请求(HTTP)接入，传给厨房，返回结果给用户)

## 1.4 初次启动项目

```shell
# 1. 进入服务目录
cd greet

# 2. 整理依赖（自动下载 go-zero 的库）
go mod tidy

# 3. 启动服务
go run greet.go
```

如遇端口占用，请将`great-api.yaml`的端口改为8889。

浏览器访问`http://localhost:port/from/me`会打印JSON信息

## 1.5 修改逻辑

修改`internal/logic/greetlogic.go`的Greet函数：

```go
//每次访问时服务时会调用Greet函数
func (l *GreetLogic) Greet(req *types.Request) (resp *types.Response, err error) {

    //给网页返回的信息结构体
	return &types.Response{
		Message: "Hello go-zero, I am " + req.Name,
	}, nil
}
```

启动服务：

```shell
go run greet.go
```

浏览器访问`http://localhost:port/from/me`，会打印`{"message":"Hello go-zero, I am me"}`

# 2. reids的使用



**配置文件**

打开 `etc/greet-api.yaml`在末尾添加：（端口占用请改此文件）

```shell
#这一步告诉了程序redis在哪里
Redis:
  Host: localhost:6379
  Type: node # 单节点模式，如果是集群选 cluster
```

**映射文件**

修改`internal/config/config.go`：

```go
//此文件为映射文件,将yaml文件映射到go里
package config

import (
    "github.com/zeromicro/go-zero/core/stores/redis" // 引入 redis 包
    "github.com/zeromicro/go-zero/rest"
)

type Config struct {
    rest.RestConf
    // 新增这一行，名字必须和 yaml 里的 key 保持一致
    Redis redis.RedisConf
}
```

下载依赖项：

```shell
go mod tidy
```

**初始化**

将`internal/svc/servicecontext.go`修改为：

```go
//此文件为工具箱，将数据库与程序连接
package svc

import (
    "zero-demo/greet/internal/config"
    "github.com/zeromicro/go-zero/core/stores/redis" // 引入包
)

type ServiceContext struct {
    Config config.Config
    // 1. 在工具箱里加一个格层放 Redis
    //RedisClient是数据库操作员
    RedisClient *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
    return &ServiceContext{
        Config: c,
        // 2. 初始化 Redis 连接
        RedisClient: redis.MustNewRedis(c.Redis),
    }
}
```

**使用数据库**

修改`internal/logic/greetlogic.go`中的`greet`函数为：

```go
func (l *GreetLogic) Greet(req *types.Request) (resp *types.Response, err error) {
    // 1. 告诉工具箱里的 RedisClient，让"greet_count"自增
    count, err := l.svcCtx.RedisClient.Incr("greet_count")
    if err != nil {
        return nil, err
    }

    // 2. 返回信息给浏览器
    return &types.Response{
        Message: fmt.Sprintf("Hello %s, 这是你第 %d 次访问!", req.Name, count),
    }, nil
}
```

若没安装redis，请执行：

```shell
sudo apt update
sudo apt install redis-server -y
```

```shell
#开机自启
sudo systemctl enable --now redis-server
```

启动：

```shell
go run greet.go
```

访问`http://localhost:redisport/from/me`会打印

```txt
{"message":"Hello me, 这是你第 6 次访问!"}
```

# 3. 数据库模式

核心模式是**旁路缓存模式**：

1.先查缓存，命中返回

2.未命中，查数据库

3.将数据库结果写入缓存并返回

## 3.1 使用goctl连接mysql（ORM）

**下载mySQL:**

```shell
# 启动一个 MySQL 容器，密码设为 root,名称为mysql-demo
sudo docker run --name mysql-demo -e MYSQL_ROOT_PASSWORD=root -p 3306:3306 -d mysql:8.0
```

**定义数据表：**在zero-demo下创建名为`user.sql`的文件，写入以下内容：

(若想看懂，可以自行学习mysql)

```sql
CREATE TABLE `user` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `number` varchar(255) NOT NULL COMMENT '学号/工号',
  `name` varchar(255) NOT NULL COMMENT '用户名称',
  `password` varchar(255) NOT NULL COMMENT '用户密码',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `number_unique` (`number`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**导入表：**

```shell
# 创建数据库gozero
sudo docker exec -it mysql-demo mysql -uroot -proot -e "CREATE DATABASE IF NOT EXISTS gozero;"

# 导入数据
sudo docker exec -i mysql-demo mysql -uroot -proot gozero < user.sql
```



**生成代码：**

```shell
# /zero-demo下执行
# 根据 sql 生成带缓存逻辑的 Go 代码
# -src: sql源文件
# -dir: 代码输出目录
# -c: 生成带缓存(cache)的代码
goctl model mysql ddl -src user.sql -dir internal/model -c

#下载依赖项
go mod tidy

# 将zero-demo下的internal/model移至/zero-demo/greet
mv internal/model greet
```

我们发现model下有三个文件：`usermodel_gen.go`、`usermodel.go`、`vars.go`都是goctl根据user.sql生成的操作函数等。

**将model装入工具箱：**

修改`etc/greet.api.yaml`：（配置文件）

```yaml
# ... Redis 配置下面 ...
# 格式: 用户:密码@tcp(地址:端口)/数据库名?参数
DataSource: root:root@tcp(127.0.0.1:3306)/gozero?charset=utf8mb4&parseTime=true&loc=Local

Cache:
  - Host: localhost:6379
```

修改`internal/config/config.go`:（映射文件）

```go
type Config struct {
    rest.RestConf
    Redis redis.RedisConf
    // 新增：数据库连接字符串
    DataSource string 
    //缓存
    Cache cache.CacheConf
}
```

修改`internal/svc/servicecontext.go`:

```go
package svc

import (
    "github.com/zeromicro/go-zero/core/stores/sqlx" // 引入 sqlx
    "zero-demo/internal/config"
    "zero-demo/greet/internal/model" // 引入刚才生成的 model
    // ...
)

type ServiceContext struct {
    Config config.Config
    RedisClient *redis.Redis
    // 新增：User Model 接口
    UserModel model.UserModel
}

func NewServiceContext(c config.Config) *ServiceContext {
    // 建立 MySQL 连接
    conn := sqlx.NewMysql(c.DataSource)
    
    return &ServiceContext{
        Config: c,
        RedisClient: redis.MustNewRedis(c.Redis),
        // 初始化 Model，把 MySQL 连接和 Redis 连接都传进去（因为它要同时操作两个）
        UserModel: model.NewUserModel(conn, c.Cache),
    }
}
```

## 3.2 旁路缓存模式的操作

看`usermodel_gen.go`的：

```go
func (m *defaultUserModel) FindOne(ctx context.Context, id int64) (*User, error) {
    //生成要查的关键字
	userIdKey := fmt.Sprintf("%s%v", cacheUserIdPrefix, id)
	var resp User
    //将关键字userIdkey在缓存中查询，并存储在resp中，并返回
    //若缓存中不存在，则调用匿名函数
	err := m.QueryRowCtx(ctx, &resp, userIdKey, func(ctx context.Context, conn sqlx.SqlConn, v any) error {
        //查询缓存
		query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", userRows, m.table)
        //调用匿名函数查询数据库，若不存在则默认存入并写入缓存
		return conn.QueryRowCtx(ctx, v, query, id)
	})
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
```

## 3.3 每个文件

系统地描述一下各个文件的作用：

```txt
zero-demo/greet/
├── greet.api                  <-- 【图纸】接口定义文件
├── greet.go                   <-- 【入口】main 函数启动点
├── etc/
│   └── greet-api.yaml         <-- 【配置】运维关心的配置文件 (端口、DB密码)
├── internal/
│   ├── config/
│   │   └── config.go          <-- 【配置映射】把 yaml 读进 Go 结构体
│   ├── svc/
│   │   └── servicecontext.go <-- 【工具箱】装载 DB、Redis 连接的地方
│   ├── handler/               <-- 【服务员】HTTP 层，解析参数
│   ├── logic/                 <-- 【厨师】业务逻辑层 (你写代码的地方)
│   ├── model/                 <-- 【仓库】数据层 (自动生成的 CRUD)
│   └── types/                 <-- 【类型】定义请求/响应的数据结构
```

**辨析每个变量和函数：**

`greet.api`:项目的源头，`goctl`读取该文件生成代码

`greet.go`:项目的入口(main函数),读取`greet-api.yaml`,创建`servicecontext`(初始化redis和mysql),启动`HTTP server`

`greet-api.yaml`:运维配置

`config.go`:将YAML文字映射为go的变量，如`Redis`映射为`redis.RedisConf`

`servicecontext.go`:将数据库连接、Redis客户端初始化并打包为`svcCtx`

`greethandler.go`:接收`HTTP`请求，解析参数到go结构体(详见`l := logic.NewGreetLogic(r.Context(), svcCtx)`),传给logic处理(详见`l.Greet(&req)`)，返回结果

`routes.go`:为`greethandler.go`制定路径

`greetlogic.go`:拿到`greethandler.go`传来的参数并处理

`_gen.go`:工具生成的标准的CRUD(增删改查)

`_model.go`:SQL实现

## 3.4 使用存储数据库

我们已经连接mysql和准备好访问模式了，现在我们使用一下吧。

下面我们将模拟注册并将其写入mysql。

修改`greet.api`：

```go
syntax = "v1"

// 1. 定义请求结构体
type RegisterReq {
    Username string `json:"username"`
    Password string `json:"password"`
    Number   string `json:"number"` // 学号/工号，作为唯一标识
}

// 2. 定义响应结构体
type RegisterResp {
    Message string `json:"message"`
}

// 3. 定义服务接口
service greet-api {
    // 定义一个叫 Register 的 Handler (处理器)
    @handler Register
    // 这是一个 POST 请求，路由是 /user/register
    post /user/register (RegisterReq) returns (RegisterResp)
}
```

删除`greethandler.go`和`greetlogic.go`

重新生成代码：

```shell
goctl api go -api greet.api -dir .
```

修改`registerlogic.go`:

```go
package logic

import (
    "context"
    "database/sql" // 需要引入这个标准库来处理 sql.NullString 等类型
    "fmt"

    "zero-demo/greet/internal/svc"
    "zero-demo/greet/internal/types"
    "zero-demo/greet/internal/model" // 引入 model 包

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
    // 把浏览器收到的数据，转换成 Model 层需要的 User 结构体
    newUser := &model.User{
        Name:     req.Username,
        Password: req.Password, // 实际项目中这里要加密，现在先存明文
        Number:   req.Number,
    }

    // 写入mysql,UserModel同样是操作员
    res, err := l.svcCtx.UserModel.Insert(l.ctx, newUser)
    
    // 错误处理
    if err != nil {
        // 如果是唯一键冲突（比如学号重复），需要特殊处理
        // 这里简单返回错误信息
        return nil, fmt.Errorf("注册失败: %v", err)
    }

    // 获取新插入的 ID (可选)
    id, _ := res.LastInsertId()

    // 5. 返回成功响应
    return &types.RegisterResp{
        Message: fmt.Sprintf("注册成功! 用户ID: %d", id),
    }, nil
}
```

运行服务：

```shell
go run greet.go
```

模拟注册：

```shell
curl -i -X POST http://localhost:port/user/register \
   -H "Content-Type: application/json" \
   -d '{"username": "User", "password": "123", "number": "NO111"}'
```

这时会提示注册成功。

若提示数据库不存在，请：

```shell
# 1. 强制删除旧容器
sudo docker rm -f mysql-demo

# 2. 重新启动 (注意 -p 3306:3306 是关键)
sudo docker run --name mysql-demo -e MYSQL_ROOT_PASSWORD=root -p 3306:3306 -d mysql:8.0

# 3. 等待 10 秒钟让 MySQL 初始化

# 4. 重新创建数据库 (因为容器重建了，数据没了)
sudo docker exec -it mysql-demo mysql -uroot -proot -e "CREATE DATABASE IF NOT EXISTS gozero;"

# 5. 重新导入表结构
sudo docker exec -i mysql-demo mysql -uroot -proot gozero < user.sql
```

# 4. 登录与鉴权(JWT)

我们已经注册好了，现在应该实现登录功能和鉴权功能了。

## 4.1 登录

修改`greet.api`：

```go
syntax = "v1"

type (
    // ... 之前的 RegisterReq/Resp 保留 ...

    // 1. 定义登录请求
    LoginReq {
        Number   string `json:"number"`
        Password string `json:"password"`
    }

    // 2. 定义登录响应
    LoginResp {
        AccessToken  string `json:"accessToken"`
        AccessExpire int64  `json:"accessExpire"`
    }
)

service greet-api {
    // ... 之前的 Register 接口保留 ...

    // 3. 新增登录接口
    @handler Login
    post /user/login (LoginReq) returns (LoginResp)
}
```

生成代码：

```shell
goctl api go -api greet.api -dir .
```

修改`greet-api.yaml`：（配置文件）

```yaml
#鉴权
Auth:
  AccessSecret: "u8wG9d2s@6#v1lA" # 随便写一串复杂的字符串作为密钥
  AccessExpire: 86400             # Token 有效期，单位秒 (这里是1天)
```

修改`config.go`：（配置映射）

```
type Config struct {
    rest.RestConf
    Redis      redis.RedisConf
    DataSource string
    Cache      cache.CacheConf
    
    // 新增：对应 YAML 里的 Auth
    Auth struct {
        AccessSecret string
        AccessExpire int64
    }
}
```

修改登录逻辑，修改`loginlogic.go`：

```go
func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
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
```

启动服务：

```shell
go run greet.go
```

尝试登录：

```shell
#尝试登录
curl -i -X POST http://localhost:port/user/login \
   -H "Content-Type: application/json" \
   -d '{"number": "NO111", "password": "123"}'
```

此时会打印类似于：

```txt
{"accessToken":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzQ3OTA1NTIsImlhdCI6MTc3NDcwNDE1MiwidXNlcklkIjoxfQ.xHzr9okVo5jHH66BPpsnEME_7n75PdBkpX684Vq3Ix8","accessExpire":1774790552}
```

的信息，可打开[jwt](jwt.io)查看

## 4.2 鉴权

修改`greet.api`:

```go
// ... 之前的 LoginReq/Resp ...

type UserInfoResp {
    Id       int64  `json:"id"`
    Name     string `json:"name"`
    Number   string `json:"number"`
}

// 注意这里：我们将服务分了个组，并开启了 jwt 认证
// jwt: Auth 对应的是 yaml 配置里的 Auth
@server(
    jwt: Auth
)
service greet-api {
    @handler UserInfo
    get /user/info returns (UserInfoResp)
}
```

生成代码：

```shell
goctl api go -api greet.api -dir .
```

修改`userinflogic.go`:

```go
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

	// 4. 返回信息，会打印
	return &types.UserInfoResp{
		Id:     userInfo.Id,
		Name:   userInfo.Name,
		Number: userInfo.Number,
	}, nil
}
```

运行代码

```shell
go run greet.go
```

```shell
#尝试登录，此时会打印Token,复制下来
curl -i -X POST http://localhost:8889/user/login \
   -H "Content-Type: application/json" \
   -d '{"number": "NO111", "password": "123"}'

#尝试使用Token查询信息
# 注意 Header 里的格式是：Authorization: Bearer <token>
curl -i http://localhost:8889/user/info \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

```

此时会打印`No111`的信息。

# 5. 关于ctx

看最顶层，它是由一个`r *http.Request`传递的`r.Context()`，它有三类作用：

**一、控制生命周期**

当浏览器或curl断开时，ctx收到信息，通知其它断开

**二、追踪信息**

我们在交互时，会返回如：

```shell
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Traceparent: 00-243b2d8ced20618df6162c922402d381-1d323ca8563ce1ae-00
Date: Wed, 14 Jan 2026 08:04:14 GMT
Content-Length: 45
```

的信息，其中`Traceparent`是从ctx中取出来的

**三、数据存储**

其最内层是空的，在**http**层添加了`http.Request`、在**Trace**层添加了`TraceID`，然后添加**JWT**组件，最后我们拿到的是`l.ctx`

# 6. 打包

将我们写的代码打包为一个不用下载mysql和redis也能启动的项目

在`go-zero/下创建一个文件名为Dockerfile`，写入内容：(镜像配置文件)

```dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 1. 现在的上下文是 zero-demo 根目录，所以可以直接复制 go.mod
COPY go.mod go.sum ./
# 更换源，防止timeout
ENV GOPROXY=https://goproxy.cn,direct

RUN go mod download

# 2. 复制 greet 服务代码到容器的 /app/greet 目录
COPY greet greet/

# 3. 编译 
RUN go build -ldflags="-s -w" -o greet-api greet/greet.go

# --- 运行阶段 ---
FROM alpine:latest

RUN apk update --no-cache && apk add --no-cache tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

# 4. 从 builder 拿二进制文件
COPY --from=builder /app/greet-api .

# 5. 从 builder 拿配置文件 (注意源路径变了)
# 我们把它复制到容器的 etc/ 目录下，保持结构清晰
COPY --from=builder /app/greet/etc/greet-api.yaml etc/greet-api.yaml

EXPOSE port

CMD ["./greet-api", "-f", "etc/greet-api.yaml"]
```

在`zero-demo/`创建`docker-compose.yaml`，填入：(操作命令文件)

```yaml
version: '3'

services:
  greet-api:
    build: .             # 在当前目录(zero-demo)构建
    ports:
      - "8888:8888"
    depends_on:
      - mysql
      - redis
    restart: always

  mysql:
    image: mysql:8.0
    container_name: mysql-svc
    environment:
      MYSQL_ROOT_PASSWORD: root
      MYSQL_DATABASE: gozero
    ports:
      - "3306:3306"
    volumes:
      # 注意：现在我们在根目录，所以路径多了一层 greet/
      # 如果你的 user.sql 在 greet 目录下，要写成 ./greet/user.sql
      # 如果 user.sql 在根目录，就写 ./user.sql
      # 这里假设你还没移动 user.sql，它还在 zero-demo/ 下(根据之前的操作)
      - ./data/mysql:/var/lib/mysql
      # 请确认你的 user.sql 到底在哪？如果报错找不到文件，请检查这个路径
      - ./user.sql:/docker-entrypoint-initdb.d/user.sql

  redis:
    image: redis:6-alpine
    container_name: redis-svc
    ports:
      - "6379:6379"
```

修改`greet-api.yaml`为:（）

```yaml
Name: greet-api
Host: 0.0.0.0
Port: port

# 修改 DataSource: 把 127.0.0.1 改成 mysql (对应 docker-compose 里的 service 名)
DataSource: root:root@tcp(mysql:3306)/gozero?charset=utf8mb4&parseTime=true&loc=Local

# 修改 Redis: 把 localhost 改成 redis
Redis:
  Host: redis:6379
  Type: node

# 修改 Cache: 把 localhost 改成 redis
Cache:
  - Host: redis:6379
    Pass: ""

Auth:
  AccessSecret: "u8wG9d2s@6#v1lA"
  AccessExpire: 86400
```

**执行：**

停止旧容器：

```shell
sudo docker rm -f mysql-demo redis-demo
# 如果你有其他占用 8889 的 Go 程序，也关掉
```

在`zero-demo/greet/`下执行：

```shell
sudo docker-compose up -d --build
```

**若出现端口占用：**（注意看是哪个端口，以6379为例）

​	1.查看此端口的服务:

```shell
# 看看有哪些容器占用了 6379
sudo docker ps -a | grep 6379
```

​	若出现占用，则：

```shell
# 这里的 $(...) 是个小技巧，意思是“把所有正在运行的容器ID都删掉”
# 如果你怕误删，可以手动 sudo docker rm -f <容器ID>
sudo docker rm -f redis-demo redis-svc
```

​	2.检查本机安装了redis，默认占用了：

```shell
#停止redis
sudo systemctl stop redis
# 或者
sudo systemctl stop redis-server
```

​	3.检查端口：

```shell
sudo lsof -i :6379
#没有打印则为未占用状态
```

​	4.重新启动

```shell
sudo docker-compose up -d --build
```

**查看状态：**

```shell
sudo docker-compose ps
#此时应该是三个服务：mysql-svc redis-svc zero-demo-api-1
```

**测试mysql注册：**

```shell
curl -i -X POST http://localhost:port/user/register \
   -H "Content-Type: application/json" \
   -d '{"username": "DockerUser", "password": "123", "number": "DOCKER01"}'
```

**测试redis登录：**

```shell
curl -i -X POST http://localhost:port/user/login \
   -H "Content-Type: application/json" \
   -d '{"number": "DOCKER01", "password": "123"}'
#返回accessToken就是成功。
```

# 7. 上传项目

```shell
# 初始化 Git 仓库
git init
```

编写`.ignore`文件

```ignore
# 1. 编译输出的二进制文件 (不要把生成物放进源码库)
greet-api
*.exe
*.out

# 2. 数据库和缓存挂载的本地数据目录 (绝对不能提交！)
data/

# 3. 依赖包目录 (Go module 模式下通常不提交 vendor)
vendor/

# 4. IDE 和操作系统产生的配置/缓存文件
.idea/
.vscode/
*.DS_Store
```

向git提交

```shell
# 1. 查看当前状态（你会看到标红的文件，但没有 data/ 目录了）
git status

# 2. 将所有合法文件添加到暂存区
git add .

# 3. 再次查看状态（文件应该都变绿了，说明准备好提交了）
git status

# 4. 提交并写明注释（遵循业界标准的 commit message 格式）
git commit -m "feat: 初始化 go-zero 微服务与 docker-compose 架构"
```

推到云端：

```shell
# 1. 把你的本地仓库和云端仓库关联起来 (把下面的 URL 换成你真实的地址)
git remote add origin https://github.com/你的用户名/zero-demo.git

# 2. 确保主分支名字叫 main (现在业界都用 main 代替 master 了)
git branch -M main

# 3. 把本地代码推送到云端 (第一次推送需要加 -u 记住关联)
git push -u origin main
```

# 8. 清理空间

```shell
# 这会删掉所有未使用的容器、网络和悬空镜像
sudo docker system prune -a -f
# 自动卸载不再需要的依赖，并清空下载缓存
sudo apt autoremove -y && sudo apt clean
# 只保留最近 3 天的日志，剩下的全删
sudo journalctl --vacuum-time=3d
```

可视化管理：

```shell
sudo apt install ncdu
# -x 参数极其重要！它告诉 ncdu “只扫描当前的系统盘，不要扫我刚刚挂载的新盘(Data)”
sudo ncdu -x /
```

