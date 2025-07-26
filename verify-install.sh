#!/bin/bash

# DigWis 面板安装验证脚本
# 用于检查安装是否成功

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 打印函数
print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }

echo -e "${BLUE}DigWis 面板安装验证${NC}"
echo "=========================="

# 检查二进制文件
print_info "检查二进制文件..."
if [ -f "/opt/digwis/digwis" ]; then
    print_success "二进制文件存在: /opt/digwis/digwis"

    # 检查文件权限
    if [ -x "/opt/digwis/digwis" ]; then
        print_success "二进制文件可执行"
    else
        print_error "二进制文件不可执行"
    fi

    # 获取文件大小
    SIZE=$(du -h /opt/digwis/digwis | cut -f1)
    print_info "文件大小: ${SIZE}"
else
    print_error "二进制文件不存在"
fi

# 检查配置文件
print_info "检查配置文件..."
if [ -f "/etc/digwis/config.yaml" ]; then
    print_success "配置文件存在: /etc/digwis/config.yaml"
    
    # 检查配置文件内容
    if grep -q "server:" /etc/digwis/config.yaml; then
        print_success "配置文件格式正确"
    else
        print_warning "配置文件格式可能有问题"
    fi
else
    print_error "配置文件不存在"
fi

# 检查目录结构
print_info "检查目录结构..."
DIRS=("/var/lib/digwis" "/var/log/digwis")

for dir in "${DIRS[@]}"; do
    if [ -d "$dir" ]; then
        print_success "目录存在: $dir"
    else
        print_error "目录不存在: $dir"
    fi
done

# 检查系统服务
print_info "检查系统服务..."
if [ -f "/etc/systemd/system/digwis.service" ]; then
    print_success "服务文件存在: /etc/systemd/system/digwis.service"
    
    # 检查服务状态
    if systemctl is-enabled digwis >/dev/null 2>&1; then
        print_success "服务已启用"
    else
        print_warning "服务未启用"
    fi
    
    if systemctl is-active digwis >/dev/null 2>&1; then
        print_success "服务正在运行"
    else
        print_warning "服务未运行"
    fi
else
    print_error "服务文件不存在"
fi

# 检查端口监听
print_info "检查端口监听..."
if netstat -tlnp 2>/dev/null | grep -q ":8080"; then
    print_success "端口8080正在监听"
    
    # 显示监听详情
    LISTEN_INFO=$(netstat -tlnp 2>/dev/null | grep ":8080" | head -1)
    print_info "监听详情: $LISTEN_INFO"
elif ss -tlnp 2>/dev/null | grep -q ":8080"; then
    print_success "端口8080正在监听"
    
    # 显示监听详情
    LISTEN_INFO=$(ss -tlnp 2>/dev/null | grep ":8080" | head -1)
    print_info "监听详情: $LISTEN_INFO"
else
    print_warning "端口8080未监听"
fi

# 检查HTTP响应
print_info "检查HTTP响应..."
if command -v curl >/dev/null 2>&1; then
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080 2>/dev/null || echo "000")
    
    if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "302" ] || [ "$HTTP_CODE" = "401" ]; then
        print_success "HTTP服务响应正常 (状态码: $HTTP_CODE)"
    else
        print_warning "HTTP服务响应异常 (状态码: $HTTP_CODE)"
    fi
else
    print_warning "curl未安装，无法测试HTTP响应"
fi

# 检查日志
print_info "检查最近日志..."
if command -v journalctl >/dev/null 2>&1; then
    RECENT_LOGS=$(journalctl -u digwis --no-pager -n 3 2>/dev/null | tail -3)
    if [ -n "$RECENT_LOGS" ]; then
        print_info "最近日志:"
        echo "$RECENT_LOGS"
    else
        print_warning "无法获取日志"
    fi
else
    print_warning "journalctl未可用"
fi

# 获取服务器IP
print_info "获取访问地址..."
SERVER_IP=$(curl -s ifconfig.me 2>/dev/null || curl -s ipinfo.io/ip 2>/dev/null || echo "YOUR_SERVER_IP")

echo ""
echo -e "${GREEN}验证完成！${NC}"
echo "=========================="
echo ""
echo -e "${BLUE}访问地址:${NC}"
echo "  本地: http://localhost:8080"
echo "  外网: http://${SERVER_IP}:8080"
echo ""
echo -e "${BLUE}管理命令:${NC}"
echo "  查看状态: systemctl status digwis"
echo "  启动服务: systemctl start digwis"
echo "  停止服务: systemctl stop digwis"
echo "  重启服务: systemctl restart digwis"
echo "  查看日志: journalctl -u digwis -f"
echo ""

# 检查防火墙提醒
if command -v ufw >/dev/null 2>&1; then
    if ufw status | grep -q "Status: active"; then
        if ! ufw status | grep -q "8080"; then
            print_warning "UFW防火墙已启用，请确保开放8080端口: sudo ufw allow 8080/tcp"
        fi
    fi
fi

if command -v firewall-cmd >/dev/null 2>&1; then
    if systemctl is-active --quiet firewalld; then
        if ! firewall-cmd --list-ports | grep -q "8080/tcp"; then
            print_warning "firewalld已启用，请确保开放8080端口:"
            echo "  sudo firewall-cmd --permanent --add-port=8080/tcp"
            echo "  sudo firewall-cmd --reload"
        fi
    fi
fi

echo -e "${YELLOW}如果无法访问面板，请检查防火墙设置${NC}"
