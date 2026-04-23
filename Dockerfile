# =====================================================
# NginxOps - 应用镜像构建
# 包含：Go后端 + 前端 + Nginx + Supervisor
# 数据库：外部 PostgreSQL
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
RUN apk add --no-cache git ca-certificates tzdata
COPY backend/go.mod backend/go.sum ./
COPY backend/ ./
RUN go mod tidy && go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/nginxops ./cmd/server

# ==================== 最终镜像 ====================
FROM alpine:3.19

LABEL maintainer="nginxops"
LABEL version="2.0.0"
LABEL description="NginxOps - Application image with Go backend, Frontend, Nginx (PostgreSQL as external service)"

ENV LANG=C.UTF-8 \
    DATA_DIR=/data

RUN apk add --no-cache \
    nginx \
    supervisor \
    curl \
    bash \
    tzdata \
    ca-certificates \
    openssl \
    && mkdir -p /var/log/supervisor \
    && mkdir -p /var/log/nginx \
    && mkdir -p /usr/share/nginx/html

RUN mkdir -p ${DATA_DIR}/logs/app \
    && mkdir -p ${DATA_DIR}/logs/nginx \
    && mkdir -p ${DATA_DIR}/nginx/conf.d \
    && mkdir -p ${DATA_DIR}/nginx/ssl \
    && mkdir -p ${DATA_DIR}/data \
    && mkdir -p ${DATA_DIR}/backups

COPY --from=frontend-builder /build/frontend/dist /usr/share/nginx/html
COPY --from=backend-builder /app/nginxops /app/nginxops
COPY backend/migrations /app/migrations

COPY docker/nginx.conf /etc/nginx/nginx.conf
COPY docker/supervisord.conf /etc/supervisord.conf
COPY docker/entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh /app/nginxops

RUN ln -sf ${DATA_DIR}/nginx/conf.d /etc/nginx/conf.d \
    && ln -sf ${DATA_DIR}/nginx/ssl /etc/nginx/ssl \
    && ln -sf ${DATA_DIR}/logs/nginx /var/log/nginx

EXPOSE 8899 80 443

HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:8899/api/health && curl -f http://localhost:80/ || exit 1

ENTRYPOINT ["/entrypoint.sh"]
