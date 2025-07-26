#!/bin/bash

# DigWis 服务器管理面板一键安装脚本（源码编译版）
# 支持 Ubuntu/Debian/CentOS/RHEL 系统
# 自动安装Go环境，从GitHub拉取源码并编译安装
# 使用方法: curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install-remote.sh | bash

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# 版本信息
PANEL_VERSION="1.0.0"
GITHUB_REPO="digwis/digwis-panel"
GITHUB_URL="https://github.com/${GITHUB_REPO}.git"

# 安装目录
INSTALL_DIR="/opt/digwis"
CONFIG_DIR="/etc/server-panel"
DATA_DIR="/var/lib/server-panel"
LOG_DIR="/var/log/server-panel"
SOURCE_DIR="/tmp/digwis-panel-source"

# 打印函数
print_logo() {
    echo -e "${CYAN}"
    echo "  ____  _       __        ___      "
    echo " |  _ \(_) __ _/ /  ___  / (_)___  "
    echo " | | | | |/ _\` | | / _ \| | / __| "
    echo " | |_| | | (_| | |/ (_) | | \__ \ "
    echo " |____/|_|\__, |_|\___/|_|_|___/ "
    echo "          |___/                   "
    echo -e "${NC}"
    echo -e "${BLUE}DigWis 服务器管理面板 v${PANEL_VERSION}${NC}"
    echo -e "${BLUE}一键安装脚本${NC}"
    echo "================================"
}

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_step() {
    echo -e "${PURPLE}[STEP]${NC} $1"
}

# 检查系统要求
check_system() {
    print_step "检查系统环境..."
    
    # 检查是否为root用户
    if [ "$EUID" -ne 0 ]; then
        print_error "请使用root权限运行此脚本"
        echo "使用方法: sudo bash install.sh"
        exit 1
    fi
    
    # 检测操作系统
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$NAME
        VER=$VERSION_ID
    else
        print_error "无法检测操作系统版本"
        exit 1
    fi
    
    print_info "检测到系统: $OS $VER"
    
    # 检查架构
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
    
    # 检查内存
    MEMORY=$(free -m | awk 'NR==2{printf "%.0f", $2}')
    if [ $MEMORY -lt 512 ]; then
        print_warning "系统内存较低 (${MEMORY}MB)，建议至少512MB"
    fi
    
    # 检查磁盘空间
    DISK_SPACE=$(df / | awk 'NR==2 {print $4}')
    if [ $DISK_SPACE -lt 1048576 ]; then  # 1GB in KB
        print_warning "磁盘空间不足，建议至少1GB可用空间"
    fi
    
    print_success "系统检查完成"
}

# 安装依赖
install_dependencies() {
    print_step "安装系统依赖..."

    if command -v apt-get >/dev/null 2>&1; then
        # Debian/Ubuntu
        apt-get update -qq
        apt-get install -y curl wget git build-essential systemd
        PACKAGE_MANAGER="apt"
    elif command -v yum >/dev/null 2>&1; then
        # CentOS/RHEL
        yum update -y -q
        yum install -y curl wget git gcc systemd
        PACKAGE_MANAGER="yum"
    elif command -v dnf >/dev/null 2>&1; then
        # Fedora
        dnf update -y -q
        dnf install -y curl wget git gcc systemd
        PACKAGE_MANAGER="dnf"
    else
        print_error "不支持的包管理器"
        exit 1
    fi

    print_success "系统依赖安装完成"
}

# 安装Go语言环境
install_go() {
    print_step "检查Go语言环境..."

    # 检查Go是否已安装
    if command -v go >/dev/null 2>&1; then
        GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        print_info "检测到Go版本: $GO_VERSION"

        # 检查版本是否满足要求（至少1.19）
        if [ "$(printf '%s\n' "1.19" "$GO_VERSION" | sort -V | head -n1)" = "1.19" ]; then
            print_success "Go版本满足要求"
            return 0
        else
            print_warning "Go版本过低，需要升级"
        fi
    fi

    print_step "安装Go语言环境..."

    # 检测系统架构
    case $ARCH in
        amd64) GOARCH="amd64" ;;
        arm64) GOARCH="arm64" ;;
        arm) GOARCH="armv6l" ;;
        *)
            print_error "不支持的Go架构: $ARCH"
            exit 1
            ;;
    esac

    # 下载并安装Go
    GO_VERSION="1.21.5"
    GO_TARBALL="go${GO_VERSION}.linux-${GOARCH}.tar.gz"

    print_info "下载Go ${GO_VERSION} for ${GOARCH}..."
    cd /tmp

    if command -v wget >/dev/null 2>&1; then
        wget -q --show-progress "https://golang.org/dl/${GO_TARBALL}"
    elif command -v curl >/dev/null 2>&1; then
        curl -L "https://golang.org/dl/${GO_TARBALL}" -o "${GO_TARBALL}"
    else
        print_error "未找到wget或curl命令"
        exit 1
    fi

    # 安装Go
    print_info "安装Go到/usr/local/go..."
    rm -rf /usr/local/go
    tar -C /usr/local -xzf "${GO_TARBALL}"

    # 设置环境变量
    export PATH=$PATH:/usr/local/go/bin
    export GOPATH=/root/go
    export GOPROXY=https://goproxy.cn,direct

    # 添加到系统环境变量
    cat >> /etc/profile << 'EOF'
export PATH=$PATH:/usr/local/go/bin
export GOPATH=/root/go
export GOPROXY=https://goproxy.cn,direct
EOF

    # 清理下载文件
    rm -f "${GO_TARBALL}"

    # 验证安装
    if /usr/local/go/bin/go version >/dev/null 2>&1; then
        print_success "Go安装成功: $(/usr/local/go/bin/go version)"
    else
        print_error "Go安装失败"
        exit 1
    fi
}

# 下载并编译面板程序
download_and_build_panel() {
    print_step "下载源代码..."

    # 清理可能存在的源码目录
    rm -rf $SOURCE_DIR

    # 克隆源代码
    print_info "从GitHub克隆源代码: ${GITHUB_URL}"
    git clone --depth=1 "${GITHUB_URL}" "$SOURCE_DIR"

    if [ ! -d "$SOURCE_DIR" ]; then
        print_error "源代码下载失败"
        exit 1
    fi

    print_success "源代码下载完成"

    print_step "编译面板程序..."

    # 进入源码目录
    cd "$SOURCE_DIR"

    # 设置Go环境变量
    export PATH=$PATH:/usr/local/go/bin
    export GOPATH=/root/go
    export GOPROXY=https://goproxy.cn,direct
    export CGO_ENABLED=0
    export GOOS=linux
    export GOARCH=$ARCH

    # 下载依赖
    print_info "下载Go依赖..."
    /usr/local/go/bin/go mod tidy

    # 编译程序
    print_info "编译二进制文件..."
    /usr/local/go/bin/go build -a -ldflags '-extldflags "-static" -s -w' -o digwis-panel main.go

    # 验证编译结果
    if [ ! -f "digwis-panel" ]; then
        print_error "编译失败"
        exit 1
    fi

    # 创建安装目录
    mkdir -p $INSTALL_DIR

    # 复制二进制文件
    cp digwis-panel $INSTALL_DIR/
    chmod +x $INSTALL_DIR/digwis-panel

    # 获取文件大小
    SIZE=$(du -h $INSTALL_DIR/digwis-panel | cut -f1)
    print_info "二进制文件大小: ${SIZE}"

    # 清理源码目录
    cd /
    rm -rf $SOURCE_DIR

    print_success "程序编译完成"
}

# 创建配置文件
create_config() {
    print_step "创建配置文件..."
    
    # 创建配置目录
    mkdir -p $CONFIG_DIR $DATA_DIR $LOG_DIR
    
    # 生成随机密钥
    SECRET_KEY=$(openssl rand -hex 32 2>/dev/null || head -c 32 /dev/urandom | xxd -p)
    
    # 创建配置文件
    cat > $CONFIG_DIR/config.yaml << EOF
# DigWis 服务器管理面板配置文件
debug: false

auth:
  session_timeout: 3600  # 1小时
  max_login_attempts: 5
  lockout_duration: 900   # 15分钟
  secret_key: "${SECRET_KEY}"

server:
  port: "8080"
  read_timeout: 30
  write_timeout: 30
  idle_timeout: 120

paths:
  data_dir: "${DATA_DIR}"
  log_dir: "${LOG_DIR}"
  temp_dir: "/tmp/server-panel"
  backup_dir: "/var/backups/server-panel"
EOF
    
    print_success "配置文件创建完成"
}

# 创建系统服务
create_service() {
    print_step "创建系统服务..."
    
    cat > /etc/systemd/system/digwis-panel.service << EOF
[Unit]
Description=DigWis Server Management Panel
Documentation=https://github.com/${GITHUB_REPO}
After=network.target
Wants=network.target

[Service]
Type=simple
User=root
Group=root
ExecStart=${INSTALL_DIR}/digwis-panel -config ${CONFIG_DIR}/config.yaml -port 8080
ExecReload=/bin/kill -HUP \$MAINPID
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=digwis-panel

# 安全设置
NoNewPrivileges=false

# 环境变量
Environment=SERVER_PANEL_SECRET=${SECRET_KEY}

[Install]
WantedBy=multi-user.target
EOF
    
    # 重新加载systemd
    systemctl daemon-reload
    systemctl enable digwis-panel
    
    print_success "系统服务创建完成"
}

# 配置防火墙
configure_firewall() {
    print_step "配置防火墙..."
    
    # UFW (Ubuntu)
    if command -v ufw >/dev/null 2>&1; then
        if ufw status | grep -q "Status: active"; then
            ufw allow 8080/tcp >/dev/null 2>&1
            print_info "UFW防火墙已允许8080端口"
        fi
    fi
    
    # firewalld (CentOS/RHEL)
    if command -v firewall-cmd >/dev/null 2>&1; then
        if systemctl is-active --quiet firewalld; then
            firewall-cmd --permanent --add-port=8080/tcp >/dev/null 2>&1
            firewall-cmd --reload >/dev/null 2>&1
            print_info "firewalld防火墙已允许8080端口"
        fi
    fi
    
    # iptables
    if command -v iptables >/dev/null 2>&1; then
        if iptables -L | grep -q "Chain INPUT"; then
            iptables -I INPUT -p tcp --dport 8080 -j ACCEPT >/dev/null 2>&1 || true
            print_info "iptables防火墙已允许8080端口"
        fi
    fi
    
    print_success "防火墙配置完成"
}

# 创建管理命令
create_management_script() {
    print_step "创建管理命令..."
    
    cat > /usr/local/bin/digwis << 'EOF'
#!/bin/bash

SERVICE_NAME="digwis-panel"

case "$1" in
    start)
        systemctl start $SERVICE_NAME
        echo "DigWis面板已启动"
        ;;
    stop)
        systemctl stop $SERVICE_NAME
        echo "DigWis面板已停止"
        ;;
    restart)
        systemctl restart $SERVICE_NAME
        echo "DigWis面板已重启"
        ;;
    status)
        systemctl status $SERVICE_NAME
        ;;
    logs)
        journalctl -u $SERVICE_NAME -f
        ;;
    update)
        echo "更新功能开发中..."
        ;;
    uninstall)
        echo "确定要卸载DigWis面板吗？(y/N)"
        read -r response
        if [[ "$response" =~ ^[Yy]$ ]]; then
            systemctl stop $SERVICE_NAME
            systemctl disable $SERVICE_NAME
            rm -rf /opt/digwis
            rm -rf /etc/server-panel
            rm -rf /var/lib/server-panel
            rm -rf /var/log/server-panel
            rm -f /etc/systemd/system/digwis-panel.service
            rm -f /usr/local/bin/digwis
            systemctl daemon-reload
            echo "DigWis面板已卸载"
        fi
        ;;
    *)
        echo "DigWis 服务器管理面板控制脚本"
        echo "用法: $0 {start|stop|restart|status|logs|update|uninstall}"
        echo ""
        echo "命令说明:"
        echo "  start     - 启动面板"
        echo "  stop      - 停止面板"
        echo "  restart   - 重启面板"
        echo "  status    - 查看状态"
        echo "  logs      - 查看日志"
        echo "  update    - 更新面板"
        echo "  uninstall - 卸载面板"
        ;;
esac
EOF
    
    chmod +x /usr/local/bin/digwis
    
    print_success "管理命令创建完成"
}

# 启动服务
start_service() {
    print_step "启动面板服务..."

    systemctl start digwis-panel

    # 等待服务启动
    sleep 3

    if systemctl is-active --quiet digwis-panel; then
        print_success "面板服务启动成功"
    else
        print_error "面板服务启动失败"
        print_info "查看错误日志: journalctl -u digwis-panel -n 20"
        exit 1
    fi
}

# 显示安装结果
show_result() {
    # 获取服务器IP
    SERVER_IP=$(curl -s ifconfig.me 2>/dev/null || curl -s ipinfo.io/ip 2>/dev/null || hostname -I | awk '{print $1}')

    clear
    print_logo

    echo -e "${GREEN}🎉 DigWis面板安装成功！${NC}"
    echo "================================"
    echo ""
    echo -e "${BLUE}📋 安装信息:${NC}"
    echo "   版本: v${PANEL_VERSION}"
    echo "   安装目录: ${INSTALL_DIR}"
    echo "   配置文件: ${CONFIG_DIR}/config.yaml"
    echo "   数据目录: ${DATA_DIR}"
    echo "   日志目录: ${LOG_DIR}"
    echo ""
    echo -e "${BLUE}🌐 访问地址:${NC}"
    echo "   内网访问: http://localhost:8080"
    if [ ! -z "$SERVER_IP" ]; then
        echo "   外网访问: http://${SERVER_IP}:8080"
        echo "   HTTPS访问: https://${SERVER_IP}:8080"
    fi
    echo ""
    echo -e "${BLUE}🔐 登录信息:${NC}"
    echo "   使用系统用户账户登录"
    echo "   支持的管理员组: sudo, wheel, admin 或 root 用户"
    echo ""
    echo -e "${BLUE}🔧 管理命令:${NC}"
    echo "   启动面板: digwis start"
    echo "   停止面板: digwis stop"
    echo "   重启面板: digwis restart"
    echo "   查看状态: digwis status"
    echo "   查看日志: digwis logs"
    echo "   卸载面板: digwis uninstall"
    echo ""
    echo -e "${BLUE}🔒 SSL证书:${NC}"
    echo "   登录面板后，在设置页面可以："
    echo "   • 生成自签名证书"
    echo "   • 申请Let's Encrypt免费证书"
    echo "   • 自动续期证书"
    echo ""
    echo -e "${YELLOW}⚠️  重要提示:${NC}"
    echo "   1. 首次登录后请及时修改默认配置"
    echo "   2. 建议配置SSL证书以提高安全性"
    echo "   3. 定期备份重要数据"
    echo ""
    echo -e "${GREEN}感谢使用 DigWis 服务器管理面板！${NC}"
    echo "文档地址: https://github.com/${GITHUB_REPO}"
}

# 主安装流程
main() {
    print_logo

    # 检查参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --version)
                echo "DigWis Panel v${PANEL_VERSION}"
                exit 0
                ;;
            --help)
                echo "DigWis 服务器管理面板一键安装脚本"
                echo ""
                echo "使用方法:"
                echo "  curl -sSL https://raw.githubusercontent.com/${GITHUB_REPO}/main/install-remote.sh | bash"
                echo "  wget -qO- https://raw.githubusercontent.com/${GITHUB_REPO}/main/install-remote.sh | bash"
                echo ""
                echo "参数:"
                echo "  --version    显示版本信息"
                echo "  --help       显示帮助信息"
                exit 0
                ;;
            *)
                print_error "未知参数: $1"
                exit 1
                ;;
        esac
        shift
    done

    # 执行安装步骤
    check_system
    install_dependencies
    install_go
    download_and_build_panel
    create_config
    create_service
    configure_firewall
    create_management_script
    start_service
    show_result
}

# 错误处理
trap 'print_error "安装过程中发生错误，请检查日志"; exit 1' ERR

# 开始安装
main "$@"
