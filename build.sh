#!/bin/bash

# 服务器面板构建脚本
# 用于编译Go服务器面板为单一二进制文件

set -e

echo "🚀 开始构建服务器管理面板..."

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "❌ Go未安装，正在安装..."
    
    # 检测系统架构
    ARCH=$(uname -m)
    case $ARCH in
        x86_64) GOARCH="amd64" ;;
        aarch64) GOARCH="arm64" ;;
        armv7l) GOARCH="arm" ;;
        *) echo "❌ 不支持的架构: $ARCH"; exit 1 ;;
    esac
    
    # 下载并安装Go
    GO_VERSION="1.21.5"
    echo "📥 下载Go ${GO_VERSION}..."
    wget -q https://golang.org/dl/go${GO_VERSION}.linux-${GOARCH}.tar.gz
    
    echo "📦 安装Go..."
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-${GOARCH}.tar.gz
    
    # 设置环境变量
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    
    rm go${GO_VERSION}.linux-${GOARCH}.tar.gz
    echo "✅ Go安装完成"
fi

# 显示Go版本
echo "📋 Go版本: $(go version)"

# 初始化Go模块（如果需要）
if [ ! -f "go.mod" ]; then
    echo "📦 初始化Go模块..."
    go mod init server-panel
fi

# 下载依赖
echo "📥 下载依赖..."
go mod tidy

# 构建二进制文件
echo "🔨 编译服务器面板..."
CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static" -s -w' -o /opt/digwis/digwis-panel main.go

# 设置权限
chmod +x /opt/digwis/digwis-panel

# 获取文件大小
SIZE=$(du -h /opt/digwis/digwis-panel | cut -f1)

echo ""
echo "✅ 构建完成！"
echo "📁 二进制文件: /opt/digwis/digwis-panel (${SIZE})"
echo ""
echo "🎯 使用方法："
echo "   sudo /opt/digwis/digwis-panel"
echo ""
echo "🌐 然后访问: https://localhost:8080"
echo ""
echo "💡 优势："
echo "   ✓ 单一二进制文件，无需任何依赖"
echo "   ✓ 直接系统用户认证"
echo "   ✓ 内置Web服务器"
echo "   ✓ 高性能，低资源占用"
echo "   ✓ 支持环境一键安装"
echo ""
echo "🔧 安装为系统服务："
echo "   sudo ./install.sh"
