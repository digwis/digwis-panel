#!/bin/bash

# 简化的发布构建脚本
set -e

VERSION="v0.3.2"
RELEASE_DIR="releases/$VERSION"

echo "🚀 构建 DigWis Panel $VERSION"

# 创建发布目录
mkdir -p "$RELEASE_DIR"

# 生成模板
echo "📝 生成模板文件..."
go run github.com/a-h/templ/cmd/templ@latest generate

# 构建CSS
echo "🎨 构建CSS..."
npm run build-css-prod

# 构建AMD64版本
echo "🔨 构建AMD64版本..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "digwis-panel-$VERSION-linux-amd64" .

# 创建压缩包
echo "📦 创建压缩包..."
tar -czf "$RELEASE_DIR/digwis-panel-$VERSION-linux-amd64.tar.gz" "digwis-panel-$VERSION-linux-amd64"

# 生成校验和
echo "🔐 生成校验和..."
cd "$RELEASE_DIR"
sha256sum *.tar.gz > checksums.txt
cd - > /dev/null

# 清理二进制文件
rm "digwis-panel-$VERSION-linux-amd64"

# 清理备份文件
echo "🧹 清理备份文件..."
find releases/ -name "*.backup" -type f -delete
find releases/ -name "version.json.backup" -type f -delete

echo "✅ 构建完成！"
echo "📁 发布文件位于: $RELEASE_DIR"
ls -la "$RELEASE_DIR"
