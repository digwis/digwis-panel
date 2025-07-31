# DigWis Panel Makefile
# 用于生产构建和 CI/CD，开发时请使用 Air: /home/parallels/go/bin/air

.PHONY: help clean install test release-build deploy-local rollback dev install-air

# 默认目标
help:
	@echo "DigWis Panel 构建工具"
	@echo ""
	@echo "🚀 开发环境:"
	@echo "  air            - 启动热重载开发服务器 (推荐)"
	@echo ""
	@echo "📦 发布构建:"
	@echo "  release-build  - 构建发布包到 releases 目录"
	@echo ""
	@echo "� 快速部署:"
	@echo "  deploy-local   - 编译并部署到本地生产环境"
	@echo "  rollback       - 回滚到上一个版本"
	@echo ""
	@echo "📦 版本发布:"
	@echo "  release-build  - 构建发布包到 releases 目录"
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



# 构建发布包
release-build:
	@echo "📦 构建发布包..."
	@chmod +x build_release.sh
	@./build_release.sh
	@echo "✅ 发布包构建完成"

# 清理构建文件
clean:
	@echo "🧹 清理构建文件..."
	rm -f digwis-panel
	rm -f assets/css/output.css
	rm -rf tmp/
	@echo "✅ 清理完成"

# 清理所有文件（包括 releases 和 tools）
clean-all:
	@echo "🧹 清理所有构建文件..."
	rm -f digwis-panel
	rm -f assets/css/output.css
	rm -rf tmp/
	rm -rf releases/
	rm -rf tools/
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

# Air 固定位置
AIR_PATH = ./tools/air

# 安装 Air 到项目本地
install-air:
	@echo "📦 安装 Air 到项目本地..."
	@mkdir -p tools
	@export GOBIN=$$(pwd)/tools && /usr/local/go/bin/go install github.com/air-verse/air@latest
	@echo "✅ Air 安装完成: $(AIR_PATH)"

# 开发环境
dev:
	@echo "🚀 启动 DigWis Panel 开发环境"
	@echo "=================================="
	@export PATH=$$HOME/local/go/bin:$$HOME/local/node-v20.18.0-linux-arm64/bin:$$HOME/local:$$PATH; \
	export GOPATH=$$HOME/go; \
	echo "✅ 环境变量设置完成"; \
	echo "📁 Go 路径: $$(which go 2>/dev/null || echo '未找到')"; \
	echo "📁 Node 路径: $$(which node 2>/dev/null || echo '未找到')"; \
	echo "📁 npm 路径: $$(which npm 2>/dev/null || echo '未找到')"; \
	echo "=================================="; \
	echo "🔍 查找 Air 命令..."; \
	if [ -f "$(AIR_PATH)" ]; then \
		echo "✅ 使用项目本地 Air: $(AIR_PATH)"; \
		echo "🔥 启动 Air 热重载..."; \
		$(AIR_PATH); \
	elif command -v air >/dev/null 2>&1; then \
		echo "✅ 使用系统 Air: $$(which air)"; \
		echo "🔥 启动 Air 热重载..."; \
		air; \
	elif [ -f "$$HOME/go/bin/air" ]; then \
		echo "✅ 使用 Go Air: $$HOME/go/bin/air"; \
		echo "🔥 启动 Air 热重载..."; \
		$$HOME/go/bin/air; \
	else \
		echo "❌ 未找到 air 命令"; \
		echo "🔧 自动安装 Air 到项目本地..."; \
		$(MAKE) install-air; \
		echo "🔥 启动 Air 热重载..."; \
		$(AIR_PATH); \
	fi

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

# Screen 开发环境管理
dev-screen:
	@echo "🚀 启动 Screen 开发环境"
	@echo "=================================="
	@if screen -list | grep -q "digwis-dev"; then \
		echo "⚠️  Screen 会话 'digwis-dev' 已存在"; \
		echo "🔗 连接到现有会话: screen -r digwis-dev"; \
		screen -r digwis-dev; \
	else \
		echo "📱 创建新的 Screen 会话..."; \
		screen -S digwis-dev -c /dev/null bash -c 'cd /media/psf/Linux-86/digwis-panel && make dev; exec bash'; \
	fi

dev-screen-detach:
	@echo "🔌 分离 Screen 会话 (服务继续运行)"
	@screen -S digwis-dev -X detach 2>/dev/null || echo "❌ 没有找到活跃的 digwis-dev 会话"

dev-screen-attach:
	@echo "🔗 连接到 Screen 开发会话"
	@screen -r digwis-dev || echo "❌ 没有找到 digwis-dev 会话，请先运行 'make dev-screen'"

dev-screen-status:
	@echo "📊 Screen 会话状态:"
	@screen -list | grep digwis || echo "❌ 没有找到 digwis 相关会话"
	@echo ""
	@echo "🔍 检查服务状态:"
	@if pgrep -f "digwis-panel.*9090" > /dev/null; then \
		echo "✅ DigWis Panel 服务正在运行 (端口 9090)"; \
		echo "🌐 访问地址: http://$(hostname -I | awk '{print $$1}'):9090"; \
	else \
		echo "❌ DigWis Panel 服务未运行"; \
	fi

dev-screen-stop:
	@echo "🛑 停止 Screen 开发环境"
	@screen -S digwis-dev -X quit 2>/dev/null && echo "✅ Screen 会话已终止" || echo "❌ 没有找到活跃的会话"

dev-screen-restart:
	@echo "🔄 重启 Screen 开发环境"
	@make dev-screen-stop
	@sleep 2
	@make dev-screen

# 开发提示
dev-help:
	@echo "💡 开发环境使用方法:"
	@echo ""
	@echo "🖥️  本地开发:"
	@echo "   make dev                    # 启动 Air 热重载 (推荐)"
	@echo "   air                         # 直接使用 Air (如果已安装)"
	@echo "   go run main.go              # 直接运行 (无热重载)"
	@echo ""
	@echo "📱 Screen 开发 (VPS推荐):"
	@echo "   make dev-screen             # 启动/连接 Screen 开发环境"
	@echo "   make dev-screen-attach      # 连接到现有 Screen 会话"
	@echo "   make dev-screen-detach      # 分离 Screen 会话 (保持运行)"
	@echo "   make dev-screen-status      # 查看 Screen 和服务状态"
	@echo "   make dev-screen-stop        # 停止 Screen 开发环境"
	@echo "   make dev-screen-restart     # 重启 Screen 开发环境"
	@echo ""
	@echo "⌨️  Screen 快捷键:"
	@echo "   Ctrl+A, D                   # 分离会话 (服务继续运行)"
	@echo "   Ctrl+A, K                   # 终止会话"
	@echo "   Ctrl+A, ?                   # 显示帮助"

# 部署相关命令 (使用 releases 版本)
deploy: release-build
	@echo "🚀 开始部署..."
	@chmod +x scripts/deploy.sh
	@./scripts/deploy.sh

# 生成安装脚本
install-script:
	@echo "📦 生成安装脚本..."
	@chmod +x scripts/deploy.sh
	@./scripts/deploy.sh --install-script-only

# 推送代码到远程仓库 (使用 releases 版本)
push: release-build
	@echo "📤 推送代码到远程仓库..."
	@chmod +x scripts/simple-push.sh
	@./scripts/simple-push.sh

# 快速推送（带自定义提交信息）
push-msg: release-build
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
release-version: release-build
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
	@echo "📦 发布命令:"
	@echo "   make release-build  - 构建发布包到 releases 目录"
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
