#!/usr/bin/env python3
"""
修复 MCP Server 工具文件中的 jsonschema 标签格式
将 'description=xxx,minimum=yyy' 格式改为只保留描述文本
"""

import re
import glob
import os

def fix_jsonschema_tag(match):
    """修复单个 jsonschema 标签"""
    full_tag = match.group(1)
    
    # 提取 description 的值
    desc_match = re.search(r'description=([^,]+)', full_tag)
    if desc_match:
        description = desc_match.group(1)
        # 移除可能的引号
        description = description.strip('"\'')
        return f'jsonschema:"{description}"'
    
    # 如果没有 description,返回原样
    return f'jsonschema:"{full_tag}"'

def fix_file(filepath):
    """修复单个文件"""
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
        
        original_content = content
        
        # 匹配 jsonschema:"description=..." 格式
        # 注意:需要处理可能包含逗号的复杂情况
        pattern = r'jsonschema:"([^"]+)"'
        
        def replace_func(match):
            tag_content = match.group(1)
            # 只处理包含 description= 的标签
            if 'description=' in tag_content:
                # 提取 description 的值(到第一个逗号或结尾)
                desc_match = re.match(r'description=([^,]+)', tag_content)
                if desc_match:
                    description = desc_match.group(1)
                    return f'jsonschema:"{description}"'
            return match.group(0)  # 保持原样
        
        content = re.sub(pattern, replace_func, content)
        
        if content != original_content:
            with open(filepath, 'w', encoding='utf-8') as f:
                f.write(content)
            return True, filepath
        return False, filepath
    except Exception as e:
        return False, f"{filepath}: {e}"

def main():
    """主函数"""
    tools_dir = "/Users/duheling/Downloads/AI-Waf/mcp-server/tools"
    go_files = glob.glob(os.path.join(tools_dir, "*.go"))
    
    modified_count = 0
    total_count = 0
    
    print(f"正在扫描 {len(go_files)} 个 Go 文件...")
    
    for filepath in sorted(go_files):
        total_count += 1
        modified, info = fix_file(filepath)
        if modified:
            modified_count += 1
            filename = os.path.basename(filepath)
            print(f"✅ 修复: {filename}")
    
    print(f"\n修复完成:")
    print(f"  总文件数: {total_count}")
    print(f"  已修复: {modified_count}")
    print(f"  未修改: {total_count - modified_count}")

if __name__ == "__main__":
    main()
