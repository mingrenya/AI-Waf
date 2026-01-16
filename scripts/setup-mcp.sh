#!/bin/bash
# 一键配置MCP Server

cd "$(dirname "$0")/.."

echo "🚀 AI-Waf MCP Server 一键配置"
echo "================================"
echo ""

# 1. 获取token
echo "📍 步骤 1/3: 获取API Token"
./scripts/get-token.sh

echo ""
echo "📍 步骤 2/3: 验证配置"
if grep -q "^MCP_API_TOKEN=.*[a-zA-Z0-9]" .env 2>/dev/null; then
    echo "✅ Token已配置"
else
    echo "❌ Token未配置，请手动运行: ./scripts/get-token.sh"
    exit 1
fi

echo ""
echo "📍 步骤 3/3: 配置Claude Desktop"
echo ""
echo "请根据你的操作系统编辑配置文件："
echo ""
echo "macOS/Linux:"
echo "  编辑: ~/.config/Claude/claude_desktop_config.json"
echo ""
echo "Windows:"
echo "  编辑: %APPDATA%\\Claude\\claude_desktop_config.json"
echo ""
echo "配置内容（复制以下内容）:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat << 'EOF'
{
  "mcpServers": {
    "ai-waf": {
      "command": "docker",
      "args": [
        "exec",
        "-i",
        "ai-waf-mcp-server",
        "/app/ai-waf-mcp"
      ]
    }
  }
}
EOF
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "✅ 配置完成后，重启Claude Desktop即可使用"
echo ""
echo "💡 测试方法: 在Claude中说 \"帮我查看WAF攻击日志\""
