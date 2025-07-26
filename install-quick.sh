#!/bin/bash

# DigWis 面板快速安装脚本
# 从源码编译安装，适用于VPS快速部署
# 使用方法: curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install-quick.sh | bash

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置信息
GITHUB_REPO="digwis/digwis-panel"
GITHUB_URL="https://github.com/${GITHUB_REPO}.git"
INSTALL_DIR="/opt/digwis"
CONFIG_DIR="/etc/digwis"
SOURCE_DIR="/tmp/digwis-source"

# 打印函数
print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }
print_step() { echo -e "${YELLOW}[STEP]${NC} $1"; }

# 检查root权限
check_root() {
    if [ "$EUID" -ne 0 ]; then
        print_error "请使用root权限运行此脚本"
        echo "使用方法: sudo bash install-quick.sh"
        exit 1
    fi
}

# 检测系统架构
detect_arch() {
    ARCH=$(uname -m)
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
        armv7l) ARCH="arm" ;;
        *) 
            print_error "不支持的系统架构: $ARCH"
            exit 1
            ;;
    esac
    print_info "系统架构: $ARCH"
}

# 安装基础依赖
install_deps() {
    print_step "安装基础依赖..."
    
    if command -v apt-get >/dev/null 2>&1; then
        apt-get update -qq
        apt-get install -y curl wget git build-essential systemd >/dev/null 2>&1
    elif command -v yum >/dev/null 2>&1; then
        yum install -y curl wget git gcc systemd >/dev/null 2>&1
    elif command -v dnf >/dev/null 2>&1; then
        dnf install -y curl wget git gcc systemd >/dev/null 2>&1
    else
        print_error "不支持的包管理器"
        exit 1
    fi
    
    print_success "基础依赖安装完成"
}

# 安装Go
install_go() {
    print_step "安装Go语言环境..."
    
    # 检查Go是否已安装且版本合适
    if command -v go >/dev/null 2>&1; then
        GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        if [ "$(printf '%s\n' "1.19" "$GO_VERSION" | sort -V | head -n1)" = "1.19" ]; then
            print_info "Go已安装，版本: $GO_VERSION"
            return 0
        fi
    fi
    
    # 下载并安装Go
    GO_VERSION="1.21.5"
    case $ARCH in
        amd64) GOARCH="amd64" ;;
        arm64) GOARCH="arm64" ;;
        arm) GOARCH="armv6l" ;;
    esac
    
    GO_TARBALL="go${GO_VERSION}.linux-${GOARCH}.tar.gz"
    
    print_info "下载Go ${GO_VERSION}..."
    cd /tmp
    curl -sL "https://golang.org/dl/${GO_TARBALL}" -o "${GO_TARBALL}"
    
    print_info "安装Go..."
    rm -rf /usr/local/go
    tar -C /usr/local -xzf "${GO_TARBALL}"
    rm -f "${GO_TARBALL}"
    
    # 设置环境变量
    export PATH=$PATH:/usr/local/go/bin
    export GOPROXY=https://goproxy.cn,direct
    
    print_success "Go安装完成"
}

# 下载并编译
build_panel() {
    print_step "下载源码并编译..."
    
    # 清理并克隆源码
    rm -rf $SOURCE_DIR
    git clone --depth=1 "${GITHUB_URL}" "$SOURCE_DIR" >/dev/null 2>&1
    
    cd "$SOURCE_DIR"
    
    # 设置Go环境
    export PATH=$PATH:/usr/local/go/bin
    export GOPROXY=https://goproxy.cn,direct
    export CGO_ENABLED=0
    export GOOS=linux
    export GOARCH=$ARCH
    
    # 编译
    print_info "编译中..."
    /usr/local/go/bin/go mod tidy >/dev/null 2>&1
    /usr/local/go/bin/go build -ldflags '-s -w' -o digwis-panel main.go
    
    # 安装
    mkdir -p $INSTALL_DIR $CONFIG_DIR
    cp digwis-panel $INSTALL_DIR/
    chmod +x $INSTALL_DIR/digwis-panel
    
    # 清理
    cd / && rm -rf $SOURCE_DIR
    
    print_success "编译安装完成"
}

# 创建配置
create_config() {
    print_step "创建配置文件..."
    
    SECRET_KEY=$(openssl rand -hex 32 2>/dev/null || head -c 32 /dev/urandom | xxd -p)
    
    cat > $CONFIG_DIR/config.yaml << EOF
debug: false
auth:
  session_timeout: 3600
  secret_key: "${SECRET_KEY}"
server:
  port: "8080"
paths:
  data_dir: "/var/lib/digwis"
  log_dir: "/var/log/digwis"
EOF
    
    mkdir -p /var/lib/digwis /var/log/digwis
    print_success "配置文件创建完成"
}

# 创建服务
create_service() {
    print_step "创建系统服务..."
    
    cat > /etc/systemd/system/digwis.service << EOF
[Unit]
Description=DigWis Server Panel
After=network.target

[Service]
Type=simple
User=root
ExecStart=${INSTALL_DIR}/digwis-panel -config ${CONFIG_DIR}/config.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
    
    systemctl daemon-reload
    systemctl enable digwis >/dev/null 2>&1
    
    print_success "系统服务创建完成"
}

# 启动服务
start_service() {
    print_step "启动服务..."
    
    systemctl start digwis
    sleep 2
    
    if systemctl is-active --quiet digwis; then
        print_success "服务启动成功"
    else
        print_error "服务启动失败"
        exit 1
    fi
}

# 显示结果
show_result() {
    SERVER_IP=$(curl -s ifconfig.me 2>/dev/null || echo "YOUR_SERVER_IP")
    
    echo ""
    echo -e "${GREEN}🎉 DigWis面板安装成功！${NC}"
    echo "================================"
    echo ""
    echo -e "${BLUE}访问地址:${NC}"
    echo "  http://localhost:8080"
    echo "  http://${SERVER_IP}:8080"
    echo ""
    echo -e "${BLUE}管理命令:${NC}"
    echo "  启动: systemctl start digwis"
    echo "  停止: systemctl stop digwis"
    echo "  重启: systemctl restart digwis"
    echo "  状态: systemctl status digwis"
    echo "  日志: journalctl -u digwis -f"
    echo ""
    echo -e "${YELLOW}请确保防火墙已开放8080端口${NC}"
    echo ""
}

# 主函数
main() {
    echo -e "${BLUE}DigWis 面板快速安装脚本${NC}"
    echo "================================"
    
    check_root
    detect_arch
    install_deps
    install_go
    build_panel
    create_config
    create_service
    start_service
    show_result
}

# 错误处理
trap 'print_error "安装失败，请检查错误信息"; exit 1' ERR

# 开始安装
main "$@"
