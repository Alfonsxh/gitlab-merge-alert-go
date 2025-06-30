.PHONY: build run test clean docker-build docker-run migrate migrate-status migrate-rollback

# 应用名称
APP_NAME=gitlab-merge-alert-go

# 构建应用
build:
	go build -o bin/$(APP_NAME) ./cmd/server

# 运行应用
run:
	go run ./cmd/server

# 运行测试
test:
	go test -v ./...

# 清理
clean:
	rm -rf bin/
	rm -f data/gitlab-merge-alert.db

# 安装依赖
deps:
	go mod download
	go mod tidy

# 格式化代码
fmt:
	go fmt ./...

# 代码检查
lint:
	golangci-lint run

# 构建Docker镜像
docker-build:
	docker build -t $(APP_NAME) .

# 构建x86_64架构的Docker镜像
docker-build-x86:
	docker build --platform linux/amd64 -t $(APP_NAME):x86_64 .

# 构建多架构Docker镜像并推送
docker-buildx:
	docker buildx build --platform linux/amd64,linux/arm64 -t $(APP_NAME):latest --push .

# 运行Docker容器
docker-run:
	docker run -d --name $(APP_NAME) \
		-p 1688:1688 \
		-v $(PWD)/data:/data \
		$(APP_NAME)

# 停止并删除Docker容器
docker-stop:
	docker stop $(APP_NAME) || true
	docker rm $(APP_NAME) || true

# 重启Docker容器
docker-restart: docker-stop docker-run

# 查看Docker日志
docker-logs:
	docker logs -f $(APP_NAME)

# 初始化数据目录
init:
	mkdir -p data
	mkdir -p logs

# 运行数据库迁移
migrate:
	@echo "Running database migrations..."
	@go run cmd/migrate/main.go

# 查看迁移状态
migrate-status:
	@echo "Checking migration status..."
	@go run cmd/migrate/main.go -status

# 回滚最后一个迁移
migrate-rollback:
	@echo "Rolling back last migration..."
	@go run cmd/migrate/main.go -rollback