#!/bin/bash

# 版本管理脚本
# 用于创建 Git 标签和发布版本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助信息
show_help() {
    echo "🏷️  版本管理脚本"
    echo "================"
    echo ""
    echo "使用方法:"
    echo "  $0 [命令] [版本号]"
    echo ""
    echo "命令:"
    echo "  tag <version>     创建新的版本标签"
    echo "  list              列出所有版本标签"
    echo "  current           显示当前版本"
    echo "  delete <version>  删除版本标签"
    echo "  push              推送标签到远程仓库"
    echo ""
    echo "示例:"
    echo "  $0 tag v1.0.1     # 创建 v1.0.1 标签"
    echo "  $0 list           # 列出所有标签"
    echo "  $0 current        # 显示当前版本"
    echo "  $0 push           # 推送标签"
    echo ""
}

# 验证版本格式
validate_version() {
    local version="$1"
    
    if [[ ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?$ ]]; then
        log_error "版本格式无效: $version"
        log_info "正确格式: v1.0.0 或 v1.0.0-beta.1"
        return 1
    fi
    
    return 0
}

# 检查是否在 Git 仓库中
check_git_repo() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        log_error "当前目录不是 Git 仓库"
        exit 1
    fi
}

# 检查工作区是否干净
check_clean_working_tree() {
    if ! git diff-index --quiet HEAD --; then
        log_error "工作区有未提交的更改，请先提交或暂存"
        git status --short
        return 1
    fi
    
    return 0
}

# 创建版本标签
create_tag() {
    local version="$1"
    
    if [[ -z "$version" ]]; then
        log_error "请提供版本号"
        show_help
        exit 1
    fi
    
    # 验证版本格式
    if ! validate_version "$version"; then
        exit 1
    fi
    
    # 检查标签是否已存在
    if git tag -l | grep -q "^${version}$"; then
        log_error "标签 $version 已存在"
        exit 1
    fi
    
    # 检查工作区
    if ! check_clean_working_tree; then
        exit 1
    fi
    
    log_info "创建版本标签: $version"
    
    # 创建带注释的标签
    git tag -a "$version" -m "Release $version"
    
    log_success "标签 $version 创建成功"
    log_info "使用 '$0 push' 推送标签到远程仓库"
}

# 列出所有标签
list_tags() {
    log_info "所有版本标签:"
    
    if ! git tag -l | grep -q .; then
        log_warning "没有找到版本标签"
        return
    fi
    
    git tag -l --sort=-version:refname | while read tag; do
        local commit_date=$(git log -1 --format=%cd --date=short "$tag")
        local commit_msg=$(git tag -l --format='%(contents:subject)' "$tag")
        echo "  📌 $tag ($commit_date) - $commit_msg"
    done
}

# 显示当前版本
show_current() {
    log_info "当前版本信息:"
    
    # 尝试获取最新标签
    if git describe --tags --exact-match HEAD 2>/dev/null; then
        local current_tag=$(git describe --tags --exact-match HEAD)
        echo "  🏷️  当前标签: $current_tag"
    else
        local latest_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "无")
        local commit_hash=$(git rev-parse --short HEAD)
        local commit_count=$(git rev-list --count HEAD)
        
        echo "  🏷️  最新标签: $latest_tag"
        echo "  📝 当前提交: $commit_hash"
        echo "  🔢 提交数量: $commit_count"
        
        if [[ "$latest_tag" != "无" ]]; then
            local commits_since=$(git rev-list --count "${latest_tag}..HEAD")
            echo "  ➕ 自上次标签: +$commits_since 提交"
        fi
    fi
}

# 删除标签
delete_tag() {
    local version="$1"
    
    if [[ -z "$version" ]]; then
        log_error "请提供要删除的版本号"
        exit 1
    fi
    
    # 检查标签是否存在
    if ! git tag -l | grep -q "^${version}$"; then
        log_error "标签 $version 不存在"
        exit 1
    fi
    
    log_warning "确定要删除标签 $version 吗？(y/N)"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        log_info "取消删除"
        exit 0
    fi
    
    # 删除本地标签
    git tag -d "$version"
    log_success "本地标签 $version 已删除"
    
    # 询问是否删除远程标签
    log_warning "是否同时删除远程标签？(y/N)"
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        git push origin ":refs/tags/$version"
        log_success "远程标签 $version 已删除"
    fi
}

# 推送标签
push_tags() {
    log_info "推送标签到远程仓库..."
    
    # 推送所有标签
    git push origin --tags
    
    log_success "所有标签已推送到远程仓库"
}

# 主函数
main() {
    local command="$1"
    local version="$2"
    
    # 检查 Git 仓库
    check_git_repo
    
    case "$command" in
        tag)
            create_tag "$version"
            ;;
        list)
            list_tags
            ;;
        current)
            show_current
            ;;
        delete)
            delete_tag "$version"
            ;;
        push)
            push_tags
            ;;
        help|--help|-h)
            show_help
            ;;
        "")
            show_help
            ;;
        *)
            log_error "未知命令: $command"
            show_help
            exit 1
            ;;
    esac
}

# 如果直接执行此脚本
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
