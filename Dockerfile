# 多阶段构建
# 阶段1: 构建Node.js前端
# 生产环境使用固定版本
FROM node:23.10.0-alpine AS frontend-builder
# 安装pnpm（使用最新稳定版本）
RUN npm install -g pnpm@10.28.0
# 设置环境变量禁用更新检查
ENV PNPM_CHECK_VERSION=false
# 设置工作目录
WORKDIR /app
# 复制前端项目文件
COPY web/ ./
# 安装依赖并构建前端
RUN pnpm install
RUN pnpm build

# 阶段2: 构建Go后端
FROM golang:1.24.1-alpine AS backend-builder
# 设置Go环境变量
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64
# 设置工作目录
WORKDIR /build
# 复制整个项目结构
COPY coraza-spoa/ ./coraza-spoa/
COPY pkg/ ./pkg/
COPY server/ ./server/
COPY mcp-server/ ./mcp-server/
COPY go.work ./
COPY geo-ip/ ./geo-ip/
# 复制前端构建产物到正确位置
COPY --from=frontend-builder /app/dist ./server/public/dist
# 使用Go的工作区功能进行构建
RUN go work use ./coraza-spoa ./pkg ./server ./mcp-server
RUN cd server && go build -o ../mrya-waf main.go

# 阶段3: 最终镜像 - 使用官方 HAProxy 3.0.10 镜像
FROM haproxy:3.0.10

# 确保以root用户进行初始化设置
USER root

# 安装Linux capabilities管理工具
RUN apt-get update && apt-get install -y libcap2-bin && \
    rm -rf /var/lib/apt/lists/*

# 创建 mrya 用户和组
RUN groupadd --gid 1000 mrya && \
    useradd --uid 1000 --gid mrya --home-dir /home/mrya --create-home --shell /bin/bash mrya

# 将 mrya 用户添加到 haproxy 组，以便有权限执行 haproxy 相关操作
RUN usermod -a -G haproxy mrya

# 创建应用目录并设置权限
WORKDIR /app
RUN chown mrya:mrya /app

# 从构建器复制Go二进制文件
COPY --from=backend-builder /build/mrya-waf .

# 复制Swagger文档文件
COPY --from=backend-builder /build/server/docs/ ./docs/

# 设置应用文件权限
RUN chown -R mrya:mrya /app && chmod +x /app/mrya-waf

# 创建 mrya 用户家目录下的 mrya-waf 目录并复制 geo-ip 文件夹
RUN mkdir -p /home/mrya/mrya-waf/geo-ip && \
    mkdir -p /home/mrya/mrya-waf/haproxy/conf && \
    mkdir -p /home/mrya/mrya-waf/haproxy/cert && \
    mkdir -p /home/mrya/mrya-waf/haproxy/spoe && \
    mkdir -p /home/mrya/mrya-waf/haproxy/conf/transaction && \
    mkdir -p /home/mrya/mrya-waf/haproxy/spoe/transaction
COPY --from=backend-builder /build/geo-ip/ /home/mrya/mrya-waf/geo-ip/
RUN chown -R mrya:mrya /home/mrya

# 🔑 关键步骤：给HAProxy和应用程序添加绑定特权端口的能力
RUN setcap 'cap_net_bind_service=+ep' /usr/local/sbin/haproxy && \
    setcap 'cap_net_bind_service=+ep' /app/mrya-waf

# 验证capabilities设置（可选，用于调试）
RUN getcap /usr/local/sbin/haproxy /app/mrya-waf

# 现在可以安全地切换到 mrya 用户
USER mrya

# 设置环境变量
ENV GIN_MODE=release

# 重置 ENTRYPOINT（覆盖基础镜像的 docker-entrypoint.sh）
ENTRYPOINT []

# 暴露端口：2333（应用程序）
EXPOSE 2333

# 运行应用
CMD ["/app/mrya-waf"]