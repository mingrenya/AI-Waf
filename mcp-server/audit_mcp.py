#!/usr/bin/env python3
"""
MCP 工具审计脚本
检查工具注册、函数定义和结构体的一致性
"""

import re
import os
from pathlib import Path

def extract_registered_tools(main_go_path):
    """从 main.go 提取所有注册的工具名称"""
    with open(main_go_path, 'r') as f:
        content = f.read()
    
    # 匹配 Name: "ai_waf_xxx"
    pattern = r'Name:\s*"(ai_waf_\w+)"'
    tools = re.findall(pattern, content)
    return sorted(set(tools))

def tool_to_func_name(tool_name):
    """将工具名转换为函数名: ai_waf_list_attack_logs -> CreateListAttackLogs"""
    # 移除前缀 ai_waf_
    name = tool_name.replace('ai_waf_', '')
    # 转换为驼峰命名，特殊处理缩写词
    parts = name.split('_')
    camel_parts = []
    for word in parts:
        # 特殊缩写词保持全大写
        if word.upper() in ['IP', 'IPS', 'QPS', 'AI', 'WAF', 'ID', 'URL', 'API']:
            camel_parts.append(word.upper())
        else:
            camel_parts.append(word.capitalize())
    return f'Create{"".join(camel_parts)}'

def tool_to_struct_name(tool_name, suffix):
    """将工具名转换为结构体名: ai_waf_list_attack_logs -> ListAttackLogsInput/Output"""
    name = tool_name.replace('ai_waf_', '')
    parts = name.split('_')
    camel_parts = []
    for word in parts:
        # 特殊缩写词保持全大写
        if word.upper() in ['IP', 'IPS', 'QPS', 'AI', 'WAF', 'ID', 'URL', 'API']:
            camel_parts.append(word.upper())
        else:
            camel_parts.append(word.capitalize())
    return f'{"".join(camel_parts)}{suffix}'

def search_in_files(pattern, directory):
    """在目录中的所有 .go 文件中搜索模式"""
    for go_file in Path(directory).glob('*.go'):
        with open(go_file, 'r') as f:
            if re.search(pattern, f.read()):
                return True
    return False

def audit_mcp_tools():
    """执行审计"""
    print("=" * 60)
    print("MCP 工具审计报告")
    print("=" * 60)
    print()
    
    # 1. 提取注册的工具
    main_go = Path(__file__).parent / 'main.go'
    tools_dir = Path(__file__).parent / 'tools'
    
    registered_tools = extract_registered_tools(main_go)
    print(f"1. 在 main.go 中找到 {len(registered_tools)} 个注册的工具")
    print()
    
    # 2. 检查函数定义
    print("2. 检查工具创建函数...")
    missing_functions = []
    for tool in registered_tools:
        func_name = tool_to_func_name(tool)
        pattern = f'func {func_name}\\('
        
        if not search_in_files(pattern, tools_dir):
            missing_functions.append((tool, func_name))
    
    if missing_functions:
        print(f"   ⚠️  发现 {len(missing_functions)} 个缺失的函数:")
        for tool, func in missing_functions:
            print(f"      ❌ {tool} -> {func}")
    else:
        print("   ✅ 所有工具创建函数都存在")
    print()
    
    # 3. 检查Input结构体
    print("3. 检查输入结构体定义...")
    missing_inputs = []
    for tool in registered_tools:
        struct_name = tool_to_struct_name(tool, 'Input')
        pattern = f'type {struct_name} struct'
        
        if not search_in_files(pattern, tools_dir):
            missing_inputs.append((tool, struct_name))
    
    if missing_inputs:
        print(f"   ⚠️  发现 {len(missing_inputs)} 个缺失的输入结构体:")
        for tool, struct in missing_inputs:
            print(f"      ❌ {tool} -> {struct}")
    else:
        print("   ✅ 所有输入结构体都存在")
    print()
    
    # 4. 检查Output结构体
    print("4. 检查输出结构体定义...")
    missing_outputs = []
    for tool in registered_tools:
        struct_name = tool_to_struct_name(tool, 'Output')
        pattern = f'type {struct_name} struct'
        
        if not search_in_files(pattern, tools_dir):
            missing_outputs.append((tool, struct_name))
    
    if missing_outputs:
        print(f"   ⚠️  发现 {len(missing_outputs)} 个缺失的输出结构体:")
        for tool, struct in missing_outputs:
            print(f"      ❌ {tool} -> {struct}")
    else:
        print("   ✅ 所有输出结构体都存在")
    print()
    
    # 5. 检查API路径一致性
    print("5. 检查API路径使用...")
    api_paths = {}
    for go_file in tools_dir.glob('*.go'):
        with open(go_file, 'r') as f:
            content = f.read()
            # 查找API路径: Get("/api/v1/xxx") 或 Post("/api/v1/xxx")
            paths = re.findall(r'(?:Get|Post|Put|Delete)(?:WithContext)?\("(/api/[^"]+)"', content)
            for path in paths:
                api_paths[path] = api_paths.get(path, 0) + 1
    
    print(f"   找到 {len(api_paths)} 个不同的API端点:")
    for path, count in sorted(api_paths.items()):
        print(f"      {path} (使用 {count} 次)")
    print()
    
    # 总结
    print("=" * 60)
    print("审计总结")
    print("=" * 60)
    print(f"注册工具数: {len(registered_tools)}")
    print(f"缺失函数数: {len(missing_functions)}")
    print(f"缺失输入结构体: {len(missing_inputs)}")
    print(f"缺失输出结构体: {len(missing_outputs)}")
    print(f"API端点数: {len(api_paths)}")
    print()
    
    if not (missing_functions or missing_inputs or missing_outputs):
        print("✅ 审计通过: 所有工具注册正确，接口定义完整")
        return 0
    else:
        print("⚠️  审计发现问题，请检查上述缺失项")
        return 1

if __name__ == '__main__':
    exit(audit_mcp_tools())
