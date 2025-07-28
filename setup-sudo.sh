#!/bin/bash

# DigWis Panel - Setup passwordless sudo for development
echo "Setting up passwordless sudo for DigWis Panel development..."

# 检查是否为 root 用户
if [ "$EUID" -eq 0 ]; then
    echo "Please run this script as a regular user, not root."
    exit 1
fi

# 获取当前用户名
USERNAME=$(whoami)

# 创建 sudoers 配置文件
SUDOERS_FILE="/etc/sudoers.d/digwis-panel"

echo "Creating sudoers configuration for user: $USERNAME"

# 需要 sudo 权限来创建文件
sudo tee "$SUDOERS_FILE" > /dev/null << EOF
# DigWis Panel - Allow passwordless sudo for package management
$USERNAME ALL=(ALL) NOPASSWD: /usr/bin/apt, /usr/bin/systemctl, /usr/bin/service
EOF

# 设置正确的权限
sudo chmod 440 "$SUDOERS_FILE"

# 验证配置
if sudo visudo -c; then
    echo "✅ Sudoers configuration is valid"
    echo "✅ Passwordless sudo configured for:"
    echo "   - apt (package management)"
    echo "   - systemctl (service management)"
    echo "   - service (service management)"
    echo ""
    echo "You can now use DigWis Panel to install packages without entering passwords."
else
    echo "❌ Sudoers configuration is invalid, removing..."
    sudo rm -f "$SUDOERS_FILE"
    exit 1
fi
