#!/bin/bash

# DigWis Panel Air 启动脚本
# 包含错误处理和自动重启功能

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$PROJECT_ROOT"

# Air 可执行文件路径
AIR_BIN="/home/parallels/go/bin/air"

# 日志文件
LOG_FILE="$PROJECT_ROOT/air.log"

# 检查函数
check_memory() {
    local available=$(free -m | awk 'NR==2{printf "%.0f", $7}')
    if [ "$available" -lt 200 ]; then
        echo -e "${YELLOW}⚠️  警告: 可用内存不足 ${available}MB，建议释放一些内存${NC}"
        return 1
    fi
    return 0
}

check_air_binary() {
    if [ ! -f "$AIR_BIN" ]; then
        echo -e "${RED}❌ Air 二进制文件不存在: $AIR_BIN${NC}"
        echo -e "${BLUE}💡 请运行: go install github.com/air-verse/air@latest${NC}"
        exit 1
    fi
}

cleanup() {
    echo -e "\n${YELLOW}🧹 清理进程...${NC}"
    # 杀死所有相关进程
    pkill -f "tmp/main" 2>/dev/null || true
    pkill -f "air" 2>/dev/null || true
    exit 0
}

# 信号处理
trap cleanup SIGINT SIGTERM

# 主函数
main() {
    echo -e "${BLUE}🚀 DigWis Panel Air 启动脚本${NC}"
    echo -e "${BLUE}================================${NC}"
    
    # 检查 Air 二进制文件
    check_air_binary
    
    # 检查内存
    if ! check_memory; then
        echo -e "${YELLOW}💡 提示: 可以尝试运行 'sudo sync && sudo sysctl vm.drop_caches=3' 清理缓存${NC}"
    fi
    
    # 清理旧的进程和文件
    echo -e "${YELLOW}🧹 清理旧进程和临时文件...${NC}"
    pkill -f "tmp/main" 2>/dev/null || true
    rm -rf tmp/
    
    # 设置环境变量
    export DIGWIS_MODE=development
    export DIGWIS_DATA_DIR="$PROJECT_ROOT/data"
    
    echo -e "${GREEN}✅ 环境准备完成${NC}"
    echo -e "${BLUE}📁 数据目录: $DIGWIS_DATA_DIR${NC}"
    echo -e "${BLUE}📝 日志文件: $LOG_FILE${NC}"
    echo -e "${BLUE}🔧 Air 配置: .air.toml${NC}"
    echo ""
    
    # 启动 Air
    echo -e "${GREEN}🚀 启动 Air...${NC}"
    echo -e "${YELLOW}💡 按 Ctrl+C 停止服务${NC}"
    echo ""
    
    # 使用 tee 同时输出到控制台和日志文件
    "$AIR_BIN" 2>&1 | tee "$LOG_FILE"
}

# 运行主函数
main "$@"
