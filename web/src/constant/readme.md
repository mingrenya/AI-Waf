```
// 使用示例
import {
  getConstant,
  setConstant,
  hasConstant,
  getCategoryConstants,
  setBatchConstants,
  ConstantCategory
} from '@/constant';

// 获取常量
const apiTimeout = getConstant(ConstantCategory.SYSTEM, 'API_TIMEOUT', 5000);
const pageSize = getConstant(ConstantCategory.CONFIG, 'DEFAULT_PAGE_SIZE', 10);

// 设置常量
setConstant(ConstantCategory.UI, 'SIDEBAR_WIDTH', 250);

// 检查常量是否存在
if (hasConstant(ConstantCategory.FEATURE, 'WAF_MODES')) {
  // 使用常量...
}

// 获取某个分类下的所有常量
const allUiConstants = getCategoryConstants(ConstantCategory.UI);

// 批量设置常量
setBatchConstants(ConstantCategory.CONFIG, {
  'MAX_FILE_SIZE': 10485760, // 10MB
  'ALLOWED_FILE_TYPES': 'jpg,png,pdf',
  'MAX_UPLOAD_FILES': 5
});
```
