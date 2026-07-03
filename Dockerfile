# 多阶段构建
# 阶段1: 构建Node.js前端
# 生产环境使用固定版本
ARG DOCKER_REGISTRY=docker.io
FROM ${DOCKER_REGISTRY}/library/node:23.10.0-alpine AS frontend-builder
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
FROM ${DOCKER_REGISTRY}/library/golang:1.24.2-alpine AS backend-builder
# 安装 gopacket 所需的 libpcap 开发库（CGo 编译需要）
RUN apk add --no-cache gcc musl-dev libpcap-dev
# 设置Go环境变量
ENV GO111MODULE=on \
    CGO_ENABLED=1
# 设置工作目录
WORKDIR /build
# 先复制 go.mod/go.sum 下载依赖（利用 Docker 层缓存）
COPY go.work ./
COPY coraza-spoa/go.mod coraza-spoa/go.sum ./coraza-spoa/
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY server/go.mod server/go.sum ./server/
COPY mcp-server/go.mod mcp-server/go.sum ./mcp-server/
RUN go work use ./coraza-spoa ./pkg ./server ./mcp-server
RUN go work sync
# 预热：下载全部依赖（此层会被缓存，除非 go.mod/go.sum 变化）
RUN cd server && go mod download && cd ../coraza-spoa && go mod download && cd ../mcp-server && go mod download && cd ../pkg && go mod download
# 复制源码（此层以下在代码变更时失效）
COPY coraza-spoa/ ./coraza-spoa/
COPY pkg/ ./pkg/
COPY server/ ./server/
COPY mcp-server/ ./mcp-server/
COPY geo-ip/ ./geo-ip/
# 复制前端构建产物到正确位置
COPY --from=frontend-builder /app/dist ./server/public/dist
# FTW 测试文件
COPY server/public/ftw-tests/ ./server/public/ftw-tests/
# 构建
RUN cd server && go build -o ../mrya-waf main.go

# 阶段3: 最终镜像 - 使用官方 HAProxy 3.0.10 镜像
FROM ${DOCKER_REGISTRY}/library/haproxy:3.0.10

# 确保以root用户进行初始化设置
USER root

# 安装Linux capabilities管理工具及常用调试工具
RUN apt-get update && apt-get install -y \
    libcap2-bin \
    curl \
    iputils-ping \
    iproute2 \
    net-tools \
    dnsutils \
    && rm -rf /var/lib/apt/lists/*

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
# 从构建器复制 alpine libpcap.so（Go CGo 编译时链接的是 musl 版本的 libpcap.so.1）
COPY --from=backend-builder /usr/lib/libpcap.so.1 /usr/lib/aarch64-linux-gnu/libpcap.so.1

# 复制Swagger文档文件
COPY --from=backend-builder /build/server/docs/ ./docs/

# 设置应用文件权限
RUN chown -R mrya:mrya /app && chmod +x /app/mrya-waf

# 创建 mrya 用户家目录下的 mrya-waf 目录并复制 geo-ip 文件夹
# 创建 FTW 测试文件目录
COPY --from=backend-builder /build/server/public/ftw-tests/ /app/server/public/ftw-tests/

RUN mkdir -p /home/mrya/mrya-waf/geo-ip && \
    mkdir -p /home/mrya/mrya-waf/haproxy/conf && \
    mkdir -p /home/mrya/mrya-waf/haproxy/cert && \
    mkdir -p /home/mrya/mrya-waf/haproxy/spoe && \
    mkdir -p /home/mrya/mrya-waf/haproxy/conf/transaction && \
    mkdir -p /home/mrya/mrya-waf/haproxy/spoe/transaction
COPY --from=backend-builder /build/geo-ip/ /home/mrya/mrya-waf/geo-ip/
RUN chown -R mrya:mrya /home/mrya

# 创建根目录下的HAProxy目录（应用程序需要）
RUN mkdir -p /haproxy/conf && \
    mkdir -p /haproxy/cert && \
    mkdir -p /haproxy/spoe && \
    mkdir -p /haproxy/conf/transaction && \
    mkdir -p /haproxy/spoe/transaction && \
    chown -R mrya:mrya /haproxy

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