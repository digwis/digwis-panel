#!/bin/bash

# DigWis 面板发布脚本
# 用于构建多平台版本并打包发布

set -e

# 版本信息
VERSION="1.0.0"
APP_NAME="digwis-panel"

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
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

echo "🚀 DigWis 面板发布脚本 v${VERSION}"
echo "================================"

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ Go未安装，请先安装Go环境"
    exit 1
fi

print_info "Go版本: $(go version)"

# 创建发布目录
RELEASE_DIR="release"
rm -rf $RELEASE_DIR
mkdir -p $RELEASE_DIR

# 支持的平台
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "linux/arm"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

print_info "开始构建多平台版本..."

for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$platform"
    
    print_info "构建 ${GOOS}/${GOARCH}..."
    
    # 设置输出文件名
    OUTPUT_NAME="${APP_NAME}-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi
    
    # 构建
    env GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build \
        -a -ldflags '-extldflags "-static" -s -w' \
        -o "${RELEASE_DIR}/${OUTPUT_NAME}" main.go
    
    # 创建压缩包
    cd $RELEASE_DIR
    if [ "$GOOS" = "windows" ]; then
        zip "${APP_NAME}-${GOOS}-${GOARCH}.zip" "${OUTPUT_NAME}"
    else
        tar -czf "${APP_NAME}-${GOOS}-${GOARCH}.tar.gz" "${OUTPUT_NAME}"
    fi
    rm "${OUTPUT_NAME}"
    cd ..
    
    print_success "完成 ${GOOS}/${GOARCH}"
done

# 复制安装脚本
print_info "复制安装脚本..."
cp install.sh "${RELEASE_DIR}/install.sh"

# 创建版本信息文件
print_info "创建版本信息..."
cat > "${RELEASE_DIR}/version.json" << EOF
{
    "version": "${VERSION}",
    "build_time": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "platforms": [
        "linux/amd64",
        "linux/arm64", 
        "linux/arm",
        "darwin/amd64",
        "darwin/arm64",
        "windows/amd64"
    ]
}
EOF

# 创建README
print_info "创建发布说明..."
cat > "${RELEASE_DIR}/README.md" << EOF
# DigWis 服务器管理面板 v${VERSION}

## 一键安装

### Linux/macOS
\`\`\`bash
curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash
\`\`\`

或者

\`\`\`bash
wget -qO- https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash
\`\`\`

### 手动安装

1. 下载对应平台的二进制文件
2. 解压到 \`/opt/digwis/\` 目录
3. 运行安装脚本

## 支持的平台

- Linux (amd64, arm64, arm)
- macOS (amd64, arm64)
- Windows (amd64)

## 功能特性

- 🚀 单一二进制文件，无需依赖
- 🔒 内置SSL证书管理
- 🌐 支持Let's Encrypt自动证书
- 📊 系统监控和管理
- 🛠️ 环境一键安装
- 📁 文件管理
- 📝 日志查看
- ⚙️ 项目管理

## 管理命令

\`\`\`bash
digwis start      # 启动面板
digwis stop       # 停止面板
digwis restart    # 重启面板
digwis status     # 查看状态
digwis logs       # 查看日志
digwis uninstall  # 卸载面板
\`\`\`

## 访问地址

- HTTP: http://your-server-ip:8080
- HTTPS: https://your-server-ip:8080 (配置SSL证书后)

## 登录方式

使用系统用户账户登录，支持的管理员组：
- sudo
- wheel  
- admin
- root 用户

## 技术支持

- GitHub: https://github.com/digwis/digwis-panel
- Issues: https://github.com/digwis/digwis-panel/issues
EOF

# 显示发布信息
print_success "发布包构建完成！"
echo ""
echo "📦 发布文件:"
ls -la "${RELEASE_DIR}/"
echo ""
echo "📋 发布信息:"
echo "   版本: v${VERSION}"
echo "   构建时间: $(date)"
echo "   支持平台: ${#PLATFORMS[@]} 个"
echo ""
echo "🚀 下一步操作:"
echo "   1. 将 ${RELEASE_DIR}/ 目录下的文件上传到 GitHub Releases"
echo "   2. 创建 GitHub Release 标签"
echo "   3. 测试一键安装脚本"
echo ""
print_warning "记得更新 GitHub 仓库地址和版本号！"
