# =====================================================
# NginxOps - 统一镜像构建
# 包含：Go后端 + 前端 + PostgreSQL + Nginx
# 配置文件：/data/config.yml
# =====================================================

# ==================== 前端构建阶段 ====================
FROM node:20-alpine AS frontend-builder
WORKDIR /build/frontend
COPY frontend/package*.json ./
RUN npm ci --registry=https://mirrors.tencent.com/npm/
COPY frontend/ ./
RUN npm run build

# ==================== 后端构建阶段 ====================
FROM golang:1.24-alpine AS backend-builder
WORKDIR /build/backend
# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata
# 复制 go.mod 和 go.sum 并下载依赖
COPY backend/go.mod backend/go.sum ./
# 复制源代码（go mod tidy 需要读取源码中的导入）
COPY backend/ ./
# 整理依赖并下载
RUN go mod tidy && go mod download
# 构建
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/nginxops ./cmd/server

# ==================== 最终镜像 ====================
FROM alpine:3.19

LABEL maintainer="nginxops"
LABEL version="1.0.0"
LABEL description="NginxOps - Unified image with Go backend, Frontend, PostgreSQL, Nginx"

# 设置环境变量
ENV LANG=C.UTF-8 \
    DATA_DIR=/data \
    PGDATA=/data/postgresql

# 安装依赖
RUN apk add --no-cache \
    nginx \
    postgresql \
    postgresql-contrib \
    curl \
    bash \
    tzdata \
    ca-certificates \
    openssl \
    psmisc \
    && mkdir -p /var/log/nginx \
    && mkdir -p /var/lib/postgresql \
    && mkdir -p /run/postgresql \
    && mkdir -p /usr/share/nginx/html

# 创建目录结构
RUN mkdir -p ${DATA_DIR}/logs/app \
    && mkdir -p ${DATA_DIR}/logs/nginx \
    && mkdir -p ${DATA_DIR}/nginx/conf.d \
    && mkdir -p ${DATA_DIR}/nginx/ssl \
    && mkdir -p ${DATA_DIR}/postgresql \
    && mkdir -p ${DATA_DIR}/data \
    && mkdir -p ${DATA_DIR}/backups

# 复制前端构建产物
COPY --from=frontend-builder /build/frontend/dist /usr/share/nginx/html

# 复制后端构建产物
COPY --from=backend-builder /app/nginxops /app/nginxops
# 复制数据库迁移文件
COPY backend/migrations /app/migrations

# 复制配置文件
COPY docker/nginx.conf /etc/nginx/nginx.conf
COPY docker/entrypoint.sh /entrypoint.sh
COPY docker/init-db.sh /init-db.sh

# 设置权限
RUN chmod +x /entrypoint.sh /init-db.sh /app/nginxops \
    && chown -R postgres:postgres /var/lib/postgresql /run/postgresql ${DATA_DIR}/postgresql

# 创建符号链接
RUN ln -sf ${DATA_DIR}/nginx/conf.d /etc/nginx/conf.d \
    && ln -sf ${DATA_DIR}/nginx/ssl /etc/nginx/ssl \
    && ln -sf ${DATA_DIR}/logs/nginx /var/log/nginx

# 暴露端口
# 8899 - 管理面板
# 80/443 - 业务代理
EXPOSE 8899 80 443

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=120s --retries=3 \
    CMD curl -f http://localhost:8899/api/health && curl -f http://localhost:80/ || exit 1

# 启动入口
ENTRYPOINT ["/entrypoint.sh"]
