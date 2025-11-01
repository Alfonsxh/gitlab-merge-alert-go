# 应用名称
APP_NAME := gitlab-merge-alert-go
BIN_DIR := bin
FRONTEND_DIR := frontend
FRONTEND_DIST := internal/web/frontend_dist
FRONTEND_PLACEHOLDER_DIR := internal/web/frontend_placeholder

# 默认目标
.DEFAULT_GOAL := build

# 伪目标声明
.PHONY: build run clean install \
	frontend frontend-install \
	backend-install \
	docker-build-x86 \
	frontend-reset

frontend-install:
	cd $(FRONTEND_DIR) && npm install

# 前端构建
frontend: frontend-install
	@echo "Building frontend..."
	cd $(FRONTEND_DIR) && npm run build
	@echo "Copying frontend dist to embed directory..."
	rm -rf $(FRONTEND_DIST)
	cp -r $(FRONTEND_DIR)/dist $(FRONTEND_DIST)
	@echo "Frontend build complete"

backend-install:
	go mod tidy && go mod vendor

# 安装依赖
install: frontend-install backend-install
	@echo "All dependencies installed"

# 主构建目标
build: frontend
	@echo "Building backend with embedded frontend for linux/amd64..."
	mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 go build -tags embed -o $(BIN_DIR)/$(APP_NAME) ./cmd
	@echo "Cleaning up temporary frontend files..."
	$(MAKE) frontend-reset
	@echo "Build complete: $(BIN_DIR)/$(APP_NAME)"

# 运行应用
run:
	go run ./cmd/main.go

# Docker构建
docker-build-x86:
	docker build --platform linux/amd64 -t $(APP_NAME):x86_64 .

# 清理目标
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BIN_DIR)/$(APP_NAME)
	$(MAKE) frontend-reset
	@echo "Clean complete"

frontend-reset:
	@echo "Restoring embedded frontend placeholder..."
	rm -rf $(FRONTEND_DIST)
	mkdir -p $(FRONTEND_DIST)
	cp -r $(FRONTEND_PLACEHOLDER_DIR)/. $(FRONTEND_DIST)/
