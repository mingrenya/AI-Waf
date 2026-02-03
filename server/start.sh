#!/bin/bash

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "================================================"
echo "   AI-WAF 服务器快速启动"
echo "================================================"
echo ""

# 检查是否在 server 目录
if [ ! -f "main.go" ]; then
    echo -e "${RED}❌ 错误: 请在 server 目录下运行此脚本${NC}"
    echo "   cd /Users/duheling/Downloads/AI-Waf/server"
    exit 1
fi

# 检查 .env 文件
if [ ! -f ".env" ]; then
    echo -e "${YELLOW}⚠️  .env 文件不存在，从模板创建...${NC}"
    cp .env.template .env
    echo -e "${GREEN}✓ 已创建 .env 文件${NC}"
fi

# 验证 JWT_SECRET 长度
JWT_SECRET=$(grep "^JWT_SECRET=" .env | cut -d'=' -f2)
JWT_LEN=${#JWT_SECRET}

echo "检查 JWT_SECRET 配置..."
echo "  当前长度: $JWT_LEN 字符"

if [ $JWT_LEN -lt 32 ]; then
    echo -e "${RED}❌ JWT_SECRET 长度不足 32 字符${NC}"
    echo ""
    echo "正在生成新的安全密钥..."
    
    # 生成一个安全的随机密钥
    NEW_SECRET="aiwaf-secure-jwt-secret-key-2026-$(openssl rand -hex 32)"
    
    # 更新 .env 文件
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s/^JWT_SECRET=.*/JWT_SECRET=$NEW_SECRET/" .env
    else
        sed -i "s/^JWT_SECRET=.*/JWT_SECRET=$NEW_SECRET/" .env
    fi
    
    echo -e "${GREEN}✓ 已生成并更新 JWT_SECRET (长度: ${#NEW_SECRET} 字符)${NC}"
else
    echo -e "${GREEN}✓ JWT_SECRET 配置正确 (长度: $JWT_LEN 字符)${NC}"
fi

echo ""
echo "================================================"
echo "   启动服务器..."
echo "================================================"
echo ""

# 检查是否强制重置配置
if [ "$FORCE_RESET_CONFIG" = "true" ]; then
    echo "⚠️  FORCE_RESET_CONFIG=true，将更新数据库中的配置"
    echo ""
fi

# 检查 MongoDB
echo "检查 MongoDB 连接..."
if ! nc -z localhost 27017 2>/dev/null; then
    echo -e "${YELLOW}⚠️  警告: MongoDB 可能未运行 (localhost:27017)${NC}"
    echo "   请确保 MongoDB 已启动或使用 docker-compose"
fi

echo ""
echo -e "${GREEN}🚀 启动 AI-WAF 服务器...${NC}"
echo ""

# 启动服务器
go run main.go
