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

# 检测系统架构
detect_arch() {
    local arch=$(uname -m)
    case $arch in
        x86_64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        armv7l|armv6l)
            echo "arm"
            ;;
        *)
            log_error "不支持的架构: $arch"
            exit 1
            ;;
    esac
}

# 获取最新版本
get_latest_version() {
    log_info "获取最新版本信息..."

    # 尝试从 GitHub API 获取最新版本
    if command -v curl &> /dev/null; then
        LATEST_VERSION=$(curl -s "https://api.github.com/repos/digwis/digwis-panel/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget &> /dev/null; then
        LATEST_VERSION=$(wget -qO- "https://api.github.com/repos/digwis/digwis-panel/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    fi

    # 如果获取失败，使用默认版本
    if [[ -z "$LATEST_VERSION" ]]; then
        LATEST_VERSION="v1.0.0"
        log_warning "无法获取最新版本，使用默认版本: $LATEST_VERSION"
    else
        log_info "最新版本: $LATEST_VERSION"
    fi

    echo "$LATEST_VERSION"
}

# 下载预编译二进制文件
download_binary() {
    log_info "下载预编译二进制文件..."

    local version=$(get_latest_version)
    local arch=$(detect_arch)
    local download_url="https://github.com/digwis/digwis-panel/releases/download/${version}/digwis-panel-${version}-linux-${arch}.tar.gz"

    log_info "下载地址: $download_url"

    # 创建安装目录
    if [[ -d "$INSTALL_DIR" ]]; then
        log_info "备份现有安装..."
        sudo mv "$INSTALL_DIR" "${INSTALL_DIR}.backup.$(date +%s)"
    fi
    sudo mkdir -p "$INSTALL_DIR"

    # 下载文件
    cd /tmp
    if command -v curl &> /dev/null; then
        curl -L -o "digwis-panel.tar.gz" "$download_url"
    elif command -v wget &> /dev/null; then
        wget -O "digwis-panel.tar.gz" "$download_url"
    else
        log_error "需要 curl 或 wget 来下载文件"
        exit 1
    fi

    # 解压文件
    tar -xzf "digwis-panel.tar.gz"

    # 移动到安装目录
    sudo mv digwis "$INSTALL_DIR/digwis-panel"
    sudo chmod +x "$INSTALL_DIR/digwis-panel"
    sudo chown root:root "$INSTALL_DIR/digwis-panel"

    # 清理临时文件
    rm -f "digwis-panel.tar.gz"

    log_success "二进制文件下载完成"
}

# 备用方案：克隆并构建
clone_and_build() {
    log_warning "使用备用方案：克隆源码并构建"

    # 如果目录已存在，先备份
    if [[ -d "$INSTALL_DIR" ]]; then
        log_info "备份现有安装..."
        sudo mv "$INSTALL_DIR" "${INSTALL_DIR}.backup.$(date +%s)"
    fi

    # 克隆项目
    sudo git clone "$REPO_URL" "$INSTALL_DIR"
    sudo chown -R $USER:$USER "$INSTALL_DIR"

    cd "$INSTALL_DIR"

    # 下载依赖
    go mod download

    # 构建项目
    go build -o digwis-panel .

    # 设置执行权限
    chmod +x digwis-panel

    log_success "源码构建完成"
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

    # 尝试下载预编译版本，失败则构建
    if ! download_binary; then
        log_warning "预编译版本下载失败，尝试源码构建..."
        clone_and_build
    fi

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
