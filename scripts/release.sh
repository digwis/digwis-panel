#!/bin/bash

# DigWis Panel 发布管理脚本
# 用于创建新版本发布包

set -e

# 配置
RELEASE_DIR="releases"
PLATFORMS=("linux/amd64")
ARCHS=("amd64")

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

# 检查版本号格式
check_version() {
    if [[ ! $1 =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        print_error "版本号格式错误，应为 vX.Y.Z 格式，如 v1.0.0"
        exit 1
    fi
}

# 检查是否在项目根目录
check_project_root() {
    if [ ! -f "main.go" ] || [ ! -f "go.mod" ]; then
        print_error "请在项目根目录运行此脚本"
        exit 1
    fi
}

# 构建所有平台版本
build_all_platforms() {
    local version=$1
    local version_dir="$RELEASE_DIR/$version"
    
    print_step "构建所有平台版本..."
    
    # 创建版本目录
    mkdir -p "$version_dir"
    
    # 确保模板已生成
    if command -v templ >/dev/null 2>&1; then
        print_info "重新生成模板..."
        templ generate
    fi
    
    # 构建每个平台
    for i in "${!PLATFORMS[@]}"; do
        local platform="${PLATFORMS[$i]}"
        local arch="${ARCHS[$i]}"
        local os=$(echo $platform | cut -d'/' -f1)
        local arch_name=$(echo $platform | cut -d'/' -f2)
        
        print_info "构建 $platform..."
        
        # 设置环境变量
        export GOOS=$os
        export GOARCH=$arch_name
        export CGO_ENABLED=1  # 启用 CGO 以支持 sqlite3
        
        # 构建
        local binary_name="digwis-panel-$version-$os-$arch"
        if [ "$arch_name" = "amd64" ]; then
            CC=x86_64-linux-gnu-gcc go build -ldflags "-s -w" -o "$binary_name" .
        else
            go build -ldflags "-s -w" -o "$binary_name" .
        fi
        
        # 创建压缩包
        local package_name="$binary_name.tar.gz"
        tar -czf "$version_dir/$package_name" "$binary_name"
        
        # 清理二进制文件
        rm "$binary_name"
        
        print_success "$platform 构建完成"
    done
    
    # 重置环境变量
    unset GOOS GOARCH CGO_ENABLED
}

# 更新 version.json
update_version_json() {
    local version=$1
    local changelog="$2"
    
    print_step "更新版本信息..."
    
    # 备份原文件
    cp "$RELEASE_DIR/version.json" "$RELEASE_DIR/version.json.backup"
    
    # 创建新的 version.json
    cat > "$RELEASE_DIR/version.json" << EOF
{
    "latest": "$version",
    "versions": {
        "$version": {
            "build_time": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
            "platforms": [
                "linux/amd64"
            ],
            "stable": true,
            "changelog": "$changelog"
        }
    },
    "download_base": "https://github.com/digwis/digwis-panel/releases/download",
    "fallback_base": "https://raw.githubusercontent.com/digwis/digwis-panel/main/releases"
}
EOF
    
    print_success "版本信息已更新"
}

# 生成校验和
generate_checksums() {
    local version=$1
    local version_dir="$RELEASE_DIR/$version"
    
    print_step "生成校验和..."
    
    cd "$version_dir"
    sha256sum *.tar.gz > checksums.txt
    cd - > /dev/null
    
    # 更新全局校验和
    cd "$RELEASE_DIR"
    find . -name "*.tar.gz" -exec sha256sum {} \; > checksums.txt
    cd - > /dev/null
    
    print_success "校验和已生成"
}

# 清理旧版本（保留最近3个版本）
cleanup_old_versions() {
    print_step "清理旧版本..."
    
    local versions=($(ls -1 "$RELEASE_DIR" | grep "^v[0-9]" | sort -V))
    local total=${#versions[@]}
    
    if [ $total -gt 3 ]; then
        local to_remove=$((total - 3))
        for ((i=0; i<to_remove; i++)); do
            local old_version="${versions[$i]}"
            print_info "删除旧版本: $old_version"
            rm -rf "$RELEASE_DIR/$old_version"
        done
        print_success "已清理 $to_remove 个旧版本"
    else
        print_info "版本数量未超过限制，无需清理"
    fi
}

# 推送到 Git 仓库
push_to_git() {
    local version=$1
    local changelog="$2"

    print_step "推送源代码到 Git 仓库..."

    # 添加 releases 目录的更改
    git add releases/
    git commit -m "Release $version: $changelog"
    git push origin main

    print_success "源代码已推送到 Git 仓库"
}

# 创建 GitHub Release
create_github_release() {
    local version=$1
    local changelog="$2"
    local version_dir="$RELEASE_DIR/$version"

    print_step "创建 GitHub Release..."

    # 检查是否安装了 gh CLI
    if ! command -v gh &> /dev/null; then
        print_warning "GitHub CLI (gh) 未安装，跳过自动创建 Release"
        print_info "手动创建命令:"
        echo "  gh release create $version $version_dir/*.tar.gz --title \"Release $version\" --notes \"$changelog\""
        return
    fi

    # 检查是否已登录
    if ! gh auth status &> /dev/null; then
        print_warning "GitHub CLI 未登录，跳过自动创建 Release"
        print_info "请先登录: gh auth login"
        print_info "然后手动创建: gh release create $version $version_dir/*.tar.gz --title \"Release $version\" --notes \"$changelog\""
        return
    fi

    # 创建 Release
    if gh release create "$version" "$version_dir"/*.tar.gz \
        --title "Release $version" \
        --notes "$changelog"; then
        print_success "GitHub Release 创建成功"
    else
        print_error "GitHub Release 创建失败"
        print_info "手动创建命令:"
        echo "  gh release create $version $version_dir/*.tar.gz --title \"Release $version\" --notes \"$changelog\""
    fi
}

# 显示发布信息
show_release_info() {
    local version=$1
    local version_dir="$RELEASE_DIR/$version"

    echo ""
    echo "=================================================="
    print_success "版本 $version 发布完成！"
    echo "=================================================="
    echo ""
    print_info "📦 发布包位置: $version_dir"
    print_info "📋 包含文件:"
    ls -la "$version_dir"
    echo ""
    print_info "🌐 远程安装命令:"
    echo "  curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/install.sh | sudo bash"
    echo ""
    print_info "🔄 远程升级命令:"
    echo "  curl -sSL https://raw.githubusercontent.com/digwis/digwis-panel/main/upgrade.sh | sudo bash"
    echo ""
}

# 主函数
main() {
    local version="$1"
    local changelog="$2"
    local auto_push="$3"

    if [ -z "$version" ]; then
        echo "使用方法: $0 <version> [changelog] [--auto]"
        echo "示例: "
        echo "  $0 v1.0.0 \"新增功能：支持嵌入式静态文件\""
        echo "  $0 v1.0.0 \"新增功能：支持嵌入式静态文件\" --auto"
        echo ""
        echo "选项:"
        echo "  --auto    自动推送到 Git 和创建 GitHub Release"
        exit 1
    fi

    if [ -z "$changelog" ]; then
        changelog="Version $version release"
    fi

    echo "=================================================="
    echo "    DigWis Panel 发布脚本"
    echo "=================================================="
    echo ""

    check_version "$version"
    check_project_root

    print_info "准备发布版本: $version"
    print_info "更新日志: $changelog"
    if [ "$auto_push" = "--auto" ]; then
        print_info "模式: 自动化发布"
    else
        print_info "模式: 本地构建"
    fi
    echo ""

    build_all_platforms "$version"
    update_version_json "$version" "$changelog"
    generate_checksums "$version"
    cleanup_old_versions

    # 自动化选项
    if [ "$auto_push" = "--auto" ]; then
        push_to_git "$version" "$changelog"
        create_github_release "$version" "$changelog"
    fi

    show_release_info "$version" "$changelog"
}

# 执行主函数
main "$@"
