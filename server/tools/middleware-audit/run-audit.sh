#!/bin/bash

# Gin 中间件审计工具
# 用于检查路由是否正确使用了安全中间件

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVER_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "🔍 Running Gin Middleware Audit..."
echo "Server directory: $SERVER_DIR"
echo ""

cd "$SERVER_DIR/tools/middleware-audit"

# 编译审计工具
echo "📦 Building audit tool..."
go build -o audit-tool main.go

# 运行审计
echo ""
echo "🚀 Running audit..."
echo ""

cd "$SERVER_DIR"
./tools/middleware-audit/audit-tool

# 清理
rm -f ./tools/middleware-audit/audit-tool

echo ""
echo "✅ Audit complete!"
