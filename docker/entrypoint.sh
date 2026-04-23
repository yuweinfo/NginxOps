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
mkdir -p /var/log/supervisor

chmod 755 ${DATA_DIR}/logs

touch ${DATA_DIR}/logs/app/stdout.log
chmod 644 ${DATA_DIR}/logs/app/stdout.log

/usr/bin/tail -f ${DATA_DIR}/logs/app/stdout.log 2>/dev/null &
TAIL_PID=$!

trap "kill $TAIL_PID 2>/dev/null" EXIT

if [ ! -f ${DATA_DIR}/nginx/ssl/default.crt ]; then
    echo "Generating default SSL certificate..."
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout ${DATA_DIR}/nginx/ssl/default.key \
        -out ${DATA_DIR}/nginx/ssl/default.crt \
        -subj "/CN=default" 2>/dev/null
fi

chmod -R 755 ${DATA_DIR}/nginx

echo "================================================"
echo "Starting Supervisor..."
echo "================================================"

exec /usr/bin/supervisord -c /etc/supervisord.conf
