# DigWis Panel Makefile
# 用于生产构建和 CI/CD，开发时请使用 Air: /home/parallels/go/bin/air

.PHONY: help build build-css-prod clean install test release-build release-auto deploy-local rollback

# 默认目标
help:
	@echo "DigWis Panel 构建工具"
	@echo ""
	@echo "🚀 开发环境:"
	@echo "  air            - 启动热重载开发服务器 (推荐)"
	@echo ""
	@echo "🏭 生产构建:"
	@echo "  build          - 构建完整项目 (生产版)"
	@echo "  build-css-prod - 构建 CSS (生产版，压缩)"
	@echo "  build-release  - 构建到 releases 目录"
	@echo ""
	@echo "� 快速部署:"
	@echo "  deploy-local   - 编译并部署到本地生产环境"
	@echo "  rollback       - 回滚到上一个版本"
	@echo ""
	@echo "📦 版本发布:"
	@echo "  release-build  - 构建多平台发布包 (本地)"
	@echo "  release-auto   - 自动发布到 GitHub (需要 VERSION 和 CHANGELOG)"
	@echo ""
	@echo "🛠️  工具:"
	@echo "  install        - 安装依赖"
	@echo "  clean          - 清理构建文件"
	@echo "  test           - 运行测试"
	@echo "  size           - 显示静态文件大小"

# 安装依赖
install:
	@echo "📦 安装 npm 依赖..."
	npm install
	@echo "✅ 依赖安装完成"

# 构建 CSS (生产版)
build-css-prod:
	@echo "🎨 构建 CSS (生产版)..."
	npm run build-css-prod
	@echo "✅ CSS 生产版构建完成"

# 构建 Go 程序 (开发版 - 根目录)
build-go:
	@echo "🔨 构建 Go 程序 (开发版)..."
	@export GOPROXY=https://goproxy.cn,direct && CGO_ENABLED=1 go build -ldflags="-s -w" -o digwis-panel .
	@echo "✅ Go 程序构建完成: ./digwis-panel"

# 构建 Go 程序到 releases 目录
build-go-release:
	@echo "🔨 构建 Go 程序到 releases 目录..."
	@mkdir -p releases
	@export GOPROXY=https://goproxy.cn,direct && CGO_ENABLED=1 go build -ldflags="-s -w" -o releases/digwis-panel .
	@echo "✅ Go 程序构建完成: ./releases/digwis-panel"

# 快速构建 (开发版 - 根目录)
build: build-css-prod build-go
	@echo "🚀 开发构建完成 (./digwis-panel)"

# 构建到 releases 目录
build-release: build-css-prod build-go-release
	@echo "🚀 发布构建完成 (./releases/digwis-panel)"

# 构建多平台发布包（本地）
release-build:
	@echo "📦 构建多平台发布包..."
	@chmod +x scripts/release.sh
	@./scripts/release.sh $(VERSION) "$(CHANGELOG)"
	@echo "✅ 多平台发布包构建完成"

# 正式发布（自动化）
release-auto:
	@echo "🚀 正式发布版本..."
	@chmod +x scripts/release.sh
	@./scripts/release.sh $(VERSION) "$(CHANGELOG)" --auto
	@echo "✅ 正式发布完成"

# 清理构建文件
clean:
	@echo "🧹 清理构建文件..."
	rm -f digwis-panel
	rm -f assets/css/output.css
	rm -rf tmp/
	@echo "✅ 清理完成"

# 清理所有文件（包括 releases）
clean-all:
	@echo "🧹 清理所有构建文件..."
	rm -f digwis-panel
	rm -f assets/css/output.css
	rm -rf tmp/
	rm -rf releases/
	@echo "✅ 全部清理完成"

# 显示文件大小
size:
	@echo "📊 静态文件大小:"
	@ls -lh assets/css/output.css assets/js/*.js 2>/dev/null || echo "请先构建 CSS"
	@echo ""
	@echo "📊 总计大小:"
	@du -ch assets/css/output.css assets/js/*.js 2>/dev/null | tail -1 || echo "请先构建 CSS"

# 运行测试
test:
	@echo "🧪 运行测试..."
	go test ./...

# 开发环境
dev:
	@echo "🚀 启动开发环境 (Air 热重载)..."
	./scripts/dev/start-air.sh

# 快速部署到本地生产环境
deploy-local:
	@echo "🔨 编译嵌入式版本..."
	CGO_ENABLED=1 go build -o digwis-panel .
	@echo "🛑 停止服务..."
	sudo systemctl stop digwis-panel
	@echo "📦 备份当前版本..."
	sudo cp /opt/digwis-panel/digwis-panel /opt/digwis-panel/digwis-panel.backup.$$(date +%s) 2>/dev/null || true
	@echo "🔄 替换程序文件..."
	sudo cp ./digwis-panel /opt/digwis-panel/digwis-panel
	@echo "🚀 启动服务..."
	sudo systemctl start digwis-panel
	@echo "✅ 部署完成！访问: http://localhost:8080"
	@echo ""
	@echo "📊 服务状态："
	@systemctl status digwis-panel --no-pager -l

# 回滚到上一个版本
rollback:
	@echo "🔍 查找备份文件..."
	@BACKUP_FILE=$$(ls -t /opt/digwis-panel/digwis-panel.backup.* 2>/dev/null | head -1); \
	if [ -z "$$BACKUP_FILE" ]; then \
		echo "❌ 没有找到备份文件"; \
		exit 1; \
	fi; \
	echo "🛑 停止服务..."; \
	sudo systemctl stop digwis-panel; \
	echo "🔄 恢复备份版本: $$BACKUP_FILE"; \
	sudo cp "$$BACKUP_FILE" /opt/digwis-panel/digwis-panel; \
	echo "🚀 启动服务..."; \
	sudo systemctl start digwis-panel; \
	echo "✅ 回滚完成！"; \
	systemctl status digwis-panel --no-pager -l

# 开发提示
dev-help:
	@echo "💡 开发环境使用方法:"
	@echo "   make dev                    # 启动 Air 热重载 (推荐)"
	@echo "   ./scripts/dev/start-air.sh  # 直接使用脚本"
	@echo "   /home/parallels/go/bin/air  # 直接使用 Air"

# 部署相关命令 (使用 releases 版本)
deploy: build-release
	@echo "🚀 开始部署..."
	@chmod +x scripts/deploy.sh
	@./scripts/deploy.sh

# 生成安装脚本
install-script:
	@echo "📦 生成安装脚本..."
	@chmod +x scripts/deploy.sh
	@./scripts/deploy.sh --install-script-only

# 推送代码到远程仓库 (使用 releases 版本)
push: build-release
	@echo "📤 推送代码到远程仓库..."
	@chmod +x scripts/simple-push.sh
	@./scripts/simple-push.sh

# 快速推送（带自定义提交信息）
push-msg: build-release
	@echo "📤 推送代码到远程仓库..."
	@chmod +x scripts/simple-push.sh
	@read -p "请输入提交信息: " msg; \
	./scripts/simple-push.sh "$$msg"

# 配置 Git 凭据
git-config:
	@echo "🔐 配置 Git 凭据..."
	@chmod +x scripts/git-push.sh
	@./scripts/git-push.sh --config-only

# 初始化仓库
git-init:
	@echo "🔧 初始化 Git 仓库..."
	@git init
	@chmod +x scripts/git-push.sh
	@./scripts/git-push.sh --config-only
	@echo "✅ Git 仓库初始化完成"

# 版本管理
version:
	@echo "🏷️  版本管理..."
	@chmod +x scripts/version.sh
	@./scripts/version.sh

# 创建版本标签
tag:
	@echo "🏷️  创建版本标签..."
	@chmod +x scripts/version.sh
	@read -p "请输入版本号 (如 v1.0.1): " version; \
	./scripts/version.sh tag "$$version"

# 发布版本 (标签 + 构建 + 推送)
release-version: build-release
	@echo "🚀 发布新版本..."
	@chmod +x scripts/version.sh
	@read -p "请输入版本号 (如 v1.0.1): " version; \
	./scripts/version.sh tag "$$version" && \
	./scripts/version.sh push && \
	make release && \
	echo "✅ 版本 $$version 发布完成"

# 检查 Git 状态
git-status:
	@echo "📋 Git 状态:"
	@git status --short

# 显示帮助信息
help:
	@echo "🛠️  DigWis Panel 构建工具"
	@echo "=========================="
	@echo ""
	@echo "📦 构建命令:"
	@echo "   make build          - 快速构建 (开发版，输出到根目录)"
	@echo "   make build-release  - 发布构建 (输出到 releases/ 目录)"
	@echo "   make release        - 多平台发布包构建"
	@echo "   make build-dev      - 开发构建"
	@echo "   make build-css      - 构建 CSS (开发版)"
	@echo "   make build-css-prod - 构建 CSS (生产版)"
	@echo "   make build-go       - 构建 Go 程序 (根目录)"
	@echo "   make build-go-release - 构建 Go 程序 (releases/ 目录)"
	@echo ""
	@echo "🚀 运行命令:"
	@echo "   make run            - 运行程序"
	@echo "   make dev            - 启动开发环境 (Air 热重载)"
	@echo ""
	@echo "🧪 测试命令:"
	@echo "   make test           - 运行测试"
	@echo ""
	@echo "📊 信息命令:"
	@echo "   make size           - 显示文件大小"
	@echo "   make git-status     - 显示 Git 状态"
	@echo ""
	@echo "🚀 部署命令:"
	@echo "   make deploy         - 完整部署 (构建 + 推送代码 + 生成安装脚本)"
	@echo "   make push           - 推送代码到远程仓库"
	@echo "   make push-msg       - 推送代码到远程仓库 (自定义提交信息)"
	@echo "   make install-script - 仅生成安装脚本"
	@echo ""
	@echo "🔧 Git 命令:"
	@echo "   make git-config     - 配置 Git 凭据"
	@echo "   make git-init       - 初始化 Git 仓库"
	@echo ""
	@echo "🏷️  版本管理:"
	@echo "   make version        - 版本管理工具"
	@echo "   make tag            - 创建版本标签"
	@echo "   make release-version - 发布新版本 (标签 + 构建 + 推送)"
	@echo ""
	@echo "🧹 清理命令:"
	@echo "   make clean          - 清理构建文件 (保留 releases/)"
	@echo "   make clean-all      - 清理所有文件 (包括 releases/)"
	@echo ""
	@echo "💡 开发帮助:"
	@echo "   make dev-help       - 显示开发环境使用方法"
