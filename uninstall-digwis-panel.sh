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
