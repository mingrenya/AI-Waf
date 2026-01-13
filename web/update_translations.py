#!/usr/bin/env python3
import json
import sys

def update_translations():
    # 英文翻译
    en_file = 'public/locales/en/translation.json'
    zh_file = 'public/locales/zh/translation.json'
    
    alert_en = {
        "management": "Alert Management",
        "channels": "Alert Channels",
        "rules": "Alert Rules",
        "history": "Alert History",
        "channelManagement": "Channel Management",
        "ruleManagement": "Rule Management",
        "historyManagement": "Alert History",
        "addChannel": "Add Channel",
        "addRule": "Add Rule",
        "channelName": "Channel Name",
        "channelType": "Channel Type",
        "ruleName": "Rule Name",
        "severity": "Severity",
        "conditions": "Conditions",
        "enabled": "Enabled",
        "disabled": "Disabled",
        "status": "Status",
        "testChannel": "Test Channel",
        "testChannelSuccess": "Test message sent successfully",
        "testChannelFailed": "Failed to send test message",
        "testChannelDescription": "Send a test message to verify channel configuration",
        "testMessage": "Test Message",
        "testMessagePlaceholder": "Enter your test message...",
        "sendTest": "Send Test",
        "deleteChannelTitle": "Delete Channel",
        "deleteChannelDescription": "Are you sure you want to delete this channel? This action cannot be undone.",
        "deleteChannelSuccess": "Channel deleted successfully",
        "deleteChannelFailed": "Failed to delete channel",
        "deleteRuleTitle": "Delete Rule",
        "deleteRuleDescription": "Are you sure you want to delete this alert rule?",
        "deleteRuleSuccess": "Alert rule deleted successfully",
        "deleteRuleFailed": "Failed to delete alert rule",
        "updateSuccess": "Updated successfully",
        "updateFailed": "Update failed",
        "channelType": {
            "webhook": "Webhook",
            "slack": "Slack",
            "discord": "Discord",
            "dingtalk": "DingTalk",
            "wecom": "Enterprise WeChat"
        },
        "severity": {
            "low": "Low",
            "medium": "Medium",
            "high": "High",
            "critical": "Critical"
        },
        "dialog": {
            "createChannelTitle": "Create Alert Channel",
            "createChannelDescription": "Configure a new notification channel for alert messages",
            "updateChannelTitle": "Update Alert Channel",
            "updateChannelDescription": "Modify alert channel configuration",
            "createRuleTitle": "Create Alert Rule",
            "createRuleDescription": "Configure a new alert rule with conditions and channels",
            "updateRuleTitle": "Update Alert Rule",
            "updateRuleDescription": "Modify alert rule configuration"
        }
    }
    
    alert_zh = {
        "management": "告警管理",
        "channels": "告警通道",
        "rules": "告警规则",
        "history": "告警历史",
        "channelManagement": "通道管理",
        "ruleManagement": "规则管理",
        "historyManagement": "告警历史",
        "addChannel": "添加通道",
        "addRule": "添加规则",
        "channelName": "通道名称",
        "channelType": "通道类型",
        "ruleName": "规则名称",
        "severity": "严重程度",
        "conditions": "触发条件",
        "enabled": "已启用",
        "disabled": "已禁用",
        "status": "状态",
        "testChannel": "测试通道",
        "testChannelSuccess": "测试消息发送成功",
        "testChannelFailed": "测试消息发送失败",
        "testChannelDescription": "发送测试消息以验证通道配置",
        "testMessage": "测试消息",
        "testMessagePlaceholder": "输入测试消息内容...",
        "sendTest": "发送测试",
        "deleteChannelTitle": "删除通道",
        "deleteChannelDescription": "确定要删除此通道吗?此操作无法撤销。",
        "deleteChannelSuccess": "通道删除成功",
        "deleteChannelFailed": "通道删除失败",
        "deleteRuleTitle": "删除规则",
        "deleteRuleDescription": "确定要删除此告警规则吗?",
        "deleteRuleSuccess": "告警规则删除成功",
        "deleteRuleFailed": "告警规则删除失败",
        "updateSuccess": "更新成功",
        "updateFailed": "更新失败",
        "channelType": {
            "webhook": "Webhook",
            "slack": "Slack",
            "discord": "Discord",
            "dingtalk": "钉钉",
            "wecom": "企业微信"
        },
        "severity": {
            "low": "低",
            "medium": "中",
            "high": "高",
            "critical": "紧急"
        },
        "dialog": {
            "createChannelTitle": "创建告警通道",
            "createChannelDescription": "配置新的通知渠道以接收告警消息",
            "updateChannelTitle": "更新告警通道",
            "updateChannelDescription": "修改告警通道配置",
            "createRuleTitle": "创建告警规则",
            "createRuleDescription": "配置新的告警规则和触发条件",
            "updateRuleTitle": "更新告警规则",
            "updateRuleDescription": "修改告警规则配置"
        }
    }
    
    try:
        # 更新英文翻译
        with open(en_file, 'r', encoding='utf-8') as f:
            en_data = json.load(f)
        en_data['alert'] = alert_en
        en_data['sidebar']['alerts'] = 'Alerts'
        en_data['breadcrumb']['alerts'] = {
            "channel": "Channels",
            "rule": "Rules",
            "history": "History"
        }
        with open(en_file, 'w', encoding='utf-8') as f:
            json.dump(en_data, f, indent=4, ensure_ascii=False)
        print(f"✓ {en_file} 更新成功")
        
        # 更新中文翻译
        with open(zh_file, 'r', encoding='utf-8') as f:
            zh_data = json.load(f)
        zh_data['alert'] = alert_zh
        zh_data['sidebar']['alerts'] = '告警'
        zh_data['breadcrumb']['alerts'] = {
            "channel": "通道",
            "rule": "规则",
            "history": "历史"
        }
        with open(zh_file, 'w', encoding='utf-8') as f:
            json.dump(zh_data, f, indent=4, ensure_ascii=False)
        print(f"✓ {zh_file} 更新成功")
        
        return 0
    except Exception as e:
        print(f"✗ 更新失败: {e}", file=sys.stderr)
        return 1

if __name__ == '__main__':
    sys.exit(update_translations())
