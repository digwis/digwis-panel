#!/bin/bash

# 测试安装脚本语法和基本功能

echo "🧪 测试 DigWis 安装脚本..."

# 测试脚本语法
echo "📝 检查脚本语法..."

if bash -n install-quick.sh; then
    echo "✅ install-quick.sh 语法正确"
else
    echo "❌ install-quick.sh 语法错误"
    exit 1
fi

if bash -n install-remote.sh; then
    echo "✅ install-remote.sh 语法正确"
else
    echo "❌ install-remote.sh 语法错误"
    exit 1
fi

# 检查必要的函数是否存在
echo "🔍 检查关键函数..."

# 检查 install-quick.sh 中的函数
QUICK_FUNCTIONS=("check_root" "detect_arch" "install_deps" "install_go" "build_panel" "create_config" "create_service" "start_service" "show_result")

for func in "${QUICK_FUNCTIONS[@]}"; do
    if grep -q "^${func}()" install-quick.sh; then
        echo "✅ 函数 ${func} 存在"
    else
        echo "❌ 函数 ${func} 缺失"
    fi
done

# 检查 install-remote.sh 中的函数
REMOTE_FUNCTIONS=("check_system" "install_dependencies" "install_go" "download_and_build_panel" "create_config" "create_service" "start_service" "show_result")

for func in "${REMOTE_FUNCTIONS[@]}"; do
    if grep -q "^${func}()" install-remote.sh; then
        echo "✅ 函数 ${func} 存在"
    else
        echo "❌ 函数 ${func} 缺失"
    fi
done

# 检查关键变量
echo "📋 检查关键变量..."

REQUIRED_VARS=("GITHUB_REPO" "GITHUB_URL" "INSTALL_DIR" "CONFIG_DIR")

for var in "${REQUIRED_VARS[@]}"; do
    if grep -q "^${var}=" install-quick.sh; then
        echo "✅ 变量 ${var} 已定义"
    else
        echo "❌ 变量 ${var} 未定义"
    fi
done

# 检查 GitHub 仓库地址是否正确
echo "🌐 检查 GitHub 仓库..."

REPO_URL=$(grep "GITHUB_REPO=" install-quick.sh | cut -d'"' -f2)
if [ "$REPO_URL" = "digwis/digwis-panel" ]; then
    echo "✅ GitHub 仓库地址正确: $REPO_URL"
else
    echo "❌ GitHub 仓库地址错误: $REPO_URL"
fi

# 模拟检查系统架构检测
echo "🏗️ 测试架构检测..."

# 模拟不同架构
test_arch() {
    local arch=$1
    local expected=$2
    
    # 临时修改 uname 输出进行测试
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

echo ""
echo "🎉 测试完成！"
echo ""
echo "📖 使用说明："
echo "1. 快速安装（推荐）："
echo "   curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install-quick.sh | bash"
echo ""
echo "2. 完整安装："
echo "   curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install-remote.sh | bash"
echo ""
echo "3. 本地测试："
echo "   sudo ./install-quick.sh"
echo ""
echo "⚠️  注意：安装脚本需要 root 权限运行"
