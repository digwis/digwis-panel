#!/bin/bash

# DigWis Panel 部署脚本
# 用于将项目推送到远程仓库并提供一键安装功能

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目信息
PROJECT_NAME="DigWis Panel"
REPO_URL="https://github.com/digwis/digwis-panel.git"
INSTALL_DIR="/opt/digwis-panel"
SERVICE_NAME="digwis-panel"
SERVICE_PORT="9091"

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        log_error "$1 命令未找到，请先安装 $1"
        exit 1
    fi
}

# 检查 Git 仓库状态
check_git_status() {
    log_info "检查 Git 仓库状态..."
    
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        log_error "当前目录不是 Git 仓库"
        exit 1
    fi
    
    # 检查是否有未提交的更改
    if ! git diff-index --quiet HEAD --; then
        log_warning "检测到未提交的更改"
        return 1
    fi
    
    return 0
}

# 提交并推送代码
commit_and_push() {
    log_info "准备提交并推送代码..."
    
    # 添加所有文件
    git add .
    
    # 检查是否有更改需要提交
    if git diff-index --quiet HEAD --; then
        log_info "没有新的更改需要提交"
        return 0
    fi
    
    # 生成提交信息
    COMMIT_MSG="Deploy: $(date '+%Y-%m-%d %H:%M:%S')"
    
    # 提交更改
    git commit -m "$COMMIT_MSG"
    log_success "代码已提交: $COMMIT_MSG"
    
    # 推送到远程仓库
    git push origin main
    log_success "代码已推送到远程仓库"
}

# 生成一键安装脚本
generate_install_script() {
    log_info "生成一键安装脚本..."
    
    cat > install-digwis-panel.sh << 'EOF'
#!/bin/bash

# DigWis Panel 一键安装脚本
# 使用方法: curl -fsSL https://raw.githubusercontent.com/your-username/digwis-panel/main/install-digwis-panel.sh | bash

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置
REPO_URL="https://github.com/digwis/digwis-panel.git"
INSTALL_DIR="/opt/digwis-panel"
SERVICE_NAME="digwis-panel"
SERVICE_PORT="9091"

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查系统要求
check_requirements() {
    log_info "检查系统要求..."
    
    # 检查操作系统
    if [[ "$OSTYPE" != "linux-gnu"* ]]; then
        log_error "仅支持 Linux 系统"
        exit 1
    fi
    
    # 检查是否为 root 用户或有 sudo 权限
    if [[ $EUID -ne 0 ]] && ! sudo -n true 2>/dev/null; then
        log_error "需要 root 权限或 sudo 权限"
        exit 1
    fi
    
    log_success "系统要求检查通过"
}

# 安装依赖
install_dependencies() {
    log_info "安装依赖..."
    
    # 更新包列表
    sudo apt update
    
    # 安装必要的包
    sudo apt install -y git curl wget unzip
    
    # 检查并安装 Go
    if ! command -v go &> /dev/null; then
        log_info "安装 Go..."
        GO_VERSION="1.23.1"
        wget -q "https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz"
        sudo tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"
        echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee -a /etc/profile
        export PATH=$PATH:/usr/local/go/bin
        rm "go${GO_VERSION}.linux-amd64.tar.gz"
        log_success "Go 安装完成"
    else
        log_info "Go 已安装: $(go version)"
    fi
}

# 克隆项目
clone_project() {
    log_info "克隆项目..."
    
    # 如果目录已存在，先备份
    if [[ -d "$INSTALL_DIR" ]]; then
        log_info "备份现有安装..."
        sudo mv "$INSTALL_DIR" "${INSTALL_DIR}.backup.$(date +%s)"
    fi
    
    # 克隆项目
    sudo git clone "$REPO_URL" "$INSTALL_DIR"
    sudo chown -R $USER:$USER "$INSTALL_DIR"
    
    log_success "项目克隆完成"
}

# 构建项目
build_project() {
    log_info "构建项目..."
    
    cd "$INSTALL_DIR"
    
    # 下载依赖
    go mod download
    
    # 构建项目
    go build -o digwis-panel .
    
    # 设置执行权限
    chmod +x digwis-panel
    
    log_success "项目构建完成"
}

# 创建系统服务
create_service() {
    log_info "创建系统服务..."
    
    sudo tee /etc/systemd/system/${SERVICE_NAME}.service > /dev/null << EOL
[Unit]
Description=DigWis Panel - Web-based Server Management Panel
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/digwis-panel -port ${SERVICE_PORT}
Restart=always
RestartSec=5
Environment=PATH=/usr/local/go/bin:/usr/bin:/bin
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
EOL

    # 重新加载 systemd
    sudo systemctl daemon-reload
    
    # 启用服务
    sudo systemctl enable ${SERVICE_NAME}
    
    log_success "系统服务创建完成"
}

# 配置防火墙
configure_firewall() {
    log_info "配置防火墙..."
    
    if command -v ufw &> /dev/null; then
        sudo ufw allow ${SERVICE_PORT}/tcp
        log_success "UFW 防火墙规则已添加"
    elif command -v firewall-cmd &> /dev/null; then
        sudo firewall-cmd --permanent --add-port=${SERVICE_PORT}/tcp
        sudo firewall-cmd --reload
        log_success "Firewalld 防火墙规则已添加"
    else
        log_info "未检测到防火墙，跳过配置"
    fi
}

# 启动服务
start_service() {
    log_info "启动服务..."
    
    sudo systemctl start ${SERVICE_NAME}
    
    # 检查服务状态
    if sudo systemctl is-active --quiet ${SERVICE_NAME}; then
        log_success "DigWis Panel 启动成功"
        log_success "访问地址: http://$(hostname -I | awk '{print $1}'):${SERVICE_PORT}"
    else
        log_error "服务启动失败"
        sudo systemctl status ${SERVICE_NAME}
        exit 1
    fi
}

# 主函数
main() {
    echo "🚀 DigWis Panel 一键安装脚本"
    echo "================================"
    
    check_requirements
    install_dependencies
    clone_project
    build_project
    create_service
    configure_firewall
    start_service
    
    echo ""
    echo "🎉 安装完成！"
    echo "================================"
    echo "服务状态: sudo systemctl status ${SERVICE_NAME}"
    echo "查看日志: sudo journalctl -u ${SERVICE_NAME} -f"
    echo "停止服务: sudo systemctl stop ${SERVICE_NAME}"
    echo "重启服务: sudo systemctl restart ${SERVICE_NAME}"
    echo ""
}

main "$@"
EOF

    chmod +x install-digwis-panel.sh
    log_success "一键安装脚本已生成: install-digwis-panel.sh"
}

# 生成卸载脚本
generate_uninstall_script() {
    log_info "生成卸载脚本..."
    
    cat > uninstall-digwis-panel.sh << 'EOF'
#!/bin/bash

# DigWis Panel 卸载脚本

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

INSTALL_DIR="/opt/digwis-panel"
SERVICE_NAME="digwis-panel"

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

main() {
    echo "🗑️  DigWis Panel 卸载脚本"
    echo "=========================="
    
    # 确认卸载
    read -p "确定要卸载 DigWis Panel 吗？(y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "取消卸载"
        exit 0
    fi
    
    # 停止服务
    if sudo systemctl is-active --quiet ${SERVICE_NAME}; then
        log_info "停止服务..."
        sudo systemctl stop ${SERVICE_NAME}
    fi
    
    # 禁用服务
    if sudo systemctl is-enabled --quiet ${SERVICE_NAME}; then
        log_info "禁用服务..."
        sudo systemctl disable ${SERVICE_NAME}
    fi
    
    # 删除服务文件
    if [[ -f "/etc/systemd/system/${SERVICE_NAME}.service" ]]; then
        log_info "删除服务文件..."
        sudo rm "/etc/systemd/system/${SERVICE_NAME}.service"
        sudo systemctl daemon-reload
    fi
    
    # 删除安装目录
    if [[ -d "$INSTALL_DIR" ]]; then
        log_info "删除安装目录..."
        sudo rm -rf "$INSTALL_DIR"
    fi
    
    log_success "DigWis Panel 已完全卸载"
}

main "$@"
EOF

    chmod +x uninstall-digwis-panel.sh
    log_success "卸载脚本已生成: uninstall-digwis-panel.sh"
}

# 显示使用帮助
show_help() {
    echo "🚀 $PROJECT_NAME 部署脚本"
    echo "=========================="
    echo ""
    echo "使用方法:"
    echo "  $0 [选项]"
    echo ""
    echo "选项:"
    echo "  --install-script-only    仅生成安装脚本，不推送代码"
    echo "  --no-push               生成脚本但不推送到远程仓库"
    echo "  --force                 强制部署，忽略未提交的更改"
    echo "  --help, -h              显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0                      # 完整部署"
    echo "  $0 --install-script-only # 仅生成安装脚本"
    echo "  $0 --force              # 强制部署"
    echo ""
}

# 主函数
main() {
    local INSTALL_SCRIPT_ONLY=false
    local NO_PUSH=false
    local FORCE=false

    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --install-script-only)
                INSTALL_SCRIPT_ONLY=true
                shift
                ;;
            --no-push)
                NO_PUSH=true
                shift
                ;;
            --force)
                FORCE=true
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                log_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done

    echo "🚀 $PROJECT_NAME 部署脚本"
    echo "=========================="

    # 如果只生成安装脚本
    if [[ "$INSTALL_SCRIPT_ONLY" == true ]]; then
        log_info "仅生成安装脚本模式"
        generate_install_script
        generate_uninstall_script
        log_success "安装脚本生成完成！"
        return 0
    fi

    # 检查必要命令
    check_command "git"
    check_command "go"

    # 检查 Git 状态
    if ! check_git_status && [[ "$FORCE" != true ]]; then
        read -p "检测到未提交的更改，是否继续？(y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "部署已取消"
            exit 0
        fi
    fi

    # 提交并推送代码 (除非指定不推送)
    if [[ "$NO_PUSH" != true ]]; then
        commit_and_push
    else
        log_info "跳过代码推送"
    fi

    # 生成安装脚本
    generate_install_script
    generate_uninstall_script

    echo ""
    log_success "部署完成！"
    echo "=========================="
    echo "一键安装命令:"
    echo "curl -fsSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install-digwis-panel.sh | bash"
    echo ""
    echo "或者下载后执行:"
    echo "wget https://raw.githubusercontent.com/digwis/digwis-panel/main/install-digwis-panel.sh"
    echo "chmod +x install-digwis-panel.sh"
    echo "./install-digwis-panel.sh"
    echo ""
    echo "📁 生成的文件:"
    echo "  - install-digwis-panel.sh   (一键安装脚本)"
    echo "  - uninstall-digwis-panel.sh (卸载脚本)"
    echo ""
}

# 如果直接执行此脚本
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
