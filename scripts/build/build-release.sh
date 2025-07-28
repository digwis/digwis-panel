#!/bin/bash

# DigWis 面板发布构建脚本
# 构建预编译二进制包，避免用户端编译

set -e

# 版本信息
# 优先使用 Git 标签，如果没有则使用参数，最后使用默认值
if git describe --tags --exact-match HEAD 2>/dev/null; then
    VERSION=$(git describe --tags --exact-match HEAD)
elif [[ -n "$1" ]]; then
    VERSION="$1"
else
    # 使用 Git 提交信息生成版本
    COMMIT_HASH=$(git rev-parse --short HEAD)
    COMMIT_DATE=$(git log -1 --format=%cd --date=format:%Y%m%d)
    VERSION="v1.0.0-dev.${COMMIT_DATE}.${COMMIT_HASH}"
fi

APP_NAME="digwis-panel"

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
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

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

echo "🚀 DigWis 面板发布构建脚本 ${VERSION}"
echo "================================"

# 检查Go环境
if ! command -v go &> /dev/null; then
    print_error "Go未安装，请先安装Go环境"
    exit 1
fi

print_info "Go版本: $(go version)"

# 获取项目根目录（脚本在 scripts/build/ 中）
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

print_info "脚本目录: $SCRIPT_DIR"
print_info "项目根目录: $PROJECT_ROOT"

# 创建发布目录
RELEASE_DIR="${PROJECT_ROOT}/releases"
rm -rf "$RELEASE_DIR"
mkdir -p "$RELEASE_DIR"

# 支持的平台
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "linux/arm"
)

print_info "开始构建多平台版本..."

for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$platform"
    
    print_info "构建 ${GOOS}/${GOARCH}..."
    
    # 设置输出文件名
    BINARY_NAME="digwis"
    PACKAGE_NAME="${APP_NAME}-${VERSION}-${GOOS}-${GOARCH}"
    
    # 创建临时目录
    TEMP_DIR="/tmp/${PACKAGE_NAME}"
    rm -rf "$TEMP_DIR"
    mkdir -p "$TEMP_DIR"
    
    # 构建二进制文件（在项目根目录执行）
    cd "$PROJECT_ROOT"
    env GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build \
        -a -ldflags '-extldflags "-static" -s -w' \
        -o "${TEMP_DIR}/${BINARY_NAME}" .
    
    if [ ! -f "${TEMP_DIR}/${BINARY_NAME}" ]; then
        print_error "构建 ${GOOS}/${GOARCH} 失败"
        continue
    fi
    
    # 复制其他必要文件
    cp "${PROJECT_ROOT}/README.md" "${TEMP_DIR}/" 2>/dev/null || true
    cp "${PROJECT_ROOT}/LICENSE" "${TEMP_DIR}/" 2>/dev/null || true
    
    # 创建安装说明
    cat > "${TEMP_DIR}/INSTALL.txt" << EOF
DigWis 面板 ${VERSION} - ${GOOS}/${GOARCH}

安装说明：
1. 解压此文件到 /opt/digwis/ 目录
2. 给予执行权限: chmod +x digwis
3. 运行: ./digwis

或者使用一键安装脚本：
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash

访问地址: http://your-server-ip:8080
EOF
    
    # 创建压缩包
    cd "$TEMP_DIR"
    tar -czf "${PACKAGE_NAME}.tar.gz" *

    # 移动到发布目录
    mv "${PACKAGE_NAME}.tar.gz" "${RELEASE_DIR}/"
    cd - > /dev/null
    
    # 清理临时目录
    rm -rf "$TEMP_DIR"
    
    print_success "完成 ${GOOS}/${GOARCH}"
done

# 创建版本信息文件
print_info "创建版本信息..."
cat > "${RELEASE_DIR}/version.json" << EOF
{
    "version": "${VERSION}",
    "build_time": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "platforms": [
        "linux/amd64",
        "linux/arm64", 
        "linux/arm"
    ]
}
EOF

# 创建校验文件
print_info "生成校验文件..."
cd "$RELEASE_DIR"
sha256sum *.tar.gz > checksums.txt
cd - > /dev/null

# 显示发布信息
print_success "发布包构建完成！"
echo ""
echo "📦 发布文件:"
ls -la "${RELEASE_DIR}/"
echo ""
echo "📋 发布信息:"
echo "   版本: ${VERSION}"
echo "   构建时间: $(date)"
echo "   支持平台: ${#PLATFORMS[@]} 个"
echo ""
echo "🚀 下一步操作:"
echo "   1. 将 ${RELEASE_DIR}/ 目录下的文件上传到 GitHub Releases"
echo "   2. 创建 GitHub Release 标签: ${VERSION}"
echo "   3. 测试一键安装脚本"
echo ""
echo "📝 发布命令示例:"
echo "   gh release create ${VERSION} ${RELEASE_DIR}/* --title \"${VERSION}\" --notes \"Release ${VERSION}\""
echo ""
print_warning "记得更新 CDN 和版本号！"
