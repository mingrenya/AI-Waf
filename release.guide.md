# RuiQi WAF 项目发版完整指南

## 🎯 发版前准备

### 1. 确保代码准备就绪
```bash
# 确保所有更改都已提交到主分支
git checkout main
git pull origin main

# 检查是否有未提交的更改
git status
```

### 2. 更新版本相关文件（可选但推荐）
```bash
# 更新 package.json、README.md、CHANGELOG.md 等文件中的版本号
# 例如在 package.json 中：
{
  "version": "1.2.0"
}
```

### 3. 提交版本更新
```bash
git add .
git commit -m "chore: bump version to v1.2.0"
git push origin main
```

## 🏷️ 创建和推送标签

### 方法一：命令行方式（推荐）

```bash
# 1. 创建标签（推荐使用带注释的标签）
git tag -a v1.2.0 -m "Release version 1.2.0

- 新增功能A
- 修复bug B
- 性能优化C"

# 2. 推送标签到远程仓库（这会触发Release工作流）
git push origin v1.2.0

# 或者推送所有标签
git push origin --tags
```

### 方法二：GitHub Web界面

1. 访问您的GitHub仓库
2. 点击右侧的 **"Releases"**
3. 点击 **"Create a new release"**
4. 在 **"Choose a tag"** 中输入 `v1.2.0`
5. 选择 **"Create new tag: v1.2.0 on publish"**
6. 填写Release标题和描述
7. 点击 **"Publish release"**

### 方法三：GitHub CLI（如果已安装）

```bash
# 安装 GitHub CLI
# macOS: brew install gh
# Ubuntu: sudo apt install gh

# 登录
gh auth login

# 创建release（会自动创建标签）
gh release create v1.2.0 \
  --title "RuiQi WAF v1.2.0" \
  --notes "发布说明内容" \
  --latest
```

## 🔄 发版流程自动化

### 标签推送后会自动发生：

1. **🏷️ 标签检测** - GitHub检测到新的`v*.*.*`格式标签
2. **🚀 触发工作流** - Release workflow自动启动
3. **🏗️ Docker构建** - 构建多平台Docker镜像
4. **📦 镜像推送** - 推送到Docker Hub
5. **📝 生成Changelog** - 自动生成更改日志
6. **🎉 创建Release** - 在GitHub创建正式Release

### 自动生成的内容：

- **Docker镜像标签：**
  - `username/ruiqi-waf:v1.2.0`
  - `username/ruiqi-waf:1.2.0` (不带v前缀)
  - `username/ruiqi-waf:latest`

- **Release页面包含：**
  - 完整的发布说明
  - Docker拉取命令
  - 变更日志
  - 支持的平台信息

## 📋 版本号规范（语义化版本）

```
v主版本.次版本.修订版本[-预发布版本]

示例：
v1.0.0      # 正式版本
v1.1.0      # 新增功能
v1.1.1      # 修复bug
v2.0.0      # 重大更新
v1.0.0-beta.1   # 测试版本
v1.0.0-rc.1     # 候选版本
```

### 版本号含义：
- **主版本号**：不兼容的API修改
- **次版本号**：向下兼容的功能性新增
- **修订版本号**：向下兼容的问题修正

## 🛠️ 实际操作示例

### 场景1：发布新功能版本 v1.2.0

```bash
# 1. 确保在主分支且代码最新
git checkout main
git pull origin main

# 2. 创建带注释的标签
git tag -a v1.2.0 -m "Release v1.2.0

新功能：
- 添加了新的WAF规则管理界面
- 支持自定义安全策略配置
- 新增API接口用于批量操作

改进：
- 优化了HAProxy配置生成逻辑
- 提升了Web界面响应速度
- 更新了文档和示例

修复：
- 修复了在某些环境下的配置同步问题
- 解决了Docker容器启动时的权限问题"

# 3. 推送标签（触发自动发版）
git push origin v1.2.0

# 4. 等待GitHub Actions完成构建和发布
```

### 场景2：紧急修复版本 v1.1.1

```bash
# 1. 基于当前release分支或主分支修复
git checkout main
# 进行bug修复...
git add .
git commit -m "fix: 修复安全规则解析错误"
git push origin main

# 2. 创建修复版本标签
git tag -a v1.1.1 -m "Release v1.1.1 - 紧急修复

修复：
- 修复了安全规则解析导致的服务启动失败问题
- 解决了特定情况下的内存泄漏问题"

# 3. 推送标签
git push origin v1.1.1
```

## 🔍 监控发版进度

### 在GitHub界面查看：

1. **Actions页面** - 查看工作流运行状态
   - 访问 `https://github.com/your-username/RuiQi/actions`
   - 找到对应的Release工作流

2. **Releases页面** - 查看发布结果
   - 访问 `https://github.com/your-username/RuiQi/releases`

3. **Docker Hub** - 确认镜像推送
   - 访问您的Docker Hub仓库页面

### 命令行检查：

```bash
# 查看所有标签
git tag -l

# 查看特定标签信息
git show v1.2.0

# 检查Docker镜像
docker pull your-username/ruiqi-waf:v1.2.0
docker images | grep ruiqi-waf
```

## ⚠️ 注意事项和故障排除

### 常见问题：

1. **权限错误**
   ```
   解决方案：确保仓库的Actions权限设置正确
   Settings > Actions > General > Workflow permissions
   ```

2. **Docker Hub登录失败**
   ```
   解决方案：检查 DOCKERHUB_USERNAME 和 DOCKERHUB_TOKEN
   Settings > Secrets and variables > Actions
   ```

3. **标签格式错误**
   ```
   正确格式：v1.2.0 (必须以v开头)
   错误格式：1.2.0, version-1.2.0
   ```

### 回滚操作：

```bash
# 删除错误的标签（本地）
git tag -d v1.2.0

# 删除远程标签（谨慎操作）
git push origin :refs/tags/v1.2.0

# 重新创建正确的标签
git tag -a v1.2.0 -m "正确的发布信息"
git push origin v1.2.0
```

## 📚 最佳实践

1. **发版前测试**
   - 在开发分支充分测试
   - 运行所有自动化测试
   - 手动验证关键功能

2. **标签信息详细**
   - 使用带注释的标签（`git tag -a`）
   - 包含详细的更改说明
   - 遵循一致的格式

3. **版本计划**
   - 制定版本发布计划
   - 维护CHANGELOG.md文件
   - 定期发布，避免功能堆积

4. **回归测试**
   - 发版后验证Docker镜像功能
   - 测试升级路径
   - 监控生产环境表现

## 🎯 自动化建议

可以创建额外的脚本来简化发版流程：

```bash
#!/bin/bash
# release.sh - 发版自动化脚本

VERSION=$1
if [ -z "$VERSION" ]; then
    echo "用法: ./release.sh v1.2.0"
    exit 1
fi

echo "准备发布 $VERSION..."

# 确保代码最新
git checkout main
git pull origin main

# 创建标签
git tag -a $VERSION -m "Release $VERSION"

# 推送标签
git push origin $VERSION

echo "✅ 发版完成！请查看 GitHub Actions 进度。"
```

使用方式：
```bash
chmod +x release.sh
./release.sh v1.2.0
```