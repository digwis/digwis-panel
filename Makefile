# DigWis Panel Makefile
# 用于生产构建和 CI/CD，开发时请使用 Air: /home/parallels/go/bin/air

.PHONY: help build build-css-prod clean install test release

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
	@echo "  release        - 构建发布包"
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
	@export GOPROXY=https://goproxy.cn,direct && go build -ldflags="-s -w" -o digwis-panel .
	@echo "✅ Go 程序构建完成: ./digwis-panel"

# 构建 Go 程序到 releases 目录
build-go-release:
	@echo "🔨 构建 Go 程序到 releases 目录..."
	@mkdir -p releases
	@export GOPROXY=https://goproxy.cn,direct && go build -ldflags="-s -w" -o releases/digwis-panel .
	@echo "✅ Go 程序构建完成: ./releases/digwis-panel"

# 快速构建 (开发版 - 根目录)
build: build-css-prod build-go
	@echo "🚀 开发构建完成 (./digwis-panel)"

# 构建到 releases 目录
build-release: build-css-prod build-go-release
	@echo "🚀 发布构建完成 (./releases/digwis-panel)"

# 构建多平台发布包
release:
	@echo "📦 构建多平台发布包..."
	@chmod +x scripts/build/build-release.sh
	@./scripts/build/build-release.sh
	@echo "✅ 多平台发布包构建完成"

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
	@chmod +x scripts/git-push.sh
	@./scripts/git-push.sh

# 快速推送（带自定义提交信息）
push-msg: build-release
	@echo "📤 推送代码到远程仓库..."
	@chmod +x scripts/git-push.sh
	@read -p "请输入提交信息: " msg; \
	./scripts/git-push.sh "$$msg"

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
	@echo "🧹 清理命令:"
	@echo "   make clean          - 清理构建文件 (保留 releases/)"
	@echo "   make clean-all      - 清理所有文件 (包括 releases/)"
	@echo ""
	@echo "💡 开发帮助:"
	@echo "   make dev-help       - 显示开发环境使用方法"
