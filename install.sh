#!/bin/bash

# DigWis Panel 一键安装脚本
# 从 GitHub 下载二进制文件并自动安装

set -e

# 配置
GITHUB_REPO="${GITHUB_REPO:-digwis/digwis-panel}"
GITHUB_RAW_URL="https://raw.githubusercontent.com/$GITHUB_REPO/main"
INSTALL_DIR="/opt/digwis-panel"
BINARY_NAME="digwis-panel"

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

print_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

print_step() {
    echo -e "${YELLOW}[STEP]${NC} $1"
}

# 显示欢迎信息
echo ""
echo -e "${BLUE}=================================="
echo "   DigWis Panel 一键安装程序"
echo "==================================${NC}"
echo ""

# 检查是否为root用户
if [ "$EUID" -ne 0 ]; then
    print_error "请使用 root 用户或 sudo 运行此脚本"
    exit 1
fi

# 检测操作系统
print_step "检测操作系统..."
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS_ID=$ID
    print_success "检测到操作系统: $NAME"
else
    print_error "无法检测操作系统类型"
    exit 1
fi

# 检查系统依赖
print_step "检查系统依赖..."

# 检查并安装tmux
if ! command -v tmux >/dev/null 2>&1; then
    print_info "tmux未安装，正在安装..."
    case $OS_ID in
        ubuntu|debian)
            apt-get update -qq
            apt-get install -y tmux curl
            ;;
        centos|rhel|fedora)
            yum install -y tmux curl
            ;;
        *)
            print_error "不支持的操作系统，请手动安装tmux"
            exit 1
            ;;
    esac
    print_success "tmux安装成功"
else
    print_success "tmux已安装"
fi

# 检查并安装Node.js和npm（AI功能依赖）
if ! command -v npm >/dev/null 2>&1; then
    print_info "npm未安装，正在安装Node.js和npm（AI功能需要）..."
    case $OS_ID in
        ubuntu|debian)
            curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
            apt-get install -y nodejs curl
            ;;
        centos|rhel)
            curl -fsSL https://rpm.nodesource.com/setup_18.x | bash -
            yum install -y nodejs curl
            ;;
        fedora)
            dnf install -y nodejs npm curl
            ;;
        *)
            print_info "无法自动安装npm，将跳过CLI工具安装"
            ;;
    esac
    
    if command -v npm >/dev/null 2>&1; then
        print_success "npm安装成功"
    else
        print_info "npm安装失败，将跳过CLI工具安装"
    fi
else
    print_success "npm已安装"
fi

# 下载二进制文件
print_step "从 GitHub 下载 DigWis Panel..."
BINARY_URL="$GITHUB_RAW_URL/$BINARY_NAME"
TEMP_FILE="/tmp/$BINARY_NAME"

print_info "下载地址: $BINARY_URL"

if command -v curl >/dev/null 2>&1; then
    if ! curl -fsSL -o "$TEMP_FILE" "$BINARY_URL"; then
        print_error "下载失败，请检查网络连接或仓库地址"
        exit 1
    fi
elif command -v wget >/dev/null 2>&1; then
    if ! wget -q -O "$TEMP_FILE" "$BINARY_URL"; then
        print_error "下载失败，请检查网络连接或仓库地址"
        exit 1
    fi
else
    print_error "未找到 curl 或 wget 工具"
    exit 1
fi

print_success "二进制文件下载完成"

# 停止现有服务
print_step "检查服务状态..."
if systemctl is-active --quiet digwis-panel 2>/dev/null; then
    print_info "停止现有服务..."
    systemctl stop digwis-panel
fi

# 创建安装目录
print_step "创建安装目录..."
mkdir -p "$INSTALL_DIR/data"
mkdir -p "/etc/digwis-panel"
mkdir -p "/var/log/digwis-panel"

# 复制程序文件
print_step "安装程序文件..."
cp "$TEMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

# 设置权限
chown -R root:root "$INSTALL_DIR"
chmod -R 750 "$INSTALL_DIR"

# 创建系统服务
print_step "创建系统服务..."
cat > /etc/systemd/system/digwis-panel.service << 'SERVICE_EOF'
[Unit]
Description=DigWis Server Management Panel
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/digwis-panel
Environment=DIGWIS_MODE=production
Environment=DIGWIS_DATA_DIR=/opt/digwis-panel/data
ExecStart=/opt/digwis-panel/digwis-panel -port 8080
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
SERVICE_EOF

systemctl daemon-reload
systemctl enable digwis-panel

# 检查面板安装状态
print_step "检查面板安装状态..."
if [ -f '/etc/digwis-panel/install.conf' ]; then
    print_info "检测到已安装的面板，执行更新模式..."
    
    # 读取现有配置
    source /etc/digwis-panel/install.conf
    
    # 更新 sudoers 配置
    print_info "更新 sudo 权限配置..."
    cat > /etc/sudoers.d/digwis-panel << SUDOERS_EOF
# digwis-panel sudo permissions
$USERNAME ALL=(ALL) NOPASSWD: /usr/bin/systemctl
$USERNAME ALL=(ALL) NOPASSWD: /usr/sbin/nginx
$USERNAME ALL=(ALL) NOPASSWD: /usr/bin/apt
$USERNAME ALL=(ALL) NOPASSWD: /usr/sbin/service
$USERNAME ALL=(ALL) NOPASSWD: /usr/bin/mysql
$USERNAME ALL=(ALL) NOPASSWD: /usr/bin/mariadb
$USERNAME ALL=(ALL) NOPASSWD: /usr/sbin/phpdismod
$USERNAME ALL=(ALL) NOPASSWD: /usr/sbin/phpenmod
$USERNAME ALL=(ALL) NOPASSWD: /bin/rm
$USERNAME ALL=(ALL) NOPASSWD: /bin/mkdir
$USERNAME ALL=(ALL) NOPASSWD: /bin/sh
$USERNAME ALL=(ALL) NOPASSWD: /usr/bin/cp
$USERNAME ALL=(ALL) NOPASSWD: /bin/chmod
$USERNAME ALL=(ALL) NOPASSWD: /bin/chown
$USERNAME ALL=(ALL) NOPASSWD: /usr/bin/ln
SUDOERS_EOF
    chmod 0440 /etc/sudoers.d/digwis-panel
    
    if visudo -c -f /etc/sudoers.d/digwis-panel >/dev/null 2>&1; then
        print_success "sudo 权限配置已更新"
    fi
    
else
    print_info "首次安装，生成随机凭据..."
    
    cd "$INSTALL_DIR"
    
    # 首次安装：生成新的随机凭据
    timeout 30s ./$BINARY_NAME -install &
    PANEL_PID=$!
    
    # 等待安装完成
    sleep 10
    
    # 停止面板
    kill $PANEL_PID 2>/dev/null || true
    wait $PANEL_PID 2>/dev/null || true
fi

# 检查安装配置
if [ -f '/etc/digwis-panel/install.conf' ]; then
    print_success "面板初始化完成，读取安装信息..."
    
    # 读取安装配置
    source /etc/digwis-panel/install.conf
    
    # 更新systemd服务配置
    sed -i "s|ExecStart=.*|ExecStart=$INSTALL_DIR/digwis-panel -port=$PORT|" /etc/systemd/system/digwis-panel.service
    sed -i "s|User=.*|User=$USERNAME|" /etc/systemd/system/digwis-panel.service
    
    # 添加环境变量
    sed -i "/Environment=DIGWIS_DATA_DIR/a Environment=SERVER_PANEL_SECRET=$SECRET_KEY" /etc/systemd/system/digwis-panel.service
    sed -i "/Environment=SERVER_PANEL_SECRET/a Environment=DIGWIS_PANEL_PORT=$PORT" /etc/systemd/system/digwis-panel.service
    sed -i "/Environment=DIGWIS_PANEL_PORT/a Environment=DIGWIS_PANEL_USER=$USERNAME" /etc/systemd/system/digwis-panel.service
    
    # 确保专用用户对安装目录有访问权限
    chown -R $USERNAME:root $INSTALL_DIR
    chmod -R 755 $INSTALL_DIR
    
    # 确保专用用户对Web目录有访问权限
    mkdir -p /var/www
    chown -R $USERNAME:www-data /var/www
    chmod -R 755 /var/www
    usermod -a -G www-data $USERNAME
    
    systemctl daemon-reload
    
    # 配置防火墙端口
    print_step "配置防火墙端口 $PORT..."
    if command -v firewall-cmd >/dev/null 2>&1; then
        firewall-cmd --permanent --add-port=$PORT/tcp >/dev/null 2>&1 || true
        firewall-cmd --reload >/dev/null 2>&1 || true
        print_success "firewalld 端口配置完成"
    elif command -v ufw >/dev/null 2>&1; then
        ufw allow $PORT/tcp >/dev/null 2>&1 || true
        print_success "ufw 端口配置完成"
    else
        print_info "未检测到防火墙管理工具，请手动开放端口 $PORT"
    fi
    
    # 启动面板服务
    print_step "启动面板服务..."
    systemctl start digwis-panel
    
    # 等待服务启动
    sleep 5
    
    if systemctl is-active --quiet digwis-panel; then
        echo ""
        
        # 获取服务器 IP 地址
        SERVER_IP=$(hostname -I 2>/dev/null | awk '{print $1}')
        if [ -z "$SERVER_IP" ]; then
            SERVER_IP=$(ip route get 1 2>/dev/null | awk '{print $7; exit}')
        fi
        if [ -z "$SERVER_IP" ]; then
            SERVER_IP="YOUR_SERVER_IP"
        fi
        
        # 检查是否为首次安装
        INSTALL_TIME_UNIX=$(date -d "$INSTALL_TIME" +%s 2>/dev/null || echo "0")
        CURRENT_TIME_UNIX=$(date +%s)
        TIME_DIFF=$((CURRENT_TIME_UNIX - INSTALL_TIME_UNIX))
        
        if [ $TIME_DIFF -lt 300 ]; then
            # 首次安装（5分钟内）- 只显示一次
            echo -e "${GREEN}=================================="
            echo "   DigWis Panel 安装完成！"
            echo "==================================${NC}"
            echo ""
            echo "🌐 面板地址: http://${SERVER_IP}:${PORT}"
            echo "👤 用户名: ${USERNAME}"
            echo "🔑 密码: ${PASSWORD}"
            echo "🚪 端口: ${PORT}"
            echo "📁 安装路径: ${INSTALL_DIR}"
            echo "⏰ 安装时间: $(date -d "${INSTALL_TIME}" '+%Y-%m-%d %H:%M:%S' 2>/dev/null || echo "${INSTALL_TIME}")"
            echo ""
            echo "📋 重要提示:"
            echo "  • 请妥善保管上述登录信息"
            echo "  • 配置文件: /etc/digwis-panel/install.conf"
            echo "  • 服务管理: systemctl {start|stop|restart} digwis-panel"
            echo "  • 查看日志: journalctl -u digwis-panel -f"
            echo ""
        else
            # 更新模式
            echo -e "${GREEN}=================================="
            echo "   DigWis Panel 更新完成！"
            echo "==================================${NC}"
            echo ""
            echo "🌐 面板地址: http://${SERVER_IP}:${PORT}"
            echo "📁 安装路径: ${INSTALL_DIR}"
            echo "⏰ 更新时间: $(date '+%Y-%m-%d %H:%M:%S')"
            echo ""
            echo "📋 提示:"
            echo "  • 登录信息保持不变"
            echo "  • 查看配置: cat /etc/digwis-panel/install.conf"
            echo "  • 服务状态: systemctl status digwis-panel"
            echo ""
        fi
    else
        print_error "服务启动失败"
        echo "查看日志: journalctl -u digwis-panel -f"
        exit 1
    fi
else
    print_error "面板初始化失败，未找到安装配置文件"
    exit 1
fi

# 清理临时文件
rm -f "$TEMP_FILE"

print_success "安装完成"
