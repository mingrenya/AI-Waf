#!/bin/bash
# AI-Waf 获取API Token工具

WAF_URL="http://localhost:2333"

echo "🔐 AI-Waf API Token 获取工具"
echo "================================"
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

# 提示输入用户名和密码
read -p "请输入用户名 (默认: admin): " username
username=${username:-admin}

read -sp "请输入密码 (默认: admin123): " password
password=${password:-admin123}
echo ""
echo ""

# 调用登录API
echo "正在登录..."
response=$(curl -s -X POST "${WAF_URL}/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"${username}\",\"password\":\"${password}\"}")

# 检查是否成功
if echo "$response" | grep -q '"code":200'; then
    # 提取token
    token=$(echo "$response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    
    if [ -n "$token" ]; then
        echo "✅ 登录成功！"
        echo ""
        echo "📋 你的API Token:"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "$token"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""
        echo "📝 下一步操作:"
        echo "1. 复制上面的token"
        echo "2. 编辑 .env 文件："
        echo "   MCP_API_TOKEN=$token"
        echo ""
        echo "3. 重启MCP Server："
        echo "   docker compose restart mcp-server"
        echo ""
        
        # 询问是否自动更新.env
        read -p "是否自动更新 .env 文件? (y/n): " update_env
        if [ "$update_env" = "y" ] || [ "$update_env" = "Y" ]; then
            if [ -f ".env" ]; then
                # 备份原文件
                cp .env .env.backup
                # 更新或添加MCP_API_TOKEN
                if grep -q "^MCP_API_TOKEN=" .env; then
                    sed -i.bak "s|^MCP_API_TOKEN=.*|MCP_API_TOKEN=$token|" .env
                else
                    echo "MCP_API_TOKEN=$token" >> .env
                fi
                echo "✅ .env 文件已更新 (原文件备份为 .env.backup)"
                echo ""
                echo "现在重启MCP Server..."
                docker compose restart mcp-server
                echo "✅ 完成！"
            else
                echo "❌ 未找到 .env 文件"
            fi
        fi
    else
        echo "❌ 无法提取token"
        echo "响应: $response"
    fi
else
    echo "❌ 登录失败"
    echo "响应: $response"
    echo ""
    echo "💡 提示: 如果是首次使用，可能需要先创建管理员账号"
fi
