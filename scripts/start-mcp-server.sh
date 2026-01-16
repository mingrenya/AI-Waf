#!/bin/bash
#
# AI-WAF MCP Server 启动脚本
# 用于在Claude Desktop中配置MCP服务器
#

set -e

# 获取脚本所在目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

# 配置
MONGO_URI="${MONGO_URI:-mongodb://localhost:27017}"
DATABASE="${DATABASE:-waf}"

# 日志文件 (记录到stderr,不污染stdout)
LOG_FILE="${LOG_FILE:-$PROJECT_ROOT/logs/mcp-server.log}"
mkdir -p "$(dirname "$LOG_FILE")"

# 构建MCP Server (如果需要)
cd "$PROJECT_ROOT/coraza-spoa"
if [ ! -f "./mcp-server" ] || [ "$1" == "--rebuild" ]; then
    echo "[INFO] 编译 MCP Server..." >&2
    go build -o mcp-server ./cmd/mcp-server
fi

# 启动MCP Server
echo "[INFO] 启动 AI-WAF MCP Server..." >&2
echo "[INFO] MongoDB: $MONGO_URI" >&2
echo "[INFO] Database: $DATABASE" >&2

exec ./mcp-server \
    -mongo "$MONGO_URI" \
    -db "$DATABASE" \
    2>> "$LOG_FILE"
