.PHONY: build run test clean docker-build docker-run migrate migrate-status migrate-rollback frontend-install frontend-dev frontend-build

# 应用名称
APP_NAME=gitlab-merge-alert-go

# 构建应用
build: frontend-build
	@echo "Building backend with embedded frontend for linux/amd64..."
	cd backend && GOOS=linux GOARCH=amd64 go build -tags embed -o ../bin/$(APP_NAME) ./cmd/main.go
	@echo "Cleaning up temporary frontend files..."
	rm -rf backend/internal/web/frontend_dist
	@echo "Build complete: bin/$(APP_NAME)"

# 运行应用（后端）
run:
	cd backend && go run ./cmd/main.go

# 构建x86_64架构的Docker镜像
docker-build-x86:
	docker build --platform linux/amd64 -t $(APP_NAME):x86_64 .

# 前端相关命令
frontend-install:
	cd frontend && npm install

frontend-build:
	@echo "Building frontend..."
	cd frontend && npm install && npm run build
	@echo "Frontend build complete"
	@echo "Copying frontend dist to backend for embedding..."
	rm -rf backend/internal/web/frontend_dist
	cp -r frontend/dist backend/internal/web/frontend_dist