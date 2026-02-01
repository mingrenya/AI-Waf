#!/bin/bash
# AI-Waf MCP Server 环境变量修复与验证脚本
# 用途: 自动检测和修复401认证问题

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 图标
CHECK="✅"
CROSS="❌"
WARN="⚠️"
INFO="ℹ️"

echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}AI-Waf MCP Server 环境修复工具${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""

# 1. 检查.env文件
echo -e "${INFO} [1/6] 检查 .env 配置文件..."
if [ ! -f .env ]; then
    echo -e "${CROSS} ${RED}错误: .env 文件不存在${NC}"
    exit 1
fi

# 检查MCP_API_TOKEN
TOKEN=$(grep "^MCP_API_TOKEN=" .env | cut -d'=' -f2)
if [ -z "$TOKEN" ]; then
    echo -e "${CROSS} ${RED}错误: .env 中未配置 MCP_API_TOKEN${NC}"
    echo -e "${INFO} 请先设置有效的JWT token"
    exit 1
fi

# 验证token格式 (JWT应该包含两个点)
if [[ "$TOKEN" == *.*.* ]]; then
    echo -e "${CHECK} ${GREEN}找到有效的 MCP_API_TOKEN (${#TOKEN} 字符)${NC}"
else
    echo -e "${WARN} ${YELLOW}警告: Token格式可能不正确 (JWT应包含两个点)${NC}"
fi

# 2. 检查docker-compose.yaml配置
echo -e "\n${INFO} [2/6] 检查 docker-compose.yaml 配置..."
if ! grep -q "env_file:" docker-compose.yaml; then
    echo -e "${CROSS} ${RED}错误: docker-compose.yaml 缺少 env_file 配置${NC}"
    echo -e "${INFO} 请确保 mcp-server 服务包含以下配置:"
    echo -e "  ${YELLOW}env_file:${NC}"
    echo -e "  ${YELLOW}  - .env${NC}"
    exit 1
fi
echo -e "${CHECK} ${GREEN}docker-compose.yaml 配置正确${NC}"

# 3. 停止现有容器
echo -e "\n${INFO} [3/6] 停止现有服务..."
docker compose down
echo -e "${CHECK} ${GREEN}服务已停止${NC}"

# 4. 重新构建MCP Server镜像
echo -e "\n${INFO} [4/6] 重新构建 mcp-server 镜像..."
docker compose build --no-cache mcp-server
echo -e "${CHECK} ${GREEN}镜像构建完成${NC}"

# 5. 启动服务
echo -e "\n${INFO} [5/6] 启动服务..."
docker compose up -d
echo -e "${CHECK} ${GREEN}服务已启动${NC}"

# 等待服务启动
echo -e "\n${INFO} 等待服务启动 (5秒)..."
sleep 5

# 6. 验证环境变量注入
echo -e "\n${INFO} [6/6] 验证环境变量注入..."
CONTAINER_TOKEN=$(docker compose exec -T mcp-server env | grep "^WAF_API_TOKEN=" | cut -d'=' -f2 || echo "")

if [ -z "$CONTAINER_TOKEN" ]; then
    echo -e "${CROSS} ${RED}失败: 容器内 WAF_API_TOKEN 为空${NC}"
    echo -e "${INFO} 请检查 docker-compose.yaml 配置"
    exit 1
fi

if [ "$TOKEN" = "$CONTAINER_TOKEN" ]; then
    echo -e "${CHECK} ${GREEN}环境变量注入成功！${NC}"
    echo -e "   ${GREEN}Token长度: ${#CONTAINER_TOKEN} 字符${NC}"
else
    echo -e "${WARN} ${YELLOW}警告: 容器内token与.env不一致${NC}"
fi

# 显示容器状态
echo -e "\n${BLUE}======================================${NC}"
echo -e "${BLUE}容器状态${NC}"
echo -e "${BLUE}======================================${NC}"
docker compose ps

# 显示MCP Server日志预览
echo -e "\n${BLUE}======================================${NC}"
echo -e "${BLUE}MCP Server 日志 (最近20行)${NC}"
echo -e "${BLUE}======================================${NC}"
docker compose logs --tail=20 mcp-server

# 最终验证建议
echo -e "\n${BLUE}======================================${NC}"
echo -e "${BLUE}验证建议${NC}"
echo -e "${BLUE}======================================${NC}"
echo -e "${INFO} 1. 检查日志中是否有 '未设置 WAF_API_TOKEN' 警告"
echo -e "${INFO} 2. 测试工具调用: ./test-mcp.sh (如果存在)"
echo -e "${INFO} 3. 实时查看日志: docker compose logs -f mcp-server"
echo -e "${INFO} 4. 验证后端连通: docker compose exec mcp-server wget -O- http://mrya:2333"
echo ""
echo -e "${GREEN}修复完成！${NC}"
echo ""
