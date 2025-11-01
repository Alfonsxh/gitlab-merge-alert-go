# syntax=docker/dockerfile:1.6

# 前端构建阶段
FROM node:20-alpine AS frontend-builder

# 配置 Alpine 中国大陆镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 设置 npm 镜像
RUN npm config set registry https://registry.npmmirror.com

# 设置前端工作目录
WORKDIR /app/frontend

# 复制前端依赖文件
COPY frontend/package*.json ./

# 安装前端依赖（缓存 npm 包）
RUN --mount=type=cache,target=/root/.npm npm ci --no-audit

# 复制前端源代码并构建
COPY frontend/ ./
RUN --mount=type=cache,target=/root/.npm npm run build

# 后端构建阶段
FROM golang:1.23-alpine AS backend-builder

# 配置 Alpine 中国大陆镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装基础依赖
RUN apk add --no-cache ca-certificates

# 设置Go代理
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn

# 设置后端工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 预下载依赖（缓存 go modules）
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build go mod download

# 复制后端源代码
COPY . ./

# 从前端构建阶段复制构建好的前端文件（用于嵌入）
COPY --from=frontend-builder /app/frontend/dist ./internal/web/frontend_dist

# 构建应用（启用 embed 标签，纯 Go 构建，禁用 CGO）
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -tags embed -a -ldflags '-extldflags "-static"' -o main ./cmd

# 最终镜像 - 最小化 Alpine
FROM alpine:latest

# 配置 Alpine 中国大陆镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装运行时依赖
RUN apk add --no-cache ca-certificates

WORKDIR /app

# 仅复制构建的二进制文件（已包含嵌入的前端资源）
COPY --from=backend-builder /app/main .

# 复制配置文件示例
COPY config.example.yaml ./config.yaml

# 创建必要的目录
RUN mkdir -p ./data ./logs

# 暴露端口
EXPOSE 1688

# 设置环境变量
ENV GMA_DATABASE_PATH=/data/gitlab-merge-alert.db
ENV GMA_LOG_PATH=/logs/app.log

# 运行应用
ENTRYPOINT ["./main"]
