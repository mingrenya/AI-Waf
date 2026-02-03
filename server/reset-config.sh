#!/bin/bash

# 重置 AI-WAF 配置脚本

echo "================================================"
echo "   AI-WAF 配置重置工具"
echo "================================================"
echo ""

echo "此脚本将重置 MongoDB 中的配置数据，使用项目的最新默认配置"
echo ""
read -p "确定要继续吗? (y/N): " confirm

if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
    echo "操作已取消"
    exit 0
fi

echo ""
echo "连接到 MongoDB..."

# 获取 MongoDB 连接信息
DB_URI=${DB_URI:-"mongodb://root:example@localhost:27017"}

# 删除现有配置
echo "删除旧配置..."
mongosh "$DB_URI/waf" --quiet --eval '
db.configs.deleteMany({});
print("✓ 已删除所有配置文档");

// 统计剩余文档
var count = db.configs.countDocuments({});
print("当前配置文档数量:", count);
'

if [ $? -eq 0 ]; then
    echo ""
    echo "✓ 配置已重置"
    echo ""
    echo "现在重新启动服务器以创建新的配置..."
else
    echo ""
    echo "❌ 配置重置失败"
    echo "请检查 MongoDB 连接是否正常"
    exit 1
fi
