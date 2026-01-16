#!/bin/bash

# MCP Server 测试脚本

echo "=== 测试 MCP Server ==="
echo ""

echo "1. 检查容器状态"
docker ps --filter "name=mcp-server" --format "{{.Names}}: {{.Status}}"
echo ""

echo "2. 测试工具列表 (tools/list)"
echo '{"jsonrpc":"2.0","method":"tools/list","params":{},"id":1}' | \
  timeout 3 docker exec -i ai-waf-mcp-server ./mcp-server \
    -mongo mongodb://root:example@mongodb:27017 -db waf 2>&1 | \
  grep -A 50 "result" || echo "等待响应超时或无响应"
echo ""

echo "3. 查看最近日志"
docker logs --tail=10 ai-waf-mcp-server 2>&1
echo ""

echo "=== 测试完成 ==="
