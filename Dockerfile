FROM golang:1.22-alpine AS builder

WORKDIR /app

# 1. 现在的上下文是 zero-demo 根目录，所以可以直接复制 go.mod
COPY go.mod go.sum ./

ENV GOPROXY=https://goproxy.cn,direct

RUN go mod download

# 2. 复制 greet 服务代码到容器的 /app/greet 目录
COPY greet greet/

# 3. 编译 (注意路径变化：编译 greet 目录下的 greet.go)
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

EXPOSE 8888

CMD ["./greet-api", "-f", "etc/greet-api.yaml"]