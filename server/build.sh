#!/bin/bash

set -e

echo "🚀 开始构建 RuiQi WAF..."

# 环境检查
echo "🔍 检查构建环境..."

# 检查 Node.js 版本
REQUIRED_NODE="23.10.0"
if command -v node &> /dev/null; then
    NODE_VERSION=$(node --version | sed 's/v//')
    echo "📦 Node.js 版本: $NODE_VERSION (要求: $REQUIRED_NODE)"
    if [ "$NODE_VERSION" != "$REQUIRED_NODE" ]; then
        echo "⚠️  警告: Node.js 版本不匹配，建议使用 v$REQUIRED_NODE"
    fi
else
    echo "❌ 错误: 未找到 Node.js，请先安装 Node.js $REQUIRED_NODE"
    exit 1
fi

# 检查 pnpm 版本
REQUIRED_PNPM="10.11.0"
if command -v pnpm &> /dev/null; then
    PNPM_VERSION=$(pnpm --version)
    echo "📦 pnpm 版本: $PNPM_VERSION (要求: $REQUIRED_PNPM)"
    if [ "$PNPM_VERSION" != "$REQUIRED_PNPM" ]; then
        echo "⚠️  警告: pnpm 版本不匹配，建议使用 $REQUIRED_PNPM"
    fi
else
    echo "❌ 错误: 未找到 pnpm，请先安装 pnpm $REQUIRED_PNPM"
    echo "💡 安装命令: npm install -g pnpm@$REQUIRED_PNPM"
    exit 1
fi

# 检查 Go 版本
REQUIRED_GO="1.24.1"
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    echo "🔧 Go 版本: $GO_VERSION (要求: $REQUIRED_GO)"
    if [ "$GO_VERSION" != "$REQUIRED_GO" ]; then
        echo "⚠️  警告: Go 版本不匹配，建议使用 $REQUIRED_GO"
    fi
else
    echo "❌ 错误: 未找到 Go，请先安装 Go $REQUIRED_GO"
    exit 1
fi

echo "✅ 环境检查完成"
echo ""

# 1. 构建前端
echo "📦 构建前端资源..."
cd ../web
pnpm install
pnpm build
cd ../server

# 2. 复制前端资源到嵌入目录
echo "📋 复制前端资源..."
mkdir -p public/dist
cp -r ../web/dist/* public/dist/

# 3. 构建后端
echo "🔧 构建后端服务..."
go mod tidy
go build -o ruiqi-waf .

echo "✅ 构建完成！"
echo "📍 可执行文件位置: server/ruiqi-waf" 