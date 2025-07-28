#!/bin/bash

# DigWis 面板一键安装脚本
# 支持 Ubuntu/Debian/CentOS/RHEL/Fedora 系统
# 使用方法: curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash

set -e

# 默认配置
VERBOSE=false
QUIET=false
GITHUB_REPO="digwis/digwis-panel"
INSTALL_DIR="/opt/digwis-panel"
CONFIG_DIR="/etc/digwis-panel"
TEMP_DIR="/tmp/digwis-panel-install"

# 下载节点配置
DOWNLOAD_NODES=(
    "https://raw.githubusercontent.com/digwis/digwis-panel/main/releases"
    "https://github.com/digwis/digwis-panel/releases/download"
)

# 版本配置
VERSION="latest"
ARCH=""

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
    local machine_arch=$(uname -m)
    case $machine_arch in
        x86_64) ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
        armv7l) ARCH="arm" ;;
        *)
            print_error "不支持的系统架构: $machine_arch"
            exit 1
            ;;
    esac
    print_verbose "检测到系统架构: $machine_arch -> $ARCH"
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

# 选择下载节点
select_download_node() {
    print_step "选择下载节点..."

    # 优先使用GitHub仓库中的releases目录
    DOWNLOAD_URL="https://raw.githubusercontent.com/${GITHUB_REPO}/main/releases"

    print_verbose "测试主要下载节点..."
    local response_code=$(curl -o /dev/null -s -w "%{http_code}" --connect-timeout 5 --max-time 10 "${DOWNLOAD_URL}/version.json" 2>/dev/null || echo "000")

    if [ "$response_code" = "200" ]; then
        print_success "使用GitHub仓库下载节点: $DOWNLOAD_URL"
    else
        print_warning "GitHub仓库不可用，使用GitHub Releases作为备用"
        DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download"
    fi
}

# 获取最新版本号
get_latest_version() {
    print_verbose "获取最新版本号..."

    # 尝试从仓库的version.json获取版本号
    local version_file="${DOWNLOAD_URL}/version.json"
    VERSION=$(curl -s --connect-timeout 5 --max-time 10 "$version_file" 2>/dev/null | grep -o '"version":"[^"]*"' | cut -d'"' -f4)

    # 如果失败，尝试从GitHub API获取
    if [ -z "$VERSION" ]; then
        print_verbose "version.json获取失败，尝试GitHub API..."
        VERSION=$(curl -s --connect-timeout 5 --max-time 10 "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep -o '"tag_name":"[^"]*"' | cut -d'"' -f4)
    fi

    # 如果仍然失败，使用默认版本
    if [ -z "$VERSION" ]; then
        print_warning "无法获取最新版本，使用默认版本"
        VERSION="v1.0.0"
    fi

    print_verbose "目标版本: $VERSION"
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

# 下载预编译的面板程序
download_panel() {
    print_step "下载面板程序..."

    # 创建临时目录
    mkdir -p "$TEMP_DIR"
    cd "$TEMP_DIR"

    # 构建下载文件名
    local package_name="digwis-panel-${VERSION}-linux-${ARCH}.tar.gz"
    local download_url="${DOWNLOAD_URL}/${package_name}"

    print_verbose "下载地址: $download_url"

    # 检查是否已经下载过
    if [ -f "$package_name" ]; then
        print_verbose "发现已下载的安装包，验证完整性..."
        # 这里可以添加校验逻辑
    else
        print_info "正在下载面板安装包..."

        # 尝试多种下载方式
        local download_success=false

        # 方式1: 使用wget
        if command -v wget >/dev/null 2>&1 && [ "$download_success" = "false" ]; then
            if [ "$VERBOSE" = "true" ]; then
                if wget --timeout=30 --tries=3 --progress=bar "$download_url" -O "$package_name"; then
                    download_success=true
                fi
            else
                if wget --timeout=30 --tries=3 -q "$download_url" -O "$package_name"; then
                    download_success=true
                fi
            fi
        fi

        # 方式2: 使用curl
        if [ "$download_success" = "false" ]; then
            if [ "$VERBOSE" = "true" ]; then
                if curl --connect-timeout 30 --max-time 300 --retry 3 --retry-delay 2 -L "$download_url" -o "$package_name"; then
                    download_success=true
                fi
            else
                if curl --connect-timeout 30 --max-time 300 --retry 3 --retry-delay 2 -sL "$download_url" -o "$package_name"; then
                    download_success=true
                fi
            fi
        fi

        # 验证下载是否成功
        if [ "$download_success" = "false" ] || [ ! -f "$package_name" ] || [ ! -s "$package_name" ]; then
            print_warning "CDN下载失败，尝试从GitHub下载..."

            # 备用方案：从GitHub Releases下载
            local github_url="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${package_name}"
            print_verbose "GitHub下载地址: $github_url"

            if curl --connect-timeout 30 --max-time 600 --retry 3 --retry-delay 5 -sL "$github_url" -o "$package_name"; then
                if [ -f "$package_name" ] && [ -s "$package_name" ]; then
                    print_success "GitHub下载成功"
                else
                    print_error "GitHub下载失败"
                    exit 1
                fi
            else
                print_error "所有下载源均失败，请检查网络连接或稍后重试"
                exit 1
            fi
        fi

        print_success "面板安装包下载完成"
    fi

    # 解压安装包
    print_verbose "解压安装包..."
    if [ "$VERBOSE" = "true" ]; then
        tar -xzf "$package_name"
    else
        tar -xzf "$package_name" >/dev/null 2>&1
    fi

    # 验证解压结果
    if [ ! -f "digwis" ]; then
        print_error "安装包解压失败或文件损坏"
        exit 1
    fi

    print_success "面板程序准备完成"
}

# 安装面板
install_panel() {
    print_step "安装面板..."

    # 创建目录
    print_verbose "创建安装目录..."
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$INSTALL_DIR/data"        # 数据目录
    mkdir -p "$CONFIG_DIR"
    mkdir -p "/var/log/digwis-panel"

    # 复制文件
    print_verbose "复制程序文件..."
    cp "$TEMP_DIR/digwis" "$INSTALL_DIR/digwis-panel"
    chmod +x "$INSTALL_DIR/digwis-panel"

    # 删除原始文件名，只保留统一命名的版本
    rm -f "$INSTALL_DIR/digwis"

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
  file: "/var/log/digwis-panel/digwis-panel.log"
EOF

    # 安装管理工具
    print_verbose "安装管理工具..."
    if curl -sSL "https://raw.githubusercontent.com/${GITHUB_REPO}/main/scripts/management/digwis" -o /usr/local/bin/digwis; then
        chmod +x /usr/local/bin/digwis
        print_verbose "管理工具安装成功"
    else
        print_warning "管理工具下载失败，将使用备用方案"
        # 创建简化版本的管理工具
        cat > /usr/local/bin/digwis << 'EOF'
#!/bin/bash
echo "DigWis 面板管理工具"
echo ""
echo "可用命令:"
echo "  systemctl start digwis-panel    # 启动服务"
echo "  systemctl stop digwis-panel     # 停止服务"
echo "  systemctl restart digwis-panel  # 重启服务"
echo "  systemctl status digwis-panel   # 查看状态"
echo "  journalctl -u digwis-panel -f   # 查看日志"
echo ""
echo "访问地址: http://localhost:8080"
EOF
        chmod +x /usr/local/bin/digwis
    fi

    print_success "面板安装完成"
}

# 创建系统服务
create_service() {
    print_step "配置系统服务..."

    print_verbose "创建systemd服务文件..."
    cat > /etc/systemd/system/digwis-panel.service << EOF
[Unit]
Description=DigWis Server Management Panel
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
Environment=DIGWIS_MODE=production
Environment=DIGWIS_DATA_DIR=$INSTALL_DIR/data
ExecStart=$INSTALL_DIR/digwis-panel -config $CONFIG_DIR/config.yaml -port 8080
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

    print_verbose "重新加载systemd配置..."
    systemctl daemon-reload >/dev/null 2>&1
    systemctl enable digwis-panel >/dev/null 2>&1

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

    print_verbose "启动digwis-panel服务..."
    systemctl start digwis-panel

    # 等待服务启动
    sleep 3

    if systemctl is-active --quiet digwis-panel; then
        print_success "面板服务启动成功"
    else
        print_error "面板服务启动失败"
        print_info "查看日志: journalctl -u digwis-panel -f"
        exit 1
    fi
}

# 清理临时文件
cleanup() {
    print_verbose "清理临时文件..."
    rm -rf "$TEMP_DIR"
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
        echo "   面板管理: digwis                        # 打开管理菜单"
        echo "   启动服务: systemctl start digwis-panel"
        echo "   停止服务: systemctl stop digwis-panel"
        echo "   重启服务: systemctl restart digwis-panel"
        echo "   查看状态: systemctl status digwis-panel"
        echo "   查看日志: journalctl -u digwis-panel -f"
        echo ""
        echo "📁 安装目录: $INSTALL_DIR"
        echo "📁 配置目录: $CONFIG_DIR"
        echo "📁 数据目录: $INSTALL_DIR/data"
        echo "📁 日志目录: /var/log/digwis-panel"
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
    select_download_node
    get_latest_version
    install_dependencies
    download_panel
    install_panel
    create_service
    configure_firewall
    start_service
    cleanup
    show_result
}

# 执行主函数
main "$@"
