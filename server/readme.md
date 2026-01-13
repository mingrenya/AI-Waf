swagger 文档地址：http://localhost:2333/swagger/index.html

生成 swagger 文档：

`go install github.com/swaggo/swag/cmd/swag@latest`
`swag init --dir ./,../pkg`
`swag fmt --dir ./,../pkg`

format 格式化
`gofmt -w .`
