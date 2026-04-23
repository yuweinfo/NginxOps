#!/bin/bash
# =====================================================
# NginxOps - 应用启动入口
# 数据库：外部 PostgreSQL
# 配置文件: /data/config.yml
# =====================================================

set -e

DATA_DIR="${DATA_DIR:-/data}"
CONFIG_FILE="/data/config.yml"

echo "================================================"
echo "NginxOps Starting..."
echo "================================================"
echo "DATA_DIR: ${DATA_DIR}"
echo "CONFIG_FILE: ${CONFIG_FILE}"
echo "================================================"

mkdir -p ${DATA_DIR}/logs/app
mkdir -p ${DATA_DIR}/logs/nginx
mkdir -p ${DATA_DIR}/nginx/conf.d
mkdir -p ${DATA_DIR}/nginx/ssl
mkdir -p ${DATA_DIR}/data
mkdir -p ${DATA_DIR}/backups
mkdir -p /run/nginx

chmod 755 ${DATA_DIR}/logs

if [ ! -f ${DATA_DIR}/nginx/ssl/default.crt ]; then
    echo "Generating default SSL certificate..."
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout ${DATA_DIR}/nginx/ssl/default.key \
        -out ${DATA_DIR}/nginx/ssl/default.crt \
        -subj "/CN=default" 2>/dev/null
fi

chmod -R 755 ${DATA_DIR}/nginx

echo "================================================"
echo "Starting Nginx..."
echo "================================================"
nginx

echo "================================================"
echo "Starting Go Backend..."
echo "================================================"
cd /app
exec /app/nginxops
