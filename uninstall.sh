#!/bin/bash

# DigWis 面板彻底卸载脚本
# 删除所有相关文件、配置和服务

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_step() {
    echo -e "${YELLOW}[STEP]${NC} $1"
}

# 显示Logo
show_logo() {
    echo -e "${BLUE}"
    echo "=================================="
    echo "    DigWis 面板卸载脚本"
    echo "=================================="
    echo -e "${NC}"
}

# 检查root权限
check_root() {
    if [ "$EUID" -ne 0 ]; then
        echo -e "${RED}[ERROR]${NC} 请使用root权限运行此脚本"
        echo ""
        echo "使用方法："
        echo "  sudo bash uninstall.sh"
        echo ""
        exit 1
    fi
}

# 停止并禁用服务
stop_service() {
    print_step "停止并禁用服务..."
    
    # 停止新命名的服务
    systemctl stop digwis-panel 2>/dev/null || true
    systemctl disable digwis-panel 2>/dev/null || true
    
    # 停止旧命名的服务（兼容性）
    systemctl stop digwis 2>/dev/null || true
    systemctl disable digwis 2>/dev/null || true
    
    print_success "服务已停止"
}

# 删除服务文件
remove_service_files() {
    print_step "删除服务文件..."
    
    # 删除新命名的服务文件
    rm -f /etc/systemd/system/digwis-panel.service
    
    # 删除旧命名的服务文件（兼容性）
    rm -f /etc/systemd/system/digwis.service
    
    # 重新加载systemd配置
    systemctl daemon-reload
    systemctl reset-failed
    
    print_success "服务文件已删除"
}

# 删除程序文件
remove_program_files() {
    print_step "删除程序文件..."
    
    # 删除新命名的安装目录
    rm -rf /opt/digwis-panel
    
    # 删除旧命名的安装目录（兼容性）
    rm -rf /opt/digwis
    
    print_success "程序文件已删除"
}

# 删除配置文件
remove_config_files() {
    print_step "删除配置文件..."
    
    # 删除所有可能的配置目录
    rm -rf /etc/digwis-panel
    rm -rf /etc/digwis
    rm -rf /etc/server-panel
    
    print_success "配置文件已删除"
}

# 删除日志文件
remove_log_files() {
    print_step "删除日志文件..."
    
    # 删除新命名的日志目录
    rm -rf /var/log/digwis-panel
    
    # 删除旧命名的日志目录（兼容性）
    rm -rf /var/log/digwis
    
    print_success "日志文件已删除"
}

# 删除临时文件
remove_temp_files() {
    print_step "删除临时文件..."
    
    # 删除新命名的临时目录
    rm -rf /tmp/digwis-panel-install
    
    # 删除旧命名的临时目录（兼容性）
    rm -rf /tmp/digwis-install
    
    print_success "临时文件已删除"
}

# 清理防火墙规则
cleanup_firewall() {
    print_step "清理防火墙规则..."
    
    # Ubuntu/Debian 使用 ufw
    if command -v ufw >/dev/null 2>&1; then
        ufw delete allow 8080/tcp 2>/dev/null || true
        ufw delete allow 8443/tcp 2>/dev/null || true
    fi
    
    # CentOS/RHEL 使用 firewalld
    if command -v firewall-cmd >/dev/null 2>&1; then
        firewall-cmd --permanent --remove-port=8080/tcp 2>/dev/null || true
        firewall-cmd --permanent --remove-port=8443/tcp 2>/dev/null || true
        firewall-cmd --reload 2>/dev/null || true
    fi
    
    print_success "防火墙规则已清理"
}

# 显示卸载结果
show_result() {
    echo ""
    echo -e "${GREEN}=================================="
    echo "    DigWis 面板卸载完成！"
    echo "==================================${NC}"
    echo ""
    echo "✅ 已删除的内容："
    echo "   - 系统服务 (digwis-panel.service)"
    echo "   - 程序文件 (/opt/digwis-panel)"
    echo "   - 配置文件 (/etc/digwis-panel)"
    echo "   - 日志文件 (/var/log/digwis-panel)"
    echo "   - 临时文件 (/tmp/digwis-panel-install)"
    echo "   - 防火墙规则 (端口 8080/8443)"
    echo ""
    echo "🔄 如需重新安装，请运行："
    echo "   sudo bash install.sh"
    echo ""
}

# 主函数
main() {
    show_logo
    check_root
    stop_service
    remove_service_files
    remove_program_files
    remove_config_files
    remove_log_files
    remove_temp_files
    cleanup_firewall
    show_result
}

# 执行主函数
main "$@"
