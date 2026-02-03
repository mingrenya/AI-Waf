#!/bin/bash
# 为MCP服务生成90天长期Token

WAF_URL="http://localhost:2333"

echo "🔐 MCP服务 - 90天长期Token生成工具"
echo "========================================"
echo ""

# 检查后端是否运行
echo "检查后端服务..."
if ! curl -s "${WAF_URL}/health" > /dev/null 2>&1; then
    echo "❌ 错误: 后端服务未运行或无法访问"
    echo "请先启动服务: docker compose up -d"
    exit 1
fi
echo "✅ 后端服务正常"
echo ""

# 提示输入用户名和密码（默认admin账号）
read -p "请输入用户名 (默认: admin): " username
username=${username:-admin}

read -sp "请输入密码 (默认: admin@123): " password
password=${password:-admin@123}
echo ""
echo ""

# 调用服务账号登录API（90天有效期）
echo "正在生成长期Token..."
response=$(curl -s -X POST "${WAF_URL}/api/v1/auth/login-service" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"${username}\",\"password\":\"${password}\"}")

# 检查是否成功
if echo "$response" | grep -q '"code":200'; then
    # 提取token
    token=$(echo "$response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    
    if [ -n "$token" ]; then
        echo "✅ Token生成成功！"
        echo ""
        echo "📋 90天有效期MCP Token:"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "$token"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""
        echo "📝 下一步操作:"
        echo "1. 复制上面的Token"
        echo "2. 编辑 .env 文件："
        echo "   MCP_API_TOKEN=<粘贴Token>"
        echo ""
        echo "3. 重启MCP服务："
        echo "   docker-compose restart mcp-server"
        echo ""
        echo "⚠️  注意: 此Token有效期90天，过期后需重新生成"
        echo ""
        
        # 自动更新.env文件（如果存在）
        if [ -f ".env" ]; then
            read -p "是否自动更新.env文件? (y/n): " update_env
            if [ "$update_env" = "y" ] || [ "$update_env" = "Y" ]; then
                if grep -q "^MCP_API_TOKEN=" .env; then
                    # 使用sed更新（兼容macOS和Linux）
                    if [[ "$OSTYPE" == "darwin"* ]]; then
                        sed -i '' "s|^MCP_API_TOKEN=.*|MCP_API_TOKEN=$token|" .env
                    else
                        sed -i "s|^MCP_API_TOKEN=.*|MCP_API_TOKEN=$token|" .env
                    fi
                    echo "✅ .env文件已更新"
                else
                    echo "MCP_API_TOKEN=$token" >> .env
                    echo "✅ 已添加MCP_API_TOKEN到.env"
                fi
                
                echo ""
                echo "🚀 执行以下命令重启MCP服务:"
                echo "   docker-compose restart mcp-server"
            fi
        fi
    else
        echo "❌ 错误: 无法从响应中提取Token"
        echo "服务器响应: $response"
    fi
else
    echo "❌ 登录失败"
    echo "服务器响应: $response"
    exit 1
fi
