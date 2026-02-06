const fs = require('fs');
const file = 'src/feature/ai-assistant/components/AIAssistantDialog.tsx';
let content = fs.readFileSync(file, 'utf8');

// 1. 简化 isChatMessageResponse 函数定义
const oldChecker = `const isChatMessageResponse = (v: unknown): v is ChatMessageResponse => {
        if (!isObject(v)) return false
        const msg = v['message']
        const ts = v['timestamp']
        const tc = v['toolCalls']
        return typeof msg === 'string' && (typeof ts === 'string' || ts === undefined) && (tc === undefined || Array.isArray(tc))
      }`;

const newChecker = `const isChatMessageResponse = (v: unknown): v is ChatMessageResponse => {
        if (!isObject(v)) return false
        // 只要有 message 字段是字符串即可
        return typeof v['message'] === 'string'
      }`;

content = content.replace(oldChecker, newChecker);

// 2. 添加调试日志
if (!content.includes('🔍 AI Chat 原始响应')) {
  content = content.replace(
    '      const resAny: unknown = response',
    `      // 添加调试日志
      if (import.meta.env.DEV) {
        console.log('🔍 AI Chat 原始响应:', response)
        console.log('🔍 响应类型:', typeof response)
        if (typeof response === 'object' && response !== null) {
          console.log('🔍 响应键:', Object.keys(response))
        }
      }

      const resAny: unknown = response`
  );
}

// 3. 在提取消息后添加调试
if (!content.includes('💬 最终提取的AI消息')) {
  const oldExtract = `      const aiMessage = chatData?.message || '无响应'
      const aiToolCalls = chatData?.toolCalls || []`;

  const newExtract = `      const aiMessage = chatData?.message || '无响应'
      const aiToolCalls = chatData?.toolCalls || []

      if (import.meta.env.DEV) {
        console.log('💬 最终提取的AI消息:', aiMessage)
        console.log('🔧 工具调用:', aiToolCalls)
      }`;

  content = content.replace(oldExtract, newExtract);
}

fs.writeFileSync(file, content, 'utf8');
console.log('✅ 已添加调试日志到 AIAssistantDialog.tsx');
console.log('📝 已简化类型检查逻辑');
