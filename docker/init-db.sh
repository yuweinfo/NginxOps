#!/bin/bash
# =====================================================
# NginxOps - 数据库初始化脚本 (Kubernetes Job)
# 仅创建数据库，Schema 迁移由 golang-migrate 自动处理
# =====================================================

set -e

USE_EXTERNAL_DB="${USE_EXTERNAL_DB:-false}"
DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-nginxops}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
MAX_RETRIES="${MAX_RETRIES:-30}"
RETRY_INTERVAL="${RETRY_INTERVAL:-5}"

echo "================================================"
echo "NginxOps - Database Init Job"
echo "================================================"
echo "USE_EXTERNAL_DB: ${USE_EXTERNAL_DB}"
echo "DB_HOST: ${DB_HOST}"
echo "DB_PORT: ${DB_PORT}"
echo "DB_NAME: ${DB_NAME}"
echo "DB_USER: ${DB_USER}"
echo "================================================"

# 如果使用内部数据库，直接成功退出
if [ "${USE_EXTERNAL_DB}" = "false" ]; then
    echo "Using internal database, skipping init job..."
    echo "Internal database will be initialized by main container."
    echo "Schema migration will be handled by golang-migrate on application startup."
    echo "================================================"
    echo "Init job completed successfully!"
    echo "================================================"
    exit 0
fi

# 等待数据库就绪
echo "Waiting for database to be ready..."
count=0
while ! pg_isready -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USER}; do
    count=$((count + 1))
    if [ $count -ge $MAX_RETRIES ]; then
        echo "ERROR: Database is not available after ${MAX_RETRIES} retries"
        exit 1
    fi
    echo "Database not ready, retrying in ${RETRY_INTERVAL}s... (${count}/${MAX_RETRIES})"
    sleep ${RETRY_INTERVAL}
done

echo "Database is ready!"

# 创建数据库（如果不存在）
echo "Checking if database '${DB_NAME}' exists..."
if ! PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -p ${DB_PORT} -U postgres -lqt | cut -d \| -f 1 | grep -qw ${DB_NAME}; then
    echo "Database '${DB_NAME}' does not exist, creating..."
    PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -p ${DB_PORT} -U postgres -c "CREATE DATABASE \"${DB_NAME}\";"
    echo "Database created successfully!"
else
    echo "Database '${DB_NAME}' already exists."
fi

echo "================================================"
echo "Init job completed successfully!"
echo "Schema migration will be handled by golang-migrate."
echo "================================================"
