# ChainSpace Backend Dockerfile
# 多阶段构建

# ============================================
# 构建阶段
# ============================================
FROM golang:1.24-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /build

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
    -o chainspace-backend \
    ./cmd/server

# ============================================
# 运行阶段
# ============================================
FROM alpine:3.19

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN adduser -D -g '' appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/chainspace-backend .

# 复制配置文件目录（运行时挂载或通过环境变量配置）
COPY --from=builder /build/configs ./configs

# 设置文件权限
RUN chown -R appuser:appuser /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动命令
ENTRYPOINT ["./chainspace-backend"]
