#!/bin/bash
# =====================================================
# NginxOps - 统一启动入口
# 配置文件: /data/config.yml
# =====================================================

set -e

DATA_DIR="${DATA_DIR:-/data}"
PGDATA="${PGDATA:-/data/postgresql}"
CONFIG_FILE="${DATA_DIR}/config.yml"

# 信号处理 - 优雅关闭所有服务
cleanup() {
    echo "================================================"
    echo "Shutting down services..."
    echo "================================================"

    # 停止 PostgreSQL
    if [ -f "${PGDATA}/postmaster.pid" ]; then
        echo "Stopping PostgreSQL..."
        su - postgres -c "pg_ctl -D ${PGDATA} -m fast stop" 2>/dev/null || true
    fi

    # 停止 Nginx
    echo "Stopping Nginx..."
    nginx -s quit 2>/dev/null || pkill -9 nginx 2>/dev/null || true

    # 停止 Go 后端
    echo "Stopping Go Backend..."
    if [ -n "$BACKEND_PID" ]; then
        kill -TERM $BACKEND_PID 2>/dev/null || true
        wait $BACKEND_PID 2>/dev/null || true
    fi

    echo "All services stopped."
    exit 0
}

trap cleanup SIGTERM SIGINT SIGHUP

echo "================================================"
echo "NginxOps Starting..."
echo "================================================"
echo "DATA_DIR: ${DATA_DIR}"
echo "CONFIG_FILE: ${CONFIG_FILE}"
echo "================================================"

# 创建必要的目录
mkdir -p ${DATA_DIR}/logs/app
mkdir -p ${DATA_DIR}/logs/nginx
mkdir -p ${DATA_DIR}/nginx/conf.d
mkdir -p ${DATA_DIR}/nginx/ssl
mkdir -p ${DATA_DIR}/postgresql
mkdir -p ${DATA_DIR}/data
mkdir -p ${DATA_DIR}/backups
mkdir -p /run/nginx

# 设置日志目录权限
chown -R postgres:postgres ${DATA_DIR}/logs 2>/dev/null || true
chmod 755 ${DATA_DIR}/logs

# 创建空的日志文件
touch ${DATA_DIR}/logs/app/stdout.log
chmod 644 ${DATA_DIR}/logs/app/stdout.log

# 创建默认SSL证书
if [ ! -f ${DATA_DIR}/nginx/ssl/default.crt ]; then
    echo "Generating default SSL certificate..."
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout ${DATA_DIR}/nginx/ssl/default.key \
        -out ${DATA_DIR}/nginx/ssl/default.crt \
        -subj "/CN=default" 2>/dev/null
fi

# =====================================================
# 检查配置文件是否存在
# =====================================================
if [ ! -f "${CONFIG_FILE}" ]; then
    echo "================================================"
    echo "Config file not found: ${CONFIG_FILE}"
    echo "System will enter setup mode."
    echo "Please access the web interface to complete setup."
    echo "================================================"

    SETUP_MODE=true
    USE_EXTERNAL_DB=false

    # 预先创建 PostgreSQL 数据目录（setup 模式下也需要）
    mkdir -p ${DATA_DIR}/postgresql
    chown -R postgres:postgres ${DATA_DIR}/postgresql 2>/dev/null || true
    chmod 700 ${DATA_DIR}/postgresql 2>/dev/null || true
else
    echo "Config file found, loading configuration..."

    # 解析配置文件获取数据库设置
    USE_EXTERNAL_DB=$(grep -E "^\s*useExternal:" ${CONFIG_FILE} | awk '{print $2}' || echo "false")

    if [ "${USE_EXTERNAL_DB}" = "true" ]; then
        echo "Using external database"
    else
        echo "Using internal database"

        # 设置权限
        chown -R postgres:postgres ${DATA_DIR}/postgresql
        chmod 700 ${DATA_DIR}/postgresql

        # 创建PostgreSQL日志文件
        touch ${DATA_DIR}/logs/postgresql.log
        chown postgres:postgres ${DATA_DIR}/logs/postgresql.log
        chmod 644 ${DATA_DIR}/logs/postgresql.log

        # 初始化数据库集群
        if [ ! -d "${PGDATA}/base" ]; then
            echo "Initializing PostgreSQL database cluster..."
            su - postgres -c "initdb -D ${PGDATA} --encoding=UTF8 --locale=en_US.UTF-8" 2>/dev/null || \
            su - postgres -c "initdb -D ${PGDATA}" 2>/dev/null || true

            # 配置PostgreSQL
            cat >> ${PGDATA}/postgresql.conf << EOF
listen_addresses = '127.0.0.1'
port = 5432
log_directory = '${DATA_DIR}/logs'
log_filename = 'postgresql.log'
EOF

            # 配置访问权限
            cat >> ${PGDATA}/pg_hba.conf << EOF
local   all             all                                     trust
host    all             all             127.0.0.1/32            trust
EOF
        fi

        # 从配置文件读取数据库配置
        DB_NAME=$(grep -E "^\s*name:" ${CONFIG_FILE} | awk '{print $2}' || echo "nginxops")
        DB_USER=$(grep -E "^\s*user:" ${CONFIG_FILE} | awk '{print $2}' || echo "postgres")
        DB_PASSWORD=$(grep -E "^\s*password:" ${CONFIG_FILE} | awk '{print $2}' || echo "postgres")

        # 启动PostgreSQL
        echo "Starting PostgreSQL..."
        su - postgres -c "pg_ctl -D ${PGDATA} -l ${DATA_DIR}/logs/postgresql.log start"
        sleep 3

        # 创建数据库和用户（如果不存在）
        if ! su - postgres -c "psql -h 127.0.0.1 -d postgres -lqt | cut -d \| -f 1 | grep -qw ${DB_NAME}"; then
            echo "Creating database and user..."
            su - postgres -c "psql -h 127.0.0.1 -d postgres -c \"CREATE USER \\\"${DB_USER}\\\" WITH PASSWORD '${DB_PASSWORD}';\"" || true
            su - postgres -c "psql -h 127.0.0.1 -d postgres -c \"CREATE DATABASE \\\"${DB_NAME}\\\" OWNER \\\"${DB_USER}\\\";\"" || true
            su - postgres -c "psql -h 127.0.0.1 -d postgres -c \"GRANT ALL PRIVILEGES ON DATABASE \\\"${DB_NAME}\\\" TO \\\"${DB_USER}\\\";\"" || true
        fi

        echo "Database ready."
    fi
fi

# 设置Nginx配置目录权限
chmod -R 755 ${DATA_DIR}/nginx

# =====================================================
# 启动服务
# =====================================================
echo "================================================"
echo "Starting services..."
echo "================================================"

# 启动 Go 后端
echo "Starting Go Backend..."
cd /app
./nginxops >> ${DATA_DIR}/logs/app/stdout.log 2>&1 &
BACKEND_PID=$!
echo "Go Backend started (PID: $BACKEND_PID)"

# 启动 Nginx
echo "Starting Nginx..."
nginx
echo "Nginx started"

echo "================================================"
echo "All services started successfully!"
echo "================================================"

# 等待所有后台进程
wait
