#!/bin/bash

# DigWis 面板一键升级脚本
# 用于升级已安装的 DigWis Panel 到最新版本
# 使用方法: curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/upgrade.sh | sudo bash

set -e

# 默认配置
VERBOSE=false
QUIET=false
GITHUB_REPO="digwis/digwis-panel"
INSTALL_DIR="/opt/digwis-panel"
TEMP_DIR="/tmp/digwis-panel-upgrade"
SERVICE_NAME="digwis-panel"

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
        --version)
            VERSION="$2"
            shift 2
            ;;
        --help|-h)
            echo "DigWis 面板升级脚本"
            echo ""
            echo "使用方法:"
            echo "  本地: sudo bash upgrade.sh [选项]"
            echo "  远程: curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/upgrade.sh | sudo bash"
            echo ""
            echo "选项:"
            echo "  --verbose, -v    显示详细升级信息"
            echo "  --quiet, -q      静默升级模式"
            echo "  --version VER    指定升级版本 (默认: latest)"
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

# 显示Logo
show_logo() {
    if [ "$QUIET" != "true" ]; then
        echo -e "${BLUE}"
        echo "=================================="
        echo "    DigWis 面板升级脚本"
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
        echo "  sudo bash upgrade.sh"
        echo ""
        exit 1
    fi
}

# 检查现有安装
check_existing_installation() {
    print_step "检查现有安装..."

    # 检查安装目录
    if [ ! -d "$INSTALL_DIR" ]; then
        print_error "未找到现有安装，请先运行安装脚本"
        print_info "运行: sudo bash install.sh"
        exit 1
    fi

    # 检查程序文件
    if [ ! -f "$INSTALL_DIR/digwis-panel" ]; then
        print_error "未找到程序文件，安装可能不完整"
        exit 1
    fi

    # 检查系统服务
    if ! systemctl list-unit-files | grep -q "$SERVICE_NAME.service"; then
        print_error "未找到系统服务，安装可能不完整"
        exit 1
    fi

    print_success "现有安装检查通过"
}

# 获取当前版本
get_current_version() {
    print_verbose "获取当前版本信息..."
    
    # 尝试从程序获取版本信息
    CURRENT_VERSION="unknown"
    if [ -f "$INSTALL_DIR/digwis-panel" ]; then
        # 这里可以添加版本检测逻辑
        CURRENT_VERSION=$(stat -c %Y "$INSTALL_DIR/digwis-panel" 2>/dev/null || echo "unknown")
    fi
    
    print_verbose "当前版本: $CURRENT_VERSION"
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

# 获取最新版本号
get_latest_version() {
    print_verbose "获取最新版本号..."

    # 尝试从GitHub API获取
    VERSION=$(curl -s --connect-timeout 5 --max-time 10 "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep -o '"tag_name":"[^"]*"' | cut -d'"' -f4)

    # 如果失败，使用默认版本
    if [ -z "$VERSION" ]; then
        print_warning "无法获取最新版本，使用默认版本"
        VERSION="v0.2.0"
    fi

    print_verbose "目标版本: $VERSION"
}

# 停止服务
stop_service() {
    print_step "停止服务..."

    if systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
        print_verbose "正在停止服务..."
        systemctl stop "$SERVICE_NAME"
        print_success "服务已停止"
    else
        print_info "服务未运行"
    fi
}

# 备份当前版本
backup_current_version() {
    print_step "备份当前版本..."

    local backup_file="$INSTALL_DIR/digwis-panel.backup.$(date +%s)"
    print_verbose "备份到: $backup_file"
    
    cp "$INSTALL_DIR/digwis-panel" "$backup_file"
    print_success "当前版本已备份"
}

# 下载新版本
download_new_version() {
    print_step "下载新版本..."

    # 创建临时目录
    mkdir -p "$TEMP_DIR"
    cd "$TEMP_DIR"

    # 构建下载文件名
    local package_name="digwis-panel-${VERSION}-linux-${ARCH}.tar.gz"
    local download_url="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${package_name}"

    print_verbose "下载地址: $download_url"

    # 下载文件
    if curl --connect-timeout 30 --max-time 600 --retry 3 --retry-delay 5 -sL "$download_url" -o "$package_name"; then
        if [ -f "$package_name" ] && [ -s "$package_name" ]; then
            print_success "新版本下载成功"
        else
            print_error "下载文件无效"
            exit 1
        fi
    else
        print_error "下载失败，请检查网络连接"
        exit 1
    fi

    # 解压文件
    print_verbose "解压安装包..."
    tar -xzf "$package_name" >/dev/null 2>&1

    # 验证解压结果
    local binary_file=$(find . -name "digwis-panel-*" -type f ! -name "*.tar.gz" | head -1)
    if [ -z "$binary_file" ]; then
        print_error "安装包解压失败或文件损坏"
        exit 1
    fi

    # 重命名为统一的文件名
    mv "$binary_file" "digwis-panel"

    print_success "新版本准备完成"
}

# 安装新版本
install_new_version() {
    print_step "安装新版本..."

    # 复制新版本
    print_verbose "复制新版本文件..."
    cp "$TEMP_DIR/digwis-panel" "$INSTALL_DIR/digwis-panel"
    chmod +x "$INSTALL_DIR/digwis-panel"

    print_success "新版本安装完成"
}

# 启动服务
start_service() {
    print_step "启动服务..."

    print_verbose "启动服务..."
    systemctl start "$SERVICE_NAME"

    # 等待服务启动
    sleep 3

    if systemctl is-active --quiet "$SERVICE_NAME"; then
        print_success "服务启动成功"
    else
        print_error "服务启动失败"
        print_info "查看日志: journalctl -u $SERVICE_NAME -f"
        exit 1
    fi
}

# 清理临时文件
cleanup() {
    print_verbose "清理临时文件..."
    rm -rf "$TEMP_DIR"
    print_verbose "清理完成"
}

# 显示升级结果
show_result() {
    if [ "$QUIET" != "true" ]; then
        echo ""
        echo -e "${GREEN}=================================="
        echo "    DigWis 面板升级完成！"
        echo "==================================${NC}"
        echo ""
        echo "🌐 访问地址:"
        echo "   本地: http://localhost:8080"
        echo "   外网: http://$(curl -s ifconfig.me 2>/dev/null || echo "YOUR_SERVER_IP"):8080"
        echo ""
        echo "🔧 管理命令:"
        echo "   查看状态: systemctl status $SERVICE_NAME"
        echo "   查看日志: journalctl -u $SERVICE_NAME -f"
        echo "   重启服务: systemctl restart $SERVICE_NAME"
        echo ""
        echo "📁 安装目录: $INSTALL_DIR"
        echo ""
        echo -e "${YELLOW}升级已完成，请测试功能是否正常${NC}"
        echo ""
    fi
}

# 主函数
main() {
    show_logo
    check_root
    check_existing_installation
    get_current_version
    detect_arch
    get_latest_version
    stop_service
    backup_current_version
    download_new_version
    install_new_version
    start_service
    cleanup
    show_result
}

# 执行主函数
main "$@"
