#!/bin/bash

# 简单的 Git 推送脚本
# 不包含任何敏感信息

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

# 主函数
main() {
    local commit_message="$1"
    
    if [[ -z "$commit_message" ]]; then
        commit_message="Update: $(date '+%Y-%m-%d %H:%M:%S')"
    fi
    
    echo "📤 简单 Git 推送脚本"
    echo "===================="
    
    # 检查是否在 Git 仓库中
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        log_error "当前目录不是 Git 仓库"
        exit 1
    fi
    
    # 配置用户信息
    log_info "配置 Git 用户信息..."
    git config user.name "digwis"
    git config user.email "support@digwis.com"
    
    # 设置远程仓库
    log_info "设置远程仓库..."
    git remote set-url origin https://github.com/digwis/digwis-panel.git
    
    # 检查状态
    log_info "检查文件状态..."
    if [[ -n $(git status --porcelain) ]]; then
        log_info "发现未提交的更改"
        git status --short
        
        # 添加文件
        log_info "添加文件到暂存区..."
        git add .
        
        # 提交更改
        log_info "提交更改: $commit_message"
        git commit -m "$commit_message"
        log_success "提交成功"
    else
        log_info "没有新的更改需要提交"
    fi
    
    # 推送到远程仓库
    log_info "推送到远程仓库..."
    log_warning "请输入 GitHub 用户名和 Personal Access Token"
    
    if git push origin main; then
        log_success "推送成功"
        
        # 显示仓库信息
        echo ""
        log_success "仓库信息:"
        echo "  📁 仓库: https://github.com/digwis/digwis-panel"
        echo "  🌿 分支: $(git branch --show-current)"
        echo "  📝 最新提交: $(git log -1 --pretty=format:'%h - %s (%cr)')"
        echo ""
    else
        log_error "推送失败"
        exit 1
    fi
}

# 如果直接执行此脚本
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
