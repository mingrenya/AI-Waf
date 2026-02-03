#!/usr/bin/env python3
"""
批量更新 tools/ 目录下所有 .go 文件的错误处理
将 fmt.Errorf 替换为更友好的错误处理函数
"""

import os
import re

# 需要处理的文件目录
TOOLS_DIR = "/Users/duheling/Downloads/AI-Waf/mcp-server/tools"

# 错误处理替换规则
REPLACEMENTS = [
    # 查询/获取类操作错误
    (r'fmt\.Errorf\("查询(\w+)失败: %w", err\)', r'WrapError(err, "查询\1")'),
    (r'fmt\.Errorf\("获取(\w+)失败: %w", err\)', r'WrapError(err, "获取\1")'),
    (r'fmt\.Errorf\("创建(\w+)失败: %w", err\)', r'WrapError(err, "创建\1")'),
    (r'fmt\.Errorf\("更新(\w+)失败: %w", err\)', r'WrapError(err, "更新\1")'),
    (r'fmt\.Errorf\("删除(\w+)失败: %w", err\)', r'WrapError(err, "删除\1")'),
    
    # 解析错误
    (r'fmt\.Errorf\("解析(\w+)失败: %w", err\)', r'FormatParseError("\1", err)'),
    (r'fmt\.Errorf\("解析响应失败: %w", err\)', r'FormatParseError("响应", err)'),
    
    # 验证错误 - 保留 NewValidationError,暂不替换（需要手动添加建议）
]

def update_file(filepath):
    """更新单个文件"""
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    original_content = content
    
    # 应用所有替换规则
    for pattern, replacement in REPLACEMENTS:
        content = re.sub(pattern, replacement, content)
    
    # 只有内容变化时才写回
    if content != original_content:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(content)
        return True
    return False

def main():
    """主函数"""
    updated_files = []
    
    # 遍历所有 .go 文件
    for filename in os.listdir(TOOLS_DIR):
        if filename.endswith('.go') and filename not in ['errors.go', 'helpers.go', 'client.go']:
            filepath = os.path.join(TOOLS_DIR, filename)
            if update_file(filepath):
                updated_files.append(filename)
                print(f"✓ Updated: {filename}")
    
    print(f"\n总计更新 {len(updated_files)} 个文件")
    if updated_files:
        print("更新的文件:")
        for f in updated_files:
            print(f"  - {f}")

if __name__ == "__main__":
    main()
