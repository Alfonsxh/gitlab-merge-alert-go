FROM golang:1.23-alpine AS builder

# 配置 Alpine 中国大陆镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装基础依赖（纯 Go 构建无需 CGO 依赖）
RUN apk add --no-cache ca-certificates

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

# 构建应用（纯 Go 构建，禁用 CGO）
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o main ./cmd/server

# 最终镜像 - 使用相同的 Alpine 基础
FROM alpine:latest

# 配置 Alpine 中国大陆镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装运行时依赖
RUN apk add --no-cache ca-certificates


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
ENTRYPOINT ["./main"]