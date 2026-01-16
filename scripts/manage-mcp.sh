#!/bin/bash

# AI-WAF MCP Server 启动脚本
# 使用方式:
#   1. 直接启动 (本地开发): ./start-mcp-server.sh
#   2. Docker 启动: docker compose up -d mcp-server
#   3. 连接到运行中的容器: docker exec -it ai-waf-mcp-server sh

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== AI-WAF MCP Server 管理脚本 ===${NC}"

# 默认配置
MONGO_URI="${MONGO_URI:-mongodb://root:example@localhost:27017}"
DATABASE="${DATABASE:-waf}"
MODE="${1:-local}"

case "$MODE" in
    "local")
        echo -e "${YELLOW}启动模式: 本地开发${NC}"
        echo -e "MongoDB: $MONGO_URI"
        echo -e "Database: $DATABASE"
        
        # 检查二进制文件
        if [ ! -f "./coraza-spoa/mcp-server" ]; then
            echo -e "${YELLOW}编译 MCP Server...${NC}"
            cd coraza-spoa
            go build -o mcp-server ./cmd/mcp-server
            cd ..
        fi
        
        echo -e "${GREEN}启动 MCP Server (STDIO 模式)...${NC}"
        exec ./coraza-spoa/mcp-server -mongo "$MONGO_URI" -db "$DATABASE"
        ;;
        
    "docker")
        echo -e "${YELLOW}启动模式: Docker${NC}"
        docker compose up -d mcp-server
        echo -e "${GREEN}MCP Server 已在 Docker 中启动${NC}"
        echo -e "查看日志: ${YELLOW}docker logs -f ai-waf-mcp-server${NC}"
        ;;
        
    "logs")
        echo -e "${YELLOW}查看 MCP Server 日志${NC}"
        docker logs -f ai-waf-mcp-server
        ;;
        
    "stop")
        echo -e "${YELLOW}停止 MCP Server${NC}"
        docker compose stop mcp-server
        ;;
        
    "restart")
        echo -e "${YELLOW}重启 MCP Server${NC}"
        docker compose restart mcp-server
        ;;
        
    "shell")
        echo -e "${YELLOW}进入 MCP Server 容器${NC}"
        docker exec -it ai-waf-mcp-server sh
        ;;
        
    *)
        echo -e "${RED}未知模式: $MODE${NC}"
        echo ""
        echo "使用方法:"
        echo "  $0 [local|docker|logs|stop|restart|shell]"
        echo ""
        echo "模式说明:"
        echo "  local   - 本地启动 (默认)"
        echo "  docker  - Docker 容器启动"
        echo "  logs    - 查看 Docker 日志"
        echo "  stop    - 停止 Docker 容器"
        echo "  restart - 重启 Docker 容器"
        echo "  shell   - 进入 Docker 容器"
        exit 1
        ;;
esac
