#!/bin/bash

# WAF快速测试脚本 - 验证基本拦截功能

WAF_URL="http://localhost:2333"

echo "========================================="
echo "WAF 快速测试"
echo "========================================="
echo ""

# 测试1: 正常请求
echo "测试1: 正常请求 (应该通过)"
curl -s -o /dev/null -w "HTTP状态码: %{http_code}\n" "$WAF_URL/"
echo ""

# 测试2: SQL注入攻击
echo "测试2: SQL注入攻击 (应该被拦截)"
curl -s -o /dev/null -w "HTTP状态码: %{http_code}\n" "$WAF_URL/?id=1' OR '1'='1"
echo ""

# 测试3: XSS攻击
echo "测试3: XSS跨站脚本攻击 (应该被拦截)"
curl -s -o /dev/null -w "HTTP状态码: %{http_code}\n" "$WAF_URL/?search=<script>alert('XSS')</script>"
echo ""

# 测试4: 路径遍历
echo "测试4: 路径遍历攻击 (应该被拦截)"
curl -s -o /dev/null -w "HTTP状态码: %{http_code}\n" "$WAF_URL/files?path=../../../../etc/passwd"
echo ""

# 测试5: 命令注入
echo "测试5: 命令注入攻击 (应该被拦截)"
curl -s -o /dev/null -w "HTTP状态码: %{http_code}\n" "$WAF_URL/ping?host=127.0.0.1|cat /etc/passwd"
echo ""

echo "========================================="
echo "说明:"
echo "- 200/301/302/404: 请求通过"
echo "- 403/406: 请求被WAF拦截"
echo "========================================="
