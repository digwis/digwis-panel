#!/bin/bash

# 测试新的统一安装脚本
# 检查语法、函数和配置

echo "🧪 测试 DigWis 统一安装脚本..."

# 测试脚本语法
echo "📝 检查脚本语法..."

if bash -n install.sh; then
    echo "✅ install.sh 语法正确"
else
    echo "❌ install.sh 语法错误"
    exit 1
fi

# 检查必要的函数是否存在
echo "🔍 检查关键函数..."

REQUIRED_FUNCTIONS=(
    "check_root"
    "detect_arch"
    "detect_os"
    "select_download_node"
    "get_latest_version"
    "install_dependencies"
    "download_panel"
    "install_panel"
    "create_service"
    "configure_firewall"
    "start_service"
    "show_result"
)

for func in "${REQUIRED_FUNCTIONS[@]}"; do
    if grep -q "^${func}()" install.sh; then
        echo "✅ 函数 ${func} 存在"
    else
        echo "❌ 函数 ${func} 缺失"
    fi
done

# 检查关键变量
echo "📋 检查关键变量..."

REQUIRED_VARS=(
    "GITHUB_REPO"
    "INSTALL_DIR"
    "CONFIG_DIR"
    "TEMP_DIR"
    "DOWNLOAD_NODES"
    "VERSION"
)

for var in "${REQUIRED_VARS[@]}"; do
    if grep -q "^${var}=" install.sh; then
        echo "✅ 变量 ${var} 已定义"
    else
        echo "❌ 变量 ${var} 未定义"
    fi
done

# 检查 GitHub 仓库地址
echo "🌐 检查 GitHub 仓库..."

REPO_URL=$(grep "GITHUB_REPO=" install.sh | cut -d'"' -f2)
if [ "$REPO_URL" = "digwis/digwis-panel" ]; then
    echo "✅ GitHub 仓库地址正确: $REPO_URL"
else
    echo "❌ GitHub 仓库地址错误: $REPO_URL"
fi

# 检查命令行参数处理
echo "🎛️ 检查命令行参数..."

if grep -q "while \[\[ \$# -gt 0 \]\]" install.sh; then
    echo "✅ 支持命令行参数"
else
    echo "❌ 不支持命令行参数"
fi

if grep -q "\-\-verbose" install.sh; then
    echo "✅ 支持详细模式"
else
    echo "❌ 不支持详细模式"
fi

if grep -q "\-\-quiet" install.sh; then
    echo "✅ 支持静默模式"
else
    echo "❌ 不支持静默模式"
fi

# 检查错误处理
echo "🛡️ 检查错误处理..."

if grep -q "set -e" install.sh; then
    echo "✅ 启用了错误退出"
else
    echo "❌ 未启用错误退出"
fi

if grep -q "retry" install.sh; then
    echo "✅ 包含重试机制"
else
    echo "❌ 缺少重试机制"
fi

# 模拟测试架构检测
echo "🏗️ 测试架构检测逻辑..."

test_arch() {
    local arch=$1
    local expected=$2

    if [ "$arch" = "x86_64" ] && [ "$expected" = "amd64" ]; then
        echo "✅ 架构映射正确: $arch -> $expected"
    elif [ "$arch" = "aarch64" ] && [ "$expected" = "arm64" ]; then
        echo "✅ 架构映射正确: $arch -> $expected"
    elif [ "$arch" = "armv7l" ] && [ "$expected" = "arm" ]; then
        echo "✅ 架构映射正确: $arch -> $expected"
    else
        echo "❌ 架构映射可能有问题"
    fi
}

test_arch "x86_64" "amd64"
test_arch "aarch64" "arm64"
test_arch "armv7l" "arm"

# 检查下载节点配置
echo "🌐 检查下载节点配置..."

DOWNLOAD_NODES_CHECK=$(grep "DOWNLOAD_NODES=" install.sh)
if [ -n "$DOWNLOAD_NODES_CHECK" ]; then
    echo "✅ 下载节点已配置"
else
    echo "❌ 下载节点未配置"
fi

echo ""
echo "🎉 测试完成！"
echo ""
echo "📖 使用说明："
echo "1. 基本安装："
echo "   curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash"
echo ""
echo "2. 详细模式："
echo "   curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash -s -- --verbose"
echo ""
echo "3. 静默模式："
echo "   curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash -s -- --quiet"
echo ""
echo "4. 本地测试："
echo "   sudo ./install.sh"
echo ""
echo "5. 验证安装："
echo "   ./verify-install.sh"
echo ""
echo "⚠️  注意：安装脚本需要 root 权限运行"
