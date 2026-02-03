# AI-WAF MCP Server 评估指南

**创建日期**: 2026-02-03  
**评估文件**: [evaluation.xml](evaluation.xml)

---

## 评估概览

本评估包含 10 个复杂的问题，用于测试 AI-WAF MCP Server 的有效性。这些问题设计用于验证 LLM 是否能够有效使用 MCP 工具来完成真实的 WAF 管理任务。

---

## 评估要求

所有问题均符合 MCP Builder Skill 的评估标准：

- ✅ **只读操作** - 所有问题只需要非破坏性的查询操作
- ✅ **独立性** - 每个问题独立，不依赖其他问题的答案
- ✅ **复杂性** - 需要多次工具调用和深度探索
- ✅ **真实性** - 基于真实的 WAF 管理场景
- ✅ **可验证** - 答案可通过直接字符串比较验证
- ✅ **稳定性** - 答案基于历史数据，不会随时间变化

---

## 问题分类

### 1. 日志分析类（问题 1, 5）

**问题 1**: 统计 SQL 注入攻击最多的源 IP
- **难度**: ⭐⭐⭐
- **涉及工具**: `ai_waf_list_attack_logs`
- **技能**: 分页查询、数据过滤、统计聚合

**问题 5**: AI 攻击模式分析
- **难度**: ⭐⭐⭐⭐
- **涉及工具**: `ai_waf_list_attack_logs`, `ai_waf_analyze_attack_patterns`
- **技能**: 数据采集、AI 分析、结果解析

### 2. 规则管理类（问题 2, 6, 10）

**问题 2**: 查找特定优先级和状态的规则
- **难度**: ⭐⭐⭐
- **涉及工具**: `ai_waf_list_micro_rules`
- **技能**: 多条件筛选、文本匹配

**问题 6**: 黑名单规则中优先级最低的规则
- **难度**: ⭐⭐⭐⭐
- **涉及工具**: `ai_waf_list_micro_rules`
- **技能**: 多条件筛选、条件字段解析、优先级比较

**问题 10**: 导出规则并找出最早创建的
- **难度**: ⭐⭐⭐⭐⭐
- **涉及工具**: `ai_waf_list_micro_rules`, `ai_waf_export_rules`
- **技能**: 数据导出、JSON 解析、时间戳比较

### 3. 封禁 IP 管理类（问题 3）

**问题 3**: 封禁时间最早的恶意 IP
- **难度**: ⭐⭐⭐
- **涉及工具**: `ai_waf_list_blocked_ips`
- **技能**: 分页查询、原因筛选、时间排序

### 4. 监控指标类（问题 4, 7, 9）

**问题 4**: 7天内最高 QPS
- **难度**: ⭐⭐⭐
- **涉及工具**: `ai_waf_get_time_series_data`
- **技能**: 时间序列数据分析、最大值查找

**问题 7**: 系统健康状态判断
- **难度**: ⭐⭐
- **涉及工具**: `ai_waf_get_system_health`
- **技能**: 健康状态阈值理解

**问题 9**: 攻击拦截率计算
- **难度**: ⭐⭐⭐
- **涉及工具**: `ai_waf_get_security_metrics`
- **技能**: 比例计算、格式化输出

### 5. 站点管理类（问题 8）

**问题 8**: 统计启用 WAF 的 .com 域名站点数
- **难度**: ⭐⭐⭐
- **涉及工具**: `ai_waf_list_sites`
- **技能**: 多条件筛选、计数统计

---

## 评估难度分布

| 难度等级 | 问题数量 | 问题编号 |
|---------|---------|---------|
| ⭐⭐ | 1 | 7 |
| ⭐⭐⭐ | 4 | 1, 2, 3, 8, 9 |
| ⭐⭐⭐⭐ | 3 | 5, 6 |
| ⭐⭐⭐⭐⭐ | 2 | 10 |

---

## 使用的 MCP 工具统计

| 工具名称 | 使用次数 | 涉及问题 |
|---------|---------|---------|
| `ai_waf_list_attack_logs` | 2 | 1, 5 |
| `ai_waf_list_micro_rules` | 4 | 2, 6, 10 |
| `ai_waf_list_blocked_ips` | 1 | 3 |
| `ai_waf_get_time_series_data` | 1 | 4 |
| `ai_waf_analyze_attack_patterns` | 1 | 5 |
| `ai_waf_get_system_health` | 1 | 7 |
| `ai_waf_list_sites` | 1 | 8 |
| `ai_waf_get_security_metrics` | 1 | 9 |
| `ai_waf_export_rules` | 1 | 10 |

---

## 工具覆盖率

本评估覆盖了以下工具类别：

- ✅ **日志查询** - `ai_waf_list_attack_logs`, `ai_waf_get_log_stats`
- ✅ **规则管理** - `ai_waf_list_micro_rules`, `ai_waf_export_rules`
- ✅ **封禁 IP** - `ai_waf_list_blocked_ips`
- ✅ **AI 分析** - `ai_waf_analyze_attack_patterns`
- ✅ **监控指标** - `ai_waf_get_time_series_data`, `ai_waf_get_security_metrics`, `ai_waf_get_system_health`
- ✅ **站点管理** - `ai_waf_list_sites`

**未覆盖的工具类别**（可在后续评估中添加）：
- ⬜ 规则创建/更新/删除（写操作，不适合评估）
- ⬜ 批量操作（写操作，不适合评估）
- ⬜ 规则优化和效果评估
- ⬜ 合规性检查
- ⬜ 安全报告生成

---

## 预期工具调用序列示例

### 问题 1: SQL 注入攻击最多的源 IP

```
1. ai_waf_list_attack_logs(
     startTime: "2026-01-01T00:00:00Z",
     endTime: "2026-01-31T23:59:59Z",
     page: 1,
     pageSize: 100
   )
2. [可能需要] ai_waf_list_attack_logs(page: 2, pageSize: 100)
3. [可能需要] ai_waf_list_attack_logs(page: 3, pageSize: 100)
4. 筛选 attackType 包含 "SQL" 或 "Injection"
5. 统计每个 srcIp 的出现次数
6. 返回出现次数最多的 IP
```

### 问题 10: 最早创建的启用规则的创建者

```
1. ai_waf_list_micro_rules(
     page: 1,
     size: 100
   )
2. [分页] ai_waf_list_micro_rules(page: 2, size: 100)
3. 筛选 status="enabled"
4. ai_waf_export_rules(
     ruleIds: [筛选出的规则ID列表],
     format: "json"
   )
5. 解析 JSON 输出
6. 比较每个规则的 createdAt 时间
7. 找出最早的规则
8. 返回其 creator 字段
```

---

## 运行评估

### 前置条件

1. **启动 MCP Server**
   ```bash
   cd /Users/duheling/Downloads/AI-Waf/mcp-server
   go run . --http
   ```

2. **配置认证**
   ```bash
   export AI_WAF_TOKEN=$(./scripts/get-mcp-token.sh)
   export AI_WAF_BASE_URL="http://localhost:8080"
   ```

3. **准备测试数据**
   - 确保系统中有 2026年1月的攻击日志数据
   - 确保有多个不同类型和优先级的 MicroRule 规则
   - 确保有被封禁的 IP 记录
   - 确保有多个站点配置

### 手动验证步骤

对于每个问题：

1. **阅读问题** - 理解问题要求和预期答案格式
2. **规划工具调用** - 确定需要使用哪些工具
3. **执行查询** - 调用 MCP 工具获取数据
4. **处理数据** - 筛选、聚合、计算
5. **验证答案** - 将结果与 evaluation.xml 中的答案对比

### 自动化评估（未来）

可以创建评估脚本自动运行这些问题：

```python
# evaluation_runner.py (未实现)
import xml.etree.ElementTree as ET
from mcp_client import MCPClient

def run_evaluation(evaluation_file):
    tree = ET.parse(evaluation_file)
    results = []
    
    for qa_pair in tree.findall('.//qa_pair'):
        question = qa_pair.find('question').text
        expected_answer = qa_pair.find('answer').text
        
        # 使用 LLM + MCP 工具回答问题
        actual_answer = ask_llm_with_mcp(question)
        
        # 验证答案
        is_correct = (actual_answer.strip() == expected_answer.strip())
        results.append({
            'question': question,
            'expected': expected_answer,
            'actual': actual_answer,
            'correct': is_correct
        })
    
    return results
```

---

## 评估指标

### 成功标准

- **准确率**: ≥ 80% 的问题答案完全正确
- **工具效率**: 平均每个问题 ≤ 20 次工具调用
- **响应时间**: 每个问题 ≤ 60 秒

### 失败分析

如果问题回答不正确，可能原因：

1. **工具描述不清** - 工具的 JSON Schema 或描述需要改进
2. **数据结构复杂** - 返回的数据格式过于复杂，LLM 难以解析
3. **缺少必要工具** - 某些操作需要的工具不存在
4. **错误处理不当** - 错误消息不够友好，无法指导 LLM 恢复
5. **分页支持不足** - 大数据集分页机制不完善

---

## 改进建议

基于评估结果，可以考虑以下改进：

### 1. 工具描述优化
- 在工具描述中添加更多使用示例
- 明确说明字段的可能取值范围
- 提供常见查询模式的建议

### 2. 响应格式优化
- 简化 JSON 响应结构
- 提供 Markdown 格式的可选输出
- 减少不必要的嵌套层级

### 3. 新增辅助工具
- `ai_waf_search_logs` - 语义化搜索攻击日志
- `ai_waf_get_rule_by_condition` - 按条件快速查找规则
- `ai_waf_get_top_attackers` - 直接获取攻击者排行

### 4. 增强现有工具
- 支持更复杂的筛选条件
- 提供聚合统计功能
- 优化大数据集的分页性能

---

## 相关文档

- [evaluation.xml](evaluation.xml) - 评估问题和答案
- [GO_MCP_BEST_PRACTICES.md](GO_MCP_BEST_PRACTICES.md) - Go MCP 最佳实践
- [ERROR_HANDLING_IMPROVEMENTS.md](ERROR_HANDLING_IMPROVEMENTS.md) - 错误处理框架
- [MCP Tools Complete List](MCP_TOOLS_COMPLETE_LIST.md) - 所有可用工具列表

---

## 版本历史

- **v1.0** (2026-02-03) - 初始版本，10 个评估问题
  - 覆盖日志、规则、IP、监控、站点 5 大类
  - 难度分布：简单 1 个，中等 4 个，困难 3 个，极难 2 个
  - 工具覆盖率：9/47 核心工具

---

**下一步**: 在真实 LLM 环境中运行评估，收集结果并分析改进方向。
