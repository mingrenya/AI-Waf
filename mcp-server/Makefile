.PHONY: build run clean test

# 编译
build:
	go build -o ai-waf-mcp .

# 运行（需要先设置环境变量）
run:
	go run main.go

# 清理
clean:
	rm -f ai-waf-mcp

# 测试
test:
	go test ./... -v

# 安装依赖
deps:
	go mod download

# 更新依赖
update:
	go get -u ./...
	go mod tidy

# 生成二进制文件（多平台）
build-all:
	GOOS=darwin GOARCH=amd64 go build -o bin/ai-waf-mcp-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o bin/ai-waf-mcp-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build -o bin/ai-waf-mcp-linux-amd64 .
	GOOS=windows GOARCH=amd64 go build -o bin/ai-waf-mcp-windows-amd64.exe .
