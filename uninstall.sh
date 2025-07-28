#!/bin/bash

# DigWis 面板卸载脚本
# 彻底清除所有安装的组件和配置
# 使用方法: curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/uninstall.sh | sudo bash

set -e

# 默认配置
INSTALL_DIR="/opt/digwis-panel"
CONFIG_DIR="/etc/digwis-panel"
LOG_DIR="/var/log/digwis-panel"
SERVICE_NAME="digwis-panel"
AUTO_CONFIRM=false

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 打印函数
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_step() {
    echo -e "${YELLOW}[STEP]${NC} $1"
}

# 解析命令行参数
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --yes|-y)
                AUTO_CONFIRM=true
                shift
                ;;
            --help|-h)
                echo "DigWis 面板卸载脚本"
                echo ""
                echo "使用方法:"
                echo "  本地: sudo ./uninstall.sh [选项]"
                echo "  远程: curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/uninstall.sh | sudo bash"
                echo ""
                echo "选项:"
                echo "  --yes, -y        自动确认卸载，不显示交互提示"
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
}

# 检查是否以root权限运行
check_root() {
    if [ "$EUID" -ne 0 ]; then
        print_error "请使用 sudo 或以 root 用户身份运行此脚本"
        exit 1
    fi
}

# 确认卸载
confirm_uninstall() {
    echo ""
    print_warning "这将完全卸载 DigWis 面板及其所有数据！"
    print_warning "包括："
    echo "  - 停止并删除系统服务"
    echo "  - 删除安装目录: $INSTALL_DIR"
    echo "  - 删除配置目录: $CONFIG_DIR"
    echo "  - 删除日志目录: $LOG_DIR"
    echo "  - 终止所有相关进程"
    echo ""

    if [ "$AUTO_CONFIRM" = "true" ]; then
        print_info "自动确认模式，继续卸载..."
        return
    fi

    read -p "确认要继续卸载吗？(y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "卸载已取消"
        exit 0
    fi
}

# 停止服务
stop_service() {
    print_step "停止 $SERVICE_NAME 服务..."

    if systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
        print_info "正在停止服务..."
        systemctl stop "$SERVICE_NAME"
        print_success "服务已停止"
    else
        print_info "服务未运行"
    fi
}

# 禁用服务
disable_service() {
    print_step "禁用 $SERVICE_NAME 服务..."

    if systemctl is-enabled --quiet "$SERVICE_NAME" 2>/dev/null; then
        print_info "正在禁用服务..."
        systemctl disable "$SERVICE_NAME"
        print_success "服务已禁用"
    else
        print_info "服务未启用"
    fi
}

# 删除服务文件
remove_service_file() {
    print_step "删除服务文件..."

    local service_file="/etc/systemd/system/${SERVICE_NAME}.service"
    if [ -f "$service_file" ]; then
        print_info "删除服务文件: $service_file"
        rm -f "$service_file"
        systemctl daemon-reload
        print_success "服务文件已删除"
    else
        print_info "服务文件不存在"
    fi
}

# 终止残留进程
kill_processes() {
    print_step "终止残留进程..."

    if pgrep -f "$SERVICE_NAME" > /dev/null; then
        print_info "发现残留进程，正在终止..."
        pkill -f "$SERVICE_NAME" || true
        sleep 3

        # 强制终止仍然存在的进程
        if pgrep -f "$SERVICE_NAME" > /dev/null; then
            print_warning "强制终止残留进程..."
            pkill -9 -f "$SERVICE_NAME" || true
        fi
        print_success "进程已终止"
    else
        print_info "没有发现残留进程"
    fi
}

# 备份配置文件
backup_config() {
    if [ -d "$CONFIG_DIR" ]; then
        local backup_dir="/tmp/digwis-panel-config-backup-$(date +%s)"
        print_step "备份配置文件..."
        print_info "备份配置文件到: $backup_dir"
        cp -r "$CONFIG_DIR" "$backup_dir"
        print_success "配置文件已备份到: $backup_dir"
    fi
}

# 删除目录
remove_directories() {
    print_step "删除程序目录..."

    # 删除安装目录
    if [ -d "$INSTALL_DIR" ]; then
        print_info "删除安装目录: $INSTALL_DIR"
        rm -rf "$INSTALL_DIR"
        print_success "安装目录已删除"
    else
        print_info "安装目录不存在"
    fi

    # 删除配置目录
    if [ -d "$CONFIG_DIR" ]; then
        print_info "删除配置目录: $CONFIG_DIR"
        rm -rf "$CONFIG_DIR"
        print_success "配置目录已删除"
    else
        print_info "配置目录不存在"
    fi

    # 删除日志目录
    if [ -d "$LOG_DIR" ]; then
        print_info "删除日志目录: $LOG_DIR"
        rm -rf "$LOG_DIR"
        print_success "日志目录已删除"
    else
        print_info "日志目录不存在"
    fi
}

# 清理临时文件
cleanup_temp() {
    print_step "清理临时文件..."

    # 清理可能的临时安装目录
    rm -rf "/tmp/digwis-panel-install" 2>/dev/null || true
    rm -rf "/tmp/digwis-panel-*" 2>/dev/null || true

    print_success "临时文件已清理"
}

# 删除用户（如果存在）
remove_user() {
    local user_name="digwis"

    if id "$user_name" &>/dev/null; then
        print_step "删除系统用户: $user_name"
        userdel -r "$user_name" 2>/dev/null || true
        print_success "用户已删除"
    fi
}

# 清理防火墙规则（可选）
cleanup_firewall() {
    print_step "清理防火墙规则..."

    # UFW
    if command -v ufw &> /dev/null; then
        ufw --force delete allow 8080 2>/dev/null || true
        ufw --force delete allow 8443 2>/dev/null || true
    fi

    # Firewalld
    if command -v firewall-cmd &> /dev/null; then
        firewall-cmd --permanent --remove-port=8080/tcp 2>/dev/null || true
        firewall-cmd --permanent --remove-port=8443/tcp 2>/dev/null || true
        firewall-cmd --reload 2>/dev/null || true
    fi

    print_info "防火墙规则已清理"
}

# 显示卸载结果
show_result() {
    echo ""
    echo "=================================================="
    print_success "DigWis 面板卸载完成！"
    echo "=================================================="
    echo ""
    print_info "已完成以下操作："
    echo "  ✓ 停止并删除系统服务"
    echo "  ✓ 删除所有程序文件"
    echo "  ✓ 删除配置文件（已备份）"
    echo "  ✓ 删除日志文件"
    echo "  ✓ 终止所有相关进程"
    echo "  ✓ 清理临时文件"
    echo "  ✓ 清理防火墙规则"
    echo ""
    print_info "如需重新安装，请运行安装脚本"
    echo ""
}

# 主函数
main() {
    echo "=================================================="
    echo "         DigWis 面板卸载脚本 v1.0"
    echo "=================================================="
    echo ""

    parse_args "$@"
    check_root
    confirm_uninstall

    print_info "开始卸载 DigWis 面板..."
    echo ""

    stop_service
    disable_service
    remove_service_file
    kill_processes
    backup_config
    remove_directories
    cleanup_temp
    remove_user
    cleanup_firewall

    show_result
}

# 执行主函数
main "$@"
