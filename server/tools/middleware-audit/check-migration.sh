#!/bin/bash

# AI-WAF 路由中间件迁移脚本
# 用于检查哪些路由可以使用辅助函数优化

set -e

echo "🔍 AI-WAF 路由中间件迁移检查"
echo "================================"
echo ""

ROUTER_FILE="./router/router.go"

if [ ! -f "$ROUTER_FILE" ]; then
    echo "❌ 错误: 找不到 $ROUTER_FILE"
    echo "   请在 server 目录下运行此脚本"
    exit 1
fi

echo "📁 分析文件: $ROUTER_FILE"
echo ""

# 统计各类路由
echo "📊 路由统计"
echo "----------"

# GET/:id 路由（可使用 IDRoute）
ID_ROUTES=$(grep -E '\.(GET|DELETE)\("/:id"' "$ROUTER_FILE" | wc -l | tr -d ' ')
echo "  GET/DELETE /:id 路由: $ID_ROUTES 个（可用 IDRoute 或 DeleteRoute）"

# POST 路由（可使用 CreateRoute）
POST_ROUTES=$(grep -E '\.POST\("",.*Create' "$ROUTER_FILE" | wc -l | tr -d ' ')
echo "  POST 创建路由: $POST_ROUTES 个（可用 CreateRoute）"

# PUT/:id 路由（可使用 UpdateRoute）
PUT_ROUTES=$(grep -E '\.PUT\("/:id"' "$ROUTER_FILE" | wc -l | tr -d ' ')
echo "  PUT /:id 更新路由: $PUT_ROUTES 个（可用 UpdateRoute）"

# GET 列表路由（可使用 ListRoute）
LIST_ROUTES=$(grep -E '\.GET\("",.*Get[A-Z]' "$ROUTER_FILE" | wc -l | tr -d ' ')
echo "  GET 列表路由: $LIST_ROUTES 个（可用 ListRoute）"

echo ""

# 计算总数
TOTAL=$((ID_ROUTES + POST_ROUTES + PUT_ROUTES + LIST_ROUTES))
echo "📈 总共可优化: $TOTAL 个路由"
echo ""

# 详细分析
echo "🔎 详细分析"
echo "----------"

echo ""
echo "1️⃣  缺少 ID 验证的路由（建议使用 IDRoute/UpdateRoute/DeleteRoute）:"
echo ""
grep -n -E '\.(GET|PUT|DELETE)\("/:id"' "$ROUTER_FILE" | \
    grep -v "ValidateMongoID" | \
    sed 's/^/   Line /' || echo "   ✅ 没有发现"

echo ""
echo "2️⃣  缺少 Content-Type 验证的 POST 路由（建议使用 CreateRoute）:"
echo ""
grep -n -E '\.POST\("",.*middleware\.HasPermission' "$ROUTER_FILE" | \
    grep -v "ValidateJSONContentType" | \
    head -10 | \
    sed 's/^/   Line /' || echo "   ✅ 没有发现"

echo ""
echo "3️⃣  缺少分页验证的列表路由（建议使用 ListRoute）:"
echo ""
grep -n -E '\.GET\("",.*Get[A-Z]' "$ROUTER_FILE" | \
    grep -v "ValidatePagination" | \
    head -10 | \
    sed 's/^/   Line /' || echo "   ✅ 没有发现"

echo ""
echo "4️⃣  DELETE 操作但没有审计日志:"
echo ""
grep -n -E '\.DELETE\(' "$ROUTER_FILE" | \
    grep -v "SecurityAudit" | \
    head -10 | \
    sed 's/^/   Line /' || echo "   ✅ 没有发现"

echo ""
echo "✅ 检查完成!"
echo ""
echo "📖 下一步："
echo "   1. 查看 router/HELPERS_GUIDE.md 了解如何使用辅助函数"
echo "   2. 逐个模块迁移（建议从用户管理开始）"
echo "   3. 每次迁移后运行测试"
echo ""
echo "🛠️  示例迁移命令："
echo "   # 编辑路由文件"
echo "   vim $ROUTER_FILE"
echo ""
echo "   # 运行测试"
echo "   go test ./controller/..."
echo ""
