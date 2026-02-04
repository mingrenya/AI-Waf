#!/bin/bash

# WAF流量检测测试脚本
# 测试WAF是否能够正确探测和拦截各种攻击流量

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# WAF服务地址
WAF_HOST="${WAF_HOST:-localhost}"
WAF_PORT="${WAF_PORT:-2333}"
WAF_URL="http://${WAF_HOST}:${WAF_PORT}"
API_URL="http://${WAF_HOST}:${WAF_PORT}"

# 测试计数器
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 打印分隔线
print_separator() {
    echo -e "${BLUE}========================================${NC}"
}

# 打印测试标题
print_test_title() {
    echo ""
    print_separator
    echo -e "${YELLOW}测试 #$TOTAL_TESTS: $1${NC}"
    print_separator
}

# 打印测试结果
print_result() {
    local status=$1
    local message=$2
    
    if [ "$status" == "PASS" ]; then
        echo -e "${GREEN}✓ PASS${NC}: $message"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}✗ FAIL${NC}: $message"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

# 执行HTTP请求并检查响应
test_request() {
    local description="$1"
    local url="$2"
    local method="${3:-GET}"
    local data="$4"
    local expected_block="${5:-false}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    print_test_title "$description"
    
    echo "请求URL: $url"
    echo "请求方法: $method"
    if [ -n "$data" ]; then
        echo "请求数据: $data"
    fi
    echo "预期结果: $([ "$expected_block" == "true" ] && echo "被拦截" || echo "正常通过")"
    echo ""
    
    # 执行请求
    if [ "$method" == "POST" ]; then
        response=$(curl -s -w "\n%{http_code}" -X POST \
            -H "Content-Type: application/json" \
            -H "User-Agent: WAF-Test-Script/1.0" \
            -d "$data" \
            "$url" 2>&1)
    else
        response=$(curl -s -w "\n%{http_code}" \
            -H "User-Agent: WAF-Test-Script/1.0" \
            "$url" 2>&1)
    fi
    
    # 提取HTTP状态码
    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | sed '$d')
    
    echo "HTTP状态码: $http_code"
    echo "响应内容: ${response_body:0:200}"
    
    # 判断是否被拦截 (403 Forbidden 或 406 Not Acceptable)
    if [ "$expected_block" == "true" ]; then
        if [ "$http_code" == "403" ] || [ "$http_code" == "406" ] || echo "$response_body" | grep -qi "blocked\|forbidden\|rejected"; then
            print_result "PASS" "恶意请求已被成功拦截 (HTTP $http_code)"
        else
            print_result "FAIL" "恶意请求未被拦截 (HTTP $http_code)"
        fi
    else
        if [ "$http_code" == "200" ] || [ "$http_code" == "301" ] || [ "$http_code" == "302" ] || [ "$http_code" == "404" ]; then
            print_result "PASS" "正常请求通过 (HTTP $http_code)"
        else
            print_result "FAIL" "正常请求被误拦截 (HTTP $http_code)"
        fi
    fi
}

# 主测试函数
main() {
    echo -e "${BLUE}"
    echo "╔════════════════════════════════════════╗"
    echo "║   WAF 流量检测测试                     ║"
    echo "║   Testing WAF Traffic Detection        ║"
    echo "╚════════════════════════════════════════╝"
    echo -e "${NC}"
    
    echo "测试目标: $WAF_URL"
    echo "API地址: $API_URL"
    echo ""
    
    # 检查WAF服务是否可访问
    echo "检查WAF服务状态..."
    if ! curl -s -o /dev/null -w "%{http_code}" --max-time 5 "$WAF_URL" > /dev/null 2>&1; then
        echo -e "${RED}错误: 无法连接到WAF服务 $WAF_URL${NC}"
        echo "请确认服务已启动: docker-compose ps"
        exit 1
    fi
    echo -e "${GREEN}WAF服务正常运行${NC}"
    echo ""
    
    sleep 1
    
    # ==================== 正常流量测试 ====================
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}第一部分: 正常流量测试${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    test_request \
        "正常GET请求 - 访问首页" \
        "$WAF_URL/" \
        "GET" \
        "" \
        "false"
    
    test_request \
        "正常GET请求 - 访问API健康检查" \
        "$API_URL/health" \
        "GET" \
        "" \
        "false"
    
    test_request \
        "正常POST请求 - 合法JSON数据" \
        "$API_URL/test" \
        "POST" \
        '{"username":"testuser","email":"test@example.com"}' \
        "false"
    
    # ==================== SQL注入攻击测试 ====================
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}第二部分: SQL注入攻击测试${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    test_request \
        "SQL注入 - UNION SELECT攻击" \
        "$WAF_URL/?id=1' UNION SELECT * FROM users--" \
        "GET" \
        "" \
        "true"
    
    test_request \
        "SQL注入 - OR 1=1 攻击" \
        "$WAF_URL/login?user=admin' OR '1'='1" \
        "GET" \
        "" \
        "true"
    
    test_request \
        "SQL注入 - 注释绕过" \
        "$WAF_URL/?id=1';DROP TABLE users;--" \
        "GET" \
        "" \
        "true"
    
    test_request \
        "SQL注入 - POST请求中的注入" \
        "$API_URL/login" \
        "POST" \
        '{"username":"admin\" OR \"1\"=\"1","password":"pass"}' \
        "true"
    
    # ==================== XSS跨站脚本攻击测试 ====================
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}第三部分: XSS跨站脚本攻击测试${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    test_request \
        "XSS攻击 - 基本script标签" \
        "$WAF_URL/?search=<script>alert('XSS')</script>" \
        "GET" \
        "" \
        "true"
    
    test_request \
        "XSS攻击 - img标签onerror" \
        "$WAF_URL/?comment=<img src=x onerror=alert('XSS')>" \
        "GET" \
        "" \
        "true"
    
    test_request \
        "XSS攻击 - JavaScript协议" \
        "$WAF_URL/?url=javascript:alert('XSS')" \
        "GET" \
        "" \
        "true"
    
    test_request \
        "XSS攻击 - 事件处理器" \
        "$WAF_URL/?name=<body onload=alert('XSS')>" \
        "GET" \
        "" \
        "true"
    
    # ==================== 路径遍历攻击测试 ====================
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}第四部分: 路径遍历攻击测试${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    test_request \
        "路径遍历 - 访问/etc/passwd" \
        "$WAF_URL/files?path=../../../../etc/passwd" \
        "GET" \
        "" \
        "true"
    
    test_request \
        "路径遍历 - Windows路径" \
        "$WAF_URL/download?file=..\\..\\..\\windows\\system32\\config\\sam" \
        "GET" \
        "" \
        "true"
    
    test_request \
        "路径遍历 - URL编码绕过" \
        "$WAF_URL/view?page=%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd" \
        "GET" \
        "" \
        "true"
    
    # ==================== 命令注入攻击测试 ====================
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}第五部分: 命令注入攻击测试${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    test_request \
        "命令注入 - 管道符号" \
        "$WAF_URL/ping?host=127.0.0.1 | cat /etc/passwd" \
        "GET" \
        "" \
        "true"
    
    test_request \
        "命令注入 - 分号分隔" \
        "$WAF_URL/execute?cmd=ls;cat /etc/shadow" \
        "GET" \
        "" \
        "true"
    
    test_request \
        "命令注入 - 反引号命令替换" \
        "$WAF_URL/?input=\`whoami\`" \
        "GET" \
        "" \
        "true"
    
    # ==================== LDAP注入攻击测试 ====================
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}第六部分: LDAP注入攻击测试${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    test_request \
        "LDAP注入 - 通配符绕过" \
        "$WAF_URL/ldap?user=*)(&(password=*" \
        "GET" \
        "" \
        "true"
    
    # ==================== XXE攻击测试 ====================
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}第七部分: XXE (XML外部实体)攻击测试${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    test_request \
        "XXE攻击 - 读取本地文件" \
        "$API_URL/xml" \
        "POST" \
        '<?xml version="1.0"?><!DOCTYPE foo [<!ENTITY xxe SYSTEM "file:///etc/passwd">]><foo>&xxe;</foo>' \
        "true"
    
    # ==================== 扫描器检测测试 ====================
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}第八部分: 扫描器和爬虫检测测试${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    # 使用常见扫描器User-Agent
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    print_test_title "扫描器检测 - Nikto扫描器"
    response=$(curl -s -w "\n%{http_code}" \
        -H "User-Agent: Nikto/2.1.6" \
        "$WAF_URL/" 2>&1)
    http_code=$(echo "$response" | tail -n1)
    if [ "$http_code" == "403" ] || [ "$http_code" == "406" ]; then
        print_result "PASS" "扫描器被成功识别和拦截"
    else
        print_result "FAIL" "扫描器未被检测"
    fi
    
    # ==================== 测试总结 ====================
    echo ""
    echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║           测试结果总结                 ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
    echo ""
    echo "总测试数: $TOTAL_TESTS"
    echo -e "${GREEN}通过: $PASSED_TESTS${NC}"
    echo -e "${RED}失败: $FAILED_TESTS${NC}"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo ""
        echo -e "${GREEN}✓✓✓ 所有测试通过! WAF工作正常 ✓✓✓${NC}"
        exit 0
    else
        echo ""
        echo -e "${YELLOW}⚠ 部分测试失败,请检查WAF配置 ⚠${NC}"
        exit 1
    fi
}

# 运行主函数
main
