#!/bin/bash

# DigWis 面板一键安装脚本
# 支持 Ubuntu/Debian/CentOS/RHEL/Fedora 系统
# 使用方法: curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash

set -e

# 默认配置
VERBOSE=false
QUIET=false
GITHUB_REPO="digwis/digwis-panel"
GITHUB_URL="https://github.com/${GITHUB_REPO}.git"
INSTALL_DIR="/opt/digwis"
CONFIG_DIR="/etc/digwis"
SOURCE_DIR="/tmp/digwis-source"
GO_VERSION="1.21.5"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --quiet|-q)
            QUIET=true
            shift
            ;;
        --help|-h)
            echo "DigWis 面板安装脚本"
            echo ""
            echo "使用方法:"
            echo "  curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash"
            echo ""
            echo "选项:"
            echo "  --verbose, -v    显示详细安装信息"
            echo "  --quiet, -q      静默安装模式"
            echo "  --help, -h       显示此帮助信息"
            exit 0
            ;;
        *)
            echo "未知参数: $1"
            echo "使用 --help 查看帮助信息"
            exit 1
            ;;
    esac
done

# 打印函数
print_info() {
    if [ "$QUIET" != "true" ]; then
        echo -e "${BLUE}[INFO]${NC} $1"
    fi
}

print_success() {
    if [ "$QUIET" != "true" ]; then
        echo -e "${GREEN}[SUCCESS]${NC} $1"
    fi
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

print_warning() {
    if [ "$QUIET" != "true" ]; then
        echo -e "${YELLOW}[WARNING]${NC} $1"
    fi
}

print_step() {
    if [ "$QUIET" != "true" ]; then
        echo -e "${YELLOW}[STEP]${NC} $1"
    fi
}

print_verbose() {
    if [ "$VERBOSE" = "true" ]; then
        echo -e "${BLUE}[VERBOSE]${NC} $1"
    fi
}

# 显示Logo（仅在非静默模式）
show_logo() {
    if [ "$QUIET" != "true" ]; then
        echo -e "${BLUE}"
        echo "=================================="
        echo "    DigWis 面板安装脚本"
        echo "=================================="
        echo -e "${NC}"
    fi
}

# 检查root权限
check_root() {
    if [ "$EUID" -ne 0 ]; then
        print_error "请使用root权限运行此脚本"
        echo ""
        echo "使用方法："
        echo "  sudo bash install.sh"
        echo "  或者："
        echo "  curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash"
        echo ""
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
    print_verbose "检测到系统架构: $ARCH"
}

# 检测操作系统
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        VERSION=$VERSION_ID
    else
        print_error "无法检测操作系统"
        exit 1
    fi
    
    print_verbose "检测到操作系统: $OS $VERSION"
    
    case $OS in
        ubuntu|debian)
            PKG_MANAGER="apt"
            PKG_UPDATE="apt update"
            PKG_INSTALL="apt install -y"
            ;;
        centos|rhel|fedora|rocky|almalinux)
            if command -v dnf >/dev/null 2>&1; then
                PKG_MANAGER="dnf"
                PKG_UPDATE="dnf check-update || true"
                PKG_INSTALL="dnf install -y"
            else
                PKG_MANAGER="yum"
                PKG_UPDATE="yum check-update || true"
                PKG_INSTALL="yum install -y"
            fi
            ;;
        *)
            print_error "不支持的操作系统: $OS"
            exit 1
            ;;
    esac
}

# 安装系统依赖
install_dependencies() {
    print_step "安装系统依赖..."
    
    print_verbose "更新包管理器..."
    $PKG_UPDATE >/dev/null 2>&1 || true
    
    print_verbose "安装必要的软件包..."
    case $PKG_MANAGER in
        apt)
            $PKG_INSTALL curl wget git gcc build-essential >/dev/null 2>&1
            ;;
        dnf|yum)
            $PKG_INSTALL curl wget git gcc gcc-c++ make >/dev/null 2>&1
            ;;
    esac
    
    print_success "系统依赖安装完成"
}

# 安装Go语言环境
install_go() {
    if command -v go >/dev/null 2>&1; then
        CURRENT_GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        if [ "$CURRENT_GO_VERSION" = "$GO_VERSION" ]; then
            print_info "Go $GO_VERSION 已安装，跳过安装"
            return
        fi
    fi
    
    print_step "安装Go语言环境..."
    
    # 确定Go架构
    case $ARCH in
        amd64) GOARCH="amd64" ;;
        arm64) GOARCH="arm64" ;;
        arm) GOARCH="armv6l" ;;
    esac
    
    GO_TARBALL="go${GO_VERSION}.linux-${GOARCH}.tar.gz"
    
    print_verbose "下载Go ${GO_VERSION} for ${GOARCH}..."
    cd /tmp
    
    # 检查是否已经下载过
    if [ -f "${GO_TARBALL}" ]; then
        print_verbose "发现已下载的Go安装包，跳过下载"
    else
        print_info "正在下载Go安装包..."
        
        # 使用多种下载方式确保成功
        if command -v wget >/dev/null 2>&1; then
            if [ "$VERBOSE" = "true" ]; then
                wget --timeout=30 --tries=3 --progress=bar "https://golang.org/dl/${GO_TARBALL}" -O "${GO_TARBALL}" || {
                    print_warning "wget下载失败，尝试使用curl..."
                    curl --connect-timeout 30 --max-time 300 --retry 3 --retry-delay 2 -L "https://golang.org/dl/${GO_TARBALL}" -o "${GO_TARBALL}"
                }
            else
                wget --timeout=30 --tries=3 -q "https://golang.org/dl/${GO_TARBALL}" -O "${GO_TARBALL}" || {
                    print_warning "wget下载失败，尝试使用curl..."
                    curl --connect-timeout 30 --max-time 300 --retry 3 --retry-delay 2 -sL "https://golang.org/dl/${GO_TARBALL}" -o "${GO_TARBALL}"
                }
            fi
        else
            if [ "$VERBOSE" = "true" ]; then
                curl --connect-timeout 30 --max-time 300 --retry 3 --retry-delay 2 -L "https://golang.org/dl/${GO_TARBALL}" -o "${GO_TARBALL}"
            else
                curl --connect-timeout 30 --max-time 300 --retry 3 --retry-delay 2 -sL "https://golang.org/dl/${GO_TARBALL}" -o "${GO_TARBALL}"
            fi
        fi
        
        # 验证下载是否成功
        if [ ! -f "${GO_TARBALL}" ] || [ ! -s "${GO_TARBALL}" ]; then
            print_error "Go安装包下载失败"
            exit 1
        fi
        print_success "Go安装包下载完成"
    fi
    
    print_verbose "安装Go到 /usr/local/go..."
    rm -rf /usr/local/go
    tar -C /usr/local -xzf "${GO_TARBALL}" >/dev/null 2>&1
    rm -f "${GO_TARBALL}"
    
    # 设置环境变量
    if ! grep -q "/usr/local/go/bin" /etc/profile; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    fi
    
    export PATH=$PATH:/usr/local/go/bin
    print_success "Go语言环境安装完成"
}

# 下载并编译面板
build_panel() {
    print_step "下载源码并编译..."
    
    # 清理并克隆源码
    rm -rf $SOURCE_DIR
    print_verbose "正在从GitHub拉取源码..."
    
    # 尝试克隆源码，增加重试机制
    for i in {1..3}; do
        if [ "$VERBOSE" = "true" ]; then
            if git clone --depth=1 "${GITHUB_URL}" "$SOURCE_DIR"; then
                print_success "源码下载完成"
                break
            fi
        else
            if git clone --depth=1 "${GITHUB_URL}" "$SOURCE_DIR" >/dev/null 2>&1; then
                print_success "源码下载完成"
                break
            fi
        fi
        
        print_warning "第${i}次尝试失败，重试中..."
        rm -rf $SOURCE_DIR
        sleep 2
        if [ $i -eq 3 ]; then
            print_error "源码下载失败，请检查网络连接"
            exit 1
        fi
    done
    
    cd "$SOURCE_DIR"
    
    # 设置Go环境
    export PATH=$PATH:/usr/local/go/bin
    export GOPROXY=https://goproxy.cn,direct
    
    print_verbose "编译面板程序..."
    if [ "$VERBOSE" = "true" ]; then
        go mod tidy
        go build -ldflags="-s -w" -o digwis main.go
    else
        go mod tidy >/dev/null 2>&1
        go build -ldflags="-s -w" -o digwis main.go >/dev/null 2>&1
    fi
    
    print_success "面板编译完成"
}

# 安装面板
install_panel() {
    print_step "安装面板..."

    # 创建目录
    print_verbose "创建安装目录..."
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$CONFIG_DIR"
    mkdir -p "/var/log/digwis"

    # 复制文件
    print_verbose "复制程序文件..."
    cp "$SOURCE_DIR/digwis" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/digwis"

    # 创建配置文件
    print_verbose "创建配置文件..."
    cat > "$CONFIG_DIR/config.yaml" << EOF
server:
  port: 8080
  host: "0.0.0.0"

auth:
  session_timeout: 3600

log:
  level: "info"
  file: "/var/log/digwis/digwis.log"
EOF

    print_success "面板安装完成"
}

# 创建系统服务
create_service() {
    print_step "配置系统服务..."

    print_verbose "创建systemd服务文件..."
    cat > /etc/systemd/system/digwis.service << EOF
[Unit]
Description=DigWis Server Management Panel
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/digwis
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

    print_verbose "重新加载systemd配置..."
    systemctl daemon-reload >/dev/null 2>&1
    systemctl enable digwis >/dev/null 2>&1

    print_success "系统服务配置完成"
}

# 配置防火墙
configure_firewall() {
    print_step "配置防火墙..."

    # Ubuntu/Debian 使用 ufw
    if command -v ufw >/dev/null 2>&1; then
        print_verbose "配置ufw防火墙..."
        ufw allow 8080/tcp >/dev/null 2>&1 || true
    fi

    # CentOS/RHEL 使用 firewalld
    if command -v firewall-cmd >/dev/null 2>&1; then
        print_verbose "配置firewalld防火墙..."
        firewall-cmd --permanent --add-port=8080/tcp >/dev/null 2>&1 || true
        firewall-cmd --reload >/dev/null 2>&1 || true
    fi

    print_success "防火墙配置完成"
}

# 启动服务
start_service() {
    print_step "启动面板服务..."

    print_verbose "启动digwis服务..."
    systemctl start digwis

    # 等待服务启动
    sleep 3

    if systemctl is-active --quiet digwis; then
        print_success "面板服务启动成功"
    else
        print_error "面板服务启动失败"
        print_info "查看日志: journalctl -u digwis -f"
        exit 1
    fi
}

# 清理临时文件
cleanup() {
    print_verbose "清理临时文件..."
    rm -rf "$SOURCE_DIR"
    print_verbose "清理完成"
}

# 显示安装结果
show_result() {
    if [ "$QUIET" != "true" ]; then
        echo ""
        echo -e "${GREEN}=================================="
        echo "    DigWis 面板安装完成！"
        echo "==================================${NC}"
        echo ""
        echo "🌐 访问地址:"
        echo "   本地: http://localhost:8080"
        echo "   外网: http://$(curl -s ifconfig.me 2>/dev/null || echo "YOUR_SERVER_IP"):8080"
        echo ""
        echo "🔧 管理命令:"
        echo "   启动服务: systemctl start digwis"
        echo "   停止服务: systemctl stop digwis"
        echo "   重启服务: systemctl restart digwis"
        echo "   查看状态: systemctl status digwis"
        echo "   查看日志: journalctl -u digwis -f"
        echo ""
        echo "📁 安装目录: $INSTALL_DIR"
        echo "📁 配置目录: $CONFIG_DIR"
        echo "📁 日志目录: /var/log/digwis"
        echo ""
        echo -e "${YELLOW}请使用系统用户账号登录面板${NC}"
        echo ""
    fi
}

# 主函数
main() {
    show_logo
    check_root
    detect_arch
    detect_os
    install_dependencies
    install_go
    build_panel
    install_panel
    create_service
    configure_firewall
    start_service
    cleanup
    show_result
}

# 执行主函数
main "$@"
