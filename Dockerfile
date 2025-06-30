FROM golang:1.23.8-bullseye AS builder

# 设置工作目录
WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./

# 设置Go代理
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# 最终镜像 - 使用本地已有的 alpine 镜像
FROM node:20-alpine

# 安装 SQLite
RUN apk add --no-cache sqlite ca-certificates


WORKDIR /app

# 复制构建的二进制文件
COPY --from=builder /app/main .

# 复制配置文件和静态资源
COPY --from=builder /app/web ./web
COPY --from=builder /app/config.yaml .

# 创建数据目录
RUN mkdir -p ./data

# 暴露端口
EXPOSE 1688

# 设置环境变量
ENV GMA_DATABASE_PATH=/data/gitlab-merge-alert.db

# 运行应用
CMD ["./main"]