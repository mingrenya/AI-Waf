# ⚠️ 重要警告 - main.go 文件被覆盖问题

## 问题描述

**用户运行了错误的命令**：
```bash
go build -o main.go  # ❌ 错误！这会用二进制文件覆盖源代码
```

这导致 main.go 源代码文件被编译后的二进制文件覆盖，出现 `unexpected NUL in input` 错误。

## 已修复

✅ 已从 git 恢复 main.go 源代码：
```bash
git checkout server/main.go
```

## 正确的命令

### 本地开发编译
```bash
cd server
go build -o ai-waf-server main.go  # ✅ 正确
```

### Docker 构建
```bash
cd server && go build -o ../mrya-waf main.go  # ✅ 正确（Dockerfile 中使用）
```

## 错误原因

`-o` 参数指定**输出文件名**，不应该是源代码文件本身：
- ❌ `go build -o main.go` - 会覆盖源文件！
- ✅ `go build -o <binary-name> main.go` - 正确
- ✅ `go build main.go` - 默认输出到当前目录（不推荐在 server 目录）

## 预防措施

已更新 `server/.gitignore` 以忽略常见的二进制文件名：
```
ai-waf-server
mrya-waf
test-build
```

但这**不能**完全防止覆盖 main.go，请务必小心使用 `-o` 参数！

## 快速验证

检查 main.go 是否是文本文件：
```bash
head -5 server/main.go
# 应该看到：package main
```

如果看到乱码或错误，说明被覆盖了，恢复方法：
```bash
git checkout server/main.go
```

## 教训

1. **永远不要** 使用 `-o <源文件名>` 编译
2. 使用有意义的二进制文件名（如 `ai-waf-server`）
3. 在运行构建命令前仔细检查
4. 保持 git 状态干净，便于恢复
